package ingest

import (
	"context"
	"fmt"
	"time"
)

// dealerChecker abstracts dealer lookup — implemented by crystal/dealer.Repository.
type dealerChecker interface {
	IsActive(ctx context.Context, ccCode string) (bool, error)
	ValidateCodes(ctx context.Context, codes []string) (map[string]bool, error)
}

// periodChecker abstracts competition period lookup — implemented by crystal/period.Repository.
type periodChecker interface {
	DateInActivePeriod(ctx context.Context, d time.Time) (bool, error)
}

// ingestRepo abstracts all DB writes — implemented by Repository.
type ingestRepo interface {
	UpsertDaily(ctx context.Context, metric MetricKey, date time.Time, rows []DailyRow) (inserted, updated int, err error)
	UpsertMAKGE(ctx context.Context, req MAKGERequest, date time.Time) (inserted bool, err error)
	UpsertGoogleRating(ctx context.Context, req GoogleRatingRequest, date time.Time) (inserted bool, err error)
	UpsertTargets(ctx context.Context, monthYear time.Time, rows []TargetRow) (inserted, updated int, err error)
}

type Service struct {
	repo    ingestRepo
	dealers dealerChecker
	periods periodChecker
}

func NewService(repo ingestRepo, dealers dealerChecker, periods periodChecker) *Service {
	return &Service{repo: repo, dealers: dealers, periods: periods}
}

func (s *Service) IngestDaily(ctx context.Context, req BulkDailyRequest) (*IngestResult, error) {
	result := &IngestResult{}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date %q: use YYYY-MM-DD", req.Date)
	}

	inPeriod, err := s.periods.DateInActivePeriod(ctx, date)
	if err != nil {
		return nil, err
	}
	if !inPeriod {
		return nil, fmt.Errorf("date %s is not within the active competition period", req.Date)
	}

	switch req.Metric {
	case MetricUfill, MetricQOC, MetricMS, MetricHSD, MetricSpeed:
	default:
		return nil, fmt.Errorf("unknown metric %q: must be one of ufill, qoc, ms, hsd, speed", req.Metric)
	}

	codes := make([]string, len(req.Rows))
	for i, r := range req.Rows {
		codes[i] = r.CCCode
	}
	invalid, err := s.dealers.ValidateCodes(ctx, codes)
	if err != nil {
		return nil, err
	}

	valid := make([]DailyRow, 0, len(req.Rows))
	for _, row := range req.Rows {
		if invalid[row.CCCode] {
			result.Errors = append(result.Errors, fmt.Sprintf("unknown cc_code: %s", row.CCCode))
			result.Skipped++
			continue
		}
		if row.Value < 0 {
			result.Errors = append(result.Errors, fmt.Sprintf("cc_code %s: value cannot be negative", row.CCCode))
			result.Skipped++
			continue
		}
		valid = append(valid, row)
	}

	if len(valid) == 0 {
		return result, nil
	}

	ins, upd, err := s.repo.UpsertDaily(ctx, req.Metric, date, valid)
	if err != nil {
		return nil, err
	}
	result.Inserted = ins
	result.Updated = upd
	return result, nil
}

func (s *Service) IngestMAKGE(ctx context.Context, req MAKGERequest) (*IngestResult, error) {
	result := &IngestResult{}

	date, err := time.Parse("2006-01-02", req.ReadingDate)
	if err != nil {
		return nil, fmt.Errorf("invalid reading_date %q", req.ReadingDate)
	}
	inPeriod, err := s.periods.DateInActivePeriod(ctx, date)
	if err != nil {
		return nil, err
	}
	if !inPeriod {
		return nil, fmt.Errorf("reading_date %s is not within the active competition period", req.ReadingDate)
	}

	active, err := s.dealers.IsActive(ctx, req.CCCode)
	if err != nil {
		return nil, err
	}
	if !active {
		return nil, fmt.Errorf("unknown cc_code: %s", req.CCCode)
	}
	if req.MeterReading < 0 {
		return nil, fmt.Errorf("meter_reading cannot be negative")
	}

	inserted, err := s.repo.UpsertMAKGE(ctx, req, date)
	if err != nil {
		return nil, err
	}
	if inserted {
		result.Inserted = 1
	} else {
		result.Updated = 1
	}
	return result, nil
}

func (s *Service) IngestGoogleRating(ctx context.Context, req GoogleRatingRequest) (*IngestResult, error) {
	result := &IngestResult{}

	date, err := time.Parse("2006-01-02", req.SnapshotDate)
	if err != nil {
		return nil, fmt.Errorf("invalid snapshot_date %q", req.SnapshotDate)
	}
	inPeriod, err := s.periods.DateInActivePeriod(ctx, date)
	if err != nil {
		return nil, err
	}
	if !inPeriod {
		return nil, fmt.Errorf("snapshot_date %s is not within the active competition period", req.SnapshotDate)
	}

	active, err := s.dealers.IsActive(ctx, req.CCCode)
	if err != nil {
		return nil, err
	}
	if !active {
		return nil, fmt.Errorf("unknown cc_code: %s", req.CCCode)
	}
	if req.Rating < 1.0 || req.Rating > 5.0 {
		return nil, fmt.Errorf("rating must be between 1.0 and 5.0, got %.1f", req.Rating)
	}
	if req.ReviewCount < 0 {
		return nil, fmt.Errorf("review_count cannot be negative")
	}

	inserted, err := s.repo.UpsertGoogleRating(ctx, req, date)
	if err != nil {
		return nil, err
	}
	if inserted {
		result.Inserted = 1
	} else {
		result.Updated = 1
	}
	return result, nil
}

func (s *Service) IngestTargets(ctx context.Context, req BulkTargetsRequest) (*IngestResult, error) {
	result := &IngestResult{}

	monthYear, err := time.Parse("2006-01-02", req.MonthYear)
	if err != nil {
		return nil, fmt.Errorf("invalid month_year %q: use YYYY-MM-01", req.MonthYear)
	}
	if monthYear.Day() != 1 {
		return nil, fmt.Errorf("month_year must be the first day of the month, e.g. 2026-05-01")
	}

	codes := make([]string, len(req.Rows))
	for i, r := range req.Rows {
		codes[i] = r.CCCode
	}
	invalid, err := s.dealers.ValidateCodes(ctx, codes)
	if err != nil {
		return nil, err
	}

	valid := make([]TargetRow, 0, len(req.Rows))
	for _, row := range req.Rows {
		if invalid[row.CCCode] {
			result.Errors = append(result.Errors, fmt.Sprintf("unknown cc_code: %s", row.CCCode))
			result.Skipped++
			continue
		}
		valid = append(valid, row)
	}

	if len(valid) == 0 {
		return result, nil
	}

	ins, upd, err := s.repo.UpsertTargets(ctx, monthYear, valid)
	if err != nil {
		return nil, err
	}
	result.Inserted = ins
	result.Updated = upd
	return result, nil
}
