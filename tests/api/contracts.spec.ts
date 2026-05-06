import { test, expect } from '@playwright/test'
import { getToken } from '../fixtures/auth'
import { TOP_CC, BOTTOM_CC, PERIOD, BASE_URL, COMP_ID, ADHOC_CC } from '../fixtures/constants'

let token: string
test.beforeAll(async () => { token = await getToken('admin') })
const h = () => ({ Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' })

test('POST /auth/login valid → 200 with JWT', async () => {
  const res = await fetch(`${BASE_URL}/api/v1/auth/login`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ employee_id: 'EMP10001', password: 'Bpcl@2026' })
  })
  expect(res.status).toBe(200)
  const d = await res.json()
  expect(d.token.split('.').length).toBe(3)
  expect(d.user.role).toBe('admin')
  expect(JSON.stringify(d)).not.toContain('password_hash')
})

test('POST /auth/login invalid → 401', async () => {
  const res = await fetch(`${BASE_URL}/api/v1/auth/login`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ employee_id: 'EMP10001', password: 'wrong' })
  })
  expect(res.status).toBe(401)
})

test('GET /auth/me → correct role', async () => {
  const res = await fetch(`${BASE_URL}/api/v1/auth/me`, { headers: h() })
  expect(res.status).toBe(200)
  const d = await res.json()
  expect(d.role).toBe('admin')
  expect(JSON.stringify(d)).not.toContain('password_hash')
})

test('GET /auth/me no token → 401', async () => {
  expect((await fetch(`${BASE_URL}/api/v1/auth/me`)).status).toBe(401)
})

test('GET /outlets/:cc → outlet data', async () => {
  const res = await fetch(`${BASE_URL}/api/v1/outlets/${TOP_CC}`, { headers: h() })
  expect(res.status).toBe(200)
  const d = await res.json()
  expect(d.cc_number).toBe(TOP_CC)
  expect(typeof d.name).toBe('string')
  expect(['A+','A','B','C',null]).toContain(d.rank)
})

test('GET /outlets/999999 → 404', async () => {
  const res = await fetch(`${BASE_URL}/api/v1/outlets/999999`, { headers: h() })
  expect(res.status).toBe(404)
  const d = await res.json()
  expect(d).toHaveProperty('error')
})

test('GET /performance → fuel and non_fuel arrays', async () => {
  const res = await fetch(
    `${BASE_URL}/api/v1/outlets/${TOP_CC}/performance?period=${PERIOD}`,
    { headers: h() }
  )
  expect(res.status).toBe(200)
  const d = await res.json()
  expect(Array.isArray(d.fuel)).toBe(true)
  expect(Array.isArray(d.non_fuel)).toBe(true)
  expect(d.fuel.length).toBeGreaterThanOrEqual(3)
  expect(d.non_fuel.length).toBeGreaterThanOrEqual(5)
})

test('GET /performance → kpis are numbers not strings', async () => {
  const res = await fetch(
    `${BASE_URL}/api/v1/outlets/${TOP_CC}/performance?period=${PERIOD}`,
    { headers: h() }
  )
  const d = await res.json()
  expect(typeof d.target_achievement_pct).toBe('number')
  expect(isNaN(d.target_achievement_pct)).toBe(false)
  expect(d.target_achievement_pct).toBeGreaterThan(0)
})

test('GET /performance → MS achieved non-zero', async () => {
  const res = await fetch(
    `${BASE_URL}/api/v1/outlets/${TOP_CC}/performance?period=${PERIOD}`,
    { headers: h() }
  )
  const d = await res.json()
  const ms = d.fuel?.find((r: any) => r.product_code === 'MS')
  expect(ms).toBeDefined()
  expect(ms.achieved).toBeGreaterThan(0)
})

test('GET /performance → returns fuel_totals', async () => {
  const res = await fetch(
    `${BASE_URL}/api/v1/outlets/${TOP_CC}/performance?period=${PERIOD}`,
    { headers: h() }
  )
  const d = await res.json()
  expect(typeof d.fuel_achieved).toBe('number')
  expect(d.fuel_achieved).toBeGreaterThan(0)
})

test('GET /analysis → returns chart arrays', async () => {
  const res = await fetch(
    `${BASE_URL}/api/v1/outlets/${TOP_CC}/analysis?from=2025-10&to=${PERIOD}`,
    { headers: h() }
  )
  expect(res.status).toBe(200)
  const d = await res.json()
  expect(Array.isArray(d.fuel_mix)).toBe(true)
  expect(d.monthly_growth.length).toBeGreaterThan(0)
})

test('GET /trend → returns 6 periods', async () => {
  const res = await fetch(
    `${BASE_URL}/api/v1/outlets/${TOP_CC}/trend?months=6`,
    { headers: h() }
  )
  expect(res.status).toBe(200)
  const d = await res.json()
  expect(d.periods.length).toBe(6)
})

test('GET /targets → fuel.MS exists', async () => {
  const res = await fetch(
    `${BASE_URL}/api/v1/outlets/${TOP_CC}/targets?period=${PERIOD}`,
    { headers: h() }
  )
  expect(res.status).toBe(200)
  const d = await res.json()
  expect(d.fuel).toHaveProperty('MS')
})

test('PUT /targets as RO → 403', async () => {
  const roToken = await getToken('ro')
  const res = await fetch(`${BASE_URL}/api/v1/outlets/${TOP_CC}/targets`, {
    method: 'PUT',
    headers: { Authorization: `Bearer ${roToken}`, 'Content-Type': 'application/json' },
    body: JSON.stringify({ period: PERIOD, fuel: { MS: 3000 } })
  })
  expect(res.status).toBe(403)
})

test('GET /leaderboard → array with rank 1 first', async () => {
  const res = await fetch(
    `${BASE_URL}/api/v1/competition/leaderboard?territory=DELHI-01`,
    { headers: h() }
  )
  expect(res.status).toBe(200)
  const d = await res.json()
  expect(Array.isArray(d)).toBe(true)
  expect(d.length).toBeGreaterThan(0)
  expect(d[0].rank).toBe(1)
})

test('GET /leaderboard → ADHOC outlet excluded', async () => {
  const res = await fetch(
    `${BASE_URL}/api/v1/competition/leaderboard?territory=DELHI-01`,
    { headers: h() }
  )
  const d = await res.json()
  expect(d.find((x: any) => x.cc_number === ADHOC_CC)).toBeUndefined()
})

test('GET /leaderboard → rank 1 has score > 56.37', async () => {
  const res = await fetch(
    `${BASE_URL}/api/v1/competition/leaderboard?territory=DELHI-01`,
    { headers: h() }
  )
  const d = await res.json()
  expect(d[0].total_score).toBeGreaterThan(56.37)
})

test('GET /territory/outlets as TM → 200 with summary', async () => {
  const tmToken = await getToken('tm')
  const res = await fetch(
    `${BASE_URL}/api/v1/territory/outlets?period=${PERIOD}`,
    { headers: { Authorization: `Bearer ${tmToken}` } }
  )
  expect(res.status).toBe(200)
  const d = await res.json()
  expect(d.summary.total_outlets).toBeGreaterThan(0)
  expect(Array.isArray(d.outlets)).toBe(true)
})

test('GET /territory/outlets as RO → 403', async () => {
  const roToken = await getToken('ro')
  const res = await fetch(
    `${BASE_URL}/api/v1/territory/outlets?period=${PERIOD}`,
    { headers: { Authorization: `Bearer ${roToken}` } }
  )
  expect(res.status).toBe(403)
})

test('GET /admin/users as admin → users array no password_hash', async () => {
  const res = await fetch(`${BASE_URL}/api/v1/admin/users`, { headers: h() })
  expect(res.status).toBe(200)
  const d = await res.json()
  expect(d.total).toBeGreaterThan(0)
  expect(Array.isArray(d.users || d.items)).toBe(true)
  const userList = d.users || d.items
  userList.forEach((u: any) => {
    expect(JSON.stringify(u)).not.toContain('password_hash')
  })
})

test('GET /admin/users as RO → 403', async () => {
  const roToken = await getToken('ro')
  const res = await fetch(`${BASE_URL}/api/v1/admin/users`,
    { headers: { Authorization: `Bearer ${roToken}` } })
  expect(res.status).toBe(403)
})

test('GET /market-share-status → dealers_missing_from_source = 0', async () => {
  const res = await fetch(
    `${BASE_URL}/api/v1/competition/${COMP_ID}/market-share-status`,
    { headers: h() }
  )
  expect(res.status).toBe(200)
  const d = await res.json()
  expect(d.dealers_missing_from_source).toBe(0)
})

test('Security headers present on all responses', async () => {
  const res = await fetch(`${BASE_URL}/health`)
  expect(res.headers.get('x-content-type-options')).toBe('nosniff')
  expect(res.headers.get('x-frame-options')).toBe('DENY')
})