import { test, expect } from '@playwright/test';
import { API_URL, ADMIN_EMAIL, ADMIN_PASSWORD, screenshot, waitForLoad, assertNoErrors } from './helpers';

test.describe.serial('Admin Auth Flow', () => {
  test('should show login page', async ({ page }) => {
    await page.goto('/login');
    await waitForLoad(page);
    await expect(page.locator('input#email')).toBeVisible();
    await expect(page.locator('input#password')).toBeVisible();
    await expect(page.getByRole('button', { name: /sign in/i })).toBeVisible();
    await screenshot(page, 'auth-01-login-page');
  });

  test('should reject invalid credentials', async ({ page }) => {
    await page.goto('/login');
    await waitForLoad(page);
    await page.locator('input#email').fill('wrong@example.com');
    await page.locator('input#password').fill('wrongpassword');
    await page.getByRole('button', { name: /sign in/i }).click();
    // Should show error or stay on login page
    await page.waitForTimeout(2000);
    await expect(page).toHaveURL(/\/login/);
    await screenshot(page, 'auth-02-invalid-login');
  });

  test('should login with admin credentials', async ({ page }) => {
    await page.goto('/login');
    await waitForLoad(page);
    await page.locator('input#email').fill(ADMIN_EMAIL);
    await page.locator('input#password').fill(ADMIN_PASSWORD);
    const responsePromise = page.waitForResponse(
      (resp) => resp.url().includes('/admin/auth/login') && resp.request().method() === 'POST'
    );
    await page.getByRole('button', { name: /sign in/i }).click();
    const response = await responsePromise;
    expect(response.status()).toBe(200);
    const body = await response.json();
    expect(body.access_token).toBeTruthy();
    await page.waitForURL('/', { timeout: 10000 });
    await waitForLoad(page);
    await assertNoErrors(page);
    await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible();
    await screenshot(page, 'auth-03-dashboard-after-login');
  });

  test('should fetch current user via /admin/auth/me', async () => {
    const loginRes = await fetch(`${API_URL}/admin/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email: ADMIN_EMAIL, password: ADMIN_PASSWORD }),
    });
    const tokens = await loginRes.json();
    const meRes = await fetch(`${API_URL}/admin/auth/me`, {
      headers: { Authorization: `Bearer ${tokens.access_token}` },
    });
    expect(meRes.status).toBe(200);
    const me = await meRes.json();
    expect(me.email).toBe(ADMIN_EMAIL);
    expect(me.role).toBe('super_admin');
  });
});
