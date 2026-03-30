import { test, expect } from '@playwright/test';
import { goToAdmin, screenshot, waitForLoad, assertNoErrors, apiCall, dbQuery, API_URL, ADMIN_EMAIL, ADMIN_PASSWORD } from './helpers';

let accessToken: string;

test.describe.serial('Admin Application Permissions', () => {
  let testAppId: string;

  test.beforeAll(async () => {
    const res = await fetch(`${API_URL}/admin/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email: ADMIN_EMAIL, password: ADMIN_PASSWORD }),
    });
    accessToken = (await res.json()).access_token;

    // Create a test application
    const app = await apiCall(accessToken, 'POST', '/admin/applications', {
      name: `E2E Perm App ${Date.now()}`,
      type: 'spa',
    });
    testAppId = app.id;
  });

  test.afterAll(async () => {
    if (testAppId) await apiCall(accessToken, 'DELETE', `/admin/applications/${testAppId}`).catch(() => {});
  });

  test('should get empty permissions for new application', async () => {
    const result = await apiCall(accessToken, 'GET', `/admin/applications/${testAppId}/permissions`);
    expect(result.permissions).toBeDefined();
    expect(result.permissions.length).toBe(0);
  });

  test('should set application permissions', async () => {
    const result = await apiCall(accessToken, 'PUT', `/admin/applications/${testAppId}/permissions`, {
      permissions: ['users:read', 'users:write', 'roles:read'],
    });
    expect(result.permissions).toBeDefined();
    expect(result.permissions.length).toBe(3);
    expect(result.permissions).toContain('users:read');
    expect(result.permissions).toContain('users:write');
  });

  test('should verify app permissions in DB', async () => {
    const rows = await dbQuery(
      'SELECT permission_name FROM application_permissions WHERE application_id = $1 ORDER BY permission_name',
      [testAppId]
    );
    expect(rows.length).toBe(3);
    expect(rows.map((r: any) => r.permission_name)).toContain('users:read');
  });

  test('should update application permissions (replace)', async () => {
    const result = await apiCall(accessToken, 'PUT', `/admin/applications/${testAppId}/permissions`, {
      permissions: ['users:read'],
    });
    expect(result.permissions.length).toBe(1);

    // DB should also have only 1
    const rows = await dbQuery(
      'SELECT permission_name FROM application_permissions WHERE application_id = $1',
      [testAppId]
    );
    expect(rows.length).toBe(1);
  });

  test('should clear application permissions', async () => {
    const result = await apiCall(accessToken, 'PUT', `/admin/applications/${testAppId}/permissions`, {
      permissions: [],
    });
    expect(result.permissions.length).toBe(0);

    const rows = await dbQuery(
      'SELECT * FROM application_permissions WHERE application_id = $1',
      [testAppId]
    );
    expect(rows.length).toBe(0);
  });

  test('should display permissions tab on application detail page', async ({ page }) => {
    // Set some permissions first
    await apiCall(accessToken, 'PUT', `/admin/applications/${testAppId}/permissions`, {
      permissions: ['users:read', 'users:write'],
    });

    await goToAdmin(page, `/applications/${testAppId}`);
    await waitForLoad(page);
    await assertNoErrors(page);

    // Click on Permissions tab
    await page.getByRole('button', { name: 'Permissions', exact: true }).click();
    await waitForLoad(page);

    await expect(page.getByText('Application Permissions')).toBeVisible();
    await screenshot(page, 'app-perms-01-tab');
  });

  test('should show info banner when no permissions selected', async ({ page }) => {
    // Clear all permissions
    await apiCall(accessToken, 'PUT', `/admin/applications/${testAppId}/permissions`, {
      permissions: [],
    });

    await goToAdmin(page, `/applications/${testAppId}`);
    await waitForLoad(page);
    await assertNoErrors(page);

    await page.getByRole('button', { name: 'Permissions', exact: true }).click();
    await waitForLoad(page);

    // Should show the info banner
    await expect(page.getByText('No permissions selected')).toBeVisible({ timeout: 5000 });
    await screenshot(page, 'app-perms-02-empty-info');
  });
});
