package ingest_test

import (
	"context"
	"testing"
	"time"

	"github.com/bpcl/portal-api/internal/crystal/ingest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- mock implementations ---

type mockDealers struct {
	active map[string]bool
}

func (m *mockDealers) IsActive(_ context.Context, ccCode string) (bool, error) {
	return m.active[ccCode], nil
}

func (m *mockDealers) ValidateCodes(_ context.Context, codes []string) (map[string]bool, error) {
	invalid := make(map[string]bool)
	for _, c := range codes {
		if !m.active[c] {
			invalid[c] = true
		}
	}
	return invalid, nil
}

type mockPeriod struct {
	activeMonth time.Time
}

func (m *mockPeriod) DateInActivePeriod(_ context.Context, d time.Time) (bool, error) {
	return d.Year() == m.activeMonth.Year() && d.Month() == m.activeMonth.Month(), nil
}

type mockRepo struct {
	upsertDailyFn   func(metric ingest.MetricKey, date time.Time, rows []ingest.DailyRow) (int, int, error)
	upsertMAKGEFn   func(req ingest.MAKGERequest, date time.Time) (bool, error)
	upsertRatingFn  func(req ingest.GoogleRatingRequest, date time.Time) (bool, error)
	upsertTargetsFn func(monthYear time.Time, rows []ingest.TargetRow) (int, int, error)
}

func (m *mockRepo) UpsertDaily(_ context.Context, metric ingest.MetricKey, date time.Time, rows []ingest.DailyRow) (int, int, error) {
	if m.upsertDailyFn != nil {
		return m.upsertDailyFn(metric, date, rows)
	}
	return len(rows), 0, nil
}

func (m *mockRepo) UpsertMAKGE(_ context.Context, req ingest.MAKGERequest, date time.Time) (bool, error) {
	if m.upsertMAKGEFn != nil {
		return m.upsertMAKGEFn(req, date)
	}
	return true, nil
}

func (m *mockRepo) UpsertGoogleRating(_ context.Context, req ingest.GoogleRatingRequest, date time.Time) (bool, error) {
	if m.upsertRatingFn != nil {
		return m.upsertRatingFn(req, date)
	}
	return true, nil
}

func (m *mockRepo) UpsertTargets(_ context.Context, monthYear time.Time, rows []ingest.TargetRow) (int, int, error) {
	if m.upsertTargetsFn != nil {
		return m.upsertTargetsFn(monthYear, rows)
	}
	return len(rows), 0, nil
}

// --- helpers ---

func activePeriod() *mockPeriod {
	may2026, _ := time.Parse("2006-01-02", "2026-05-01")
	return &mockPeriod{activeMonth: may2026}
}

func knownDealers(codes ...string) *mockDealers {
	m := &mockDealers{active: make(map[string]bool)}
	for _, c := range codes {
		m.active[c] = true
	}
	return m
}

func newSvc(repo *mockRepo, dealers *mockDealers, period *mockPeriod) *ingest.Service {
	return ingest.NewService(repo, dealers, period)
}

// --- IngestDaily tests ---

func TestIngestDaily_ValidRequest_ReturnsInserted(t *testing.T) {
	repo := &mockRepo{}
	svc := newSvc(repo, knownDealers("112385", "112386"), activePeriod())

	result, err := svc.IngestDaily(context.Background(), ingest.BulkDailyRequest{
		Date:   "2026-05-07",
		Metric: ingest.MetricUfill,
		Rows: []ingest.DailyRow{
			{CCCode: "112385", Value: 52},
			{CCCode: "112386", Value: 46},
		},
	})

	require.NoError(t, err)
	assert.Equal(t, 2, result.Inserted)
	assert.Equal(t, 0, result.Updated)
	assert.Equal(t, 0, result.Skipped)
}

func TestIngestDaily_InvalidDateFormat_ReturnsError(t *testing.T) {
	svc := newSvc(&mockRepo{}, knownDealers("112385"), activePeriod())

	_, err := svc.IngestDaily(context.Background(), ingest.BulkDailyRequest{
		Date:   "07-05-2026",
		Metric: ingest.MetricUfill,
		Rows:   []ingest.DailyRow{{CCCode: "112385", Value: 52}},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid date")
}

func TestIngestDaily_DateOutsideActivePeriod_ReturnsError(t *testing.T) {
	svc := newSvc(&mockRepo{}, knownDealers("112385"), activePeriod())

	_, err := svc.IngestDaily(context.Background(), ingest.BulkDailyRequest{
		Date:   "2026-06-01", // June, not May
		Metric: ingest.MetricUfill,
		Rows:   []ingest.DailyRow{{CCCode: "112385", Value: 52}},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "active competition period")
}

func TestIngestDaily_UnknownMetric_ReturnsError(t *testing.T) {
	svc := newSvc(&mockRepo{}, knownDealers("112385"), activePeriod())

	_, err := svc.IngestDaily(context.Background(), ingest.BulkDailyRequest{
		Date:   "2026-05-07",
		Metric: "badmetric",
		Rows:   []ingest.DailyRow{{CCCode: "112385", Value: 52}},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown metric")
}

func TestIngestDaily_UnknownCCCode_SkippedWithError(t *testing.T) {
	svc := newSvc(&mockRepo{}, knownDealers("112385"), activePeriod())

	result, err := svc.IngestDaily(context.Background(), ingest.BulkDailyRequest{
		Date:   "2026-05-07",
		Metric: ingest.MetricUfill,
		Rows: []ingest.DailyRow{
			{CCCode: "112385", Value: 10},
			{CCCode: "UNKNOWN", Value: 5},
		},
	})

	require.NoError(t, err)
	assert.Equal(t, 1, result.Inserted)
	assert.Equal(t, 1, result.Skipped)
	assert.Len(t, result.Errors, 1)
	assert.Contains(t, result.Errors[0], "UNKNOWN")
}

func TestIngestDaily_NegativeValue_SkippedWithError(t *testing.T) {
	svc := newSvc(&mockRepo{}, knownDealers("112385"), activePeriod())

	result, err := svc.IngestDaily(context.Background(), ingest.BulkDailyRequest{
		Date:   "2026-05-07",
		Metric: ingest.MetricUfill,
		Rows:   []ingest.DailyRow{{CCCode: "112385", Value: -1}},
	})

	require.NoError(t, err)
	assert.Equal(t, 0, result.Inserted)
	assert.Equal(t, 1, result.Skipped)
	assert.Len(t, result.Errors, 1)
}

func TestIngestDaily_AllSkipped_ReturnsEmptyResult(t *testing.T) {
	svc := newSvc(&mockRepo{}, knownDealers(), activePeriod())

	result, err := svc.IngestDaily(context.Background(), ingest.BulkDailyRequest{
		Date:   "2026-05-07",
		Metric: ingest.MetricUfill,
		Rows:   []ingest.DailyRow{{CCCode: "NONE", Value: 5}},
	})

	require.NoError(t, err)
	assert.Equal(t, 0, result.Inserted)
	assert.Equal(t, 1, result.Skipped)
}

// --- IngestMAKGE tests ---

func TestIngestMAKGE_ValidRequest_ReturnsInserted(t *testing.T) {
	svc := newSvc(&mockRepo{}, knownDealers("112390"), activePeriod())

	result, err := svc.IngestMAKGE(context.Background(), ingest.MAKGERequest{
		CCCode:       "112390",
		ReadingDate:  "2026-05-08",
		MeterReading: 14523.50,
	})

	require.NoError(t, err)
	assert.Equal(t, 1, result.Inserted)
	assert.Equal(t, 0, result.Updated)
}

func TestIngestMAKGE_Update_ReturnsUpdated(t *testing.T) {
	repo := &mockRepo{upsertMAKGEFn: func(_ ingest.MAKGERequest, _ time.Time) (bool, error) {
		return false, nil // false = was updated
	}}
	svc := newSvc(repo, knownDealers("112390"), activePeriod())

	result, err := svc.IngestMAKGE(context.Background(), ingest.MAKGERequest{
		CCCode:       "112390",
		ReadingDate:  "2026-05-08",
		MeterReading: 14523.50,
	})

	require.NoError(t, err)
	assert.Equal(t, 0, result.Inserted)
	assert.Equal(t, 1, result.Updated)
}

func TestIngestMAKGE_InvalidDate_ReturnsError(t *testing.T) {
	svc := newSvc(&mockRepo{}, knownDealers("112390"), activePeriod())

	_, err := svc.IngestMAKGE(context.Background(), ingest.MAKGERequest{
		CCCode:       "112390",
		ReadingDate:  "not-a-date",
		MeterReading: 100,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid reading_date")
}

func TestIngestMAKGE_DateOutsidePeriod_ReturnsError(t *testing.T) {
	svc := newSvc(&mockRepo{}, knownDealers("112390"), activePeriod())

	_, err := svc.IngestMAKGE(context.Background(), ingest.MAKGERequest{
		CCCode:       "112390",
		ReadingDate:  "2026-04-30",
		MeterReading: 100,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "active competition period")
}

func TestIngestMAKGE_UnknownDealer_ReturnsError(t *testing.T) {
	svc := newSvc(&mockRepo{}, knownDealers(), activePeriod())

	_, err := svc.IngestMAKGE(context.Background(), ingest.MAKGERequest{
		CCCode:       "NONE",
		ReadingDate:  "2026-05-08",
		MeterReading: 100,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown cc_code")
}

func TestIngestMAKGE_NegativeMeterReading_ReturnsError(t *testing.T) {
	svc := newSvc(&mockRepo{}, knownDealers("112390"), activePeriod())

	_, err := svc.IngestMAKGE(context.Background(), ingest.MAKGERequest{
		CCCode:       "112390",
		ReadingDate:  "2026-05-08",
		MeterReading: -1,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be negative")
}

// --- IngestGoogleRating tests ---

func TestIngestGoogleRating_ValidRequest_ReturnsInserted(t *testing.T) {
	svc := newSvc(&mockRepo{}, knownDealers("112461"), activePeriod())

	result, err := svc.IngestGoogleRating(context.Background(), ingest.GoogleRatingRequest{
		CCCode:       "112461",
		SnapshotDate: "2026-05-08",
		Rating:       4.4,
		ReviewCount:  1469,
	})

	require.NoError(t, err)
	assert.Equal(t, 1, result.Inserted)
}

func TestIngestGoogleRating_RatingBelowMin_ReturnsError(t *testing.T) {
	svc := newSvc(&mockRepo{}, knownDealers("112461"), activePeriod())

	_, err := svc.IngestGoogleRating(context.Background(), ingest.GoogleRatingRequest{
		CCCode:       "112461",
		SnapshotDate: "2026-05-08",
		Rating:       0.9,
		ReviewCount:  100,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "rating must be between")
}

func TestIngestGoogleRating_RatingAboveMax_ReturnsError(t *testing.T) {
	svc := newSvc(&mockRepo{}, knownDealers("112461"), activePeriod())

	_, err := svc.IngestGoogleRating(context.Background(), ingest.GoogleRatingRequest{
		CCCode:       "112461",
		SnapshotDate: "2026-05-08",
		Rating:       5.1,
		ReviewCount:  100,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "rating must be between")
}

func TestIngestGoogleRating_NegativeReviewCount_ReturnsError(t *testing.T) {
	svc := newSvc(&mockRepo{}, knownDealers("112461"), activePeriod())

	_, err := svc.IngestGoogleRating(context.Background(), ingest.GoogleRatingRequest{
		CCCode:       "112461",
		SnapshotDate: "2026-05-08",
		Rating:       4.0,
		ReviewCount:  -1,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be negative")
}

// --- IngestTargets tests ---

func TestIngestTargets_ValidRequest_ReturnsInserted(t *testing.T) {
	ufill := 3000
	svc := newSvc(&mockRepo{}, knownDealers("112385"), activePeriod())

	result, err := svc.IngestTargets(context.Background(), ingest.BulkTargetsRequest{
		MonthYear: "2026-05-01",
		Rows: []ingest.TargetRow{
			{CCCode: "112385", UfillTarget: &ufill},
		},
	})

	require.NoError(t, err)
	assert.Equal(t, 1, result.Inserted)
	assert.Equal(t, 0, result.Skipped)
}

func TestIngestTargets_InvalidMonthYear_ReturnsError(t *testing.T) {
	svc := newSvc(&mockRepo{}, knownDealers("112385"), activePeriod())

	_, err := svc.IngestTargets(context.Background(), ingest.BulkTargetsRequest{
		MonthYear: "not-a-date",
		Rows:      []ingest.TargetRow{{CCCode: "112385"}},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid month_year")
}

func TestIngestTargets_MonthYearNotFirstDay_ReturnsError(t *testing.T) {
	svc := newSvc(&mockRepo{}, knownDealers("112385"), activePeriod())

	_, err := svc.IngestTargets(context.Background(), ingest.BulkTargetsRequest{
		MonthYear: "2026-05-15",
		Rows:      []ingest.TargetRow{{CCCode: "112385"}},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "first day")
}

func TestIngestTargets_UnknownCCCode_SkippedWithError(t *testing.T) {
	svc := newSvc(&mockRepo{}, knownDealers("112385"), activePeriod())

	result, err := svc.IngestTargets(context.Background(), ingest.BulkTargetsRequest{
		MonthYear: "2026-05-01",
		Rows: []ingest.TargetRow{
			{CCCode: "112385"},
			{CCCode: "UNKNOWN"},
		},
	})

	require.NoError(t, err)
	assert.Equal(t, 1, result.Inserted)
	assert.Equal(t, 1, result.Skipped)
	assert.Len(t, result.Errors, 1)
}
