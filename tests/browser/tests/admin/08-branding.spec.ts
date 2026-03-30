import { test, expect } from '@playwright/test';
import { goToAdmin, screenshot, waitForLoad, assertNoErrors, apiCall, API_URL, ADMIN_EMAIL, ADMIN_PASSWORD } from './helpers';

let accessToken: string;

test.describe.serial('Admin Branding', () => {
  test.beforeAll(async () => {
    const res = await fetch(`${API_URL}/admin/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email: ADMIN_EMAIL, password: ADMIN_PASSWORD }),
    });
    accessToken = (await res.json()).access_token;
  });

  test('should display branding page', async ({ page }) => {
    await goToAdmin(page, '/branding');
    await waitForLoad(page);
    await assertNoErrors(page);
    await screenshot(page, 'branding-01-page');
  });

  test('should fetch branding via API', async ({ page }) => {
    const branding = await apiCall(accessToken, 'GET', '/admin/branding');
    expect(branding).toBeTruthy();

    await goToAdmin(page, '/branding');
    await waitForLoad(page);
    await assertNoErrors(page);
    await screenshot(page, 'branding-02-loaded');
  });

  test('should update branding via API', async ({ page }) => {
    const updated = await apiCall(accessToken, 'PATCH', '/admin/branding', {
      branding: { primary_color: '#6366f1', company_name: 'E2E Test Co' },
    });
    expect(updated).toBeTruthy();

    await goToAdmin(page, '/branding');
    await waitForLoad(page);
    await assertNoErrors(page);
    await screenshot(page, 'branding-03-updated');
  });
});
