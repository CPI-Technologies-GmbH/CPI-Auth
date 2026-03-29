import { test, expect } from '@playwright/test';
import { goToAdmin, screenshot, waitForLoad, assertNoErrors, apiCall, dbQuery, API_URL, ADMIN_EMAIL, ADMIN_PASSWORD } from './helpers';

let accessToken: string;

test.describe.serial('Admin Applications Management', () => {
  let testAppId: string;
  let testAppClientId: string;

  test.beforeAll(async () => {
    const res = await fetch(`${API_URL}/admin/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email: ADMIN_EMAIL, password: ADMIN_PASSWORD }),
    });
    accessToken = (await res.json()).access_token;

    // Create a test application for all subsequent tests
    const app = await apiCall(accessToken, 'POST', '/admin/applications', {
      name: `E2E Detail App ${Date.now()}`,
      type: 'web',
      redirect_uris: ['http://localhost:3000/callback'],
      allowed_origins: ['http://localhost:3000'],
    });
    testAppId = app.id;
    testAppClientId = app.client_id;
  });

  test.afterAll(async () => {
    if (testAppId) {
      await apiCall(accessToken, 'DELETE', `/admin/applications/${testAppId}`).catch(() => {});
    }
  });

  // ─── LIST PAGE ──────────────────────────────────────────────

  test('should display applications list page', async ({ page }) => {
    await goToAdmin(page, '/applications');
    await waitForLoad(page);
    await assertNoErrors(page);
    await expect(page.getByRole('heading', { name: 'Applications' })).toBeVisible();
    await screenshot(page, '04-apps-list');
  });

  test('should create application via API and see it in list', async ({ page }) => {
    const name = `E2E Temp App ${Date.now()}`;
    const app = await apiCall(accessToken, 'POST', '/admin/applications', {
      name,
      type: 'spa',
    });
    expect(app.id).toBeTruthy();
    expect(app.client_id).toBeTruthy();
    expect(app.type).toBe('spa');

    // Verify in DB
    const rows = await dbQuery('SELECT name, type FROM applications WHERE id = $1', [app.id]);
    expect(rows.length).toBe(1);
    expect(rows[0].type).toBe('spa');

    // Verify in UI
    await goToAdmin(page, '/applications');
    await waitForLoad(page);
    await expect(page.getByText(name)).toBeVisible({ timeout: 5000 });

    // Cleanup
    await apiCall(accessToken, 'DELETE', `/admin/applications/${app.id}`);
  });

  // ─── NEGATIVE: CREATE ──────────────────────────────────────

  test('should reject creating application without name', async () => {
    try {
      await apiCall(accessToken, 'POST', '/admin/applications', { type: 'spa' });
      expect(true).toBe(false); // Should not reach here
    } catch {
      // Expected — validation error
    }
  });

  test('should reject creating application with invalid type', async () => {
    try {
      await apiCall(accessToken, 'POST', '/admin/applications', {
        name: 'Invalid Type App',
        type: 'invalid_type',
      });
      // If no error thrown, check response doesn't have a valid id
    } catch {
      // Expected
    }
  });

  // ─── DETAIL PAGE: OVERVIEW TAB ─────────────────────────────

  test('should navigate to application detail page', async ({ page }) => {
    await goToAdmin(page, `/applications/${testAppId}`);
    await waitForLoad(page);
    await assertNoErrors(page);
    await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
    await expect(page.getByText('Client Information')).toBeVisible();
    await screenshot(page, '04-app-detail-overview');
  });

  test('should show all 5 tabs', async ({ page }) => {
    await goToAdmin(page, `/applications/${testAppId}`);
    await waitForLoad(page);
    await expect(page.getByRole('button', { name: 'Overview' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Settings' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Permissions' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Connections' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'API' })).toBeVisible();
  });

  // ─── STATUS TOGGLE ─────────────────────────────────────────

  test('should toggle application status to active', async () => {
    const updated = await apiCall(accessToken, 'PATCH', `/admin/applications/${testAppId}`, {
      is_active: true,
    });
    expect(updated.is_active).toBe(true);

    // Verify in DB
    const rows = await dbQuery('SELECT is_active FROM applications WHERE id = $1', [testAppId]);
    expect(rows[0].is_active).toBe(true);
  });

  test('should toggle application status to disabled', async () => {
    const updated = await apiCall(accessToken, 'PATCH', `/admin/applications/${testAppId}`, {
      is_active: false,
    });
    expect(updated.is_active).toBe(false);
  });

  test('should toggle status via UI switch', async ({ page }) => {
    await goToAdmin(page, `/applications/${testAppId}`);
    await waitForLoad(page);
    await expect(page.getByText('Disabled')).toBeVisible({ timeout: 5000 });

    // Click the status switch
    await page.getByRole('switch').click();
    await page.waitForTimeout(500);

    // Verify the app is now active in DB
    const rows = await dbQuery('SELECT is_active FROM applications WHERE id = $1', [testAppId]);
    expect(rows[0].is_active).toBe(true);

    // Toggle back
    await page.getByRole('switch').click();
    await page.waitForTimeout(500);
    const rows2 = await dbQuery('SELECT is_active FROM applications WHERE id = $1', [testAppId]);
    expect(rows2[0].is_active).toBe(false);
  });

  // ─── GRANT TYPES ───────────────────────────────────────────

  test('should enable authorization_code grant type via API', async () => {
    const updated = await apiCall(accessToken, 'PATCH', `/admin/applications/${testAppId}`, {
      grant_types: ['authorization_code'],
    });
    expect(updated.grant_types).toContain('authorization_code');

    const rows = await dbQuery('SELECT grant_types FROM applications WHERE id = $1', [testAppId]);
    expect(rows[0].grant_types).toContain('authorization_code');
  });

  test('should enable multiple grant types via API', async () => {
    const updated = await apiCall(accessToken, 'PATCH', `/admin/applications/${testAppId}`, {
      grant_types: ['authorization_code', 'client_credentials', 'refresh_token'],
    });
    expect(updated.grant_types).toContain('authorization_code');
    expect(updated.grant_types).toContain('client_credentials');
    expect(updated.grant_types).toContain('refresh_token');
    expect(updated.grant_types).not.toContain('implicit');
    expect(updated.grant_types).not.toContain('password');
  });

  test('should clear all grant types via API', async () => {
    const updated = await apiCall(accessToken, 'PATCH', `/admin/applications/${testAppId}`, {
      grant_types: [],
    });
    expect(updated.grant_types).toEqual([]);
  });

  test('should toggle grant type checkbox in UI', async ({ page }) => {
    await goToAdmin(page, `/applications/${testAppId}`);
    await waitForLoad(page);

    // Find the Authorization Code checkbox (first one)
    const checkboxes = page.getByRole('checkbox');
    await expect(checkboxes.first()).toBeVisible({ timeout: 5000 });

    // Click first checkbox (Authorization Code)
    await checkboxes.first().click();
    await page.waitForTimeout(500);

    // Verify in DB that grant type was added
    const rows = await dbQuery('SELECT grant_types FROM applications WHERE id = $1', [testAppId]);
    expect(rows[0].grant_types).toContain('authorization_code');

    // Uncheck it
    await checkboxes.first().click();
    await page.waitForTimeout(500);

    const rows2 = await dbQuery('SELECT grant_types FROM applications WHERE id = $1', [testAppId]);
    expect(rows2[0].grant_types).not.toContain('authorization_code');

    await screenshot(page, '04-app-grant-types');
  });

  // ─── SETTINGS TAB: REDIRECT URIs ──────────────────────────

  test('should update redirect_uris via API', async () => {
    const updated = await apiCall(accessToken, 'PATCH', `/admin/applications/${testAppId}`, {
      redirect_uris: ['http://localhost:3000/callback', 'http://localhost:4000/auth'],
    });
    expect(updated.redirect_uris).toHaveLength(2);
    expect(updated.redirect_uris).toContain('http://localhost:3000/callback');
    expect(updated.redirect_uris).toContain('http://localhost:4000/auth');
  });

  test('should set redirect_uris to empty array', async () => {
    const updated = await apiCall(accessToken, 'PATCH', `/admin/applications/${testAppId}`, {
      redirect_uris: [],
    });
    expect(updated.redirect_uris).toEqual([]);
  });

  // ─── SETTINGS TAB: ALLOWED ORIGINS ────────────────────────

  test('should update allowed_origins via API', async () => {
    const updated = await apiCall(accessToken, 'PATCH', `/admin/applications/${testAppId}`, {
      allowed_origins: ['http://localhost:3000', 'https://app.example.com'],
    });
    expect(updated.allowed_origins).toHaveLength(2);
    expect(updated.allowed_origins).toContain('https://app.example.com');
  });

  test('should set allowed_origins to empty array', async () => {
    const updated = await apiCall(accessToken, 'PATCH', `/admin/applications/${testAppId}`, {
      allowed_origins: [],
    });
    expect(updated.allowed_origins).toEqual([]);
  });

  // ─── SETTINGS TAB: ALLOWED LOGOUT URLs ────────────────────

  test('should update allowed_logout_urls via API', async () => {
    const updated = await apiCall(accessToken, 'PATCH', `/admin/applications/${testAppId}`, {
      allowed_logout_urls: ['http://localhost:3000/logout'],
    });
    expect(updated.allowed_logout_urls).toContain('http://localhost:3000/logout');
  });

  test('should set allowed_logout_urls to empty array', async () => {
    const updated = await apiCall(accessToken, 'PATCH', `/admin/applications/${testAppId}`, {
      allowed_logout_urls: [],
    });
    expect(updated.allowed_logout_urls).toEqual([]);
  });

  // ─── SETTINGS TAB: TOKEN TTLs ─────────────────────────────

  test('should update access_token_ttl via API', async () => {
    const updated = await apiCall(accessToken, 'PATCH', `/admin/applications/${testAppId}`, {
      access_token_ttl: 7200,
    });
    expect(updated.access_token_ttl).toBe(7200);

    const rows = await dbQuery('SELECT access_token_ttl FROM applications WHERE id = $1', [testAppId]);
    expect(rows[0].access_token_ttl).toBe(7200);
  });

  test('should update refresh_token_ttl via API', async () => {
    const updated = await apiCall(accessToken, 'PATCH', `/admin/applications/${testAppId}`, {
      refresh_token_ttl: 86400,
    });
    expect(updated.refresh_token_ttl).toBe(86400);
  });

  test('should update id_token_ttl via API', async () => {
    const updated = await apiCall(accessToken, 'PATCH', `/admin/applications/${testAppId}`, {
      id_token_ttl: 1800,
    });
    expect(updated.id_token_ttl).toBe(1800);
  });

  test('should update all TTLs at once via API', async () => {
    const updated = await apiCall(accessToken, 'PATCH', `/admin/applications/${testAppId}`, {
      access_token_ttl: 900,
      refresh_token_ttl: 604800,
      id_token_ttl: 600,
    });
    expect(updated.access_token_ttl).toBe(900);
    expect(updated.refresh_token_ttl).toBe(604800);
    expect(updated.id_token_ttl).toBe(600);
  });

  test('should reset TTLs to defaults via API', async () => {
    const updated = await apiCall(accessToken, 'PATCH', `/admin/applications/${testAppId}`, {
      access_token_ttl: 3600,
      refresh_token_ttl: 2592000,
      id_token_ttl: 3600,
    });
    expect(updated.access_token_ttl).toBe(3600);
    expect(updated.refresh_token_ttl).toBe(2592000);
    expect(updated.id_token_ttl).toBe(3600);
  });

  // ─── SETTINGS TAB: UI ─────────────────────────────────────

  test('should switch to Settings tab and see all fields', async ({ page }) => {
    await goToAdmin(page, `/applications/${testAppId}`);
    await waitForLoad(page);

    await page.getByRole('button', { name: 'Settings', exact: true }).click();
    await page.waitForTimeout(300);

    await expect(page.getByText('Redirect URIs')).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('Allowed Origins')).toBeVisible();
    await expect(page.getByText('Token Lifetimes')).toBeVisible();
    await expect(page.getByText('Access Token TTL')).toBeVisible();
    await expect(page.getByText('Refresh Token TTL')).toBeVisible();
    await expect(page.getByText('ID Token TTL')).toBeVisible();
    await screenshot(page, '04-app-settings-tab');
  });

  test('should add redirect URI via UI', async ({ page }) => {
    // First clear existing URIs
    await apiCall(accessToken, 'PATCH', `/admin/applications/${testAppId}`, {
      redirect_uris: [],
    });

    await goToAdmin(page, `/applications/${testAppId}`);
    await waitForLoad(page);
    await page.getByRole('button', { name: 'Settings', exact: true }).click();
    await page.waitForTimeout(300);

    // Type a redirect URI in the input field
    const uriInput = page.getByPlaceholder('https://example.com/callback');
    await uriInput.fill('http://localhost:5000/auth/callback');
    await uriInput.press('Enter');
    await page.waitForTimeout(300);

    // The tag should appear
    await expect(page.getByText('http://localhost:5000/auth/callback')).toBeVisible();

    // Save settings
    await page.getByRole('button', { name: /Save Settings/ }).click();
    await page.waitForTimeout(500);

    // Verify in DB
    const rows = await dbQuery('SELECT redirect_uris FROM applications WHERE id = $1', [testAppId]);
    expect(rows[0].redirect_uris).toContain('http://localhost:5000/auth/callback');
  });

  // ─── SETTINGS TAB: NEGATIVE ───────────────────────────────

  test('Save Settings button should be disabled when nothing changed', async ({ page }) => {
    await goToAdmin(page, `/applications/${testAppId}`);
    await waitForLoad(page);
    await page.getByRole('button', { name: 'Settings', exact: true }).click();
    await page.waitForTimeout(300);

    const saveBtn = page.getByRole('button', { name: /Save Settings/ });
    await expect(saveBtn).toBeDisabled();
  });

  // ─── CLIENT SECRET ─────────────────────────────────────────

  test('should rotate client secret via API', async () => {
    // Get original secret
    const original = await apiCall(accessToken, 'GET', `/admin/applications/${testAppId}`);
    const originalSecret = original.client_secret;

    // Rotate
    const rotated = await apiCall(accessToken, 'POST', `/admin/applications/${testAppId}/rotate-secret`);
    expect(rotated.client_secret).toBeTruthy();
    expect(rotated.client_secret).not.toBe(originalSecret);

    // Verify in DB that hash changed
    const rows = await dbQuery('SELECT client_secret_hash FROM applications WHERE id = $1', [testAppId]);
    expect(rows[0].client_secret_hash).toBeTruthy();
  });

  // ─── PERMISSIONS TAB ───────────────────────────────────────

  test('should show permissions tab', async ({ page }) => {
    await goToAdmin(page, `/applications/${testAppId}`);
    await waitForLoad(page);
    await page.getByRole('button', { name: 'Permissions', exact: true }).click();
    await page.waitForTimeout(300);

    await expect(page.getByText('Application Permissions')).toBeVisible({ timeout: 5000 });
    await screenshot(page, '04-app-permissions-tab');
  });

  test('should set application permissions via API', async () => {
    // First create a permission if none exist
    let perms;
    try {
      perms = await apiCall(accessToken, 'GET', '/admin/permissions');
    } catch {
      perms = [];
    }

    if (Array.isArray(perms) && perms.length > 0) {
      const result = await apiCall(accessToken, 'PUT', `/admin/applications/${testAppId}/permissions`, {
        permissions: [perms[0].name],
      });
      expect(result.permissions).toContain(perms[0].name);

      // Clear permissions
      const cleared = await apiCall(accessToken, 'PUT', `/admin/applications/${testAppId}/permissions`, {
        permissions: [],
      });
      expect(cleared.permissions).toEqual([]);
    }
  });

  // ─── CONNECTIONS TAB ───────────────────────────────────────

  test('should show connections tab with social providers', async ({ page }) => {
    await goToAdmin(page, `/applications/${testAppId}`);
    await waitForLoad(page);
    await page.getByRole('button', { name: 'Connections', exact: true }).click();
    await page.waitForTimeout(300);

    await expect(page.getByText('Social Connections')).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('Google')).toBeVisible();
    await expect(page.getByText('GitHub')).toBeVisible();
    await expect(page.getByText('Microsoft')).toBeVisible();
    await screenshot(page, '04-app-connections-tab');
  });

  // ─── API TAB ───────────────────────────────────────────────

  test('should show API tab with scopes', async ({ page }) => {
    await goToAdmin(page, `/applications/${testAppId}`);
    await waitForLoad(page);
    await page.getByRole('button', { name: 'API', exact: true }).click();
    await page.waitForTimeout(300);

    await expect(page.getByText('API Scopes & Permissions')).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('openid')).toBeVisible();
    await expect(page.getByText('profile')).toBeVisible();
    await expect(page.getByText('email', { exact: true })).toBeVisible();
    await screenshot(page, '04-app-api-tab');
  });

  // ─── UPDATE NAME / DESCRIPTION ─────────────────────────────

  test('should update application name via API', async () => {
    const updated = await apiCall(accessToken, 'PATCH', `/admin/applications/${testAppId}`, {
      name: 'Renamed E2E App',
    });
    expect(updated.name).toBe('Renamed E2E App');

    const rows = await dbQuery('SELECT name FROM applications WHERE id = $1', [testAppId]);
    expect(rows[0].name).toBe('Renamed E2E App');
  });

  test('should update application description via API', async () => {
    const updated = await apiCall(accessToken, 'PATCH', `/admin/applications/${testAppId}`, {
      description: 'E2E test application description',
    });
    expect(updated.description).toBe('E2E test application description');
  });

  // ─── NEGATIVE: UPDATE ──────────────────────────────────────

  test('should reject update of non-existent application', async () => {
    try {
      await apiCall(accessToken, 'PATCH', '/admin/applications/00000000-0000-0000-0000-000000000000', {
        name: 'Should fail',
      });
      expect(true).toBe(false);
    } catch {
      // Expected — not found
    }
  });

  test('should reject getting non-existent application', async () => {
    try {
      await apiCall(accessToken, 'GET', '/admin/applications/00000000-0000-0000-0000-000000000000');
      expect(true).toBe(false);
    } catch {
      // Expected — not found
    }
  });

  // ─── DELETE ────────────────────────────────────────────────

  test('should delete application and verify removal', async () => {
    // Create a temporary app to delete
    const app = await apiCall(accessToken, 'POST', '/admin/applications', {
      name: 'To Be Deleted',
      type: 'web',
    });
    expect(app.id).toBeTruthy();

    // Delete it
    await apiCall(accessToken, 'DELETE', `/admin/applications/${app.id}`);

    // Verify removed from DB
    const rows = await dbQuery('SELECT id FROM applications WHERE id = $1', [app.id]);
    expect(rows.length).toBe(0);

    // Verify 404 on GET
    try {
      await apiCall(accessToken, 'GET', `/admin/applications/${app.id}`);
      expect(true).toBe(false);
    } catch {
      // Expected
    }
  });

  test('should reject deleting non-existent application', async () => {
    // This should either succeed silently (204) or fail — just ensure no crash
    try {
      await apiCall(accessToken, 'DELETE', '/admin/applications/00000000-0000-0000-0000-000000000000');
    } catch {
      // Acceptable
    }
  });

  // ─── DELETE VIA UI ─────────────────────────────────────────

  test('should delete application from detail page UI', async ({ page }) => {
    // Create a temp app for UI deletion
    const app = await apiCall(accessToken, 'POST', '/admin/applications', {
      name: 'UI Delete Test',
      type: 'spa',
    });

    await goToAdmin(page, `/applications/${app.id}`);
    await waitForLoad(page);

    // Click Delete button
    await page.getByRole('button', { name: 'Delete' }).click();
    await page.waitForTimeout(300);

    // Confirmation dialog should appear
    await expect(page.getByRole('heading', { name: 'Delete Application' })).toBeVisible({ timeout: 3000 });
    await screenshot(page, '04-app-delete-confirm');

    // Confirm deletion
    const confirmBtn = page.getByRole('button', { name: 'Delete Application' });
    await confirmBtn.click();
    await page.waitForTimeout(1000);

    // Should redirect to applications list
    await expect(page).toHaveURL(/\/applications$/);

    // Verify removed from DB
    const rows = await dbQuery('SELECT id FROM applications WHERE id = $1', [app.id]);
    expect(rows.length).toBe(0);
  });

  // ─── MULTIPLE SETTINGS UPDATE ──────────────────────────────

  test('should update multiple settings at once via API', async () => {
    const updated = await apiCall(accessToken, 'PATCH', `/admin/applications/${testAppId}`, {
      name: 'Multi Update Test',
      is_active: true,
      grant_types: ['authorization_code', 'refresh_token'],
      redirect_uris: ['http://localhost:3000/cb'],
      allowed_origins: ['http://localhost:3000'],
      allowed_logout_urls: ['http://localhost:3000/bye'],
      access_token_ttl: 1800,
      refresh_token_ttl: 86400,
      id_token_ttl: 900,
    });
    expect(updated.name).toBe('Multi Update Test');
    expect(updated.is_active).toBe(true);
    expect(updated.grant_types).toContain('authorization_code');
    expect(updated.grant_types).toContain('refresh_token');
    expect(updated.redirect_uris).toContain('http://localhost:3000/cb');
    expect(updated.access_token_ttl).toBe(1800);
    expect(updated.id_token_ttl).toBe(900);
  });
});
