import { test, expect } from '@playwright/test';
import {
  LOGIN_UI,
  TEST_PASSWORD,
  screenshot,
  waitForLoad,
  assertNoErrors,
  waitForApiReady,
} from './helpers';

test.describe.serial('Reset Password Page', () => {
  test.beforeAll(async () => {
    await waitForApiReady();
  });

  test('should display reset password page with token', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/reset-password?token=test-token-123`);
    await waitForLoad(page);
    await assertNoErrors(page);

    await expect(page.locator('h1')).toContainText('Set new password');
    await expect(page.locator('#password')).toBeVisible();
    await expect(page.locator('#confirmPassword')).toBeVisible();
    await expect(page.getByRole('button', { name: /reset password/i })).toBeVisible();

    await screenshot(page, 'reset-01-page-with-token');
  });

  test('should show error when no token is provided', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/reset-password`);
    await waitForLoad(page);

    // Should show "Invalid or expired reset token" error
    await expect(page.getByText(/invalid or expired/i)).toBeVisible({ timeout: 5000 });

    // Form should not be visible without token
    await expect(page.locator('#password')).not.toBeVisible();

    await screenshot(page, 'reset-02-no-token');
  });

  test('should show validation for empty password fields', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/reset-password?token=test-token`);
    await waitForLoad(page);

    await page.getByRole('button', { name: /reset password/i }).click();

    const alerts = page.locator('[role="alert"]');
    await expect(alerts.first()).toBeVisible();
    expect(await alerts.count()).toBeGreaterThanOrEqual(2);

    await screenshot(page, 'reset-03-empty-fields');
  });

  test('should show error for short password', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/reset-password?token=test-token`);
    await waitForLoad(page);

    await page.locator('#password').fill('short');
    await page.locator('#confirmPassword').fill('short');
    await page.getByRole('button', { name: /reset password/i }).click();

    await expect(page.getByText('Password must be at least 8 characters')).toBeVisible();

    await screenshot(page, 'reset-04-short-password');
  });

  test('should show error for password mismatch', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/reset-password?token=test-token`);
    await waitForLoad(page);

    await page.locator('#password').fill(TEST_PASSWORD);
    await page.locator('#confirmPassword').fill('DifferentPassword1!');
    await page.getByRole('button', { name: /reset password/i }).click();

    await expect(page.getByText('Passwords do not match')).toBeVisible();

    await screenshot(page, 'reset-05-password-mismatch');
  });

  test('should show error for invalid token on submit', async ({ page }) => {
    // Mock reset-password API to return token error
    await page.route('**/api/v1/auth/reset-password', (route) => {
      route.fulfill({
        status: 400,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'invalid_token', error_description: 'Invalid or expired reset token' }),
      });
    });

    await page.goto(`${LOGIN_UI}/reset-password?token=invalid-token-xyz`);
    await waitForLoad(page);

    await page.locator('#password').fill(TEST_PASSWORD);
    await page.locator('#confirmPassword').fill(TEST_PASSWORD);

    await page.getByRole('button', { name: /reset password/i }).click();

    await expect(page.getByText(/invalid or expired/i)).toBeVisible({ timeout: 5000 });

    await screenshot(page, 'reset-06-invalid-token');
  });

  test('should show password strength meter', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/reset-password?token=test-token`);
    await waitForLoad(page);

    await page.locator('#password').fill('abcdefgh');
    await expect(page.getByText('Weak')).toBeVisible();

    await page.locator('#password').fill('');
    await page.locator('#password').fill('Abcd1234!@xyz');
    await expect(page.getByText('Strong')).toBeVisible();

    await screenshot(page, 'reset-07-strength-meter');
  });
});
