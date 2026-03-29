import { test, expect } from '@playwright/test';
import { goToAdmin, screenshot, waitForLoad, assertNoErrors, apiCall, dbQuery, API_URL, ADMIN_EMAIL, ADMIN_PASSWORD } from './helpers';

let accessToken: string;

test.describe.serial('Admin Roles & Permissions', () => {
  test.beforeAll(async () => {
    const res = await fetch(`${API_URL}/admin/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email: ADMIN_EMAIL, password: ADMIN_PASSWORD }),
    });
    accessToken = (await res.json()).access_token;
  });

  test('should display roles page with roles list', async ({ page }) => {
    await goToAdmin(page, '/roles');
    await waitForLoad(page);
    await assertNoErrors(page);
    await expect(page.getByRole('heading', { name: 'Roles & Permissions' })).toBeVisible();
    await expect(page.getByRole('heading', { name: 'All Permissions' })).toBeVisible();
    await screenshot(page, 'roles-01-list');
  });

  test('should create role via API and verify in UI', async ({ page }) => {
    const name = `E2E Role ${Date.now()}`;
    const role = await apiCall(accessToken, 'POST', '/admin/roles', {
      name,
      permissions: ['users:read', 'users:write'],
    });
    expect(role.id).toBeTruthy();

    const rows = await dbQuery('SELECT name FROM roles WHERE id = $1', [role.id]);
    expect(rows.length).toBe(1);
    expect(rows[0].name).toBe(name);

    await goToAdmin(page, '/roles');
    await waitForLoad(page);
    await assertNoErrors(page);
    await expect(page.getByText(name)).toBeVisible({ timeout: 5000 });
    await screenshot(page, 'roles-02-with-new');

    await apiCall(accessToken, 'DELETE', `/admin/roles/${role.id}`);
  });

  test('should show permissions list', async ({ page }) => {
    const permissions = await apiCall(accessToken, 'GET', '/admin/permissions');
    expect(Array.isArray(permissions)).toBe(true);
    expect(permissions.length).toBeGreaterThan(0);

    await goToAdmin(page, '/roles');
    await waitForLoad(page);
    await assertNoErrors(page);
    await expect(page.getByRole('heading', { name: 'All Permissions' })).toBeVisible();
    await screenshot(page, 'roles-03-permissions');
  });
});
