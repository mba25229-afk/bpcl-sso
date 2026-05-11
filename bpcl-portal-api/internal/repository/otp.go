package repository

import (
	"context"
	"errors"
	"time"

	"github.com/bpcl/portal-api/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OTPRepo struct {
	pool *pgxpool.Pool
}

func NewOTPRepo(pool *pgxpool.Pool) *OTPRepo {
	return &OTPRepo{pool: pool}
}

func (r *OTPRepo) DeleteByEmail(ctx context.Context, email string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM password_reset_otps WHERE email = $1`, email)
	return err
}

func (r *OTPRepo) Insert(ctx context.Context, email, otpHash string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO password_reset_otps (email, otp_hash, expires_at) VALUES ($1, $2, $3)`,
		email, otpHash, expiresAt,
	)
	return err
}

func (r *OTPRepo) GetLatestUnused(ctx context.Context, email string) (string, time.Time, error) {
	var hash string
	var exp time.Time
	err := r.pool.QueryRow(ctx,
		`SELECT otp_hash, expires_at FROM password_reset_otps
		 WHERE email = $1 AND used_at IS NULL
		 ORDER BY created_at DESC LIMIT 1`,
		email,
	).Scan(&hash, &exp)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", time.Time{}, model.ErrNotFound
		}
		return "", time.Time{}, err
	}
	return hash, exp, nil
}

func (r *OTPRepo) MarkUsed(ctx context.Context, email string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE password_reset_otps SET used_at = NOW()
		 WHERE email = $1 AND used_at IS NULL`,
		email,
	)
	return err
}
