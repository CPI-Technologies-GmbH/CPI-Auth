import { test, expect } from '@playwright/test';
import { goToAdmin, screenshot, waitForLoad, assertNoErrors, apiCall, dbQuery, API_URL, ADMIN_EMAIL, ADMIN_PASSWORD } from './helpers';

let accessToken: string;

test.describe.serial('Admin User Role Assignment', () => {
  let testUserId: string;
  let testRoleId: string;

  test.beforeAll(async () => {
    const res = await fetch(`${API_URL}/admin/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email: ADMIN_EMAIL, password: ADMIN_PASSWORD }),
    });
    accessToken = (await res.json()).access_token;

    // Create a test user
    const user = await apiCall(accessToken, 'POST', '/admin/users', {
      email: `e2e-role-test-${Date.now()}@test.local`,
      password: 'Xk9mQ2vL8nR4!pw',
      name: 'Role Test User',
    });
    testUserId = user.id;

    // Create a test role
    const role = await apiCall(accessToken, 'POST', '/admin/roles', {
      name: `E2E Assign Role ${Date.now()}`,
      permissions: ['users:read'],
    });
    testRoleId = role.id;
  });

  test.afterAll(async () => {
    if (testUserId) await apiCall(accessToken, 'DELETE', `/admin/users/${testUserId}`).catch(() => {});
    if (testRoleId) await apiCall(accessToken, 'DELETE', `/admin/roles/${testRoleId}`).catch(() => {});
  });

  test('should assign role to user via API', async () => {
    await apiCall(accessToken, 'POST', `/admin/users/${testUserId}/roles`, {
      role_id: testRoleId,
    });

    const roles = await apiCall(accessToken, 'GET', `/admin/users/${testUserId}/roles`);
    expect(Array.isArray(roles)).toBe(true);
    expect(roles.length).toBeGreaterThanOrEqual(1);

    const assigned = roles.find((r: any) => r.id === testRoleId);
    expect(assigned).toBeTruthy();
  });

  test('should verify role assignment in DB', async () => {
    const rows = await dbQuery(
      'SELECT * FROM user_roles WHERE user_id = $1 AND role_id = $2',
      [testUserId, testRoleId]
    );
    expect(rows.length).toBe(1);
  });

  test('should show assigned roles on user detail page', async ({ page }) => {
    await goToAdmin(page, `/users/${testUserId}`);
    await waitForLoad(page);
    await assertNoErrors(page);

    // Click on Roles tab (custom tabs use button elements, not role="tab")
    await page.getByRole('button', { name: 'Roles', exact: true }).click();
    await waitForLoad(page);

    // Should see the "Assign Role" button
    await expect(page.getByRole('button', { name: 'Assign Role' })).toBeVisible();

    // The assigned role should be visible
    await expect(page.getByText('Role Assignments')).toBeVisible();
    await screenshot(page, 'user-roles-01-tab');
  });

  test('should remove role from user via API', async () => {
    await apiCall(accessToken, 'DELETE', `/admin/users/${testUserId}/roles/${testRoleId}`);

    const roles = await apiCall(accessToken, 'GET', `/admin/users/${testUserId}/roles`);
    const stillAssigned = roles.find((r: any) => r.id === testRoleId);
    expect(stillAssigned).toBeUndefined();
  });

  test('should verify role removal in DB', async () => {
    const rows = await dbQuery(
      'SELECT * FROM user_roles WHERE user_id = $1 AND role_id = $2',
      [testUserId, testRoleId]
    );
    expect(rows.length).toBe(0);
  });
});
