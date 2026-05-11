package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bpcl/portal-api/internal/model"
	"github.com/bpcl/portal-api/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// ── fakes ────────────────────────────────────────────────────────────────────

type fakeOTPRepo struct {
	stored   map[string]string
	expiries map[string]time.Time
	usedAt   map[string]bool
}

func newFakeOTPRepo() *fakeOTPRepo {
	return &fakeOTPRepo{
		stored:   make(map[string]string),
		expiries: make(map[string]time.Time),
		usedAt:   make(map[string]bool),
	}
}
func (f *fakeOTPRepo) DeleteByEmail(_ context.Context, email string) error {
	delete(f.stored, email)
	return nil
}
func (f *fakeOTPRepo) Insert(_ context.Context, email, hash string, exp time.Time) error {
	f.stored[email] = hash
	f.expiries[email] = exp
	return nil
}
func (f *fakeOTPRepo) GetLatestUnused(_ context.Context, email string) (string, time.Time, error) {
	h, ok := f.stored[email]
	if !ok || f.usedAt[email] {
		return "", time.Time{}, model.ErrNotFound
	}
	return h, f.expiries[email], nil
}
func (f *fakeOTPRepo) MarkUsed(_ context.Context, email string) error {
	f.usedAt[email] = true
	return nil
}

type fakeEmailSender struct{ sent []string }

func (f *fakeEmailSender) Send(to, _, _ string) error {
	f.sent = append(f.sent, to)
	return nil
}

type fakeUserRepoForOTP struct {
	user *model.User
}

func (f *fakeUserRepoForOTP) GetByEmail(_ context.Context, email string) (*model.User, error) {
	if f.user != nil && f.user.Email == email {
		return f.user, nil
	}
	return nil, model.ErrNotFound
}
func (f *fakeUserRepoForOTP) GetByEmployeeID(_ context.Context, _ string) (*model.User, error) {
	return nil, model.ErrNotFound
}
func (f *fakeUserRepoForOTP) GetByID(_ context.Context, _ uuid.UUID) (*model.User, error) {
	return nil, model.ErrNotFound
}
func (f *fakeUserRepoForOTP) List(_ context.Context, _, _ string, _, _ int) ([]*model.User, int, error) {
	return nil, 0, nil
}
func (f *fakeUserRepoForOTP) Create(_ context.Context, _ *model.User) error { return nil }
func (f *fakeUserRepoForOTP) Update(_ context.Context, _ uuid.UUID, _ string, _ model.UserRole, _ *string, _ bool) error {
	return nil
}
func (f *fakeUserRepoForOTP) UpdatePassword(_ context.Context, _ uuid.UUID, _ string) error {
	return nil
}
func (f *fakeUserRepoForOTP) UpdateLastLogin(_ context.Context, _ uuid.UUID) error { return nil }

// ── tests ────────────────────────────────────────────────────────────────────

func TestGenerateAndSend_UnknownEmail_Silent(t *testing.T) {
	otpRepo := newFakeOTPRepo()
	emailer := &fakeEmailSender{}
	userRepo := &fakeUserRepoForOTP{user: nil}
	svc := service.NewOTPService(otpRepo, userRepo, emailer)

	err := svc.GenerateAndSend(context.Background(), "unknown@example.com")
	require.NoError(t, err)
	assert.Empty(t, emailer.sent)
}

func TestGenerateAndSend_Success(t *testing.T) {
	otpRepo := newFakeOTPRepo()
	emailer := &fakeEmailSender{}
	user := &model.User{ID: uuid.New(), Email: "emp@bpcl.com", Name: "Test"}
	userRepo := &fakeUserRepoForOTP{user: user}
	svc := service.NewOTPService(otpRepo, userRepo, emailer)

	err := svc.GenerateAndSend(context.Background(), "emp@bpcl.com")
	require.NoError(t, err)
	assert.Equal(t, []string{"emp@bpcl.com"}, emailer.sent)
	assert.NotEmpty(t, otpRepo.stored["emp@bpcl.com"])
}

func TestResetPassword_WrongOTP(t *testing.T) {
	otpRepo := newFakeOTPRepo()
	hash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.MinCost)
	otpRepo.stored["emp@bpcl.com"] = string(hash)
	otpRepo.expiries["emp@bpcl.com"] = time.Now().Add(10 * time.Minute)

	user := &model.User{ID: uuid.New(), Email: "emp@bpcl.com"}
	userRepo := &fakeUserRepoForOTP{user: user}
	svc := service.NewOTPService(otpRepo, userRepo, &fakeEmailSender{})

	err := svc.ResetPassword(context.Background(), "emp@bpcl.com", "999999", "newpassword1")
	assert.True(t, errors.Is(err, model.ErrUnauthorized))
}

func TestResetPassword_ExpiredOTP(t *testing.T) {
	otpRepo := newFakeOTPRepo()
	hash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.MinCost)
	otpRepo.stored["emp@bpcl.com"] = string(hash)
	otpRepo.expiries["emp@bpcl.com"] = time.Now().Add(-1 * time.Minute)

	user := &model.User{ID: uuid.New(), Email: "emp@bpcl.com"}
	userRepo := &fakeUserRepoForOTP{user: user}
	svc := service.NewOTPService(otpRepo, userRepo, &fakeEmailSender{})

	err := svc.ResetPassword(context.Background(), "emp@bpcl.com", "123456", "newpassword1")
	assert.True(t, errors.Is(err, model.ErrUnauthorized))
}

func TestResetPassword_Success(t *testing.T) {
	otpRepo := newFakeOTPRepo()
	hash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.MinCost)
	otpRepo.stored["emp@bpcl.com"] = string(hash)
	otpRepo.expiries["emp@bpcl.com"] = time.Now().Add(10 * time.Minute)

	user := &model.User{ID: uuid.New(), Email: "emp@bpcl.com"}
	userRepo := &fakeUserRepoForOTP{user: user}
	svc := service.NewOTPService(otpRepo, userRepo, &fakeEmailSender{})

	err := svc.ResetPassword(context.Background(), "emp@bpcl.com", "123456", "newpassword1")
	require.NoError(t, err)
	assert.True(t, otpRepo.usedAt["emp@bpcl.com"])
}

func TestResetPassword_ShortPassword(t *testing.T) {
	svc := service.NewOTPService(newFakeOTPRepo(), &fakeUserRepoForOTP{}, &fakeEmailSender{})
	err := svc.ResetPassword(context.Background(), "emp@bpcl.com", "123456", "short")
	assert.Error(t, err)
}
