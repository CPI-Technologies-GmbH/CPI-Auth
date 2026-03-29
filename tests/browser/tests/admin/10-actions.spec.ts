import { test, expect } from '@playwright/test';
import { goToAdmin, screenshot, waitForLoad, assertNoErrors, apiCall, dbQuery, API_URL, ADMIN_EMAIL, ADMIN_PASSWORD } from './helpers';

let accessToken: string;

test.describe.serial('Admin Actions', () => {
  test.beforeAll(async () => {
    const res = await fetch(`${API_URL}/admin/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email: ADMIN_EMAIL, password: ADMIN_PASSWORD }),
    });
    accessToken = (await res.json()).access_token;
  });

  test('should display actions page (empty state)', async ({ page }) => {
    await goToAdmin(page, '/actions');
    await waitForLoad(page);
    await assertNoErrors(page);
    await expect(page.getByRole('heading', { name: 'Actions', exact: true })).toBeVisible();
    await screenshot(page, 'actions-01-empty');
  });

  test('should create action via API and verify in UI', async ({ page }) => {
    const action = await apiCall(accessToken, 'POST', '/admin/actions', {
      trigger: 'post-login',
      name: 'E2E Test Action',
      code: 'console.log("test")',
      enabled: true,
    });
    expect(action.id).toBeTruthy();

    const rows = await dbQuery('SELECT name, trigger FROM actions WHERE id = $1', [action.id]);
    expect(rows.length).toBe(1);
    expect(rows[0].name).toBe('E2E Test Action');

    await goToAdmin(page, '/actions');
    await waitForLoad(page);
    await assertNoErrors(page);
    await expect(page.getByText('E2E Test Action').first()).toBeVisible({ timeout: 5000 });
    await screenshot(page, 'actions-02-with-action');

    await apiCall(accessToken, 'DELETE', `/admin/actions/${action.id}`);
  });

  test('should update and delete action', async ({ page }) => {
    const action = await apiCall(accessToken, 'POST', '/admin/actions', {
      trigger: 'pre-registration',
      name: 'Temp Action',
      code: 'console.log("temp")',
      enabled: false,
    });
    const updated = await apiCall(accessToken, 'PATCH', `/admin/actions/${action.id}`, {
      name: 'Updated Action',
    });
    expect(updated.name).toBe('Updated Action');

    await apiCall(accessToken, 'DELETE', `/admin/actions/${action.id}`);
    const rows = await dbQuery('SELECT id FROM actions WHERE id = $1', [action.id]);
    expect(rows.length).toBe(0);

    await goToAdmin(page, '/actions');
    await waitForLoad(page);
    await assertNoErrors(page);
    await screenshot(page, 'actions-03-after-delete');
  });
});
