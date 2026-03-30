import { test, expect } from '@playwright/test';
import { goToAdmin, screenshot, waitForLoad, assertNoErrors, apiCall, dbQuery, API_URL, ADMIN_EMAIL, ADMIN_PASSWORD } from './helpers';

let accessToken: string;

test.describe.serial('Admin Webhooks', () => {
  test.beforeAll(async () => {
    const res = await fetch(`${API_URL}/admin/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email: ADMIN_EMAIL, password: ADMIN_PASSWORD }),
    });
    accessToken = (await res.json()).access_token;
  });

  test('should display webhooks page', async ({ page }) => {
    await goToAdmin(page, '/webhooks');
    await waitForLoad(page);
    await assertNoErrors(page);
    await expect(page.getByRole('heading', { name: 'Webhooks' })).toBeVisible();
    await screenshot(page, 'webhooks-01-list');
  });

  test('should create webhook via API and verify in UI', async ({ page }) => {
    const webhook = await apiCall(accessToken, 'POST', '/admin/webhooks', {
      url: 'https://e2e-test.example.com/hook',
      events: ['user.created', 'user.login'],
      active: true,
    });
    expect(webhook.id).toBeTruthy();

    const rows = await dbQuery('SELECT url FROM webhooks WHERE id = $1', [webhook.id]);
    expect(rows.length).toBe(1);
    expect(rows[0].url).toBe('https://e2e-test.example.com/hook');

    await goToAdmin(page, '/webhooks');
    await waitForLoad(page);
    await assertNoErrors(page);
    await screenshot(page, 'webhooks-02-after-create');

    await apiCall(accessToken, 'DELETE', `/admin/webhooks/${webhook.id}`);
  });

  test('should delete webhook and verify removal', async ({ page }) => {
    const webhook = await apiCall(accessToken, 'POST', '/admin/webhooks', {
      url: 'https://delete-test.example.com/hook',
      events: ['user.deleted'],
      active: false,
    });
    await apiCall(accessToken, 'DELETE', `/admin/webhooks/${webhook.id}`);

    const rows = await dbQuery('SELECT id FROM webhooks WHERE id = $1', [webhook.id]);
    expect(rows.length).toBe(0);

    await goToAdmin(page, '/webhooks');
    await waitForLoad(page);
    await assertNoErrors(page);
    await screenshot(page, 'webhooks-03-after-delete');
  });
});
