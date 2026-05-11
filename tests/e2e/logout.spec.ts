import { test, expect } from '@playwright/test';
import { loginAs } from '../fixtures/auth';

test('manual logout clears session and back button does not restore', async ({ page }) => {
  await loginAs(page, 'tm');

  // Click the logout button (Sign Out)
  await page.locator('button:has-text("Sign Out"), button:has-text("Logout"), button:has-text("Log out")').click();
  await page.waitForLoadState('networkidle');

  // Should be on login page
  await expect(page.locator('input').first()).toBeVisible({ timeout: 5000 });

  // Back button should not restore authenticated session
  await page.goBack();
  await page.waitForLoadState('networkidle');

  // Should still be on login page (no valid token in storage)
  await expect(page.locator('input').first()).toBeVisible({ timeout: 5000 });

  // Tokens should be cleared
  const token = await page.evaluate(() => localStorage.getItem('accessToken'));
  expect(token).toBeNull();
});
