# Design: Forgot Password + MAY DATA ETL Pipeline
Date: 2026-05-11

---

## Feature 1: Forgot Password with OTP

### Architecture

Three-layer addition: DB migration → Go backend (service + handler + SMTP) → React modal.

### DB Migration (000012)

```sql
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

OTP TTL: 10 minutes. A new request invalidates prior unused OTPs for the same email (delete before insert).

### API Endpoints (public — no auth middleware)

**POST /api/v1/auth/forgot-password**
- Body: `{ "email": "..." }`
- Looks up user by email. If not found, still returns 200 (prevents enumeration).
- Generates 6-digit numeric OTP, bcrypt-hashes it, inserts into `password_reset_otps`.
- Sends email via SMTP with OTP.
- Response: `{ "message": "If that email is registered, an OTP has been sent." }`

**POST /api/v1/auth/reset-password**
- Body: `{ "email": "...", "otp": "123456", "new_password": "..." }`
- Validates: email required, OTP 6 digits, password min 8 chars.
- Fetches latest unused, unexpired OTP for email.
- Compares OTP with bcrypt hash. On mismatch: 401.
- On match: updates `users.password_hash`, marks OTP `used_at = NOW()`.
- Response: 200 `{ "message": "Password reset successful." }`

### SMTP Config (new env vars)

```
BPCL_SMTP_HOST=smtp.gmail.com
BPCL_SMTP_PORT=587
BPCL_SMTP_USER=your@email.com
BPCL_SMTP_PASS=app-password
BPCL_SMTP_FROM=BPCL Insight <your@email.com>
```

### Go Implementation

New files:
- `internal/service/otp.go` — OTPService: GenerateAndSend, Verify
- `internal/service/email.go` — EmailSender interface + SMTPSender impl
- `internal/handler/otp.go` — ForgotPassword, ResetPassword handlers
- `migrations/000012_password_reset_otps.up.sql` / `.down.sql`

Interfaces extended:
- `handler/interfaces.go` adds `OTPServiceI`

Router: two new public routes added before authMw block.

### Frontend

New file: `Bpclssoportal-main/src/app/components/ForgotPasswordModal.tsx`

Two internal steps managed by local state:
1. **Step 1** — email input + "Send OTP" button
2. **Step 2** — OTP input (6 digits) + new password input + "Reset Password" button

`LoginPage.tsx` adds: "Forgot password?" link below the Sign In button that opens the modal.

`api/client.ts` adds: `forgotPassword(email)` and `resetPassword(email, otp, newPassword)`.

---

## Feature 2: MAY DATA ETL Pipeline

### Source

`docs/crystal /MAY DATA.xlsx` — May 2026 data.

### Script

`scripts/ingest_may_data.py`

CLI flags:
- `--dry-run` — parse and log without writing to DB
- `--sheet SHEET_NAME` — run only one sheet

### Execution Order (respects FK constraints)

1. Upsert dealers from TARGETS sheet → `cr_dealers`
2. Ensure competition period 2026-05-01 exists → `cr_competition_periods`
3. Load TARGETS → `cr_monthly_targets`
4. Load UFILL (daily) → `cr_daily_ufill`
5. Load QOC (daily) → `cr_daily_qoc`
6. Load SPEED (daily) → `cr_daily_speed`
7. Load MS (daily) → `cr_daily_ms`
8. Load HSD (daily) → `cr_daily_hsd`
9. Load MAK GE Reading → `cr_mak_ge_readings`
10. Load SANGAM → `cr_sangam_data`
11. Load GOOGLE RATINGS → `cr_google_ratings`
12. Log run to `cr_etl_log`

### Sheet Parsing Rules

**TARGETS**: row[0]=RO Name, row[1]=CC Code. Skip header rows (row 0 = title, row 1 = headers). Month = 2026-05-01. DSW/Nitrogen: 'YES'/'NO' → boolean.

**UFILL / QOC / MS / HSD**: row[0]=RO Name, row[1]=CC Code, cols[2..32]=daily dates (datetime objects from openpyxl), col[33]=MTD (skip). Each non-None cell → one row.

**SPEED**: row[1]=CC Code, cols[3..33]=daily dates. Product column (col[2]) ignored (only "Speed" rows relevant; skip "Speed 100" if present separately).

**MAK GE Reading**: row[1]=CC Code, cols[2..6]=reading dates from header. Skip rows with all-None readings.

**SANGAM**: row[1]=CC Code, col[3]=status, col[4]=remarks. Month = 2026-05-01.

**GOOGLE RATINGS**: row[2] is headers, col[0]=CC Code. Three date snapshots per row at offsets (col 2,3), (col 5,6), (col 8,9). Skip if rating is None.

### Upsert Strategy

All inserts use `INSERT ... ON CONFLICT (cc_code, txn_date) DO UPDATE SET ...` — safe to re-run.

### Error Handling

- Row-level errors logged and skipped (don't abort the entire run).
- Final summary: rows inserted/updated/skipped per sheet.
- `cr_etl_log` entry written regardless of partial failures.
