package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/bpcl/portal-api/internal/model"
	"github.com/bpcl/portal-api/internal/parser"
	"github.com/bpcl/portal-api/internal/repository"
	"github.com/google/uuid"
)

type MarketShareService struct {
	repo      *repository.MarketShareRepo
	outlets   OutletRepository
	uploadDir string
}

func NewMarketShareService(repo *repository.MarketShareRepo, outlets OutletRepository, uploadDir string) *MarketShareService {
	return &MarketShareService{
		repo:      repo,
		outlets:   outlets,
		uploadDir: uploadDir,
	}
}

func (s *MarketShareService) IngestDelhiMaster(ctx context.Context, filePath string, uploadedFileID uuid.UUID) (int, int, error) {
	rows, err := parser.ParseDelhiMaster(filePath)
	if err != nil {
		return 0, 0, err
	}

	var marketShareRows []model.MarketShareData
	skipped := 0

	for _, row := range rows {
		ta, err := s.repo.GetTradingAreaByName(ctx, row.TradingArea)
		if err != nil {
			skipped++
			continue
		}

		for periodStr, vol := range row.MonthlyVolumes {
			period, err := parsePeriodString(periodStr)
			if err != nil {
				continue
			}

			msVol, hsdVol := parseVolumePair(vol)
			if msVol == 0 && hsdVol == 0 {
				continue
			}

			omc := normalizeOMC(row.OMC)
			if omc == "" {
				continue
			}

			var cc *string
			if row.CCNumber != "" {
				cc = &row.CCNumber
			}

			marketShareRows = append(marketShareRows, model.MarketShareData{
				OutletName:     row.OutletName,
				CCNumber:      cc,
				OMC:            omc,
				TradingAreaID:  ta.ID,
				Period:         period,
				MSVolKL:        msVol,
				HSDVolKL:       hsdVol,
				Source:         "delhi_master",
				UploadedFileID: &uploadedFileID,
			})
		}
	}

	if err := s.repo.UpsertMarketShareData(ctx, marketShareRows); err != nil {
		return 0, 0, err
	}

	affectedPeriods := make(map[time.Time]bool)
	for _, row := range marketShareRows {
		affectedPeriods[row.Period] = true
	}

	for period := range affectedPeriods {
		if err := s.RecomputeTradingAreaTotals(ctx, period); err != nil {
			continue
		}
	}

	return len(marketShareRows), skipped, nil
}

func (s *MarketShareService) RecomputeTradingAreaTotals(ctx context.Context, period time.Time) error {
	totals, err := s.repo.AggregateByTradingArea(ctx, period)
	if err != nil {
		return err
	}

	var modelTotals []model.TradingAreaTotal
	for _, r := range totals {
		modelTotals = append(modelTotals, model.TradingAreaTotal{
			TradingAreaID: r.TradingAreaID,
			Period:        period,
			TotalMSKL:     r.TotalMS,
			TotalHSDKL:    r.TotalHSD,
			BPCLMSKL:      r.BPCLMS,
			HPCLMSKL:      r.HPCLMS,
			IOCLMSKL:      r.IOCLMS,
			BPCLHSDKL:     r.BPCLHSD,
			HPCLHSDKL:     r.HPCLHSD,
			IOCLHSDKL:     r.IOCLHSD,
			ComputedAt:    time.Now(),
		})
	}

	return s.repo.UpsertTradingAreaTotals(ctx, modelTotals)
}

func (s *MarketShareService) ComputeMarketShareGains(ctx context.Context, competitionID uuid.UUID) (int, error) {
	competition, err := s.repo.GetCompetitionByID(ctx, competitionID)
	if err != nil {
		return 0, err
	}

	outlets, err := s.repo.GetOutletsWithTradingArea(ctx, competitionID)
	if err != nil {
		return 0, err
	}

	period := competition.Period
	lastYearPeriod := time.Date(period.Year()-1, period.Month(), 1, 0, 0, 0, 0, time.UTC)

	updates := make(map[string]struct{ MS, HSD float64 })

	for _, outlet := range outlets {
		if outlet.TradingAreaID == nil {
			continue
		}

		msVolCurrent, hsdVolCurrent, err := s.repo.GetDealerVolumeForPeriod(ctx, outlet.CCNumber, period)
		if err != nil || (msVolCurrent == 0 && hsdVolCurrent == 0) {
			continue
		}

		taCurrent, err := s.repo.GetTradingAreaTotal(ctx, *outlet.TradingAreaID, period)
		if err != nil || taCurrent.TotalMSKL == 0 {
			continue
		}

		taLastYear, err := s.repo.GetTradingAreaTotal(ctx, *outlet.TradingAreaID, lastYearPeriod)
		if err != nil {
			continue
		}

		var msGainPP, hsdGainPP float64
		if taLastYear.TotalMSKL > 0 {
			msGainPP = ((msVolCurrent / taCurrent.TotalMSKL) - (msVolCurrent / taLastYear.TotalMSKL)) * 100
			hsdGainPP = ((hsdVolCurrent / taCurrent.TotalHSDKL) - (hsdVolCurrent / taLastYear.TotalHSDKL)) * 100
		} else {
			msGainPP = (msVolCurrent / taCurrent.TotalMSKL) * 100
			hsdGainPP = (hsdVolCurrent / taCurrent.TotalHSDKL) * 100
		}

		updates[outlet.CCNumber] = struct{ MS, HSD float64 }{msGainPP, hsdGainPP}
	}

	if len(updates) > 0 {
		if err := s.repo.UpdateScoresWithMSGain(ctx, competitionID, updates); err != nil {
			return 0, err
		}
	}

	return len(updates), nil
}

func (s *MarketShareService) RecomputeScoresAndRanks(ctx context.Context, competitionID uuid.UUID) error {
	if err := s.repo.RecalculateTotalScore(ctx, competitionID); err != nil {
		return err
	}
	return s.repo.RecomputeTotalScores(ctx, competitionID)
}

func (s *MarketShareService) GetMarketShareStatus(ctx context.Context, competitionID uuid.UUID) (*model.MarketShareStatusResponse, error) {
	competition, err := s.repo.GetCompetitionByID(ctx, competitionID)
	if err != nil {
		return nil, err
	}

	missing, err := s.repo.GetCountScoresWithNullMSGain(ctx, competitionID)
	if err != nil {
		return nil, err
	}

	total, err := s.repo.GetTotalRegularDealers(ctx, competitionID)
	if err != nil {
		return nil, err
	}

	missingFromSource, withPeriodGaps, err := s.repo.CategorizeMissingDealers(ctx, competitionID, competition.Period)
	if err != nil {
		missingFromSource = missing
		withPeriodGaps = 0
	}

	status := "complete"
	msg := "All dealers have market share data"
	if missing > 0 {
		status = "partial"
		msg = fmt.Sprintf("%d/%d dealers have real market share data. %d dealers not present in Delhi Master source file.", total-missing, total, missingFromSource)
	}

	return &model.MarketShareStatusResponse{
		CompetitionID:            competitionID,
		TotalDealers:             total,
		DealersWithMSData:        total - missing,
		DealersMissingFromSource: missingFromSource,
		DealersWithPeriodGaps:    withPeriodGaps,
		Message:                  msg,
		Status:                   status,
	}, nil
}

func (s *MarketShareService) FullRecompute(ctx context.Context, competitionID uuid.UUID) (*model.RecomputeResponse, error) {
	before, err := s.repo.GetTopScores(ctx, competitionID, 10)
	if err != nil {
		before = nil
	}

	updated, err := s.ComputeMarketShareGains(ctx, competitionID)
	if err != nil {
		return nil, err
	}

	outlets, err := s.repo.GetOutletsWithTradingArea(ctx, competitionID)
	if err != nil {
		return nil, err
	}

	for _, outlet := range outlets {
		if outlet.TradingAreaID == nil {
			continue
		}

		msGainPP, hsdGainPP, err := s.repo.GetDealerMSGain(ctx, outlet.CCNumber, competitionID)
		if err != nil {
			continue
		}

		msScore := computeMSGainScore(msGainPP)
		hsdScore := computeMSGainScore(hsdGainPP)

		if err := s.repo.UpdateParameterScores(ctx, competitionID, outlet.CCNumber, msScore, hsdScore); err != nil {
			continue
		}
	}

	if err := s.RecomputeScoresAndRanks(ctx, competitionID); err != nil {
		return nil, err
	}

	after, err := s.repo.GetTopScores(ctx, competitionID, 10)
	if err != nil {
		after = nil
	}

	return &model.RecomputeResponse{
		CompetitionID:  competitionID,
		ScoresUpdated:  len(outlets),
		MSGainComputed: updated,
		Top10Before:    before,
		Top10After:     after,
	}, nil
}

func computeMSGainScore(gainPP float64) float64 {
	const maxMarks = 10.0

	if gainPP >= 2.0 {
		return maxMarks
	}
	if gainPP >= -2.0 {
		return (gainPP / 2.0) * maxMarks
	}
	return -maxMarks
}

func parsePeriodString(s string) (time.Time, error) {
	parts := splitMonthYear(s)
	if len(parts) != 2 {
		return time.Time{}, nil
	}

	month, ok := monthMap[parts[0]]
	if !ok {
		return time.Time{}, nil
	}

	year, err := strconv.Atoi(parts[1])
	if err != nil || year < 100 {
		year += 2000
	}

	return time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC), nil
}

func splitMonthYear(s string) []string {
	var result []string
	var current []rune
	for _, c := range s {
		if c == '-' {
			result = append(result, string(current))
			current = nil
		} else {
			current = append(current, c)
		}
	}
	if len(current) > 0 {
		result = append(result, string(current))
	}
	return result
}

var monthMap = map[string]int{
	"jan": 1, "feb": 2, "mar": 3, "apr": 4,
	"may": 5, "jun": 6, "jul": 7, "aug": 8,
	"sep": 9, "oct": 10, "nov": 11, "dec": 12,
}

func normalizeOMC(s string) string {
	s = normalizeString(s)
	switch s {
	case "BPC", "BPCL":
		return "BPCL"
	case "HPC", "HPCL":
		return "HPCL"
	case "IOC", "IOCL":
		return "IOCL"
	}
	return ""
}

func normalizeString(s string) string {
	result := make([]rune, 0, len(s))
	for _, c := range s {
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' {
			result = append(result, c)
		}
	}
	return string(result)
}

func parseVolumePair(s string) (ms, hsd float64) {
	ms = parseVolume(s)
	hsd = parseVolume(s)
	return ms, hsd
}

func parseVolume(s string) float64 {
	s = normalizeString(s)
	if s == "" {
		return 0
	}
	v, _ := strconv.ParseFloat(s, 64)
	return v
}