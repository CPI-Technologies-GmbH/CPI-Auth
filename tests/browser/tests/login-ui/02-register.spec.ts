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

const EMAILS_TO_CLEANUP: string[] = [];

test.describe.serial('Register Page', () => {
  test.beforeAll(async () => {
    await waitForApiReady();
  });

  test.afterAll(async () => {
    for (const email of EMAILS_TO_CLEANUP) {
      await deleteTestUserByEmail(email);
    }
  });

  test('should display register page with all elements', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/register`);
    await waitForLoad(page);
    await assertNoErrors(page);

    await expect(page.locator('h1')).toContainText('Create your account');
    await expect(page.locator('#name')).toBeVisible();
    await expect(page.locator('#email')).toBeVisible();
    await expect(page.locator('#password')).toBeVisible();
    await expect(page.locator('#confirmPassword')).toBeVisible();
    await expect(page.getByText('Terms of Service')).toBeVisible();
    await expect(page.getByRole('button', { name: /create account/i })).toBeVisible();

    await screenshot(page, 'register-01-page-display');
  });

  test('should show validation errors for empty fields', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/register`);
    await waitForLoad(page);

    await page.getByRole('button', { name: /create account/i }).click();

    // name, email, password, confirm password, terms → at least 4 error alerts
    const alerts = page.locator('[role="alert"]');
    await expect(alerts.first()).toBeVisible();
    expect(await alerts.count()).toBeGreaterThanOrEqual(4);

    await screenshot(page, 'register-02-empty-validation');
  });

  test('should show error for invalid email format', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/register`);
    await waitForLoad(page);

    await page.locator('#name').fill('Test User');
    await page.locator('#email').fill('bad-email');
    await page.locator('#password').fill(TEST_PASSWORD);
    await page.locator('#confirmPassword').fill(TEST_PASSWORD);
    await page.locator('input[type="checkbox"]').first().check();
    await page.getByRole('button', { name: /create account/i }).click();

    await expect(page.getByText('Please enter a valid email address')).toBeVisible();

    await screenshot(page, 'register-03-invalid-email');
  });

  test('should show error for weak password', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/register`);
    await waitForLoad(page);

    await page.locator('#name').fill('Test User');
    await page.locator('#email').fill('test@example.com');
    await page.locator('#password').fill('short');
    await page.locator('#confirmPassword').fill('short');
    await page.locator('input[type="checkbox"]').first().check();
    await page.getByRole('button', { name: /create account/i }).click();

    await expect(page.getByText('Password must be at least 8 characters')).toBeVisible();

    await screenshot(page, 'register-04-weak-password');
  });

  test('should show error for password mismatch', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/register`);
    await waitForLoad(page);

    await page.locator('#name').fill('Test User');
    await page.locator('#email').fill('test@example.com');
    await page.locator('#password').fill(TEST_PASSWORD);
    await page.locator('#confirmPassword').fill('DifferentPassword1!');
    await page.locator('input[type="checkbox"]').first().check();
    await page.getByRole('button', { name: /create account/i }).click();

    await expect(page.getByText('Passwords do not match')).toBeVisible();

    await screenshot(page, 'register-05-password-mismatch');
  });

  test('should show error when terms not accepted', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/register`);
    await waitForLoad(page);

    await page.locator('#name').fill('Test User');
    await page.locator('#email').fill('test@example.com');
    await page.locator('#password').fill(TEST_PASSWORD);
    await page.locator('#confirmPassword').fill(TEST_PASSWORD);
    // Don't check terms
    await page.getByRole('button', { name: /create account/i }).click();

    await expect(page.getByText('You must agree to the terms of service')).toBeVisible();

    await screenshot(page, 'register-06-terms-not-accepted');
  });

  test('should show password strength meter progression', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/register`);
    await waitForLoad(page);

    // Type a weak password
    await page.locator('#password').fill('abcdefgh');
    await expect(page.getByText('Weak')).toBeVisible();

    // Type a strong password
    await page.locator('#password').fill('');
    await page.locator('#password').fill('Abcd1234!@xyz');
    await expect(page.getByText('Strong')).toBeVisible();

    await screenshot(page, 'register-07-strength-meter');
  });

  test('should register successfully and redirect to login', async ({ page }) => {
    const email = uniqueEmail('register');
    EMAILS_TO_CLEANUP.push(email);

    await page.goto(`${LOGIN_UI}/register`);
    await waitForLoad(page);

    await page.locator('#name').fill('E2E Register User');
    await page.locator('#email').fill(email);
    await page.locator('#password').fill(TEST_PASSWORD);
    await page.locator('#confirmPassword').fill(TEST_PASSWORD);
    await page.locator('input[type="checkbox"]').first().check();

    const responsePromise = page.waitForResponse(
      (resp) => resp.url().includes('/api/v1/auth/register') && resp.request().method() === 'POST'
    );

    await page.getByRole('button', { name: /create account/i }).click();

    const response = await responsePromise;
    expect(response.status()).toBe(200);

    await page.waitForURL(/\/login/, { timeout: 10000 });

    await screenshot(page, 'register-08-success-redirect');
  });

  test('should show error for duplicate email', async ({ page }) => {
    const email = uniqueEmail('dup');
    EMAILS_TO_CLEANUP.push(email);

    // Register user first via API
    await registerTestUser(email);

    await page.goto(`${LOGIN_UI}/register`);
    await waitForLoad(page);

    await page.locator('#name').fill('Duplicate User');
    await page.locator('#email').fill(email);
    await page.locator('#password').fill(TEST_PASSWORD);
    await page.locator('#confirmPassword').fill(TEST_PASSWORD);
    await page.locator('input[type="checkbox"]').first().check();

    const responsePromise = page.waitForResponse(
      (resp) => resp.url().includes('/api/v1/auth/register') && resp.request().method() === 'POST'
    );

    await page.getByRole('button', { name: /create account/i }).click();

    const response = await responsePromise;
    expect(response.status()).toBeGreaterThanOrEqual(400);

    // Error alert from API
    await expect(page.locator('[data-alert="error"], [role="alert"]').first()).toBeVisible({ timeout: 5000 });

    await screenshot(page, 'register-09-duplicate-email');
  });
});
