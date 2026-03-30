import { test, expect } from '@playwright/test';
import { goToAdmin, screenshot, waitForLoad, assertNoErrors, apiCall, dbQuery, API_URL, ADMIN_EMAIL, ADMIN_PASSWORD } from './helpers';

let accessToken: string;

test.describe.serial('Admin Tenants Management', () => {
  test.beforeAll(async () => {
    const res = await fetch(`${API_URL}/admin/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email: ADMIN_EMAIL, password: ADMIN_PASSWORD }),
    });
    accessToken = (await res.json()).access_token;
  });

  test('should display tenants page without errors', async ({ page }) => {
    await goToAdmin(page, '/tenants');
    await waitForLoad(page);
    await assertNoErrors(page);
    await expect(page.getByRole('heading', { name: 'Tenants' })).toBeVisible();
    await screenshot(page, 'tenants-01-list');
  });

  test('should create a tenant via API and verify in UI', async ({ page }) => {
    const slug = `e2e-tenant-${Date.now()}`;
    const tenant = await apiCall(accessToken, 'POST', '/admin/tenants', {
      name: 'E2E Test Tenant',
      slug,
    });
    expect(tenant.id).toBeTruthy();
    expect(tenant.slug).toBe(slug);

    const rows = await dbQuery('SELECT name, slug FROM tenants WHERE id = $1', [tenant.id]);
    expect(rows.length).toBe(1);
    expect(rows[0].slug).toBe(slug);

    await goToAdmin(page, '/tenants');
    await waitForLoad(page);
    await assertNoErrors(page);
    await expect(page.getByText(slug)).toBeVisible({ timeout: 5000 });
    await screenshot(page, 'tenants-02-with-new');

    await apiCall(accessToken, 'DELETE', `/admin/tenants/${tenant.id}`);
  });

  test('should update and delete a tenant', async ({ page }) => {
    const tenant = await apiCall(accessToken, 'POST', '/admin/tenants', {
      name: 'Temp Tenant',
      slug: `temp-${Date.now()}`,
    });
    const updated = await apiCall(accessToken, 'PATCH', `/admin/tenants/${tenant.id}`, {
      name: 'Updated Tenant Name',
    });
    expect(updated.name).toBe('Updated Tenant Name');

    await apiCall(accessToken, 'DELETE', `/admin/tenants/${tenant.id}`);
    const rows = await dbQuery('SELECT id FROM tenants WHERE id = $1', [tenant.id]);
    expect(rows.length).toBe(0);

    await goToAdmin(page, '/tenants');
    await waitForLoad(page);
    await assertNoErrors(page);
    await screenshot(page, 'tenants-03-after-delete');
  });
});
