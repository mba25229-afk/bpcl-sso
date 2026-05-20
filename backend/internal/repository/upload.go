package repository

import (
	"context"
	"errors"
	"time"

	"github.com/bpcl/portal-api/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UploadRepo struct {
	pool *pgxpool.Pool
}

func NewUploadRepo(pool *pgxpool.Pool) *UploadRepo {
	return &UploadRepo{pool: pool}
}

const uploadSelect = `SELECT id, user_id, original_name, stored_path, size_bytes,
	mime_type, upload_type, status, error_message, row_count, cc_number, period,
	created_at, updated_at FROM uploaded_files`

func (r *UploadRepo) Create(ctx context.Context, f *model.UploadedFile) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO uploaded_files
			(user_id, original_name, stored_path, size_bytes, mime_type,
			 upload_type, status, error_message, row_count, cc_number, period)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at`,
		f.UserID, f.OriginalName, f.StoredPath, f.SizeBytes, f.MimeType,
		f.UploadType, f.Status, f.ErrorMessage, f.RowCount, f.CCNumber, f.Period,
	).Scan(&f.ID, &f.CreatedAt, &f.UpdatedAt)
}

func (r *UploadRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.UploadedFile, error) {
	row := r.pool.QueryRow(ctx, uploadSelect+` WHERE id = $1`, id)
	f, err := scanUpload(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return f, nil
}

func (r *UploadRepo) ListByUser(ctx context.Context, userID uuid.UUID, cc string, limit, offset int) ([]*model.UploadedFile, int, error) {
	var (
		rows pgx.Rows
		err  error
	)
	countQ := `SELECT COUNT(*) FROM uploaded_files WHERE user_id = $1 AND status != 'deleted'`
	listQ := uploadSelect + ` WHERE user_id = $1 AND status != 'deleted'`

	var total int
	if cc != "" {
		countQ += ` AND cc_number = $2`
		listQ += ` AND cc_number = $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4`
		if err = r.pool.QueryRow(ctx, countQ, userID, cc).Scan(&total); err != nil {
			return nil, 0, err
		}
		rows, err = r.pool.Query(ctx, listQ, userID, cc, limit, offset)
	} else {
		listQ += ` ORDER BY created_at DESC LIMIT $2 OFFSET $3`
		if err = r.pool.QueryRow(ctx, countQ, userID).Scan(&total); err != nil {
			return nil, 0, err
		}
		rows, err = r.pool.Query(ctx, listQ, userID, limit, offset)
	}
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []*model.UploadedFile
	for rows.Next() {
		f, scanErr := scanUpload(rows)
		if scanErr != nil {
			return nil, 0, scanErr
		}
		result = append(result, f)
	}
	return result, total, rows.Err()
}

func (r *UploadRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status, errMsg string, rowCount int) error {
	var errMsgPtr *string
	if errMsg != "" {
		errMsgPtr = &errMsg
	}
	var rowCountPtr *int
	if rowCount > 0 {
		rowCountPtr = &rowCount
	}
	_, err := r.pool.Exec(ctx,
		`UPDATE uploaded_files SET status = $1, error_message = $2, row_count = $3, updated_at = $4 WHERE id = $5`,
		status, errMsgPtr, rowCountPtr, time.Now(), id,
	)
	return err
}

func (r *UploadRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE uploaded_files SET status = 'deleted', updated_at = $1 WHERE id = $2`,
		time.Now(), id,
	)
	return err
}

func scanUpload(row pgx.Row) (*model.UploadedFile, error) {
	var f model.UploadedFile
	var rawID, rawUserID [16]byte
	var sizeBytes pgtype.Int8
	var mimeType, errMsg, ccNumber pgtype.Text
	var rowCount pgtype.Int4
	var period pgtype.Date

	err := row.Scan(
		&rawID, &rawUserID, &f.OriginalName, &f.StoredPath,
		&sizeBytes, &mimeType, &f.UploadType, &f.Status,
		&errMsg, &rowCount, &ccNumber, &period,
		&f.CreatedAt, &f.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	f.ID = uuid.UUID(rawID)
	f.UserID = uuid.UUID(rawUserID)
	if sizeBytes.Valid {
		f.SizeBytes = &sizeBytes.Int64
	}
	if mimeType.Valid {
		f.MimeType = &mimeType.String
	}
	if errMsg.Valid {
		f.ErrorMessage = &errMsg.String
	}
	if rowCount.Valid {
		v := int(rowCount.Int32)
		f.RowCount = &v
	}
	if ccNumber.Valid {
		f.CCNumber = &ccNumber.String
	}
	if period.Valid {
		t := period.Time
		f.Period = &t
	}
	return &f, nil
}
