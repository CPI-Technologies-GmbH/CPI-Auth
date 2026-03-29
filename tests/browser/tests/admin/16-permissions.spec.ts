import { test, expect } from '@playwright/test';
import { goToAdmin, screenshot, waitForLoad, assertNoErrors, apiCall, dbQuery, API_URL, ADMIN_EMAIL, ADMIN_PASSWORD } from './helpers';

let accessToken: string;

test.describe.serial('Admin Permissions CRUD', () => {
  test.beforeAll(async () => {
    const res = await fetch(`${API_URL}/admin/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email: ADMIN_EMAIL, password: ADMIN_PASSWORD }),
    });
    accessToken = (await res.json()).access_token;
  });

  test('should list permissions via API', async () => {
    const permissions = await apiCall(accessToken, 'GET', '/admin/permissions');
    expect(Array.isArray(permissions)).toBe(true);
    expect(permissions.length).toBeGreaterThan(0);

    // System permissions should include known defaults
    const names = permissions.map((p: any) => p.name);
    expect(names).toContain('users:read');
    expect(names).toContain('users:write');
  });

  test('should create a custom permission', async () => {
    const perm = await apiCall(accessToken, 'POST', '/admin/permissions', {
      name: `e2e:custom-${Date.now()}`,
      display_name: 'E2E Custom Permission',
      group_name: 'E2E',
      description: 'Created by E2E test',
    });
    expect(perm.id).toBeTruthy();
    expect(perm.is_system).toBe(false);

    // Verify in DB
    const rows = await dbQuery('SELECT name, is_system FROM permissions WHERE id = $1', [perm.id]);
    expect(rows.length).toBe(1);
    expect(rows[0].is_system).toBe(false);

    // Cleanup
    await apiCall(accessToken, 'DELETE', `/admin/permissions/${perm.id}`);
  });

  test('should update a custom permission', async () => {
    const perm = await apiCall(accessToken, 'POST', '/admin/permissions', {
      name: `e2e:update-${Date.now()}`,
      display_name: 'Original Name',
      group_name: 'E2E',
    });

    const updated = await apiCall(accessToken, 'PATCH', `/admin/permissions/${perm.id}`, {
      display_name: 'Updated Name',
    });
    expect(updated.display_name).toBe('Updated Name');

    await apiCall(accessToken, 'DELETE', `/admin/permissions/${perm.id}`);
  });

  test('should not delete a system permission', async () => {
    const permissions = await apiCall(accessToken, 'GET', '/admin/permissions');
    const systemPerm = permissions.find((p: any) => p.is_system === true);
    expect(systemPerm).toBeTruthy();

    const res = await fetch(`${API_URL}/admin/permissions/${systemPerm.id}`, {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${accessToken}`,
      },
    });
    expect(res.status).toBeGreaterThanOrEqual(400);
  });

  test('should display permissions on roles page with Create Permission button', async ({ page }) => {
    await goToAdmin(page, '/roles');
    await waitForLoad(page);
    await assertNoErrors(page);

    await expect(page.getByRole('heading', { name: 'All Permissions' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Create Permission' })).toBeVisible();
    await screenshot(page, 'perms-01-list');
  });

  test('should show system badge on system permissions', async ({ page }) => {
    await goToAdmin(page, '/roles');
    await waitForLoad(page);
    await assertNoErrors(page);

    // System permissions should have a "System" badge
    const systemBadges = page.locator('text=System');
    await expect(systemBadges.first()).toBeVisible({ timeout: 5000 });
    await screenshot(page, 'perms-02-system-badge');
  });
});
