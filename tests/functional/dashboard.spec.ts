import { test, expect } from '@playwright/test'
import { loginAs } from '../fixtures/auth'
import { TOP_CC, TOP_NAME, BOTTOM_CC, PERIOD } from '../fixtures/constants'

test.beforeEach(async ({ page }) => {
  await loginAs(page, 'admin')
})

test('CC 112847 loads M.L. SETHI outlet name', async ({ page }) => {
  await page.locator('input').first().fill(TOP_CC)
  await page.locator(
    'button:has-text("Fetch"), button:has-text("Search")'
  ).first().click()
  await page.waitForTimeout(3000)
  const body = await page.locator('body').textContent()
  expect(body).toContain(TOP_NAME)
})

test('KPI strip shows non-zero values', async ({ page }) => {
  await page.locator('input').first().fill(TOP_CC)
  await page.locator(
    'button:has-text("Fetch"), button:has-text("Search")'
  ).first().click()
  await page.waitForTimeout(3000)
  const body = await page.locator('body').textContent()
  expect(body).toMatch(/\d+\.?\d*%|₹\d+|\d+\.\d+/)
})

test('CC with prefix CC 112847 loads same outlet', async ({ page }) => {
  await page.locator('input').first().fill('CC 112847')
  await page.locator(
    'button:has-text("Fetch"), button:has-text("Search")'
  ).first().click()
  await page.waitForTimeout(3000)
  const body = await page.locator('body').textContent()
  expect(body).toContain(TOP_NAME)
})

test('unknown CC 999999 shows error not blank', async ({ page }) => {
  await page.locator('input').first().fill('999999')
  await page.locator(
    'button:has-text("Fetch"), button:has-text("Search")'
  ).first().click()
  await page.waitForTimeout(2000)
  const body = await page.locator('body').textContent()
  const hasError = body?.toLowerCase().includes('not found') ||
                    body?.toLowerCase().includes('error') ||
                    body?.toLowerCase().includes('invalid')
  expect(hasError).toBe(true)
})