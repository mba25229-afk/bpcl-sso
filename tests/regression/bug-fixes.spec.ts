import { test, expect } from '@playwright/test'
import { loginAs, getToken } from '../fixtures/auth'
import { TOP_CC, TOP_NAME, BASE_URL, COMP_ID, ADHOC_CC } from '../fixtures/constants'

test('BUG-1: market share dealers_missing_from_source = 0', async () => {
  const token = await getToken('admin')
  const res = await fetch(
    `${BASE_URL}/api/v1/competition/${COMP_ID}/market-share-status`,
    { headers: { Authorization: `Bearer ${token}` } }
  )
  const d = await res.json()
  expect(d.dealers_missing_from_source).toBe(0)
})

test('BUG-4: ADHOC dealer excluded from leaderboard', async () => {
  const token = await getToken('admin')
  const res = await fetch(
    `${BASE_URL}/api/v1/competition/leaderboard?territory=DELHI-01`,
    { headers: { Authorization: `Bearer ${token}` } }
  )
  const d = await res.json()
  expect(Array.isArray(d)).toBe(true)
  expect(d.find((x: any) => x.cc_number === ADHOC_CC)).toBeUndefined()
})

test('BUG-5: leaderboard returns array not object', async () => {
  const token = await getToken('admin')
  const res = await fetch(
    `${BASE_URL}/api/v1/competition/leaderboard?territory=DELHI-01`,
    { headers: { Authorization: `Bearer ${token}` } }
  )
  const d = await res.json()
  expect(Array.isArray(d)).toBe(true)
})

test('BUG-CC: CC prefix stripped before API call', async ({ page }) => {
  const failed: string[] = []
  page.on('response', r => {
    if (r.status() === 404 && r.url().includes('CC%20')) failed.push(r.url())
  })
  await loginAs(page, 'admin')
  const input = page.locator('input').first()
  await input.fill('CC 112847')
  await page.locator(
    'button:has-text("Fetch"), button:has-text("Search"), button[type="submit"]'
  ).first().click()
  await page.waitForTimeout(2000)
  expect(failed).toHaveLength(0)
})

test('BUG-ZEROS: performance table non-zero after valid fetch', async ({ page }) => {
  await loginAs(page, 'admin')
  const input = page.locator('input').first()
  await input.fill(TOP_CC)
  await page.locator(
    'button:has-text("Fetch"), button:has-text("Search")'
  ).first().click()
  await page.waitForTimeout(3000)
  const bodyText = await page.locator('body').textContent()
  expect(bodyText).toContain(TOP_NAME)
})

test('BUG-CORS: no CORS errors in console after login', async ({ page }) => {
  const corsErrors: string[] = []
  page.on('console', msg => {
    if (msg.type() === 'error' && msg.text().toLowerCase().includes('cors')) {
      corsErrors.push(msg.text())
    }
  })
  await loginAs(page, 'admin')
  await page.waitForTimeout(2000)
  expect(corsErrors).toHaveLength(0)
})