import { test, expect } from '@playwright/test';
import {
  LOGIN_UI,
  screenshot,
  waitForLoad,
  assertNoErrors,
  waitForApiReady,
} from './helpers';

test.describe.serial('Navigation Between Pages', () => {
  test.beforeAll(async () => {
    await waitForApiReady();
  });

  test('should navigate from login to register via Sign up link', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/login`);
    await waitForLoad(page);

    await page.getByRole('link', { name: /sign up/i }).click();
    await page.waitForURL(/\/register/, { timeout: 5000 });

    await expect(page.locator('h1')).toContainText('Create your account');

    await screenshot(page, 'nav-01-login-to-register');
  });

  test('should navigate from register to login via Sign in link', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/register`);
    await waitForLoad(page);

    await page.getByRole('link', { name: /sign in/i }).click();
    await page.waitForURL(/\/login/, { timeout: 5000 });

    await expect(page.locator('h1')).toContainText('Sign in to your account');

    await screenshot(page, 'nav-02-register-to-login');
  });

  test('should navigate from login to forgot password', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/login`);
    await waitForLoad(page);

    await page.getByText('Forgot password?').click();
    await page.waitForURL(/\/forgot-password/, { timeout: 5000 });

    await expect(page.locator('h1')).toContainText('Reset your password');

    await screenshot(page, 'nav-03-login-to-forgot');
  });

  test('should navigate from forgot password to login via Back to sign in', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/forgot-password`);
    await waitForLoad(page);

    await page.getByRole('link', { name: /back to sign in/i }).click();
    await page.waitForURL(/\/login/, { timeout: 5000 });

    await expect(page.locator('h1')).toContainText('Sign in to your account');

    await screenshot(page, 'nav-04-forgot-to-login');
  });

  test('should navigate from reset password to login', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/reset-password?token=test`);
    await waitForLoad(page);

    await page.getByRole('link', { name: /back to sign in/i }).click();
    await page.waitForURL(/\/login/, { timeout: 5000 });

    await screenshot(page, 'nav-05-reset-to-login');
  });

  test('should navigate from verify email to login', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/verify-email`);
    await waitForLoad(page);

    // Wait for error state (no token)
    await page.waitForTimeout(2000);

    const signInLink = page.getByRole('link', { name: /sign in/i });
    await expect(signInLink).toBeVisible({ timeout: 5000 });
    await signInLink.click();

    await page.waitForURL(/\/login/, { timeout: 5000 });

    await screenshot(page, 'nav-06-verify-to-login');
  });

  test('should navigate from passwordless to login', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/passwordless`);
    await waitForLoad(page);

    await page.getByRole('link', { name: /back to sign in/i }).click();
    await page.waitForURL(/\/login/, { timeout: 5000 });

    await screenshot(page, 'nav-07-passwordless-to-login');
  });

  test('should navigate from error page to login', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/error?error=test`);
    await waitForLoad(page);

    await page.getByRole('link', { name: /back to sign in/i }).click();
    await page.waitForURL(/\/login/, { timeout: 5000 });

    await screenshot(page, 'nav-08-error-to-login');
  });

  test('should redirect root to /login', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/`);
    await page.waitForURL(/\/login/, { timeout: 10000 });

    await expect(page.locator('h1')).toContainText('Sign in to your account');

    await screenshot(page, 'nav-09-root-redirect');
  });

  test('should preserve OAuth params across navigation', async ({ page }) => {
    const oauthParams = 'client_id=test-client&redirect_uri=https://example.com/cb&scope=openid&state=abc123';

    await page.goto(`${LOGIN_UI}/login?${oauthParams}`);
    await waitForLoad(page);

    // Navigate to register
    await page.getByRole('link', { name: /sign up/i }).click();
    await page.waitForURL(/\/register/, { timeout: 5000 });

    // Check OAuth params are preserved in the URL
    const registerUrl = page.url();
    expect(registerUrl).toContain('client_id=test-client');
    expect(registerUrl).toContain('redirect_uri');

    // Navigate back to login
    await page.getByRole('link', { name: /sign in/i }).click();
    await page.waitForURL(/\/login/, { timeout: 5000 });

    const loginUrl = page.url();
    expect(loginUrl).toContain('client_id=test-client');

    await screenshot(page, 'nav-10-oauth-preserved');
  });
});
