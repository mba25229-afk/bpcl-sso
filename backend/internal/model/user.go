package model

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleROManager         UserRole = "ro_manager"
	RoleTerritoryManager  UserRole = "territory_manager"
	RoleAdmin             UserRole = "admin"
)

type User struct {
	ID            uuid.UUID  `json:"id"`
	EmployeeID    string     `json:"employee_id"`
	Email         string     `json:"email"`
	Name          string     `json:"name"`
	PasswordHash  string     `json:"-"`
	Role          UserRole   `json:"role"`
	TerritoryCode *string    `json:"territory_code"`
	IsActive      bool       `json:"is_active"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	LastLoginAt   *time.Time `json:"last_login_at"`
}
