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

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

const userSelect = `SELECT id, employee_id, email, name, password_hash, role,
	territory_code, is_active, created_at, updated_at, last_login_at FROM users`

func (r *UserRepo) GetByEmployeeID(ctx context.Context, employeeID string) (*model.User, error) {
	row := r.pool.QueryRow(ctx, userSelect+` WHERE employee_id = $1`, employeeID)
	return scanUser(row)
}

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	row := r.pool.QueryRow(ctx, userSelect+` WHERE id = $1`, id)
	return scanUser(row)
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	row := r.pool.QueryRow(ctx, userSelect+` WHERE email = $1`, email)
	return scanUser(row)
}

func (r *UserRepo) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET last_login_at = $1, updated_at = $1 WHERE id = $2`, now, id)
	return err
}

func scanUser(row pgx.Row) (*model.User, error) {
	var u model.User
	var rawID [16]byte
	var role string
	var territory pgtype.Text
	var lastLogin pgtype.Timestamptz

	err := row.Scan(
		&rawID, &u.EmployeeID, &u.Email, &u.Name, &u.PasswordHash, &role,
		&territory, &u.IsActive, &u.CreatedAt, &u.UpdatedAt, &lastLogin,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	u.ID = uuid.UUID(rawID)
	u.Role = model.UserRole(role)
	if territory.Valid {
		u.TerritoryCode = &territory.String
	}
	if lastLogin.Valid {
		t := lastLogin.Time
		u.LastLoginAt = &t
	}
	return &u, nil
}

func (r *UserRepo) List(ctx context.Context, role string, territory string, limit, offset int) ([]*model.User, int, error) {
	args := []interface{}{limit, offset}
	where := "1=1"
	if role != "" {
		where += " AND role = $3"
		args = append(args, role)
	}
	if territory != "" {
		where += " AND territory_code = $4"
		args = append(args, territory)
	}

	query := userSelect + " WHERE " + where + " ORDER BY name LIMIT $1 OFFSET $2"
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []*model.User
	for rows.Next() {
		var u model.User
		var rawID [16]byte
		var rstr string
		var territory pgtype.Text
		var lastLogin pgtype.Timestamptz
		err := rows.Scan(
			&rawID, &u.EmployeeID, &u.Email, &u.Name, &u.PasswordHash, &rstr,
			&territory, &u.IsActive, &u.CreatedAt, &u.UpdatedAt, &lastLogin,
		)
		if err != nil {
			return nil, 0, err
		}
		u.ID = uuid.UUID(rawID)
		u.Role = model.UserRole(rstr)
		if territory.Valid {
			u.TerritoryCode = &territory.String
		}
		if lastLogin.Valid {
			t := lastLogin.Time
			u.LastLoginAt = &t
		}
		result = append(result, &u)
	}

	countQuery := "SELECT COUNT(*) FROM users WHERE " + where
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args[2:]...).Scan(&total); err != nil {
		return nil, 0, err
	}

	return result, total, rows.Err()
}

func (r *UserRepo) Create(ctx context.Context, u *model.User) error {
	query := `
		INSERT INTO users (id, employee_id, email, name, password_hash, role, territory_code, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.pool.Exec(ctx, query,
		u.ID, u.EmployeeID, u.Email, u.Name, u.PasswordHash, u.Role, u.TerritoryCode, u.IsActive)
	return err
}

func (r *UserRepo) Update(ctx context.Context, id uuid.UUID, name string, role model.UserRole, territory *string, isActive bool) error {
	query := `
		UPDATE users SET name = $1, role = $2, territory_code = $3, is_active = $4, updated_at = NOW()
		WHERE id = $5`
	_, err := r.pool.Exec(ctx, query, name, role, territory, isActive, id)
	return err
}

func (r *UserRepo) UpdatePassword(ctx context.Context, id uuid.UUID, hash string) error {
	query := `UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, hash, id)
	return err
}
