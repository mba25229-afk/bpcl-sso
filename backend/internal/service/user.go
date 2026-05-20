package service

import (
	"context"
	"errors"

	"github.com/bpcl/portal-api/internal/model"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	users UserRepository
}

func NewUserService(users UserRepository) *UserService {
	return &UserService{users: users}
}

type ListUsersParams struct {
	Role        string
	Territory   string
	Limit      int
	Offset     int
}

type ListUsersResponse struct {
	Users  []*model.User `json:"users"`
	Total  int         `json:"total"`
	Limit  int         `json:"limit"`
	Offset int         `json:"offset"`
}

type CreateUserInput struct {
	EmployeeID    string     `json:"employee_id"`
	Email        string     `json:"email"`
	Name         string     `json:"name"`
	Role         model.UserRole `json:"role"`
	TerritoryCode *string   `json:"territory_code"`
	Password     string     `json:"password"`
}

type UpdateUserInput struct {
	Name         string     `json:"name"`
	Role         model.UserRole `json:"role"`
	TerritoryCode *string   `json:"territory_code"`
	IsActive     bool       `json:"is_active"`
}

type ResetPasswordInput struct {
	NewPassword string `json:"new_password"`
}

func (s *UserService) ListUsers(ctx context.Context, p ListUsersParams) (*ListUsersResponse, error) {
	if p.Limit <= 0 {
		p.Limit = 20
	}
	if p.Limit > 100 {
		p.Limit = 100
	}

	users, total, err := s.users.List(ctx, p.Role, p.Territory, p.Limit, p.Offset)
	if err != nil {
		return nil, err
	}

	return &ListUsersResponse{
		Users:  users,
		Total:  total,
		Limit:  p.Limit,
		Offset: p.Offset,
	}, nil
}

func (s *UserService) CreateUser(ctx context.Context, input CreateUserInput) (*model.User, error) {
	claims, err := ExtractClaims(ctx)
	if err != nil {
		return nil, err
	}
	if model.UserRole(claims.Role) != model.RoleAdmin {
		return nil, model.ErrForbidden
	}

	if len(input.Password) < 8 {
		return nil, errors.New("password must be at least 8 characters")
	}
	if !hasValidPassword(input.Password) {
		return nil, errors.New("password must contain uppercase, lowercase, and number")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	existing, _ := s.users.GetByEmployeeID(ctx, input.EmployeeID)
	if existing != nil {
		return nil, errors.New("employee_id already exists")
	}

	u := &model.User{
		ID:            uuid.New(),
		EmployeeID:    input.EmployeeID,
		Email:         input.Email,
		Name:          input.Name,
		PasswordHash:  string(hash),
		Role:         input.Role,
		TerritoryCode: input.TerritoryCode,
		IsActive:      true,
	}

	if err := s.users.Create(ctx, u); err != nil {
		return nil, err
	}

	u.PasswordHash = ""
	return u, nil
}

func (s *UserService) UpdateUser(ctx context.Context, id uuid.UUID, input UpdateUserInput) (*model.User, error) {
	claims, err := ExtractClaims(ctx)
	if err != nil {
		return nil, err
	}
	if model.UserRole(claims.Role) != model.RoleAdmin {
		return nil, model.ErrForbidden
	}

	if err := s.users.Update(ctx, id, input.Name, input.Role, input.TerritoryCode, input.IsActive); err != nil {
		return nil, err
	}

	return s.users.GetByID(ctx, id)
}

func (s *UserService) ResetPassword(ctx context.Context, id uuid.UUID, input ResetPasswordInput) error {
	claims, err := ExtractClaims(ctx)
	if err != nil {
		return err
	}
	if model.UserRole(claims.Role) != model.RoleAdmin {
		return model.ErrForbidden
	}

	if len(input.NewPassword) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	if !hasValidPassword(input.NewPassword) {
		return errors.New("password must contain uppercase, lowercase, and number")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.users.UpdatePassword(ctx, id, string(hash))
}

func hasValidPassword(pw string) bool {
	var hasUpper, hasLower, hasDigit bool
	for _, c := range pw {
		switch {
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= '0' && c <= '9':
			hasDigit = true
		}
	}
	return hasUpper && hasLower && hasDigit
}