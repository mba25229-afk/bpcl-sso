import { test, expect } from '@playwright/test';

// Note: forgot-password OTP flow is Chunk 6 (hardening).
// This spec defines the expected UX — it will fail until the flow is built,
// serving as a failing acceptance test.

test('forgot password link is visible on login page', async ({ page }) => {
  await page.goto('/');
  await page.waitForLoadState('networkidle');
  await expect(
    page.locator('text=/forgot|reset|password/i').first()
  ).toBeVisible({ timeout: 5000 });
});

test('forgot password → OTP input step is reachable', async ({ page }) => {
  await page.goto('/');
  await page.waitForLoadState('networkidle');

  const forgotLink = page.locator('text=/forgot|reset/i').first();
  await forgotLink.click();

  // Should show an email input
  await expect(page.locator('input[type="email"], input[placeholder*="email" i]')).toBeVisible({ timeout: 5000 });
});
