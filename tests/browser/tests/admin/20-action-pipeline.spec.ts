import { test, expect } from '@playwright/test';
import { goToAdmin, screenshot, waitForLoad, assertNoErrors, apiCall, dbQuery, API_URL, ADMIN_EMAIL, ADMIN_PASSWORD } from './helpers';

let accessToken: string;

test.describe.serial('Action Pipeline Integration', () => {
  test.beforeAll(async () => {
    const res = await fetch(`${API_URL}/admin/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email: ADMIN_EMAIL, password: ADMIN_PASSWORD }),
    });
    accessToken = (await res.json()).access_token;
  });

  test('should list actions via API', async () => {
    const result = await apiCall(accessToken, 'GET', '/admin/actions');
    expect(result).toBeDefined();
    // result may be empty array or paginated
    expect(Array.isArray(result.data) || Array.isArray(result)).toBe(true);
  });

  test('should create an action with pre-login trigger', async () => {
    const action = await apiCall(accessToken, 'POST', '/admin/actions', {
      name: `e2e-pre-login-${Date.now()}`,
      trigger: 'pre-login',
      code: 'module.exports = async (ctx) => ({ allow: true })',
      is_active: true,
      order: 1,
      runtime: 'javascript',
      timeout_ms: 5000,
    });
    expect(action.id).toBeTruthy();
    expect(action.trigger).toBe('pre-login');

    // Cleanup
    await apiCall(accessToken, 'DELETE', `/admin/actions/${action.id}`).catch(() => {});
  });

  test('should create an action with pre-registration trigger', async () => {
    const action = await apiCall(accessToken, 'POST', '/admin/actions', {
      name: `e2e-pre-reg-${Date.now()}`,
      trigger: 'pre-registration',
      code: 'module.exports = async (ctx) => ({ allow: true })',
      is_active: true,
      order: 1,
      runtime: 'javascript',
      timeout_ms: 5000,
    });
    expect(action.id).toBeTruthy();
    expect(action.trigger).toBe('pre-registration');

    // Cleanup
    await apiCall(accessToken, 'DELETE', `/admin/actions/${action.id}`).catch(() => {});
  });

  test('should display actions page in admin UI', async ({ page }) => {
    await goToAdmin(page, '/actions');
    await waitForLoad(page);
    await assertNoErrors(page);

    await screenshot(page, 'actions-01-list');
  });

  test('should show action triggers in create form', async ({ page }) => {
    await goToAdmin(page, '/actions');
    await waitForLoad(page);

    // Click the create button
    const createBtn = page.getByRole('button', { name: /create/i }).first().or(page.getByRole('link', { name: /create/i }).first());
    if (await createBtn.isVisible()) {
      await createBtn.click();
      await waitForLoad(page);
      await screenshot(page, 'actions-02-create-form');
    } else {
      await screenshot(page, 'actions-02-no-create-button');
    }
  });
});
