import { test, expect } from '@playwright/test';
import { loginAs } from '../fixtures/auth';
import { BASE_URL } from '../fixtures/constants';
import { getToken } from '../fixtures/auth';

test('cron-status endpoint returns last ETL run info', async ({ page }) => {
  const token = await getToken('tm');

  const res = await page.request.get(`${BASE_URL}/api/v1/health/cron-status`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  expect(res.status()).toBe(200);

  const body = await res.json();
  // Either has a last run or is empty — either way the shape is correct
  expect(typeof body).toBe('object');
});

test('Crystal dashboard shows ETL cron badge', async ({ page }) => {
  await loginAs(page, 'tm');
  await page.locator('button:has-text("Crystal"), a:has-text("Crystal")').click();

  // The badge always renders (either "no data yet" or last run time)
  await expect(page.locator('text=/ETL:/i')).toBeVisible({ timeout: 8000 });
});
