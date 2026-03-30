import { test, expect } from '@playwright/test';
import {
  LOGIN_UI,
  TEST_PASSWORD,
  screenshot,
  waitForLoad,
  assertNoErrors,
  waitForApiReady,
  registerTestUser,
  deleteTestUserByEmail,
  uniqueEmail,
} from './helpers';

const TEST_EMAIL = uniqueEmail('login');

test.describe.serial('Login Page', () => {
  test.beforeAll(async () => {
    await waitForApiReady();
    await registerTestUser(TEST_EMAIL);
  });

  test.afterAll(async () => {
    await deleteTestUserByEmail(TEST_EMAIL);
  });

  test('should display login page with all elements', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/login`);
    await waitForLoad(page);
    await assertNoErrors(page);

    await expect(page.locator('h1')).toContainText('Sign in to your account');
    await expect(page.locator('#email')).toBeVisible();
    await expect(page.locator('#password')).toBeVisible();
    await expect(page.getByText('Remember me')).toBeVisible();
    await expect(page.getByText('Forgot password?')).toBeVisible();
    await expect(page.getByRole('button', { name: /sign in/i })).toBeVisible();
    await expect(page.getByRole('link', { name: /sign up/i })).toBeVisible();

    await screenshot(page, 'login-01-page-display');
  });

  test('should show validation errors for empty fields', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/login`);
    await waitForLoad(page);

    await page.getByRole('button', { name: /sign in/i }).click();

    // Both email and password should show "This field is required"
    const alerts = page.locator('[role="alert"]');
    await expect(alerts.first()).toBeVisible();
    expect(await alerts.count()).toBeGreaterThanOrEqual(2);

    await screenshot(page, 'login-02-empty-validation');
  });

  test('should show error for invalid email format', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/login`);
    await waitForLoad(page);

    await page.locator('#email').fill('not-an-email');
    await page.locator('#password').fill(TEST_PASSWORD);
    await page.getByRole('button', { name: /sign in/i }).click();

    await expect(page.getByText('Please enter a valid email address')).toBeVisible();

    await screenshot(page, 'login-03-invalid-email');
  });

  test('should show error for wrong credentials', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/login`);
    await waitForLoad(page);

    await page.locator('#email').fill('wrong@example.com');
    await page.locator('#password').fill(TEST_PASSWORD);

    const responsePromise = page.waitForResponse(
      (resp) => resp.url().includes('/api/v1/auth/login') && resp.request().method() === 'POST'
    );

    await page.getByRole('button', { name: /sign in/i }).click();

    const response = await responsePromise;
    expect(response.status()).toBeGreaterThanOrEqual(400);

    // Error alert should appear
    await expect(page.locator('[data-alert="error"], [role="alert"]').first()).toBeVisible({ timeout: 5000 });

    await screenshot(page, 'login-04-wrong-credentials');
  });

  test('should login successfully with valid credentials', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/login`);
    await waitForLoad(page);

    await page.locator('#email').fill(TEST_EMAIL);
    await page.locator('#password').fill(TEST_PASSWORD);

    const responsePromise = page.waitForResponse(
      (resp) => resp.url().includes('/api/v1/auth/login') && resp.request().method() === 'POST'
    );

    await page.getByRole('button', { name: /sign in/i }).click();

    const response = await responsePromise;
    expect(response.status()).toBe(200);

    const body = await response.json();
    expect(body.access_token).toBeTruthy();
    expect(body.token_type).toBe('Bearer');

    await screenshot(page, 'login-05-success');
  });

  test('should include remember_me in request body when checked', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/login`);
    await waitForLoad(page);

    await page.locator('#email').fill(TEST_EMAIL);
    await page.locator('#password').fill(TEST_PASSWORD);
    await page.getByText('Remember me').click();

    const requestPromise = page.waitForRequest(
      (req) => req.url().includes('/api/v1/auth/login') && req.method() === 'POST'
    );

    await page.getByRole('button', { name: /sign in/i }).click();

    const request = await requestPromise;
    const postData = JSON.parse(request.postData() || '{}');
    expect(postData.remember_me).toBe(true);

    await screenshot(page, 'login-06-remember-me');
  });

  test('should toggle password visibility', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/login`);
    await waitForLoad(page);

    await page.locator('#password').fill('mySecret');

    await expect(page.locator('#password')).toHaveAttribute('type', 'password');

    await page.getByRole('button', { name: /show password/i }).click();
    await expect(page.locator('#password')).toHaveAttribute('type', 'text');

    await page.getByRole('button', { name: /hide password/i }).click();
    await expect(page.locator('#password')).toHaveAttribute('type', 'password');

    await screenshot(page, 'login-07-password-toggle');
  });

  test('should pre-fill email from login_hint query parameter', async ({ page }) => {
    const hint = 'prefilled@example.com';
    await page.goto(`${LOGIN_UI}/login?login_hint=${encodeURIComponent(hint)}`);
    await waitForLoad(page);

    await expect(page.locator('#email')).toHaveValue(hint);

    await screenshot(page, 'login-08-login-hint');
  });

  test('should show error for short password', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/login`);
    await waitForLoad(page);

    await page.locator('#email').fill(TEST_EMAIL);
    await page.locator('#password').fill('short');
    await page.getByRole('button', { name: /sign in/i }).click();

    await expect(page.getByText('Password must be at least 8 characters')).toBeVisible();

    await screenshot(page, 'login-09-short-password');
  });
});
