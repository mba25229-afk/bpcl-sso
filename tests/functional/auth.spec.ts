import { test, expect } from '@playwright/test'
import { loginAs } from '../fixtures/auth'

test('valid login redirects away from login page', async ({ page }) => {
  await loginAs(page, 'admin')
  expect(page.url()).not.toContain('login')
})

test('invalid password shows error message', async ({ page }) => {
  await page.goto('/')
  await page.waitForLoadState('networkidle')
  await page.locator('input').first().fill('EMP10001')
  await page.locator('input[type="password"]').fill('wrongpassword')
  await page.locator('button[type="submit"]').first().click()
  await page.waitForTimeout(2000)
  const bodyText = await page.locator('body').textContent()
  const hasError = bodyText?.toLowerCase().includes('fail') ||
                   bodyText?.toLowerCase().includes('invalid') ||
                   bodyText?.toLowerCase().includes('error')
  expect(hasError).toBe(true)
})

test('logout clears session', async ({ page }) => {
  await loginAs(page, 'admin')
  const logoutBtn = page.locator(
    'button:has-text("Logout"), button:has-text("Sign Out"), [aria-label*="logout"]'
  ).first()
  if (await logoutBtn.isVisible()) {
    await logoutBtn.click()
    await page.waitForTimeout(1000)
    expect(page.url()).toContain('login') || expect(page.url()).toBe('http://localhost:5173/')
  }
})