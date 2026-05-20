# ETL Design — Google Sheets Cron Ingest + Health Monitor

## Architecture

```
Google Sheet (MAY_DATA)
       │
       │  googleapis / Sheets API v4  (no Excel download — live API)
       ▼
  src/etl/ingest.ts   ←── Zod validation (see db-schema.md)
       │
       │  Knex upsert (cc_code + date = unique key)
       ▼
   PostgreSQL
       │
       │  POST on completion (success or failure)
       ▼
  /api/health/cron-ping   ←── cron monitor endpoint
```

**Do not download Excel files.** Use Google Sheets API directly — it always returns current data and avoids file-format bugs.

---

## Cron Schedule

```bash
# crontab -e
# Run at 11:59 PM IST daily (18:29 UTC)
29 18 * * * /usr/local/bin/node /app/src/etl/run.ts >> /var/log/crystal-etl.log 2>&1
```

Why 11:59 PM? Dealers update data throughout the day. End-of-day capture gives the fullest MTD picture.

---

## ETL Runner: `src/etl/run.ts`

```typescript
import { ingestDailyData } from './ingest';
import { pingHealthCheck }  from './healthcheck';

async function run() {
  const startTime = Date.now();
  let status: 'ok' | 'error' = 'error';
  let detail = '';

  try {
    const result = await ingestDailyData();
    status = 'ok';
    detail = `Ingested ${result.rowsUpserted} rows for ${result.date}`;
    console.log(`[ETL] ${detail}`);
  } catch (err: any) {
    detail = err.message ?? 'Unknown ETL error';
    console.error(`[ETL] FAILED: ${detail}`);
  } finally {
    const duration = Date.now() - startTime;
    await pingHealthCheck({ status, detail, durationMs: duration });
    process.exit(status === 'ok' ? 0 : 1);
  }
}

run();
```

---

## Health Check Ping: `src/etl/healthcheck.ts`

```typescript
import https from 'https';

interface PingPayload {
  status:     'ok' | 'error';
  detail:     string;
  durationMs: number;
}

export async function pingHealthCheck(payload: PingPayload): Promise<void> {
  const endpoint = process.env.HEALTH_PING_URL; // e.g. https://yourapp.com/api/health/cron-ping

  if (!endpoint) {
    console.warn('[ETL] HEALTH_PING_URL not set — skipping ping');
    return;
  }

  const body = JSON.stringify({
    ...payload,
    timestamp: new Date().toISOString(),
    source: 'etl-cron',
  });

  return new Promise((resolve) => {
    const req = https.request(endpoint, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', 'Content-Length': Buffer.byteLength(body) },
      timeout: 5000,
    }, (res) => {
      console.log(`[ETL] Health ping → HTTP ${res.statusCode}`);
      resolve();
    });
    req.on('error', (e) => {
      console.error(`[ETL] Health ping failed: ${e.message}`);
      resolve(); // don't throw — ETL result already logged
    });
    req.write(body);
    req.end();
  });
}
```

---

## Health Check Endpoint: `src/routes/health.ts`

```typescript
import { Router } from 'express';
import db from '../db';

const router = Router();

// GET /api/health — simple uptime check (no auth required)
router.get('/', (req, res) => {
  res.json({ status: 'ok', ts: new Date().toISOString() });
});

// POST /api/health/cron-ping — receives ETL status (internal secret only)
router.post('/cron-ping', async (req, res) => {
  const secret = req.headers['x-cron-secret'];
  if (secret !== process.env.CRON_SECRET) {
    return res.status(403).json({ error: 'Forbidden' });
  }

  const { status, detail, durationMs, timestamp, source } = req.body;

  await db('etl_log').insert({ status, detail, duration_ms: durationMs, ran_at: timestamp, source });

  if (status === 'error') {
    // TODO: send Slack webhook or email alert here
    console.error(`[HEALTH] ETL reported error: ${detail}`);
  }

  res.json({ received: true });
});

// GET /api/health/cron-status — last ETL run info (auth required for frontend)
router.get('/cron-status', async (req, res) => {
  const last = await db('etl_log').orderBy('ran_at', 'desc').first();
  res.json(last ?? { status: 'never_run' });
});

export default router;
```

---

## ETL Log Table

```sql
CREATE TABLE etl_log (
  id          SERIAL PRIMARY KEY,
  ran_at      TIMESTAMPTZ NOT NULL,
  status      TEXT NOT NULL,        -- 'ok' | 'error'
  detail      TEXT,
  duration_ms INTEGER,
  source      TEXT DEFAULT 'etl-cron'
);
```

---

## Bash Healthcheck Test (manual + CI)

```bash
# Test that the cron endpoint is alive and reachable
curl -s -o /dev/null -w "%{http_code}" \
  -X POST https://yourapp.com/api/health/cron-ping \
  -H "Content-Type: application/json" \
  -H "x-cron-secret: ${CRON_SECRET}" \
  -d '{"status":"ok","detail":"manual test","durationMs":0,"timestamp":"2026-05-10T00:00:00Z","source":"manual"}'

# Expected: 200
# If not 200 — the endpoint is down before the cron even runs
```

Add this curl to your CI pipeline as a smoke test on every deploy.

---

## Env Vars Required

```bash
GOOGLE_SHEETS_ID=<your-sheet-id>
GOOGLE_SERVICE_ACCOUNT_JSON=<base64-encoded service account JSON>
HEALTH_PING_URL=https://yourapp.com/api/health/cron-ping
CRON_SECRET=<random 32-char string>
DATABASE_URL=postgresql://user:pass@host:5432/crystal
```

---

## Failure Modes + Responses

| Failure | Detection | Response |
|---|---|---|
| Sheet API down | `ingest.ts` throws | ETL exits 1, pings `status: error` |
| Zod validation fails | Schema parse throws | Log offending rows, skip bad rows, continue, ping warning |
| DB upsert fails | Knex throws | ETL exits 1, pings `status: error` |
| Cron doesn't run | No ping received in 25h | Alert (set external monitor — UptimeRobot free tier) |
| Dealer name mismatch | cc_code FK violation | Reject row, log: "Unknown cc_code: XXXX" |

**The only acceptable silent failure is if `HEALTH_PING_URL` is not set.** Everything else must log loudly.
