# Forgot Password (OTP via Email) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a "Forgot password?" OTP email flow — user submits email, receives a 6-digit OTP, enters it with a new password to reset.

**Architecture:** DB table stores bcrypt-hashed OTPs with 10-min TTL. Two public API endpoints (no auth). React modal with two steps managed by local state. SMTP via Go stdlib `net/smtp`.

**Tech Stack:** Go (net/smtp, bcrypt), PostgreSQL, React + TypeScript, shadcn Dialog + InputOTP

---

## File Map

| Action | Path |
|--------|------|
| Create | `migrations/000012_password_reset_otps.up.sql` |
| Create | `migrations/000012_password_reset_otps.down.sql` |
| Modify | `bpcl-portal-api/internal/config/config.go` |
| Modify | `bpcl-portal-api/internal/service/interfaces.go` |
| Modify | `bpcl-portal-api/internal/repository/user.go` |
| Create | `bpcl-portal-api/internal/repository/otp.go` |
| Create | `bpcl-portal-api/internal/service/email.go` |
| Create | `bpcl-portal-api/internal/service/otp.go` |
| Create | `bpcl-portal-api/internal/service/otp_test.go` |
| Modify | `bpcl-portal-api/internal/handler/interfaces.go` |
| Modify | `bpcl-portal-api/internal/handler/handler.go` |
| Create | `bpcl-portal-api/internal/handler/otp.go` |
| Create | `bpcl-portal-api/internal/handler/otp_test.go` |
| Modify | `bpcl-portal-api/internal/handler/helpers_test.go` |
| Modify | `bpcl-portal-api/internal/router/router.go` |
| Modify | `bpcl-portal-api/cmd/api/main.go` |
| Modify | `Bpclssoportal-main/src/api/client.ts` |
| Create | `Bpclssoportal-main/src/app/components/ForgotPasswordModal.tsx` |
| Modify | `Bpclssoportal-main/src/app/LoginPage.tsx` |

---

### Task 1: DB Migration

**Files:**
- Create: `migrations/000012_password_reset_otps.up.sql`
- Create: `migrations/000012_password_reset_otps.down.sql`

- [ ] **Step 1: Create up migration**

```sql
-- migrations/000012_password_reset_otps.up.sql
CREATE TABLE password_reset_otps (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email       TEXT NOT NULL,
  otp_hash    TEXT NOT NULL,
  expires_at  TIMESTAMPTZ NOT NULL,
  used_at     TIMESTAMPTZ,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_prot_email ON password_reset_otps (email);
```

- [ ] **Step 2: Create down migration**

```sql
-- migrations/000012_password_reset_otps.down.sql
DROP TABLE IF EXISTS password_reset_otps;
```

- [ ] **Step 3: Run migration**

```bash
cd /Users/meynesh/Documents/bpcl-sso
bash scripts/migrate.sh up
```

Expected: migration applies without error.

- [ ] **Step 4: Commit**

```bash
git add migrations/000012_password_reset_otps.up.sql migrations/000012_password_reset_otps.down.sql
git commit -m "feat: add password_reset_otps migration"
```

---

### Task 2: Config — Add SMTP Fields

**Files:**
- Modify: `bpcl-portal-api/internal/config/config.go`
- Modify: `bpcl-portal-api/.env`

- [ ] **Step 1: Add SMTP fields to Config struct**

In `config.go`, add to the `Config` struct after `RateLimitRPM int`:

```go
SMTPHost string
SMTPPort string
SMTPUser string
SMTPPass string
SMTPFrom string
```

- [ ] **Step 2: Load SMTP fields in Load()**

In the `Load()` function, add to the returned `&Config{...}`:

```go
SMTPHost: viper.GetString("BPCL_SMTP_HOST"),
SMTPPort: viper.GetString("BPCL_SMTP_PORT"),
SMTPUser: viper.GetString("BPCL_SMTP_USER"),
SMTPPass: viper.GetString("BPCL_SMTP_PASS"),
SMTPFrom: viper.GetString("BPCL_SMTP_FROM"),
```

- [ ] **Step 3: Add SMTP vars to .env**

Append to `bpcl-portal-api/.env`:

```
BPCL_SMTP_HOST=smtp.gmail.com
BPCL_SMTP_PORT=587
BPCL_SMTP_USER=your-gmail@gmail.com
BPCL_SMTP_PASS=your-16-char-app-password
BPCL_SMTP_FROM=BPCL Insight <your-gmail@gmail.com>
```

Also append to `bpcl-portal-api/.env.example` (same keys, empty values).

- [ ] **Step 4: Verify build compiles**

```bash
cd bpcl-portal-api && go build ./...
```

Expected: no errors.

- [ ] **Step 5: Commit**

```bash
git add bpcl-portal-api/internal/config/config.go bpcl-portal-api/.env.example
git commit -m "feat: add SMTP config fields"
```

---

### Task 3: Repository — GetByEmail + OTPRepo

**Files:**
- Modify: `bpcl-portal-api/internal/service/interfaces.go`
- Modify: `bpcl-portal-api/internal/repository/user.go`
- Create: `bpcl-portal-api/internal/repository/otp.go`

- [ ] **Step 1: Add GetByEmail to UserRepository interface**

In `bpcl-portal-api/internal/service/interfaces.go`, add `GetByEmail` to the `UserRepository` interface after `GetByEmployeeID`:

```go
GetByEmail(ctx context.Context, email string) (*model.User, error)
```

- [ ] **Step 2: Implement GetByEmail in UserRepo**

In `bpcl-portal-api/internal/repository/user.go`, add after `GetByID`:

```go
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	row := r.pool.QueryRow(ctx, userSelect+` WHERE email = $1`, email)
	return scanUser(row)
}
```

- [ ] **Step 3: Add OTPRepository interface to service/interfaces.go**

Append to `bpcl-portal-api/internal/service/interfaces.go`:

```go
type OTPRepository interface {
	DeleteByEmail(ctx context.Context, email string) error
	Insert(ctx context.Context, email, otpHash string, expiresAt time.Time) error
	GetLatestUnused(ctx context.Context, email string) (otpHash string, expiresAt time.Time, err error)
	MarkUsed(ctx context.Context, email string) error
}
```

Make sure `time` is imported at the top of interfaces.go.

- [ ] **Step 4: Create repository/otp.go**

```go
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
```

- [ ] **Step 5: Verify build**

```bash
cd bpcl-portal-api && go build ./...
```

Expected: no errors.

- [ ] **Step 6: Commit**

```bash
git add bpcl-portal-api/internal/service/interfaces.go \
        bpcl-portal-api/internal/repository/user.go \
        bpcl-portal-api/internal/repository/otp.go
git commit -m "feat: add GetByEmail and OTPRepo"
```

---

### Task 4: Service — EmailSender + OTPService

**Files:**
- Create: `bpcl-portal-api/internal/service/email.go`
- Create: `bpcl-portal-api/internal/service/otp.go`
- Create: `bpcl-portal-api/internal/service/otp_test.go`

- [ ] **Step 1: Write failing tests for OTPService**

Create `bpcl-portal-api/internal/service/otp_test.go`:

```go
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
	stored   map[string]string // email → hash
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
	assert.Empty(t, emailer.sent) // no email sent, no error revealed
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
	otpRepo.expiries["emp@bpcl.com"] = time.Now().Add(-1 * time.Minute) // already expired

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
```

- [ ] **Step 2: Run tests — verify they fail**

```bash
cd bpcl-portal-api && go test ./internal/service/... -run TestGenerateAndSend -v 2>&1 | head -20
```

Expected: compile error — `service.NewOTPService undefined`.

- [ ] **Step 3: Create service/email.go**

```go
package service

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
)

type EmailSender interface {
	Send(to, subject, body string) error
}

type SMTPConfig struct {
	Host string
	Port string
	User string
	Pass string
	From string
}

type SMTPSender struct {
	cfg SMTPConfig
}

func NewSMTPSender(cfg SMTPConfig) *SMTPSender {
	return &SMTPSender{cfg: cfg}
}

func (s *SMTPSender) Send(to, subject, body string) error {
	addr := net.JoinHostPort(s.cfg.Host, s.cfg.Port)
	auth := smtp.PlainAuth("", s.cfg.User, s.cfg.Pass, s.cfg.Host)

	msg := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		s.cfg.From, to, subject, body,
	))

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("smtp: dial: %w", err)
	}
	client, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		return fmt.Errorf("smtp: new client: %w", err)
	}
	defer client.Quit()

	tlsCfg := &tls.Config{ServerName: s.cfg.Host}
	if err := client.StartTLS(tlsCfg); err != nil {
		return fmt.Errorf("smtp: starttls: %w", err)
	}
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("smtp: auth: %w", err)
	}
	if err := client.Mail(s.cfg.User); err != nil {
		return fmt.Errorf("smtp: mail from: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp: rcpt: %w", err)
	}
	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp: data: %w", err)
	}
	if _, err = wc.Write(msg); err != nil {
		return fmt.Errorf("smtp: write: %w", err)
	}
	return wc.Close()
}
```

- [ ] **Step 4: Create service/otp.go**

```go
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
```

- [ ] **Step 5: Run tests — verify they pass**

```bash
cd bpcl-portal-api && go test ./internal/service/... -run "TestGenerateAndSend|TestResetPassword" -v
```

Expected: all 5 tests PASS.

- [ ] **Step 6: Commit**

```bash
git add bpcl-portal-api/internal/service/email.go \
        bpcl-portal-api/internal/service/otp.go \
        bpcl-portal-api/internal/service/otp_test.go
git commit -m "feat: OTPService with email sender"
```

---

### Task 5: Handler — OTPServiceI + ForgotPassword + ResetPassword

**Files:**
- Modify: `bpcl-portal-api/internal/handler/interfaces.go`
- Modify: `bpcl-portal-api/internal/handler/handler.go`
- Create: `bpcl-portal-api/internal/handler/otp.go`
- Create: `bpcl-portal-api/internal/handler/otp_test.go`
- Modify: `bpcl-portal-api/internal/handler/helpers_test.go`

- [ ] **Step 1: Write failing handler tests**

Create `bpcl-portal-api/internal/handler/otp_test.go`:

```go
package handler_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bpcl/portal-api/internal/model"
	"github.com/bpcl/portal-api/internal/handler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockOTPSvc struct{ mock.Mock }

func (m *mockOTPSvc) GenerateAndSend(ctx context.Context, email string) error {
	return m.Called(ctx, email).Error(0)
}
func (m *mockOTPSvc) ResetPassword(ctx context.Context, email, otp, newPassword string) error {
	return m.Called(ctx, email, otp, newPassword).Error(0)
}

func TestForgotPassword_MissingEmail(t *testing.T) {
	otp := &mockOTPSvc{}
	h := &handler.Handler{OTP: otp}

	r := httptest.NewRequest(http.MethodPost, "/api/v1/auth/forgot-password",
		bytes.NewBufferString(`{}`))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ForgotPassword(w, r)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestForgotPassword_Success(t *testing.T) {
	otp := &mockOTPSvc{}
	h := &handler.Handler{OTP: otp}

	otp.On("GenerateAndSend", mock.Anything, "emp@bpcl.com").Return(nil)

	r := httptest.NewRequest(http.MethodPost, "/api/v1/auth/forgot-password",
		bytes.NewBufferString(`{"email":"emp@bpcl.com"}`))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ForgotPassword(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestResetForgottenPassword_MissingFields(t *testing.T) {
	otp := &mockOTPSvc{}
	h := &handler.Handler{OTP: otp}

	r := httptest.NewRequest(http.MethodPost, "/api/v1/auth/reset-password",
		bytes.NewBufferString(`{"email":"emp@bpcl.com"}`))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ResetForgottenPassword(w, r)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestResetForgottenPassword_InvalidOTP(t *testing.T) {
	otp := &mockOTPSvc{}
	h := &handler.Handler{OTP: otp}

	otp.On("ResetPassword", mock.Anything, "emp@bpcl.com", "000000", "newpass123").
		Return(model.ErrUnauthorized)

	r := httptest.NewRequest(http.MethodPost, "/api/v1/auth/reset-password",
		bytes.NewBufferString(`{"email":"emp@bpcl.com","otp":"000000","new_password":"newpass123"}`))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ResetForgottenPassword(w, r)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestResetForgottenPassword_Success(t *testing.T) {
	otp := &mockOTPSvc{}
	h := &handler.Handler{OTP: otp}

	otp.On("ResetPassword", mock.Anything, "emp@bpcl.com", "123456", "newpass123").Return(nil)

	r := httptest.NewRequest(http.MethodPost, "/api/v1/auth/reset-password",
		bytes.NewBufferString(`{"email":"emp@bpcl.com","otp":"123456","new_password":"newpass123"}`))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ResetForgottenPassword(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
}
```

- [ ] **Step 2: Run tests — verify they fail**

```bash
cd bpcl-portal-api && go test ./internal/handler/... -run "TestForgotPassword|TestResetForgotten" -v 2>&1 | head -20
```

Expected: compile error — `handler.Handler` has no `OTP` field.

- [ ] **Step 3: Add OTPServiceI to handler/interfaces.go**

Append to `bpcl-portal-api/internal/handler/interfaces.go`:

```go
type OTPServiceI interface {
	GenerateAndSend(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, email, otp, newPassword string) error
}
```

Make sure `context` is imported.

- [ ] **Step 4: Add OTP field to Handler struct and update New()**

In `bpcl-portal-api/internal/handler/handler.go`, add `OTP OTPServiceI` to the struct and a new `otp OTPServiceI` parameter to `New()`:

```go
package handler

type Handler struct {
	Auth        AuthServiceI
	Outlet      OutletServiceI
	Performance PerformanceServiceI
	Target      TargetServiceI
	Upload      UploadServiceI
	Competition CompetitionServiceI
	MarketShare MarketShareServiceI
	Users       UserServiceI
	OTP         OTPServiceI
}

func New(
	auth AuthServiceI,
	outlet OutletServiceI,
	perf PerformanceServiceI,
	target TargetServiceI,
	upload UploadServiceI,
	competition CompetitionServiceI,
	marketShare MarketShareServiceI,
	users UserServiceI,
	otp OTPServiceI,
) *Handler {
	return &Handler{
		Auth:        auth,
		Outlet:      outlet,
		Performance: perf,
		Target:      target,
		Upload:      upload,
		Competition: competition,
		MarketShare: marketShare,
		Users:       users,
		OTP:         otp,
	}
}
```

- [ ] **Step 5: Update helpers_test.go to pass nil for otp**

In `bpcl-portal-api/internal/handler/helpers_test.go`, update the `newHandler` call to add `nil` as the last arg:

```go
func newHandler(auth *mockAuthSvc, outlet *mockOutletSvc, perf *mockPerfSvc, target *mockTargetSvc, upload *mockUploadSvc, comp *mockCompSvc, marketShare *mockMarketShareSvc) *handler.Handler {
	users := &mockUserSvc{}
	return handler.New(auth, outlet, perf, target, upload, comp, marketShare, users, nil)
}
```

- [ ] **Step 6: Create handler/otp.go**

```go
package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/bpcl/portal-api/internal/model"
)

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

type resetPasswordRequest struct {
	Email       string `json:"email"`
	OTP         string `json:"otp"`
	NewPassword string `json:"new_password"`
}

func (h *Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req forgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid JSON body", "INVALID_BODY", "")
		return
	}
	if strings.TrimSpace(req.Email) == "" {
		writeError(w, http.StatusUnprocessableEntity, "email is required", "MISSING_FIELDS", "")
		return
	}
	// Fire-and-forget: never reveal whether email exists
	_ = h.OTP.GenerateAndSend(r.Context(), req.Email)
	writeJSON(w, http.StatusOK, map[string]string{
		"message": "If that email is registered, an OTP has been sent.",
	})
}

func (h *Handler) ResetForgottenPassword(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid JSON body", "INVALID_BODY", "")
		return
	}
	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.OTP) == "" || strings.TrimSpace(req.NewPassword) == "" {
		writeError(w, http.StatusUnprocessableEntity, "email, otp, and new_password are required", "MISSING_FIELDS", "")
		return
	}
	if err := h.OTP.ResetPassword(r.Context(), req.Email, req.OTP, req.NewPassword); err != nil {
		if errors.Is(err, model.ErrUnauthorized) {
			writeError(w, http.StatusUnauthorized, "invalid or expired OTP", "UNAUTHORIZED", "")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error(), "INVALID_REQUEST", "")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Password reset successful."})
}
```

- [ ] **Step 7: Run tests — verify they pass**

```bash
cd bpcl-portal-api && go test ./internal/handler/... -run "TestForgotPassword|TestResetForgotten" -v
```

Expected: all 4 tests PASS.

- [ ] **Step 8: Commit**

```bash
git add bpcl-portal-api/internal/handler/interfaces.go \
        bpcl-portal-api/internal/handler/handler.go \
        bpcl-portal-api/internal/handler/otp.go \
        bpcl-portal-api/internal/handler/otp_test.go \
        bpcl-portal-api/internal/handler/helpers_test.go
git commit -m "feat: OTP handler endpoints"
```

---

### Task 6: Router + Wire main.go

**Files:**
- Modify: `bpcl-portal-api/internal/router/router.go`
- Modify: `bpcl-portal-api/cmd/api/main.go`

- [ ] **Step 1: Add routes to router.go**

In `bpcl-portal-api/internal/router/router.go`, update the `New()` signature to accept `h *handler.Handler` (it already does). Add two routes in the `// ── Public routes` block after the existing public routes:

```go
mux.HandleFunc("POST /api/v1/auth/forgot-password", h.ForgotPassword)
mux.HandleFunc("POST /api/v1/auth/reset-password", h.ResetForgottenPassword)
```

- [ ] **Step 2: Wire OTP service in main.go**

In `bpcl-portal-api/cmd/api/main.go`:

After the existing `userSvc := service.NewUserService(userRepo)` line, add:

```go
otpRepo := repository.NewOTPRepo(pool)
emailSender := service.NewSMTPSender(service.SMTPConfig{
    Host: cfg.SMTPHost,
    Port: cfg.SMTPPort,
    User: cfg.SMTPUser,
    Pass: cfg.SMTPPass,
    From: cfg.SMTPFrom,
})
otpSvc := service.NewOTPService(otpRepo, userRepo, emailSender)
```

Then update the `handler.New(...)` call to pass `otpSvc` as the last argument:

```go
h := handler.New(authSvc, outletSvc, perfSvc, targetSvc, uploadSvc, compSvc, marketShareSvc, userSvc, otpSvc)
```

- [ ] **Step 3: Build and verify**

```bash
cd bpcl-portal-api && go build ./...
```

Expected: no errors.

- [ ] **Step 4: Run all tests**

```bash
cd bpcl-portal-api && go test ./...
```

Expected: all pass.

- [ ] **Step 5: Commit**

```bash
git add bpcl-portal-api/internal/router/router.go \
        bpcl-portal-api/cmd/api/main.go
git commit -m "feat: wire OTP routes and service in main"
```

---

### Task 7: Frontend — API Client + Modal + LoginPage

**Files:**
- Modify: `Bpclssoportal-main/src/api/client.ts`
- Create: `Bpclssoportal-main/src/app/components/ForgotPasswordModal.tsx`
- Modify: `Bpclssoportal-main/src/app/LoginPage.tsx`

- [ ] **Step 1: Add API methods to client.ts**

In `Bpclssoportal-main/src/api/client.ts`, append inside the `api` object (after `resetPassword`):

```ts
forgotPassword: (email: string) =>
  request('/api/v1/auth/forgot-password', {
    method: 'POST',
    body: JSON.stringify({ email }),
  }),

resetForgottenPassword: (email: string, otp: string, newPassword: string) =>
  request('/api/v1/auth/reset-password', {
    method: 'POST',
    body: JSON.stringify({ email, otp, new_password: newPassword }),
  }),
```

- [ ] **Step 2: Create ForgotPasswordModal.tsx**

```tsx
import { useState } from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from './ui/dialog';
import { InputOTP, InputOTPGroup, InputOTPSlot } from './ui/input-otp';
import { api } from '../../api/client';

interface Props {
  open: boolean;
  onClose: () => void;
}

type Step = 'email' | 'otp';

export function ForgotPasswordModal({ open, onClose }: Props) {
  const [step, setStep] = useState<Step>('email');
  const [email, setEmail] = useState('');
  const [otp, setOtp] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState(false);

  function handleClose() {
    setStep('email');
    setEmail('');
    setOtp('');
    setNewPassword('');
    setError('');
    setLoading(false);
    setSuccess(false);
    onClose();
  }

  async function handleSendOTP(e: React.FormEvent) {
    e.preventDefault();
    if (!email) { setError('Email is required'); return; }
    setLoading(true);
    setError('');
    try {
      await api.forgotPassword(email);
      setStep('otp');
    } catch {
      setError('Something went wrong. Please try again.');
    } finally {
      setLoading(false);
    }
  }

  async function handleReset(e: React.FormEvent) {
    e.preventDefault();
    if (otp.length !== 6) { setError('Enter the 6-digit OTP'); return; }
    if (newPassword.length < 8) { setError('Password must be at least 8 characters'); return; }
    setLoading(true);
    setError('');
    try {
      await api.resetForgottenPassword(email, otp, newPassword);
      setSuccess(true);
    } catch (err: any) {
      if (err.status === 401) {
        setError('Invalid or expired OTP. Please try again.');
      } else {
        setError(err.message || 'Something went wrong.');
      }
    } finally {
      setLoading(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Reset Password</DialogTitle>
        </DialogHeader>

        {success ? (
          <div className="space-y-4 py-2">
            <p className="text-sm text-green-700 bg-green-50 border border-green-200 rounded-lg px-4 py-3">
              Password reset successfully. You can now sign in with your new password.
            </p>
            <button
              onClick={handleClose}
              className="w-full py-3 rounded-lg text-gray-900"
              style={{ backgroundColor: '#FFE000' }}
            >
              Back to Sign In
            </button>
          </div>
        ) : step === 'email' ? (
          <form onSubmit={handleSendOTP} className="space-y-4 py-2">
            <p className="text-sm text-gray-500">
              Enter your registered email address and we'll send you a one-time password.
            </p>
            <div>
              <label className="block text-sm text-gray-700 mb-1">Email Address</label>
              <input
                type="email"
                placeholder="your@email.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:outline-none"
                autoComplete="email"
              />
            </div>
            {error && (
              <div className="px-4 py-3 rounded-lg bg-red-50 border border-red-200 text-sm text-red-700">
                {error}
              </div>
            )}
            <button
              type="submit"
              disabled={loading}
              className="w-full py-3 rounded-lg text-gray-900 disabled:opacity-60"
              style={{ backgroundColor: '#FFE000' }}
            >
              {loading ? 'Sending…' : 'Send OTP'}
            </button>
          </form>
        ) : (
          <form onSubmit={handleReset} className="space-y-4 py-2">
            <p className="text-sm text-gray-500">
              Enter the 6-digit OTP sent to <strong>{email}</strong> and your new password.
            </p>
            <div>
              <label className="block text-sm text-gray-700 mb-2">One-Time Password</label>
              <InputOTP maxLength={6} value={otp} onChange={setOtp}>
                <InputOTPGroup>
                  <InputOTPSlot index={0} />
                  <InputOTPSlot index={1} />
                  <InputOTPSlot index={2} />
                  <InputOTPSlot index={3} />
                  <InputOTPSlot index={4} />
                  <InputOTPSlot index={5} />
                </InputOTPGroup>
              </InputOTP>
            </div>
            <div>
              <label className="block text-sm text-gray-700 mb-1">New Password</label>
              <input
                type="password"
                placeholder="Min 8 characters"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:outline-none"
                autoComplete="new-password"
              />
            </div>
            {error && (
              <div className="px-4 py-3 rounded-lg bg-red-50 border border-red-200 text-sm text-red-700">
                {error}
              </div>
            )}
            <button
              type="submit"
              disabled={loading}
              className="w-full py-3 rounded-lg text-gray-900 disabled:opacity-60"
              style={{ backgroundColor: '#FFE000' }}
            >
              {loading ? 'Resetting…' : 'Reset Password'}
            </button>
            <button
              type="button"
              onClick={() => { setStep('email'); setError(''); setOtp(''); }}
              className="w-full text-sm text-gray-500 hover:text-gray-700"
            >
              ← Back
            </button>
          </form>
        )}
      </DialogContent>
    </Dialog>
  );
}
```

- [ ] **Step 3: Add forgot password link to LoginPage.tsx**

In `Bpclssoportal-main/src/app/LoginPage.tsx`:

Add imports at the top:
```tsx
import { useState } from 'react';  // already present
import { ForgotPasswordModal } from './components/ForgotPasswordModal';
```

Add state inside the component (before `handleSubmit`):
```tsx
const [showForgot, setShowForgot] = useState(false);
```

Add the link + modal below the `<button type="submit">` element:
```tsx
<button
  type="button"
  onClick={() => setShowForgot(true)}
  className="w-full text-sm text-gray-500 hover:text-gray-700 text-center mt-1"
>
  Forgot password?
</button>

<ForgotPasswordModal open={showForgot} onClose={() => setShowForgot(false)} />
```

- [ ] **Step 4: Start dev server and manually verify the flow**

```bash
cd Bpclssoportal-main && npm run dev
```

Open http://localhost:5173. Click "Forgot password?". Verify:
- Modal opens
- Step 1: entering email and clicking "Send OTP" advances to Step 2
- Step 2: OTP input shows 6 slots, password field is present
- "← Back" returns to Step 1
- TypeScript compiles with no errors

- [ ] **Step 5: Commit**

```bash
git add Bpclssoportal-main/src/api/client.ts \
        Bpclssoportal-main/src/app/components/ForgotPasswordModal.tsx \
        Bpclssoportal-main/src/app/LoginPage.tsx
git commit -m "feat: forgot password modal with OTP flow"
```

---

## Done

The forgot password feature is complete. To test end-to-end:
1. Ensure `.env` has valid SMTP credentials
2. Start the API: `cd bpcl-portal-api && go run ./cmd/api`
3. Open the frontend, click "Forgot password?", enter a real registered email
4. Check inbox for OTP, enter it with a new password
