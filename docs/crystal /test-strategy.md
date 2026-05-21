# Test Strategy — Crystal

## Three Layers. Three Tools. Zero Overlap.

```
tests/
├── backend/        # Vitest — unit + integration (Node, no browser)
├── frontend/       # Vitest + React Testing Library (DOM, no network)
└── e2e/            # Playwright — critical UI flows only (real browser + server)
```

**Rule:** If you're testing data correctness — it's backend.
**Rule:** If you're testing a component renders correctly — it's frontend.
**Rule:** If you're testing a user flow end-to-end — it's Playwright. Max 7 flows.

---

## Backend Tests (`tests/backend/`)

**Tool:** Vitest + Supertest
**Runs:** `npm run test:backend`
**DB:** Test database (`crystal_test`) — reset before each suite

### What to test

#### 1. Zod schema validation (`tests/backend/etl/schema.test.ts`)
```typescript
// This is where your 41 data points get validated — not in Playwright
import { describe, it, expect } from 'vitest';
import { DailyActualRow, TargetRow, excelSerialToDate } from '../../../src/etl/schema';

describe('DailyActualRow schema', () => {
  it('accepts valid row', () => {
    expect(() => DailyActualRow.parse({
      cc_code: 112385, date: '2026-05-01',
      ufill: 52, qoc: 0, speed_kl: 4.09, ms_kl: 85.1, hsd_kl: 11.85
    })).not.toThrow();
  });

  it('accepts null values (empty cell = not yet entered)', () => {
    expect(() => DailyActualRow.parse({
      cc_code: 112385, date: '2026-05-01',
      ufill: null, qoc: null, speed_kl: null, ms_kl: 85.1, hsd_kl: null
    })).not.toThrow();
  });

  it('rejects negative values', () => {
    expect(() => DailyActualRow.parse({
      cc_code: 112385, date: '2026-05-01',
      ufill: -1, qoc: 0, speed_kl: null, ms_kl: 85.1, hsd_kl: null
    })).toThrow();
  });

  it('rejects unknown cc_code format', () => {
    expect(() => DailyActualRow.parse({
      cc_code: -1, date: '2026-05-01',
      ufill: 52, qoc: 0, speed_kl: null, ms_kl: 85.1, hsd_kl: null
    })).toThrow();
  });
});

describe('excelSerialToDate', () => {
  it('converts 46143 → 2026-05-01', () => {
    expect(excelSerialToDate(46143)).toBe('2026-05-01');
  });
  it('converts 46173 → 2026-05-31', () => {
    expect(excelSerialToDate(46173)).toBe('2026-05-31');
  });
});
```

#### 2. All 41 dealers present in DB (`tests/backend/db/dealers.test.ts`)
```typescript
import { describe, it, expect, beforeAll } from 'vitest';
import db from '../../../src/db';

const EXPECTED_CC_CODES = [
  112385, 112386, 112390, 151186, 198868, 200922, 147286,
  112402, 144329, 168006, 112407, 148413, 244458, 112414,
  167171, 112415, 112417, 112418, 112419, 112420, 198751,
  112422, 112424, 112431, 112438, 112439, 112440, 112442,
  112446, 127090, 163708, 112448, 136152, 112450, 128472,
  112456, 112458, 112461, 131712, 112465
];

describe('dealers table', () => {
  it('contains all 40 known dealers', async () => {
    const rows = await db('dealers').select('cc_code');
    const codes = rows.map(r => r.cc_code);
    EXPECTED_CC_CODES.forEach(cc => {
      expect(codes).toContain(cc);
    });
  });

  it('cc_code is always a positive integer', async () => {
    const rows = await db('dealers').select('cc_code');
    rows.forEach(r => expect(r.cc_code).toBeGreaterThan(0));
  });
});
```

#### 3. ETL upsert idempotency (`tests/backend/etl/upsert.test.ts`)
```typescript
// Running ETL twice must not double-count data
it('upsert is idempotent', async () => {
  const row = { cc_code: 112385, date: '2026-05-01', ufill: 52, qoc: 0, speed_kl: 4.09, ms_kl: 85.1, hsd_kl: 11.85 };
  await upsertDailyActual(row);
  await upsertDailyActual(row); // second run — same data
  const count = await db('daily_actuals').where({ cc_code: 112385, date: '2026-05-01' }).count();
  expect(Number(count[0].count)).toBe(1); // not 2
});
```

#### 4. Auth routes (`tests/backend/routes/auth.test.ts`)
```typescript
// Test token generation, expiry, refresh, and forgot-password flow
it('POST /api/auth/refresh returns new access token', async () => { ... });
it('GET /api/health returns 200 without auth', async () => { ... });
it('GET /api/dealers returns 401 without token', async () => { ... });
it('POST /api/health/cron-ping returns 403 without cron secret', async () => { ... });
```

#### 5. MTD computation (`tests/backend/services/mtd.test.ts`)
```typescript
// MTD must be computed, never stored
it('MTD sums daily_actuals correctly for partial month', async () => {
  // seed 5 days of data for cc_code 112385
  // call getMTD(112385, '2026-05-01', '2026-05-05')
  // verify sum matches manual addition
});
```

---

## Frontend Tests (`tests/frontend/`)

**Tool:** Vitest + React Testing Library
**Runs:** `npm run test:frontend`
**Mock:** All API calls mocked via `msw` — no real network

### What to test

```typescript
// tests/frontend/components/DealerCard.test.tsx
it('shows red badge when achievement < 80% of target', () => { ... });
it('shows green badge when achievement >= 100% of target', () => { ... });
it('displays "--" for null actuals, not 0', () => { ... });

// tests/frontend/hooks/useInactivityLogout.test.ts
it('clears tokens after 20 minutes of inactivity', () => {
  vi.useFakeTimers();
  // simulate no activity for 20 min
  vi.advanceTimersByTime(20 * 60 * 1000);
  expect(localStorage.getItem('accessToken')).toBeNull();
});
it('resets timer on user interaction', () => { ... });

// tests/frontend/pages/Dashboard.test.tsx
it('renders 40 dealer rows', () => { ... });
it('shows session warning banner at 19 minutes', () => { ... });
```

---

## Playwright E2E (`tests/e2e/`)

**Tool:** Playwright MCP
**Runs:** `npm run test:e2e`
**Scope:** 7 flows max. These are USER JOURNEYS, not data validations.

### The 7 flows

```
1. login.spec.ts           — valid login → dashboard visible
2. inactivity.spec.ts      — idle 20 min → redirect to /login?reason=inactivity
3. forgot-password.spec.ts — submit email → OTP → new password → login succeeds
4. dashboard-filter.spec.ts — date filter changes displayed data
5. dealer-drilldown.spec.ts — click dealer → report card loads with correct RO name
6. cron-status.spec.ts     — admin sees last ETL run status in UI
7. logout.spec.ts          — manual logout clears session, back button doesn't restore
```

### What Playwright does NOT test
- Whether UFILL = 52 for ANAND on May 1st ← backend test
- Whether Zod rejects a negative value ← backend test
- Whether a React component renders a red badge ← frontend test

---

## Running All Tests

```bash
npm run test:backend   # vitest run tests/backend
npm run test:frontend  # vitest run tests/frontend
npm run test:e2e       # playwright test tests/e2e
npm run test           # runs backend + frontend (not e2e — too slow for CI default)
```

## CI Pipeline Order
```
1. test:backend  (fast, ~10s)
2. test:frontend (fast, ~8s)
3. Build
4. Deploy to staging
5. test:e2e      (slow, ~90s — runs against staging, not localhost)
6. Deploy to prod (only if all pass)
```
