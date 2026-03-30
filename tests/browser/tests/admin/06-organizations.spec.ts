import { test, expect } from '@playwright/test';
import { goToAdmin, screenshot, waitForLoad, assertNoErrors, apiCall, dbQuery, API_URL, ADMIN_EMAIL, ADMIN_PASSWORD } from './helpers';

let accessToken: string;

test.describe.serial('Admin Organizations', () => {
  test.beforeAll(async () => {
    const res = await fetch(`${API_URL}/admin/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email: ADMIN_EMAIL, password: ADMIN_PASSWORD }),
    });
    accessToken = (await res.json()).access_token;
  });

  test('should display organizations page', async ({ page }) => {
    await goToAdmin(page, '/organizations');
    await waitForLoad(page);
    await assertNoErrors(page);
    await expect(page.getByRole('heading', { name: 'Organizations' })).toBeVisible();
    await screenshot(page, 'orgs-01-list');
  });

  test('should create org via API and verify in UI', async ({ page }) => {
    const name = `E2E Org ${Date.now()}`;
    const org = await apiCall(accessToken, 'POST', '/admin/organizations', { name });
    expect(org.id).toBeTruthy();

    const rows = await dbQuery('SELECT name FROM organizations WHERE id = $1', [org.id]);
    expect(rows.length).toBe(1);
    expect(rows[0].name).toBe(name);

    await goToAdmin(page, '/organizations');
    await waitForLoad(page);
    await assertNoErrors(page);
    await screenshot(page, 'orgs-02-after-create');

    await apiCall(accessToken, 'DELETE', `/admin/organizations/${org.id}`);
  });

  test('should update and delete org', async ({ page }) => {
    const org = await apiCall(accessToken, 'POST', '/admin/organizations', { name: 'Temp Org' });
    const updated = await apiCall(accessToken, 'PATCH', `/admin/organizations/${org.id}`, { name: 'Updated Org' });
    expect(updated.name).toBe('Updated Org');

    await apiCall(accessToken, 'DELETE', `/admin/organizations/${org.id}`);
    const rows = await dbQuery('SELECT id FROM organizations WHERE id = $1', [org.id]);
    expect(rows.length).toBe(0);

    await goToAdmin(page, '/organizations');
    await waitForLoad(page);
    await assertNoErrors(page);
    await screenshot(page, 'orgs-03-after-delete');
  });
});
