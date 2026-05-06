import { test, expect } from '@playwright/test'
import { BASE_URL } from '../fixtures/constants'

const tryLogin = (id: string) => fetch(`${BASE_URL}/api/v1/auth/login`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ employee_id: id, password: 'test' })
})

test('SQL injection in employee_id → 401 not 500', async () => {
  const r = await tryLogin("' OR '1'='1")
  expect(r.status).toBe(401)
  expect(r.status).not.toBe(500)
})

test('SQL DROP TABLE attempt → 401 not 500', async () => {
  const r = await tryLogin("'; DROP TABLE users;--")
  expect([401, 422]).toContain(r.status)
  expect(r.status).not.toBe(500)
})

test('XSS in page inputs does not fire alert', async ({ page }) => {
  let alertFired = false
  page.on('dialog', () => { alertFired = true })
  await page.goto('/')
  await page.waitForLoadState('networkidle')
  await page.locator('input').first().fill("<script>alert('xss')</script>")
  await page.keyboard.press('Tab')
  await page.waitForTimeout(500)
  expect(alertFired).toBe(false)
})

test('Invalid JWT → 401', async () => {
  const r = await fetch(`${BASE_URL}/api/v1/auth/me`, {
    headers: { Authorization: 'Bearer fakejwt.fake.fake' }
  })
  expect(r.status).toBe(401)
})

test('No auth header → 401', async () => {
  expect((await fetch(`${BASE_URL}/api/v1/outlets/112847`)).status).toBe(401)
})

test('Response never contains password_hash', async () => {
  const res = await fetch(`${BASE_URL}/api/v1/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ employee_id: 'EMP10001', password: 'Bpcl@2026' })
  })
  const text = await res.text()
  expect(text).not.toContain('password_hash')
})

test('Security headers on /health', async () => {
  const res = await fetch(`${BASE_URL}/health`)
  expect(res.headers.get('x-content-type-options')).toBe('nosniff')
  expect(res.headers.get('x-frame-options')).toBe('DENY')
})