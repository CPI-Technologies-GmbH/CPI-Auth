import { test, expect } from '@playwright/test';
import { goToAdmin, screenshot, waitForLoad, assertNoErrors, apiCall, API_URL, ADMIN_EMAIL, ADMIN_PASSWORD } from './helpers';

let accessToken: string;

test.describe.serial('Admin Settings', () => {
  test.beforeAll(async () => {
    const res = await fetch(`${API_URL}/admin/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email: ADMIN_EMAIL, password: ADMIN_PASSWORD }),
    });
    accessToken = (await res.json()).access_token;
  });

  test('should display settings page with all tabs', async ({ page }) => {
    await goToAdmin(page, '/settings');
    await waitForLoad(page);
    await assertNoErrors(page);

    // Verify page heading
    await expect(page.getByRole('heading', { name: 'Settings' })).toBeVisible();

    // Verify tabs exist (using button text since TabsTrigger renders as button)
    await expect(page.locator('button', { hasText: 'Security' }).first()).toBeVisible();
    await expect(page.locator('button', { hasText: 'MFA' })).toBeVisible();
    await expect(page.locator('button', { hasText: 'Email' }).first()).toBeVisible();
    await expect(page.locator('button', { hasText: 'Social Providers' })).toBeVisible();

    // Default tab is Security, verify Password Policy card is shown
    await expect(page.getByText('Password Policy')).toBeVisible();

    await screenshot(page, 'settings-01-page-with-tabs');
  });

  test('should fetch settings via API', async ({ page }) => {
    const settings = await apiCall(accessToken, 'GET', '/admin/settings');
    expect(settings).toBeTruthy();
    expect(settings.security).toBeTruthy();
    expect(settings.mfa).toBeTruthy();
    expect(settings.security.password_min_length).toBeGreaterThanOrEqual(1);
    expect(settings.mfa).toHaveProperty('enabled');

    // Navigate to settings page and verify no errors
    await goToAdmin(page, '/settings');
    await waitForLoad(page);
    await assertNoErrors(page);

    await screenshot(page, 'settings-02-api-fetch');
  });

  test('should navigate tabs and verify content', async ({ page }) => {
    await goToAdmin(page, '/settings');
    await waitForLoad(page);
    await assertNoErrors(page);

    // Click MFA tab and verify content
    await page.locator('button', { hasText: 'MFA' }).click();
    await waitForLoad(page);
    await expect(page.getByText('Multi-Factor Authentication')).toBeVisible();
    await assertNoErrors(page);

    // Click Email tab and verify content
    await page.locator('button', { hasText: 'Email' }).first().click();
    await waitForLoad(page);
    await expect(page.getByText('SMTP Configuration')).toBeVisible();
    await assertNoErrors(page);

    await screenshot(page, 'settings-03-tab-navigation');
  });
});
