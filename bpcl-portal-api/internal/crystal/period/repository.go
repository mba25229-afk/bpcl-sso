package period

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool  *pgxpool.Pool
	mu    sync.RWMutex
	cache *activePeriod
}

type activePeriod struct {
	monthYear  time.Time
	totalSlots int
	loadedAt   time.Time
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Active returns the current active competition period.
// Cached for 60s — admin changes take effect within a minute.
func (r *Repository) Active(ctx context.Context) (time.Time, int, error) {
	r.mu.RLock()
	if r.cache != nil && time.Since(r.cache.loadedAt) < 60*time.Second {
		t, s := r.cache.monthYear, r.cache.totalSlots
		r.mu.RUnlock()
		return t, s, nil
	}
	r.mu.RUnlock()

	var monthYear time.Time
	var totalSlots int
	err := r.pool.QueryRow(ctx,
		`SELECT month_year, total_slots FROM cr_competition_periods WHERE is_active = TRUE LIMIT 1`,
	).Scan(&monthYear, &totalSlots)
	if err != nil {
		return time.Time{}, 0, fmt.Errorf("no active competition period: %w", err)
	}

	r.mu.Lock()
	r.cache = &activePeriod{monthYear: monthYear, totalSlots: totalSlots, loadedAt: time.Now()}
	r.mu.Unlock()

	return monthYear, totalSlots, nil
}

// DateInActivePeriod returns true if d falls within the active period's calendar month.
func (r *Repository) DateInActivePeriod(ctx context.Context, d time.Time) (bool, error) {
	monthYear, _, err := r.Active(ctx)
	if err != nil {
		return false, err
	}
	return d.Year() == monthYear.Year() && d.Month() == monthYear.Month(), nil
}
