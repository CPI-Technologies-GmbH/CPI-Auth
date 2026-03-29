import { test, expect } from '@playwright/test';
import { goToAdmin, screenshot, waitForLoad, assertNoErrors, apiCall, dbQuery, API_URL, ADMIN_EMAIL, ADMIN_PASSWORD } from './helpers';

let accessToken: string;

test.describe.serial('Per-App Token Lifetimes', () => {
  let testAppId: string;

  test.beforeAll(async () => {
    const res = await fetch(`${API_URL}/admin/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email: ADMIN_EMAIL, password: ADMIN_PASSWORD }),
    });
    accessToken = (await res.json()).access_token;

    const app = await apiCall(accessToken, 'POST', '/admin/applications', {
      name: `E2E TTL App ${Date.now()}`,
      type: 'spa',
      redirect_uris: ['http://localhost:3000/callback'],
      allowed_origins: ['http://localhost:3000'],
    });
    testAppId = app.id;
  });

  test.afterAll(async () => {
    if (testAppId) await apiCall(accessToken, 'DELETE', `/admin/applications/${testAppId}`).catch(() => {});
  });

  test('should create app with default TTLs', async () => {
    const app = await apiCall(accessToken, 'GET', `/admin/applications/${testAppId}`);
    expect(app.access_token_ttl).toBe(3600);
    expect(app.refresh_token_ttl).toBe(2592000);
    expect(app.id_token_ttl).toBe(3600);
  });

  test('should update app with custom TTLs', async () => {
    const updated = await apiCall(accessToken, 'PATCH', `/admin/applications/${testAppId}`, {
      access_token_ttl: 3600,
      refresh_token_ttl: 86400,
      id_token_ttl: 7200,
    });
    expect(updated.access_token_ttl).toBe(3600);
    expect(updated.refresh_token_ttl).toBe(86400);
    expect(updated.id_token_ttl).toBe(7200);
  });

  test('should verify TTLs in database', async () => {
    const rows = await dbQuery(
      'SELECT access_token_ttl, refresh_token_ttl, id_token_ttl FROM applications WHERE id = $1',
      [testAppId]
    );
    expect(rows.length).toBe(1);
    expect(rows[0].access_token_ttl).toBe(3600);
    expect(rows[0].refresh_token_ttl).toBe(86400);
    expect(rows[0].id_token_ttl).toBe(7200);
  });

  test('should reset TTLs to defaults', async () => {
    const updated = await apiCall(accessToken, 'PATCH', `/admin/applications/${testAppId}`, {
      access_token_ttl: 3600,
      refresh_token_ttl: 2592000,
      id_token_ttl: 3600,
    });
    expect(updated.access_token_ttl).toBe(3600);
    expect(updated.refresh_token_ttl).toBe(2592000);
    expect(updated.id_token_ttl).toBe(3600);
  });

  test('should show TTL settings on application detail page', async ({ page }) => {
    await apiCall(accessToken, 'PATCH', `/admin/applications/${testAppId}`, {
      access_token_ttl: 1800,
    });

    await goToAdmin(page, `/applications/${testAppId}`);
    await waitForLoad(page);
    await assertNoErrors(page);

    // Navigate to Settings tab
    await page.getByRole('button', { name: 'Settings', exact: true }).click();
    await waitForLoad(page);

    await screenshot(page, 'token-ttl-01-settings-tab');

    // The page should display the TTL fields
    await expect(page.getByRole('heading', { name: 'Token Lifetimes' })).toBeVisible({ timeout: 5000 });
    await screenshot(page, 'token-ttl-02-ttl-fields');
  });
});
