import { test, expect } from '@playwright/test'
import { getToken } from '../fixtures/auth'
import { BASE_URL, PERIOD } from '../fixtures/constants'

test('TM gets territory outlets with summary', async () => {
  const token = await getToken('tm')
  const res = await fetch(
    `${BASE_URL}/api/v1/territory/outlets?period=${PERIOD}`,
    { headers: { Authorization: `Bearer ${token}` } }
  )
  expect(res.status).toBe(200)
  const d = await res.json()
  expect(d.summary.total_outlets).toBeGreaterThan(0)
  expect(d.summary.avg_achievement_pct).toBeGreaterThanOrEqual(0)
  expect(Array.isArray(d.outlets)).toBe(true)
})

test('RO cannot access territory endpoint', async () => {
  const token = await getToken('ro')
  const res = await fetch(
    `${BASE_URL}/api/v1/territory/outlets?period=${PERIOD}`,
    { headers: { Authorization: `Bearer ${token}` } }
  )
  expect(res.status).toBe(403)
})

test('Admin can access all territory data', async () => {
  const token = await getToken('admin')
  const res = await fetch(
    `${BASE_URL}/api/v1/territory/outlets?period=${PERIOD}`,
    { headers: { Authorization: `Bearer ${token}` } }
  )
  expect(res.status).toBe(200)
})