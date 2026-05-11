import { test, expect } from '@playwright/test';
import { loginAs } from '../fixtures/auth';

test('idle 20 min → redirect to /login?reason=inactivity', async ({ page }) => {
  await loginAs(page, 'tm');

  // Fast-forward the inactivity timer via fake clock injection
  await page.evaluate(() => {
    // Advance the Date and trigger setTimeout callbacks immediately
    const orig = window.setTimeout;
    let id = 0;
    const cbs: [Function, number][] = [];
    (window as any).__fakeTimerCbs = cbs;
    (window as any).setTimeout = (fn: Function, delay: number) => {
      cbs.push([fn, delay]);
      return ++id;
    };
  });

  // Trigger the 20-min logout callback directly
  await page.evaluate(() => {
    const cbs: [Function, number][] = (window as any).__fakeTimerCbs || [];
    // Find the 20-min timeout (1200000 ms)
    const logoutCb = cbs.find(([, d]) => d === 20 * 60 * 1000);
    if (logoutCb) logoutCb[0]();
  });

  await expect(page).toHaveURL(/reason=inactivity/, { timeout: 5000 });
});

test('session warning banner appears before logout', async ({ page }) => {
  await loginAs(page, 'tm');

  // Dispatch the session-warning event directly
  await page.evaluate(() => {
    window.dispatchEvent(new CustomEvent('session-warning'));
  });

  await expect(page.locator('[data-testid="session-warning"]')).toBeVisible({ timeout: 3000 });
});
