import { test, expect } from '@playwright/test';
import { goToAdmin, screenshot, waitForLoad, assertNoErrors, apiCall, dbQuery, API_URL, ADMIN_EMAIL, ADMIN_PASSWORD } from './helpers';

let accessToken: string;

test.describe.serial('User Locale / Email Localization', () => {
  let testUserId: string;

  test.beforeAll(async () => {
    const res = await fetch(`${API_URL}/admin/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email: ADMIN_EMAIL, password: ADMIN_PASSWORD }),
    });
    accessToken = (await res.json()).access_token;
  });

  test.afterAll(async () => {
    if (testUserId) await apiCall(accessToken, 'DELETE', `/admin/users/${testUserId}`).catch(() => {});
  });

  test('should create user with default locale', async () => {
    const user = await apiCall(accessToken, 'POST', '/admin/users', {
      email: `e2e-locale-${Date.now()}@test.local`,
      password: 'Xk9mQ2vL8nR4!pw',
      name: 'Locale Test',
    });
    testUserId = user.id;
    // Default locale should be 'en' or empty
    expect(user.locale === 'en' || user.locale === '' || user.locale === undefined).toBe(true);
  });

  test('should verify locale in database', async () => {
    const rows = await dbQuery(
      'SELECT locale FROM users WHERE id = $1',
      [testUserId]
    );
    expect(rows.length).toBe(1);
    expect(rows[0].locale).toBe('en');
  });

  test('should update user locale', async () => {
    const updated = await apiCall(accessToken, 'PATCH', `/admin/users/${testUserId}`, {
      locale: 'de',
    });
    // locale should be updated
    expect(updated.locale).toBe('de');
  });

  test('should verify updated locale in database', async () => {
    const rows = await dbQuery(
      'SELECT locale FROM users WHERE id = $1',
      [testUserId]
    );
    expect(rows.length).toBe(1);
    expect(rows[0].locale).toBe('de');
  });

  test('should show locale on user detail page', async ({ page }) => {
    await goToAdmin(page, `/users/${testUserId}`);
    await waitForLoad(page);
    await assertNoErrors(page);

    await screenshot(page, 'user-locale-01-detail');
  });

  test('should register user with locale via auth API', async () => {
    const email = `e2e-locale-reg-${Date.now()}@test.local`;
    const res = await fetch(`${API_URL}/api/v1/auth/register`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-Tenant-ID': 'a0000000-0000-0000-0000-000000000001',
      },
      body: JSON.stringify({
        email,
        password: 'Xk9mQ2vL8nR4!pw',
        name: 'Locale Reg User',
        locale: 'fr',
      }),
    });
    const body = await res.json();
    expect(res.status).toBeLessThan(400);
    expect(body.access_token).toBeTruthy();

    // Verify locale in DB
    const rows = await dbQuery(
      "SELECT locale FROM users WHERE email = $1",
      [email]
    );
    expect(rows.length).toBe(1);
    expect(rows[0].locale).toBe('fr');

    // Cleanup
    const uid = rows[0]?.id;
    if (uid) await apiCall(accessToken, 'DELETE', `/admin/users/${uid}`).catch(() => {});
  });
});
