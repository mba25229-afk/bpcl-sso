import { test, expect } from '@playwright/test';
import { loginAs } from '../fixtures/auth';

test('month filter on Crystal dashboard changes displayed data', async ({ page }) => {
  await loginAs(page, 'tm');

  // Navigate to Crystal tab
  await page.locator('button:has-text("Crystal"), a:has-text("Crystal")').click();
  await expect(page.locator('[data-testid="crystal-dashboard-table"]')).toBeVisible({ timeout: 10000 });

  // Change month to previous month
  const monthInput = page.locator('input[type="month"]');
  const prev = (() => {
    const d = new Date();
    d.setMonth(d.getMonth() - 1);
    return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`;
  })();

  await monthInput.fill(prev);
  // Trigger change event
  await monthInput.dispatchEvent('change');

  // The as_of text should reflect the previous month
  await expect(page.locator(`text=${prev}`)).toBeVisible({ timeout: 8000 });
});
