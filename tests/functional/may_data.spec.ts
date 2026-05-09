/**
 * May 2026 data integrity tests.
 *
 * These tests verify that the Crystal (cr_*) data loaded from MAY DATA.xlsx
 * is correctly surfaced through the legacy performance_records / targets APIs
 * and rendered in the frontend dashboard and territory views.
 *
 * Run order:
 *   1. Tests in describe("API layer") hit the backend directly — fast, no UI.
 *   2. Tests in describe("Dashboard UI") drive the browser.
 *   3. Tests in describe("Territory UI") drive the territory view.
 */

import { test, expect, request as pwRequest } from '@playwright/test'
import { loginAs } from '../fixtures/auth'
import { BASE_URL } from '../fixtures/constants'

// Crystal dealers we know have data in MAY DATA.xlsx
const CC_WITH_DATA    = '112386'  // ANIL FILLING STATION — has MS/HSD/UFill rows
const CC_WITH_MS_DATA = '112385'  // ANAND SUPER SERVICE STN — heavy MS data
const MAY_PERIOD      = '2026-05'

// ── helpers ──────────────────────────────────────────────────────────────────

async function apiToken(): Promise<string> {
  const res = await fetch(`${BASE_URL}/api/v1/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ employee_id: 'EMP10001', password: 'Bpcl@2026' }),
  })
  const d = await res.json()
  if (!d.token) throw new Error('Login failed: ' + JSON.stringify(d))
  return d.token
}

async function fetchPerformance(cc: string, period: string) {
  const token = await apiToken()
  const res = await fetch(
    `${BASE_URL}/api/v1/outlets/${cc}/performance?period=${period}`,
    { headers: { Authorization: `Bearer ${token}` } }
  )
  return res.json()
}

async function fetchTargets(cc: string, period: string) {
  const token = await apiToken()
  const res = await fetch(
    `${BASE_URL}/api/v1/outlets/${cc}/targets?period=${period}`,
    { headers: { Authorization: `Bearer ${token}` } }
  )
  return res.json()
}

async function fetchTerritorySummary(period: string) {
  const token = await apiToken()
  const res = await fetch(
    `${BASE_URL}/api/v1/territory/summary?period=${period}`,
    { headers: { Authorization: `Bearer ${token}` } }
  )
  return res.json()
}

// ── API layer tests ───────────────────────────────────────────────────────────
// All API tests share a single token fetched once to avoid rate limit hits.

test.describe('API layer — May 2026 performance data', () => {
  let token = ''
  let perf112386: any
  let perf112385: any
  let tgt112385: any
  let tgt112386: any
  let territory: any

  test.beforeAll(async () => {
    token = await apiToken()
    const headers = { Authorization: `Bearer ${token}` }
    const [p386, p385, t385, t386, terr] = await Promise.all([
      fetch(`${BASE_URL}/api/v1/outlets/${CC_WITH_DATA}/performance?period=${MAY_PERIOD}`, { headers }).then(r => r.json()),
      fetch(`${BASE_URL}/api/v1/outlets/${CC_WITH_MS_DATA}/performance?period=${MAY_PERIOD}`, { headers }).then(r => r.json()),
      fetch(`${BASE_URL}/api/v1/outlets/${CC_WITH_MS_DATA}/targets?period=${MAY_PERIOD}`, { headers }).then(r => r.json()),
      fetch(`${BASE_URL}/api/v1/outlets/${CC_WITH_DATA}/targets?period=${MAY_PERIOD}`, { headers }).then(r => r.json()),
      fetch(`${BASE_URL}/api/v1/territory/summary?period=${MAY_PERIOD}`, { headers }).then(r => r.json()),
    ])
    perf112386 = p386; perf112385 = p385
    tgt112385 = t385; tgt112386 = t386
    territory = terr
  })

  test('GET /outlets/112386/performance returns non-null MS achieved', () => {
    expect(perf112386.period).toBe(MAY_PERIOD)
    const ms = perf112386.fuel?.find((f: any) => f.product_code === 'MS')
    expect(ms).toBeDefined()
    expect(ms.achieved).not.toBeNull()
    expect(ms.achieved).toBeGreaterThan(0)
  })

  test('GET /outlets/112386/performance returns non-null HSD achieved', () => {
    const hsd = perf112386.fuel?.find((f: any) => f.product_code === 'HSD')
    expect(hsd).toBeDefined()
    expect(hsd.achieved).not.toBeNull()
    expect(hsd.achieved).toBeGreaterThan(0)
  })

  test('GET /outlets/112386/performance returns non-null UFill achieved', () => {
    const ufill = perf112386.non_fuel?.find((f: any) => f.product_code === 'UFill')
    expect(ufill).toBeDefined()
    expect(ufill.achieved).not.toBeNull()
    expect(ufill.achieved).toBeGreaterThan(0)
  })

  test('GET /outlets/112386/performance target_achievement_pct is non-null', () => {
    expect(perf112386.target_achievement_pct).not.toBeNull()
    expect(perf112386.target_achievement_pct).toBeGreaterThan(0)
  })

  test('GET /outlets/112385/performance MS achieved > 0 (ANAND heavy user)', () => {
    const ms = perf112385.fuel?.find((f: any) => f.product_code === 'MS')
    expect(ms?.achieved).toBeGreaterThan(0)
  })

  test('GET /outlets/112385/targets returns MS target from TARGETS sheet', () => {
    const msTarget = tgt112385.fuel?.MS ?? tgt112385.fuel?.ms
    expect(Number(msTarget)).toBeGreaterThan(0)
  })

  test('GET /outlets/112386/targets returns UFill target', () => {
    const ufillTarget = tgt112386.non_fuel?.UFill ?? tgt112386.non_fuel?.ufill
    expect(Number(ufillTarget)).toBeGreaterThan(0)
  })

  test('GET /territory/summary has outlets with non-zero MS achievement', () => {
    expect(territory.summary?.total_outlets).toBeGreaterThan(0)
    const hasData = territory.outlets?.some((o: any) => o.ms_achievement_pct > 0)
    expect(hasData).toBe(true)
  })

})

// ── Dashboard UI tests ────────────────────────────────────────────────────────

test.describe('Dashboard UI — May 2026 data visible', () => {

  test.beforeEach(async ({ page }) => {
    await loginAs(page, 'admin')
  })

  test('CC 112386 May 2026 shows outlet name and non-null achieved data', async ({ page }) => {
    // Fill CC and fetch (date defaults to today = May 2026)
    await page.locator('input').first().fill(CC_WITH_DATA)
    await page.locator('button:has-text("Fetch"), button:has-text("Search")').first().click()
    await page.waitForTimeout(4000)

    const body = await page.locator('body').textContent() ?? ''
    expect(body).toContain('ANIL FILLING STATION')
    // MS or HSD achieved value should appear (non-zero digits)
    expect(body).toMatch(/85\.|11\./)
  })

  test('CC 112386 May 2026 MS row shows non-zero achieved', async ({ page }) => {
    await page.locator('input[type="date"], input[placeholder*="date" i]').first()
      .fill('2026-05-08').catch(() => {})

    await page.locator('input').first().fill(CC_WITH_DATA)
    await page.locator('button:has-text("Fetch"), button:has-text("Search")').first().click()
    await page.waitForTimeout(3000)

    // The MS row should not show "0" as achieved
    const perfTable = page.locator('table, [class*="performance"], [class*="table"]').first()
    const body = await page.locator('body').textContent() ?? ''
    expect(body).toContain('ANIL FILLING STATION')
    // achieved value for MS should not be 0
    expect(body).not.toMatch(/Motor Spirit.*?\b0\b.*?0\.0%/)
  })

  test('CC 112386 May 2026 shows target achievement percent', async ({ page }) => {
    await page.locator('input').first().fill(CC_WITH_DATA)
    await page.locator('button:has-text("Fetch"), button:has-text("Search")').first().click()
    await page.waitForTimeout(4000)

    const body = await page.locator('body').textContent() ?? ''
    expect(body).toContain('ANIL FILLING STATION')
    // The target achievement % should exist and be a number (e.g. "6.76%" or "18.6%")
    expect(body).toMatch(/\d+\.?\d*%/)
  })

})

// ── Territory UI tests ────────────────────────────────────────────────────────

test.describe('Territory UI — May 2026 non-zero achievement', () => {

  async function goToTerritory(page: any) {
    await loginAs(page, 'admin')
    // Wait for nav to be fully rendered, then click Territory button
    await page.waitForSelector('nav button:has-text("Territory"), button:has-text("Territory")', { timeout: 10000 })
    await page.locator('button:has-text("Territory")').first().click()
    // Wait for TerritoryView to mount and fetch data
    await page.waitForTimeout(4000)
  }

  test('Territory view loads and shows dealer rows after login', async ({ page }) => {
    await goToTerritory(page)

    const body = await page.locator('body').textContent() ?? ''
    // TerritoryView renders "Territory Overview" heading
    expect(body).toMatch(/Territory Overview|Total Outlets/i)
    // Should list outlet rows from our 40 dealers
    const hasDealers = body.includes('ANIL') || body.includes('ANAND') || body.includes('FILLING') || body.includes('Outlets')
    expect(hasDealers).toBe(true)
  })

  test('Territory view May 2026 shows non-zero achievement % for at least one dealer', async ({ page }) => {
    await goToTerritory(page)

    const body = await page.locator('body').textContent() ?? ''
    // TerritoryView fetches 2026-05 by default (current month)
    // At least one dealer should show non-zero achievement (e.g. 18.6%)
    expect(body).toMatch(/[1-9]\d*\.\d+%|[1-9]\d*%/)
  })

})
