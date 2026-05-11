import { test, expect } from '@playwright/test';
import { loginAs } from '../fixtures/auth';

test('valid login → Crystal Dashboard tab visible', async ({ page }) => {
  await loginAs(page, 'tm');
  await expect(page.locator('button:has-text("Crystal"), a:has-text("Crystal")')).toBeVisible({ timeout: 8000 });
});

test('invalid credentials shows error', async ({ page }) => {
  await page.goto('/');
  await page.waitForLoadState('networkidle');
  const inputs = page.locator('input');
  await inputs.first().fill('BADUSER');
  await inputs.nth(1).fill('wrongpassword');
  await page.locator('button[type="submit"], button:has-text("Sign"), button:has-text("Login")').first().click();
  await expect(page.locator('text=/invalid|error|incorrect/i')).toBeVisible({ timeout: 5000 });
});
