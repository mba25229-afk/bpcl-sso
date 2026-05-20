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

type CompetitionRepo struct {
	pool *pgxpool.Pool
}

func NewCompetitionRepo(pool *pgxpool.Pool) *CompetitionRepo {
	return &CompetitionRepo{pool: pool}
}

// ─── CompetitionPeriod ───────────────────────────────────────────────────────

func (r *CompetitionRepo) GetActivePeriod(ctx context.Context, territoryCode string) (*model.CompetitionPeriod, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, period, territory_code, status, published_at, created_by, created_at
		FROM competition_periods
		WHERE (territory_code = $1 OR territory_code = 'DELHI-ALL')
		  AND status IN ('active','published')
		ORDER BY period DESC LIMIT 1`,
		territoryCode,
	)
	return scanPeriod(row)
}

func scanPeriod(row pgx.Row) (*model.CompetitionPeriod, error) {
	var p model.CompetitionPeriod
	var rawID [16]byte
	var publishedAt pgtype.Timestamptz
	var createdBy pgtype.UUID

	err := row.Scan(&rawID, &p.Name, &p.Period, &p.TerritoryCode, &p.Status, &publishedAt, &createdBy, &p.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	p.ID = uuid.UUID(rawID)
	if publishedAt.Valid {
		t := publishedAt.Time
		p.PublishedAt = &t
	}
	p.CreatedBy = pguuid(createdBy)
	return &p, nil
}

// ─── CompetitionScore ────────────────────────────────────────────────────────

const scoreSelect = `SELECT id, competition_id, cc_number,
	ms_vol_kl, ms_growth_pct, ms_ta_gain_pp,
	hsd_vol_kl, hsd_growth_pct, hsd_ta_gain_pp,
	qoc_count, lubricants_value, ufill_count, speed_vol_kl,
	score_ms_vol, score_ms_growth, score_ms_ta_gain,
	score_hsd_vol, score_hsd_growth, score_hsd_ta_gain,
	score_qoc, score_lubricants, score_ufill, score_speed,
	score_cleanliness, score_sangam, score_google, score_ips, score_bonus,
	total_score, rank, computed_at, created_at
	FROM competition_scores`

func (r *CompetitionRepo) GetScores(ctx context.Context, competitionID uuid.UUID) ([]*model.CompetitionScore, error) {
	rows, err := r.pool.Query(ctx,
		scoreSelect+` WHERE competition_id = $1 ORDER BY rank ASC NULLS LAST`,
		competitionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*model.CompetitionScore
	for rows.Next() {
		s, scanErr := scanScore(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, s)
	}
	return result, rows.Err()
}

func (r *CompetitionRepo) GetDealerScore(ctx context.Context, competitionID uuid.UUID, cc string) (*model.CompetitionScore, error) {
	row := r.pool.QueryRow(ctx, scoreSelect+` WHERE competition_id = $1 AND cc_number = $2`, competitionID, cc)
	s, err := scanScore(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return s, nil
}

func scanScore(row pgx.Row) (*model.CompetitionScore, error) {
	var s model.CompetitionScore
	var rawID, rawCompID [16]byte
	var rank pgtype.Int4
	var computedAt pgtype.Timestamptz

	// 30 nullable numeric fields
	var (
		msVolKL, msGrowthPct, msTAGainPP     pgtype.Numeric
		hsdVolKL, hsdGrowthPct, hsdTAGainPP  pgtype.Numeric
		qocCount, lubValue, ufillCount        pgtype.Numeric
		speedVolKL                            pgtype.Numeric
		sMSVol, sMSGrowth, sMSTAGain         pgtype.Numeric
		sHSDVol, sHSDGrowth, sHSDTAGain      pgtype.Numeric
		sQOC, sLub, sUFill, sSpeed           pgtype.Numeric
		sCleanliness, sSangam, sGoogle, sIPS pgtype.Numeric
		sBonus, totalScore                    pgtype.Numeric
	)

	err := row.Scan(
		&rawID, &rawCompID, &s.CCNumber,
		&msVolKL, &msGrowthPct, &msTAGainPP,
		&hsdVolKL, &hsdGrowthPct, &hsdTAGainPP,
		&qocCount, &lubValue, &ufillCount, &speedVolKL,
		&sMSVol, &sMSGrowth, &sMSTAGain,
		&sHSDVol, &sHSDGrowth, &sHSDTAGain,
		&sQOC, &sLub, &sUFill, &sSpeed,
		&sCleanliness, &sSangam, &sGoogle, &sIPS, &sBonus,
		&totalScore, &rank, &computedAt, &s.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	s.ID = uuid.UUID(rawID)
	s.CompetitionID = uuid.UUID(rawCompID)
	s.MSVolKL = numericToFloat64Ptr(msVolKL)
	s.MSGrowthPct = numericToFloat64Ptr(msGrowthPct)
	s.MSTAGainPP = numericToFloat64Ptr(msTAGainPP)
	s.HSDVolKL = numericToFloat64Ptr(hsdVolKL)
	s.HSDGrowthPct = numericToFloat64Ptr(hsdGrowthPct)
	s.HSDTAGainPP = numericToFloat64Ptr(hsdTAGainPP)
	s.QOCCount = numericToFloat64Ptr(qocCount)
	s.LubricantsValue = numericToFloat64Ptr(lubValue)
	s.UFillCount = numericToFloat64Ptr(ufillCount)
	s.SpeedVolKL = numericToFloat64Ptr(speedVolKL)
	s.ScoreMSVol = numericToFloat64Ptr(sMSVol)
	s.ScoreMSGrowth = numericToFloat64Ptr(sMSGrowth)
	s.ScoreMSTAGain = numericToFloat64Ptr(sMSTAGain)
	s.ScoreHSDVol = numericToFloat64Ptr(sHSDVol)
	s.ScoreHSDGrowth = numericToFloat64Ptr(sHSDGrowth)
	s.ScoreHSDTAGain = numericToFloat64Ptr(sHSDTAGain)
	s.ScoreQOC = numericToFloat64Ptr(sQOC)
	s.ScoreLubricants = numericToFloat64Ptr(sLub)
	s.ScoreUFill = numericToFloat64Ptr(sUFill)
	s.ScoreSpeed = numericToFloat64Ptr(sSpeed)
	s.ScoreCleanliness = numericToFloat64Ptr(sCleanliness)
	s.ScoreSangam = numericToFloat64Ptr(sSangam)
	s.ScoreGoogle = numericToFloat64Ptr(sGoogle)
	s.ScoreIPS = numericToFloat64Ptr(sIPS)
	s.ScoreBonus = numericToFloat64Ptr(sBonus)
	s.TotalScore = numericToFloat64Ptr(totalScore)
	if rank.Valid {
		v := int(rank.Int32)
		s.Rank = &v
	}
	if computedAt.Valid {
		t := computedAt.Time
		s.ComputedAt = &t
	}
	return &s, nil
}

// ─── DealerAuditScore ────────────────────────────────────────────────────────

func (r *CompetitionRepo) GetAuditScore(ctx context.Context, cc string, period time.Time) (*model.DealerAuditScore, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, cc_number, period, cleanliness_grade, cleanliness_status,
			sangam_count, google_rating, ips_pct, audit_status, audited_by, audited_at, created_at
		FROM dealer_audit_scores WHERE cc_number = $1 AND period = $2`,
		cc, period,
	)
	return scanAuditScore(row)
}

func scanAuditScore(row pgx.Row) (*model.DealerAuditScore, error) {
	var a model.DealerAuditScore
	var rawID [16]byte
	var cleanGrade pgtype.Text
	var sangamCount pgtype.Int4
	var googleRating, ipsPct pgtype.Numeric
	var auditedBy pgtype.UUID
	var auditedAt pgtype.Timestamptz

	err := row.Scan(
		&rawID, &a.CCNumber, &a.Period,
		&cleanGrade, &a.CleanlinessStatus,
		&sangamCount, &googleRating, &ipsPct,
		&a.AuditStatus, &auditedBy, &auditedAt, &a.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	a.ID = uuid.UUID(rawID)
	if cleanGrade.Valid {
		a.CleanlinessGrade = &cleanGrade.String
	}
	if sangamCount.Valid {
		v := int(sangamCount.Int32)
		a.SangamCount = &v
	}
	a.GoogleRating = numericToFloat64Ptr(googleRating)
	a.IPSPct = numericToFloat64Ptr(ipsPct)
	a.AuditedBy = pguuid(auditedBy)
	if auditedAt.Valid {
		t := auditedAt.Time
		a.AuditedAt = &t
	}
	return &a, nil
}

// ─── CompetitionBonus ────────────────────────────────────────────────────────

func (r *CompetitionRepo) GetBonus(ctx context.Context, competitionID uuid.UUID, cc string) (*model.CompetitionBonus, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, competition_id, cc_number, bonus_marks, remarks, awarded_by, created_at
		FROM competition_bonus WHERE competition_id = $1 AND cc_number = $2`,
		competitionID, cc,
	)
	return scanBonus(row)
}

func (r *CompetitionRepo) UpsertBonus(ctx context.Context, b *model.CompetitionBonus) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO competition_bonus (competition_id, cc_number, bonus_marks, remarks, awarded_by)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (competition_id, cc_number) DO UPDATE SET
			bonus_marks = EXCLUDED.bonus_marks,
			remarks     = EXCLUDED.remarks,
			awarded_by  = EXCLUDED.awarded_by
		RETURNING id, created_at`,
		b.CompetitionID, b.CCNumber, b.BonusMarks, b.Remarks, b.AwardedBy,
	).Scan(&b.ID, &b.CreatedAt)
}

func scanBonus(row pgx.Row) (*model.CompetitionBonus, error) {
	var b model.CompetitionBonus
	var rawID, rawCompID, rawAwardedBy [16]byte
	var bonusMarks pgtype.Numeric

	err := row.Scan(&rawID, &rawCompID, &b.CCNumber, &bonusMarks, &b.Remarks, &rawAwardedBy, &b.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	b.ID = uuid.UUID(rawID)
	b.CompetitionID = uuid.UUID(rawCompID)
	b.AwardedBy = uuid.UUID(rawAwardedBy)
	if v := numericToFloat64Ptr(bonusMarks); v != nil {
		b.BonusMarks = *v
	}
	return &b, nil
}
