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
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	User         *model.User `json:"user"`
}

type RefreshResponse struct {
	AccessToken string `json:"access_token"`
}

const refreshExpiry = 7 * 24 * time.Hour

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

	refreshClaims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(refreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID: user.ID,
		Role:   string(user.Role),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshStr, err := refreshToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("auth: sign refresh token: %w", err)
	}

	return &LoginResponse{AccessToken: tokenStr, RefreshToken: refreshStr, User: user}, nil
}

// RefreshAccessToken verifies a refresh token and returns a new short-lived access token.
func (s *AuthService) RefreshAccessToken(refreshTokenStr string) (*RefreshResponse, error) {
	token, err := jwt.ParseWithClaims(refreshTokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("auth: unexpected signing method")
		}
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, model.ErrUnauthorized
	}
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, model.ErrUnauthorized
	}

	newClaims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   claims.Subject,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.jwtExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID:        claims.UserID,
		Role:          claims.Role,
		TerritoryCode: claims.TerritoryCode,
	}
	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	newTokenStr, err := newToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("auth: sign access token: %w", err)
	}
	return &RefreshResponse{AccessToken: newTokenStr}, nil
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
