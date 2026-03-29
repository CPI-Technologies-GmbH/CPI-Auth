import { test, expect } from '@playwright/test';
import { goToAdmin, screenshot, waitForLoad, assertNoErrors, apiCall, dbQuery, API_URL, ADMIN_EMAIL, ADMIN_PASSWORD } from './helpers';

let accessToken: string;

test.describe.serial('Admin Users Management', () => {
  test.beforeAll(async () => {
    const res = await fetch(`${API_URL}/admin/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email: ADMIN_EMAIL, password: ADMIN_PASSWORD }),
    });
    accessToken = (await res.json()).access_token;
  });

  test('should display users page with table', async ({ page }) => {
    await goToAdmin(page, '/users');
    await waitForLoad(page);
    await assertNoErrors(page);
    await expect(page.getByRole('heading', { name: 'Users' })).toBeVisible();
    // Table should have rows with user data
    await expect(page.locator('table tbody tr').first()).toBeVisible({ timeout: 5000 });
    await screenshot(page, 'users-01-list');
  });

  test('should create a user via API and see in list', async ({ page }) => {
    const email = `e2e-user-${Date.now()}@test.local`;
    const user = await apiCall(accessToken, 'POST', '/admin/users', {
      email,
      password: 'Xk9mQ2vL8nR4!pw',
      name: 'E2E Test User',
    });
    expect(user.id).toBeTruthy();
    expect(user.email).toBe(email);

    // Verify in DB
    const rows = await dbQuery('SELECT email, name FROM users WHERE id = $1', [user.id]);
    expect(rows.length).toBe(1);
    expect(rows[0].email).toBe(email);

    // Verify shows in UI
    await goToAdmin(page, '/users');
    await waitForLoad(page);
    await assertNoErrors(page);
    await expect(page.getByText(email)).toBeVisible({ timeout: 10000 });
    await screenshot(page, 'users-02-after-create');

    // Cleanup
    await apiCall(accessToken, 'DELETE', `/admin/users/${user.id}`);
  });

  test('should update a user via API', async ({ page }) => {
    const email = `e2e-update-${Date.now()}@test.local`;
    const user = await apiCall(accessToken, 'POST', '/admin/users', {
      email,
      password: 'Xk9mQ2vL8nR4!pw',
      name: 'Original Name',
    });

    const updated = await apiCall(accessToken, 'PATCH', `/admin/users/${user.id}`, {
      name: 'Updated Name',
    });
    expect(updated.name).toBe('Updated Name');

    const rows = await dbQuery('SELECT name FROM users WHERE id = $1', [user.id]);
    expect(rows[0].name).toBe('Updated Name');

    await goToAdmin(page, '/users');
    await waitForLoad(page);
    await assertNoErrors(page);
    await screenshot(page, 'users-03-after-update');

    await apiCall(accessToken, 'DELETE', `/admin/users/${user.id}`);
  });

  test('should delete a user via API', async ({ page }) => {
    const email = `e2e-delete-${Date.now()}@test.local`;
    const user = await apiCall(accessToken, 'POST', '/admin/users', {
      email,
      password: 'Xk9mQ2vL8nR4!pw',
      name: 'Delete Me',
    });
    await apiCall(accessToken, 'DELETE', `/admin/users/${user.id}`);

    const rows = await dbQuery('SELECT id FROM users WHERE id = $1', [user.id]);
    expect(rows.length).toBe(0);

    await goToAdmin(page, '/users');
    await waitForLoad(page);
    await assertNoErrors(page);
    await screenshot(page, 'users-04-after-delete');
  });
});
