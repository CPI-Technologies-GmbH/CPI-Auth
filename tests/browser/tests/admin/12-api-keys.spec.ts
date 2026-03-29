import { test, expect } from '@playwright/test';
import { goToAdmin, screenshot, waitForLoad, assertNoErrors, apiCall, dbQuery, API_URL, ADMIN_EMAIL, ADMIN_PASSWORD } from './helpers';

let accessToken: string;

test.describe.serial('Admin API Keys', () => {
  test.beforeAll(async () => {
    const res = await fetch(`${API_URL}/admin/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email: ADMIN_EMAIL, password: ADMIN_PASSWORD }),
    });
    accessToken = (await res.json()).access_token;
  });

  test('should display API keys page with empty state or table', async ({ page }) => {
    await goToAdmin(page, '/api-keys');
    await waitForLoad(page);
    await assertNoErrors(page);

    await expect(page.getByRole('heading', { name: 'API Keys', exact: true })).toBeVisible();
    await expect(page.getByRole('button', { name: /create api key/i }).first()).toBeVisible();

    await screenshot(page, 'apikeys-01-page-display');
  });

  test('should create API key via API and verify in UI', async ({ page }) => {
    const keyName = `E2E Test Key ${Date.now()}`;
    const key = await apiCall(accessToken, 'POST', '/admin/api-keys', {
      name: keyName,
      scopes: ['read:users', 'write:users'],
      rate_limit: 1000,
    });
    expect(key?.id).toBeTruthy();

    // Verify in DB
    const rows = await dbQuery('SELECT name FROM api_keys WHERE id = $1', [key.id]);
    expect(rows.length).toBe(1);
    expect(rows[0].name).toBe(keyName);

    // Navigate and verify key name appears in the UI
    await goToAdmin(page, '/api-keys');
    await waitForLoad(page);
    await assertNoErrors(page);

    await expect(page.getByText(keyName)).toBeVisible();

    // Cleanup
    await apiCall(accessToken, 'DELETE', `/admin/api-keys/${key.id}`);

    await screenshot(page, 'apikeys-02-key-in-ui');
  });

  test('should create and revoke API key', async ({ page }) => {
    const keyName = `Revoke Test Key ${Date.now()}`;
    const key = await apiCall(accessToken, 'POST', '/admin/api-keys', {
      name: keyName,
      scopes: ['read:users'],
      rate_limit: 500,
    });
    expect(key?.id).toBeTruthy();

    // Verify exists in DB
    const beforeRows = await dbQuery('SELECT id FROM api_keys WHERE id = $1', [key.id]);
    expect(beforeRows.length).toBe(1);

    // Delete (revoke) via API
    await apiCall(accessToken, 'DELETE', `/admin/api-keys/${key.id}`);

    // Verify removed from DB
    const afterRows = await dbQuery('SELECT id FROM api_keys WHERE id = $1', [key.id]);
    expect(afterRows.length).toBe(0);

    // Navigate and verify the revoked key is NOT visible
    await goToAdmin(page, '/api-keys');
    await waitForLoad(page);
    await assertNoErrors(page);

    await expect(page.getByText(keyName)).not.toBeVisible();

    await screenshot(page, 'apikeys-03-after-revoke');
  });
});
