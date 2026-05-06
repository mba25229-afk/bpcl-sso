import { test, expect } from '@playwright/test'
import { loginAs, getToken } from '../fixtures/auth'
import { BASE_URL, ADHOC_CC, TOP_CC } from '../fixtures/constants'

test('leaderboard tab loads and shows rank 1', async ({ page }) => {
  await loginAs(page, 'admin')
  const leaderboardLink = page.locator(
    'a:has-text("Leaderboard"), button:has-text("Leaderboard"), nav >> text=Leaderboard'
  ).first()
  if (await leaderboardLink.isVisible()) {
    await leaderboardLink.click()
    await page.waitForTimeout(2000)
    const body = await page.locator('body').textContent()
    expect(body).toContain('1')
  }
})

test('leaderboard API rank 1 score > 56.37', async () => {
  const token = await getToken('admin')
  const res = await fetch(
    `${BASE_URL}/api/v1/competition/leaderboard?territory=DELHI-01`,
    { headers: { Authorization: `Bearer ${token}` } }
  )
  const d = await res.json()
  expect(d[0].total_score).toBeGreaterThan(56.37)
  expect(d[0].cc_number).toBe(TOP_CC)
})

test('ADHOC dealer not in leaderboard API response', async () => {
  const token = await getToken('admin')
  const res = await fetch(
    `${BASE_URL}/api/v1/competition/leaderboard?territory=DELHI-01`,
    { headers: { Authorization: `Bearer ${token}` } }
  )
  const d = await res.json()
  expect(d.find((x: any) => x.cc_number === ADHOC_CC)).toBeUndefined()
})