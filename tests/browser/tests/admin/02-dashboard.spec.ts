import { test, expect } from '@playwright/test';
import { goToAdmin, screenshot, waitForLoad, assertNoErrors } from './helpers';

test.describe.serial('Admin Dashboard', () => {
  test('should display dashboard with metrics', async ({ page }) => {
    await goToAdmin(page, '/');
    await waitForLoad(page);
    await assertNoErrors(page);
    await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible();
    await expect(page.getByText('Active Users')).toBeVisible();
    await expect(page.getByText('Login Success Rate')).toBeVisible();
    await expect(page.getByText('Total Sessions')).toBeVisible();
    await screenshot(page, 'dashboard-01-overview');
  });

  test('should show login chart with period toggle', async ({ page }) => {
    await goToAdmin(page, '/');
    await waitForLoad(page);
    await assertNoErrors(page);
    await expect(page.getByText('Logins Over Time')).toBeVisible();
    const btn30d = page.getByRole('button', { name: '30D' });
    await expect(btn30d).toBeVisible();
    await btn30d.click();
    await page.waitForTimeout(500);
    await screenshot(page, 'dashboard-02-chart-30d');
  });

  test('should show recent events and error rate', async ({ page }) => {
    await goToAdmin(page, '/');
    await waitForLoad(page);
    await assertNoErrors(page);
    await expect(page.getByText('Recent Events')).toBeVisible();
    await expect(page.getByRole('heading', { name: 'Error Rate' })).toBeVisible();
    await screenshot(page, 'dashboard-03-events');
  });
});
