package dealer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool     *pgxpool.Pool
	mu       sync.RWMutex
	cache    map[string]bool // cc_code -> active
	loadedAt time.Time
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool, cache: make(map[string]bool)}
}

// IsActive returns true if cc_code exists in cr_dealers and is active.
// Cache TTL: 5 minutes (dealer list changes rarely).
func (r *Repository) IsActive(ctx context.Context, ccCode string) (bool, error) {
	r.mu.RLock()
	if time.Since(r.loadedAt) < 5*time.Minute {
		active, ok := r.cache[ccCode]
		r.mu.RUnlock()
		if ok {
			return active, nil
		}
		return false, nil
	}
	r.mu.RUnlock()

	rows, err := r.pool.Query(ctx, `SELECT cc_code, is_active FROM cr_dealers`)
	if err != nil {
		return false, fmt.Errorf("load dealers: %w", err)
	}
	defer rows.Close()

	fresh := make(map[string]bool)
	for rows.Next() {
		var cc string
		var active bool
		if err := rows.Scan(&cc, &active); err != nil {
			return false, err
		}
		fresh[cc] = active
	}

	r.mu.Lock()
	r.cache = fresh
	r.loadedAt = time.Now()
	r.mu.Unlock()

	return fresh[ccCode], nil
}

// ValidateCodes returns a map of invalid (unknown or inactive) cc_codes.
func (r *Repository) ValidateCodes(ctx context.Context, codes []string) (map[string]bool, error) {
	invalid := make(map[string]bool)
	for _, c := range codes {
		ok, err := r.IsActive(ctx, c)
		if err != nil {
			return nil, err
		}
		if !ok {
			invalid[c] = true
		}
	}
	return invalid, nil
}
