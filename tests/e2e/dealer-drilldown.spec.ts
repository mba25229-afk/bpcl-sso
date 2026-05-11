import { test, expect } from '@playwright/test';
import { loginAs } from '../fixtures/auth';

test('click dealer row → report card shows RO name, Target / MTD / Gap columns', async ({ page }) => {
  await loginAs(page, 'tm');

  await page.locator('button:has-text("Crystal"), a:has-text("Crystal")').click();
  const table = page.locator('[data-testid="crystal-dashboard-table"]');
  await expect(table).toBeVisible({ timeout: 10000 });

  // Click the first dealer row
  const firstRow = table.locator('tbody tr').first();
  const dealerName = await firstRow.locator('td').first().locator('p').first().textContent();
  await firstRow.click();

  // Report card modal should show Target / MTD / Gap columns
  await expect(page.locator('text=Target')).toBeVisible({ timeout: 3000 });
  await expect(page.locator('text=MTD')).toBeVisible();
  await expect(page.locator('text=Gap')).toBeVisible();

  // Modal title should include the dealer name
  if (dealerName) {
    await expect(page.locator(`text=${dealerName.trim()}`).first()).toBeVisible();
  }
});
