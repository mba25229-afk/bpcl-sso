package service

import (
	"context"
	"fmt"
	"time"

	"github.com/bpcl/portal-api/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type contextKey string

const ClaimsKey contextKey = "claims"

type Claims struct {
	jwt.RegisteredClaims
	UserID        uuid.UUID `json:"user_id"`
	Role          string    `json:"role"`
	TerritoryCode *string   `json:"territory_code,omitempty"`
}

type LoginResponse struct {
	Token string     `json:"token"`
	User  *model.User `json:"user"`
}

type AuthService struct {
	users     UserRepository
	jwtSecret []byte
	jwtExpiry time.Duration
}

func NewAuthService(users UserRepository, jwtSecret string, jwtExpiry time.Duration) *AuthService {
	return &AuthService{
		users:     users,
		jwtSecret: []byte(jwtSecret),
		jwtExpiry: jwtExpiry,
	}
}

func (s *AuthService) Login(ctx context.Context, employeeID, password string) (*LoginResponse, error) {
	user, err := s.users.GetByEmployeeID(ctx, employeeID)
	if err != nil {
		return nil, model.ErrUnauthorized
	}
	if !user.IsActive {
		return nil, model.ErrUnauthorized
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, model.ErrUnauthorized
	}

	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.jwtExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID:        user.ID,
		Role:          string(user.Role),
		TerritoryCode: user.TerritoryCode,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("auth: sign token: %w", err)
	}

	_ = s.users.UpdateLastLogin(ctx, user.ID)

	return &LoginResponse{Token: tokenStr, User: user}, nil
}

func (s *AuthService) ValidateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("auth: unexpected signing method: %v", t.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, model.ErrUnauthorized
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, model.ErrUnauthorized
	}
	return claims, nil
}

func ExtractClaims(ctx context.Context) (*Claims, error) {
	c, ok := ctx.Value(ClaimsKey).(*Claims)
	if !ok || c == nil {
		return nil, model.ErrUnauthorized
	}
	return c, nil
}
