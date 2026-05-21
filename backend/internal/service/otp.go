package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/bpcl/portal-api/internal/model"
	"golang.org/x/crypto/bcrypt"
)

const otpTTL = 10 * time.Minute

type OTPService struct {
	otpRepo  OTPRepository
	userRepo UserRepository
	emailer  EmailSender
}

func NewOTPService(otpRepo OTPRepository, userRepo UserRepository, emailer EmailSender) *OTPService {
	return &OTPService{otpRepo: otpRepo, userRepo: userRepo, emailer: emailer}
}

func (s *OTPService) GenerateAndSend(ctx context.Context, email string) error {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil // silent — no email enumeration
	}

	otp, err := generateOTP()
	if err != nil {
		return fmt.Errorf("otp: generate: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(otp), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("otp: hash: %w", err)
	}

	_ = s.otpRepo.DeleteByEmail(ctx, email)
	if err := s.otpRepo.Insert(ctx, email, string(hash), time.Now().Add(otpTTL)); err != nil {
		return fmt.Errorf("otp: store: %w", err)
	}

	body := fmt.Sprintf(
		"Hello %s,\n\nYour BPCL Insight password reset OTP is: %s\n\nThis OTP expires in 10 minutes.\n\nIf you did not request this, ignore this email.",
		user.Name, otp,
	)
	return s.emailer.Send(email, "BPCL Insight — Password Reset OTP", body)
}

func (s *OTPService) ResetPassword(ctx context.Context, email, otp, newPassword string) error {
	if len(newPassword) < 8 {
		return fmt.Errorf("otp: password must be at least 8 characters")
	}

	otpHash, expiresAt, err := s.otpRepo.GetLatestUnused(ctx, email)
	if err != nil {
		return model.ErrUnauthorized
	}
	if time.Now().After(expiresAt) {
		return model.ErrUnauthorized
	}
	if err := bcrypt.CompareHashAndPassword([]byte(otpHash), []byte(otp)); err != nil {
		return model.ErrUnauthorized
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return model.ErrUnauthorized
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("otp: hash password: %w", err)
	}
	if err := s.userRepo.UpdatePassword(ctx, user.ID, string(newHash)); err != nil {
		return fmt.Errorf("otp: update password: %w", err)
	}

	_ = s.otpRepo.MarkUsed(ctx, email)
	return nil
}

func generateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
