import { test, expect } from '@playwright/test';
import {
  LOGIN_UI,
  screenshot,
  waitForLoad,
  assertNoErrors,
  waitForApiReady,
} from './helpers';

test.describe.serial('Verify Email Page', () => {
  test.beforeAll(async () => {
    await waitForApiReady();
  });

  test('should show error for invalid token', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/verify-email?token=invalid-token-xyz`);
    await waitForLoad(page);

    await expect(page.locator('h1')).toContainText('Email Verification');

    // Wait for verification to complete (auto-verifies on mount)
    await page.waitForLoadState('networkidle');

    // Should show an error state (API returns error for invalid token)
    const errorIcon = page.locator('svg path[d*="M6 18L18 6"]');
    const errorAlert = page.locator('[role="alert"]');
    await expect(errorIcon.or(errorAlert).first()).toBeVisible({ timeout: 10000 });

    await screenshot(page, 'verify-01-invalid-token');
  });

  test('should show error when no token is provided', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/verify-email`);
    await waitForLoad(page);

    await expect(page.locator('h1')).toContainText('Email Verification');

    // Without token, should show error
    await expect(page.getByText(/verification failed|error/i)).toBeVisible({ timeout: 5000 });

    await screenshot(page, 'verify-02-no-token');
  });

  test('should show resend verification button on error', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/verify-email?token=bad-token`);
    await waitForLoad(page);
    await page.waitForLoadState('networkidle');

    // Wait for error state to render
    await page.waitForTimeout(2000);

    await expect(page.getByRole('button', { name: /resend/i })).toBeVisible({ timeout: 5000 });

    await screenshot(page, 'verify-03-resend-button');
  });

  test('should have back to sign in link', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/verify-email?token=bad-token`);
    await waitForLoad(page);
    await page.waitForLoadState('networkidle');

    // Wait for error state to render
    await page.waitForTimeout(2000);

    const signInLink = page.getByRole('link', { name: /sign in/i });
    await expect(signInLink).toBeVisible({ timeout: 5000 });
    await expect(signInLink).toHaveAttribute('href', '/login');

    await screenshot(page, 'verify-04-back-link');
  });
});
