package service

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bpcl/portal-api/internal/model"
	"github.com/bpcl/portal-api/internal/parser"
	"github.com/google/uuid"
)

const maxUploadBytes = 10 * 1024 * 1024 // 10 MB default

type UploadResponse struct {
	UploadID uuid.UUID `json:"upload_id"`
	Status   string    `json:"status"`
}

type UploadListResponse struct {
	Items  []*model.UploadedFile `json:"items"`
	Total  int                   `json:"total"`
	Limit  int                   `json:"limit"`
	Offset int                   `json:"offset"`
}

type UploadService struct {
	uploads     UploadRepository
	performance PerformanceRepository
	uploadDir   string
	maxBytes    int64
}

func NewUploadService(uploads UploadRepository, perf PerformanceRepository, uploadDir string, maxMB int) *UploadService {
	if maxMB <= 0 {
		maxMB = 10
	}
	return &UploadService{
		uploads:     uploads,
		performance: perf,
		uploadDir:   uploadDir,
		maxBytes:    int64(maxMB) * 1024 * 1024,
	}
}

func (s *UploadService) HandleUpload(ctx context.Context, file multipart.File, header *multipart.FileHeader, cc string, period time.Time, userID uuid.UUID) (*UploadResponse, error) {
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".xlsx" && ext != ".xls" && ext != ".csv" {
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}
	if header.Size > s.maxBytes {
		return nil, fmt.Errorf("file too large: max %d MB", s.maxBytes/1024/1024)
	}

	if err := os.MkdirAll(s.uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("upload: mkdir: %w", err)
	}

	storedName := fmt.Sprintf("%s_%s%s", uuid.New().String(), time.Now().Format("20060102150405"), ext)
	storedPath := filepath.Join(s.uploadDir, storedName)
	out, err := os.Create(storedPath)
	if err != nil {
		return nil, fmt.Errorf("upload: create file: %w", err)
	}
	if _, err := io.Copy(out, file); err != nil {
		out.Close()
		return nil, fmt.Errorf("upload: write file: %w", err)
	}
	out.Close()

	size := header.Size
	mime := header.Header.Get("Content-Type")
	f := &model.UploadedFile{
		UserID:       userID,
		OriginalName: header.Filename,
		StoredPath:   storedPath,
		SizeBytes:    &size,
		MimeType:     &mime,
		UploadType:   "performance",
		Status:       "pending",
		CCNumber:     &cc,
		Period:       &period,
	}
	if err := s.uploads.Create(ctx, f); err != nil {
		return nil, fmt.Errorf("upload: create record: %w", err)
	}

	go s.ProcessAsync(f.ID, storedPath, cc, period, userID)

	return &UploadResponse{UploadID: f.ID, Status: "pending"}, nil
}

func (s *UploadService) ProcessAsync(uploadID uuid.UUID, storedPath, cc string, period time.Time, userID uuid.UUID) {
	ctx := context.Background()
	rows, err := parser.ParsePerformance(storedPath)
	if err != nil {
		_ = s.uploads.UpdateStatus(ctx, uploadID, "failed", err.Error(), 0)
		return
	}

	count := 0
	for _, row := range rows {
		productID := productCodeToID(row.ProductCode)
		if productID == 0 {
			continue
		}
		rec := &model.PerformanceRecord{
			CCNumber:       cc,
			ProductID:      productID,
			Period:         period,
			Source:         "upload",
			UploadedFileID: &uploadID,
		}
		if row.Achieved != "" {
			if v, err := parseFloat(row.Achieved); err == nil {
				rec.Achieved = &v
			}
		}
		if row.LastYear != "" {
			if v, err := parseFloat(row.LastYear); err == nil {
				rec.LastYear = &v
			}
		}
		if row.VolumeKL != "" {
			if v, err := parseFloat(row.VolumeKL); err == nil {
				rec.VolumeKL = &v
			}
		}
		if err := s.performance.Upsert(ctx, rec); err != nil {
			_ = s.uploads.UpdateStatus(ctx, uploadID, "failed", err.Error(), count)
			return
		}
		count++
	}

	_ = s.uploads.UpdateStatus(ctx, uploadID, "done", "", count)
}

func (s *UploadService) GetHistory(ctx context.Context, userID uuid.UUID, cc string, limit, offset int) (*UploadListResponse, error) {
	items, total, err := s.uploads.ListByUser(ctx, userID, cc, limit, offset)
	if err != nil {
		return nil, err
	}
	return &UploadListResponse{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(strings.TrimSpace(s), "%f", &f)
	return f, err
}
