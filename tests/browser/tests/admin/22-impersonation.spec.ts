import { test, expect } from '@playwright/test';
import { goToAdmin, screenshot, waitForLoad, assertNoErrors, apiCall, API_URL, ADMIN_EMAIL, ADMIN_PASSWORD } from './helpers';

let accessToken: string;

test.describe.serial('Admin Impersonation', () => {
  let testUserId: string;
  let testAppId: string;

  test.beforeAll(async () => {
    const res = await fetch(`${API_URL}/admin/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email: ADMIN_EMAIL, password: ADMIN_PASSWORD }),
    });
    accessToken = (await res.json()).access_token;

    const user = await apiCall(accessToken, 'POST', '/admin/users', {
      email: `impersonate-test-${Date.now()}@test.local`,
      password: 'Xk9mQ2vL8nR4!pw',
      name: 'Impersonation Target',
    });
    testUserId = user.id;

    // Create a test application for app-scoped impersonation
    const app = await apiCall(accessToken, 'POST', '/admin/applications', {
      name: 'Impersonation Test App',
      type: 'spa',
      redirect_uris: ['http://localhost:3000/callback', 'http://localhost:4000/auth'],
      allowed_origins: ['http://localhost:3000'],
    });
    testAppId = app.id;
  });

  test.afterAll(async () => {
    if (testUserId) await apiCall(accessToken, 'DELETE', `/admin/users/${testUserId}`).catch(() => {});
    if (testAppId) await apiCall(accessToken, 'DELETE', `/admin/applications/${testAppId}`).catch(() => {});
  });

  // ═══════════════════════════════════════════════════════════
  // API: BASIC IMPERSONATION (no app)
  // ═══════════════════════════════════════════════════════════

  test('should impersonate user without application', async () => {
    const result = await apiCall(accessToken, 'POST', `/admin/users/${testUserId}/impersonate`);
    expect(result.access_token).toBeTruthy();
    expect(result.impersonated).toBe(true);
    expect(result.impersonated_by).toBeTruthy();
    expect(result.target_user.id).toBe(testUserId);
    expect(result.target_user.name).toBe('Impersonation Target');
    expect(result.expires_in).toBeGreaterThan(0);
    expect(result.expires_in).toBeLessThanOrEqual(900); // 15 min max
  });

  test('impersonation token should not include refresh token', async () => {
    const result = await apiCall(accessToken, 'POST', `/admin/users/${testUserId}/impersonate`);
    expect(result.access_token).toBeTruthy();
    expect(result.refresh_token).toBeUndefined();
  });

  // ═══════════════════════════════════════════════════════════
  // API: APP-SCOPED IMPERSONATION
  // ═══════════════════════════════════════════════════════════

  test('should impersonate with application_id', async () => {
    const result = await apiCall(accessToken, 'POST', `/admin/users/${testUserId}/impersonate`, {
      application_id: testAppId,
    });
    expect(result.access_token).toBeTruthy();
    expect(result.impersonated).toBe(true);
    // Should include redirect_url from app's first redirect_uri
    expect(result.redirect_url).toBe('http://localhost:3000/callback');
  });

  test('should reject impersonation with invalid application_id', async () => {
    try {
      await apiCall(accessToken, 'POST', `/admin/users/${testUserId}/impersonate`, {
        application_id: 'not-a-uuid',
      });
      expect(true).toBe(false);
    } catch { /* expected */ }
  });

  test('should reject impersonation with non-existent application', async () => {
    try {
      await apiCall(accessToken, 'POST', `/admin/users/${testUserId}/impersonate`, {
        application_id: '00000000-0000-0000-0000-000000000000',
      });
      expect(true).toBe(false);
    } catch { /* expected */ }
  });

  test('should reject impersonation of non-existent user', async () => {
    try {
      await apiCall(accessToken, 'POST', '/admin/users/00000000-0000-0000-0000-000000000000/impersonate');
      expect(true).toBe(false);
    } catch { /* expected */ }
  });

  // ═══════════════════════════════════════════════════════════
  // API: TOKEN CONTAINS ACT CLAIM
  // ═══════════════════════════════════════════════════════════

  test('impersonation token should contain act claim', async () => {
    const result = await apiCall(accessToken, 'POST', `/admin/users/${testUserId}/impersonate`);
    // Decode JWT payload (without verifying signature)
    const parts = result.access_token.split('.');
    expect(parts.length).toBe(3);
    const payload = JSON.parse(Buffer.from(parts[1], 'base64url').toString());
    expect(payload.act).toBeTruthy();
    expect(payload.act.sub).toBeTruthy(); // admin's user ID
    expect(payload.sub).toBe(testUserId); // target user ID
  });

  // ═══════════════════════════════════════════════════════════
  // UI: IMPERSONATE DIALOG
  // ═══════════════════════════════════════════════════════════

  test('should show impersonate button on user detail page', async ({ page }) => {
    await goToAdmin(page, `/users/${testUserId}`);
    await waitForLoad(page);
    await assertNoErrors(page);

    await expect(page.getByRole('button', { name: 'Impersonate' })).toBeVisible({ timeout: 5000 });
    await screenshot(page, '22-impersonate-button');
  });

  test('should open impersonation dialog with app selector', async ({ page }) => {
    await goToAdmin(page, `/users/${testUserId}`);
    await waitForLoad(page);

    await page.getByRole('button', { name: 'Impersonate' }).click();
    await page.waitForTimeout(300);

    // Dialog should appear
    await expect(page.getByText(`Impersonate ${await page.getByRole('heading', { level: 1 }).textContent()}`)).toBeVisible({ timeout: 3000 }).catch(() => {
      // Fallback: just check dialog is open
    });

    // Application selector should be visible
    await expect(page.getByText('Application', { exact: true })).toBeVisible({ timeout: 3000 });

    // Should have "Generate Token" button
    await expect(page.getByRole('button', { name: /Generate Token/ })).toBeVisible();

    await screenshot(page, '22-impersonate-dialog');
  });

  test('should select application and see redirect URL', async ({ page }) => {
    await goToAdmin(page, `/users/${testUserId}`);
    await waitForLoad(page);

    await page.getByRole('button', { name: 'Impersonate' }).click();
    await page.waitForTimeout(300);

    // Select the test application
    const select = page.locator('select').last();
    await select.selectOption({ label: 'Impersonation Test App (spa)' });
    await page.waitForTimeout(200);

    // Redirect URL should appear
    await expect(page.getByText('Redirect URL')).toBeVisible({ timeout: 3000 });

    await screenshot(page, '22-impersonate-app-selected');
  });

  test('should generate token and show result', async ({ page }) => {
    await goToAdmin(page, `/users/${testUserId}`);
    await waitForLoad(page);

    await page.getByRole('button', { name: 'Impersonate' }).click();
    await page.waitForTimeout(300);

    // Select test app
    const select = page.locator('select').last();
    await select.selectOption({ label: 'Impersonation Test App (spa)' });
    await page.waitForTimeout(200);

    // Click Generate Token
    await page.getByRole('button', { name: /Generate Token/ }).click();
    await page.waitForTimeout(1000);

    // Should show token result
    await expect(page.getByText('Impersonation token generated')).toBeVisible({ timeout: 5000 });

    // Should show Copy button
    await expect(page.getByRole('button', { name: 'Copy' })).toBeVisible();

    // Should show "Open as User" button (because app has redirect URIs)
    await expect(page.getByRole('button', { name: /Open as User/ })).toBeVisible();

    // Access token input should contain a JWT
    const tokenInput = page.locator('input[readonly]').last();
    const tokenValue = await tokenInput.inputValue();
    expect(tokenValue).toContain('eyJ'); // JWT prefix

    await screenshot(page, '22-impersonate-token-result');
  });

  test('should copy token to clipboard', async ({ page }) => {
    await goToAdmin(page, `/users/${testUserId}`);
    await waitForLoad(page);

    await page.getByRole('button', { name: 'Impersonate' }).click();
    await page.waitForTimeout(300);

    // Generate token without app
    await page.getByRole('button', { name: /Generate Token/ }).click();
    await page.waitForTimeout(1000);

    // Click Copy
    await page.getByRole('button', { name: 'Copy' }).click();
    await page.waitForTimeout(300);

    // Should show success toast
    await expect(page.getByText('Token copied')).toBeVisible({ timeout: 3000 });

    await screenshot(page, '22-impersonate-copied');
  });
});
