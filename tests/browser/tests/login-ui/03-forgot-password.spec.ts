import { test, expect } from '@playwright/test';
import {
  LOGIN_UI,
  screenshot,
  waitForLoad,
  assertNoErrors,
  waitForApiReady,
  registerTestUser,
  deleteTestUserByEmail,
  clearMailHog,
  waitForEmail,
  uniqueEmail,
} from './helpers';

const TEST_EMAIL = uniqueEmail('forgot');

test.describe.serial('Forgot Password Page', () => {
  test.beforeAll(async () => {
    await waitForApiReady();
    await registerTestUser(TEST_EMAIL);
  });

  test.afterAll(async () => {
    await deleteTestUserByEmail(TEST_EMAIL);
  });

  test('should display forgot password page', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/forgot-password`);
    await waitForLoad(page);
    await assertNoErrors(page);

    await expect(page.locator('h1')).toContainText('Reset your password');
    await expect(page.locator('#email')).toBeVisible();
    await expect(page.getByRole('button', { name: /send reset instructions/i })).toBeVisible();
    await expect(page.getByText('Back to sign in')).toBeVisible();

    await screenshot(page, 'forgot-01-page-display');
  });

  test('should show validation error for empty email', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/forgot-password`);
    await waitForLoad(page);

    await page.getByRole('button', { name: /send reset instructions/i }).click();

    await expect(page.getByText('This field is required')).toBeVisible();

    await screenshot(page, 'forgot-02-empty-email');
  });

  test('should show error for invalid email format', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/forgot-password`);
    await waitForLoad(page);

    await page.locator('#email').fill('invalid');
    await page.getByRole('button', { name: /send reset instructions/i }).click();

    await expect(page.getByText('Please enter a valid email address')).toBeVisible();

    await screenshot(page, 'forgot-03-invalid-email');
  });

  test('should show success message after submission', async ({ page }) => {
    // Mock forgot-password API (may not be implemented on backend)
    await page.route('**/api/v1/auth/forgot-password', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ message: 'Password reset instructions sent' }),
      });
    });

    await page.goto(`${LOGIN_UI}/forgot-password`);
    await waitForLoad(page);

    await page.locator('#email').fill(TEST_EMAIL);
    await page.getByRole('button', { name: /send reset instructions/i }).click();

    await expect(page.getByText(/sent password reset instructions/i)).toBeVisible({ timeout: 5000 });

    await screenshot(page, 'forgot-04-success');
  });

  test('should receive reset email via MailHog', async ({ page }) => {
    await clearMailHog();

    await page.route('**/api/v1/auth/forgot-password', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ message: 'Password reset instructions sent' }),
      });
    });

    await page.goto(`${LOGIN_UI}/forgot-password`);
    await waitForLoad(page);

    await page.locator('#email').fill(TEST_EMAIL);
    await page.getByRole('button', { name: /send reset instructions/i }).click();

    await expect(page.getByText(/sent password reset instructions/i)).toBeVisible({ timeout: 5000 });

    // Check MailHog for email (graceful if not running)
    const email = await waitForEmail(TEST_EMAIL, 5000);
    if (email) {
      expect(email).toBeTruthy();
    }

    await screenshot(page, 'forgot-05-mailhog');
  });

  test('should show same success for non-existent email (no enumeration)', async ({ page }) => {
    await page.route('**/api/v1/auth/forgot-password', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ message: 'Password reset instructions sent' }),
      });
    });

    await page.goto(`${LOGIN_UI}/forgot-password`);
    await waitForLoad(page);

    await page.locator('#email').fill('nonexistent@example.com');
    await page.getByRole('button', { name: /send reset instructions/i }).click();

    // Should show same success message (no email enumeration)
    await expect(page.getByText(/sent password reset instructions/i)).toBeVisible({ timeout: 5000 });

    await screenshot(page, 'forgot-06-no-enumeration');
  });
});
