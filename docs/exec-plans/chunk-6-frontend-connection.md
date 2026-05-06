# Chunk 6 — Frontend Connection
## Connect React frontend to the live Go API

---

## Context
Backend API is running (Chunk 5 done).
Frontend is at `../Bpclssoportal-main/` — pre-built React/Vite app with mocked data.
This chunk replaces mocked data with real API calls.

Read AGENTS.md and architecture/CONTRACTS.md before starting.
The CONTRACTS.md response shapes are the contract — frontend components expect exactly these shapes.

---

## Task 1 — Discover Frontend API Usage

Before changing anything, read the frontend:
```bash
# Find all existing API calls / mock data
grep -r "fetch\|axios\|mock\|dummy\|hardcoded\|localhost" ../Bpclssoportal-main/src --include="*.js" --include="*.jsx" --include="*.ts" --include="*.tsx"
```

List every place data is fetched or mocked. You will replace each one.

---

## Task 2 — Create API Client

Create `../Bpclssoportal-main/src/api/client.js` (or .ts):
```javascript
const BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

async function request(path, options = {}) {
  const token = localStorage.getItem('bpcl_token');
  const res = await fetch(`${BASE_URL}${path}`, {
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...options.headers,
    },
    ...options,
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw Object.assign(new Error(err.error || 'API error'), { status: res.status, code: err.code });
  }
  return res.status === 204 ? null : res.json();
}

export const api = {
  login: (employeeId, password) =>
    request('/api/v1/auth/login', { method: 'POST', body: JSON.stringify({ employee_id: employeeId, password }) }),
  me: () => request('/api/v1/auth/me'),
  getOutlet: (cc) => request(`/api/v1/outlets/${cc}`),
  getPerformance: (cc, period) => request(`/api/v1/outlets/${cc}/performance?period=${period}`),
  getAnalysis: (cc, from, to) => request(`/api/v1/outlets/${cc}/analysis?from=${from}&to=${to}`),
  getTargets: (cc, period) => request(`/api/v1/outlets/${cc}/targets?period=${period}`),
  setTargets: (cc, body) => request(`/api/v1/outlets/${cc}/targets`, { method: 'PUT', body: JSON.stringify(body) }),
  getLeaderboard: () => request('/api/v1/competition/leaderboard'),
  getDealerScorecard: (cc, competitionId) => request(`/api/v1/competition/dealers/${cc}/scorecard?competition_id=${competitionId}`),
  uploadFile: (formData) => request('/api/v1/uploads', { method: 'POST', body: formData, headers: {} }),
  getUploads: (cc) => request(`/api/v1/uploads${cc ? `?cc_number=${cc}` : ''}`),
};
```

Create `../Bpclssoportal-main/.env.local`:
```
VITE_API_URL=http://localhost:8080
```

---

## Task 3 — Auth Flow

If no login page exists, create one:
- Fields: Employee ID + Password
- On submit: call `api.login()`
- Store token: `localStorage.setItem('bpcl_token', data.token)`
- Store user: `localStorage.setItem('bpcl_user', JSON.stringify(data.user))`
- Redirect to dashboard

Add logout: clear localStorage, redirect to login.

---

## Task 4 — Replace Mock Data in Each Component

For every component that uses hardcoded/mocked data:

**ROHeader / Outlet Info component:**
- Replace mock with: `api.getOutlet(cc)` on mount
- Map: `cc_number, name, location, rank`

**KPIStrip component (4 KPI cards):**
- Replace mock with: `api.getPerformance(cc, period).then(d => d.kpis)`
- Map: `total_revenue_cr, target_achievement_pct, yoy_growth_pct, fuel_vs_nonfuel_ratio`

**PerformanceTable (Fuel + Non-Fuel):**
- Replace mock with: `api.getPerformance(cc, period).then(d => ({fuel: d.fuel, nonFuel: d.non_fuel}))`
- Map column headers exactly: Product, Target, Achieved, Last Year, Volume (KL), Achievement %, YoY %

**Analysis Charts:**
- Replace mock with: `api.getAnalysis(cc, from, to)`
- Map: `fuel_mix → pie chart`, `target_vs_achieved → bar chart`, `monthly_growth → line chart`

**Leaderboard / Competition view:**
- Replace mock with: `api.getLeaderboard()`
- Display: rank, dealer name, total score, top 10 highlighted

---

## Task 5 — Error States and Loading

For every API call, handle:
- Loading state: show skeleton/spinner while fetching
- Error state: show error message (don't show blank screen)
- 401 error: clear token, redirect to login
- 403 error: show "You don't have access to this outlet"
- 404 error: show "Outlet not found"

---

## Task 6 — Period Selector

The performance view needs a period picker (YYYY-MM).
Default: current month.
On change: re-fetch performance data with new period.

Format for API: `2026-03` (YYYY-MM, the API accepts this and converts to DATE internally).

---

## Verify — Full End-to-End Test

```bash
# Terminal 1: backend running
make run

# Terminal 2: frontend
cd ../Bpclssoportal-main
npm install
npm run dev
# Open http://localhost:5173
```

Checklist:
- [ ] Login with EMP10001 / Bpcl@2026 succeeds
- [ ] Dashboard loads outlet info for CC 112847
- [ ] KPI strip shows real numbers (not all zeros)
- [ ] Fuel table shows MS: achieved ~207 KL for March 2026
- [ ] Non-fuel table shows QOC, Lubricants etc with real values
- [ ] Analysis charts render with real trend data
- [ ] Leaderboard shows M.L. SETHI SERVICE STATION at rank 1 with score 56.37
- [ ] Switching to CC 115012 (MAHADEV) shows declining performance
- [ ] RO manager login (EMP10005) can ONLY see their own outlet (112847)
- [ ] Network tab shows calls going to localhost:8080 (not mock data)

---

## Done When
All checklist items pass. No console errors. No hardcoded mock data remaining.

## Project Complete ✓
The full stack is running: PostgreSQL → Go API → React frontend.

## Ongoing Maintenance
Run `graphify .` in the repo root after major changes to keep the knowledge graph current.
