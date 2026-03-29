import { test, expect } from '@playwright/test';
import {
  LOGIN_UI,
  screenshot,
  waitForLoad,
  assertNoErrors,
  waitForApiReady,
} from './helpers';

test.describe.serial('Passwordless Page', () => {
  test.beforeAll(async () => {
    await waitForApiReady();
  });

  test('should display passwordless page with method selection', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/passwordless`);
    await waitForLoad(page);
    await assertNoErrors(page);

    await expect(page.locator('h1')).toContainText('Passwordless Sign In');
    await expect(page.locator('#email')).toBeVisible();
    await expect(page.getByText('Code')).toBeVisible();
    await expect(page.getByText('Magic Link')).toBeVisible();
    await expect(page.getByRole('button', { name: /continue/i })).toBeVisible();

    await screenshot(page, 'passwordless-01-page-display');
  });

  test('should show validation error for empty email', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/passwordless`);
    await waitForLoad(page);

    await page.getByRole('button', { name: /continue/i }).click();

    await expect(page.locator('[role="alert"]').first()).toBeVisible();

    await screenshot(page, 'passwordless-02-empty-email');
  });

  test('should pre-fill email from query parameter', async ({ page }) => {
    const email = 'prefilled@example.com';
    await page.goto(`${LOGIN_UI}/passwordless?email=${encodeURIComponent(email)}`);
    await waitForLoad(page);

    await expect(page.locator('#email')).toHaveValue(email);

    await screenshot(page, 'passwordless-03-prefilled-email');
  });

  test('should transition to OTP step when code method is used', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/passwordless`);
    await waitForLoad(page);

    // Select Code method (default)
    await page.locator('#email').fill('test@example.com');

    // Mock the API to avoid real calls
    await page.route('**/api/v1/auth/passwordless/start', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ message: 'Code sent' }),
      });
    });

    await page.getByRole('button', { name: /continue/i }).click();

    // Should transition to OTP step
    await expect(page.getByText(/check your email/i)).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#otp-0')).toBeVisible();
    await expect(page.getByRole('button', { name: /verify/i })).toBeVisible();

    await screenshot(page, 'passwordless-04-otp-step');
  });

  test('should show link sent confirmation for magic link method', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/passwordless`);
    await waitForLoad(page);

    await page.locator('#email').fill('test@example.com');

    // Select Magic Link method
    await page.getByText('Magic Link').click();

    // Mock the API
    await page.route('**/api/v1/auth/passwordless/start', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ message: 'Link sent' }),
      });
    });

    await page.getByRole('button', { name: /continue/i }).click();

    // Should show link sent confirmation
    await expect(page.getByText(/sent a sign-in link/i)).toBeVisible({ timeout: 5000 });

    await screenshot(page, 'passwordless-05-magic-link-sent');
  });

  test('should show resend countdown timer', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/passwordless`);
    await waitForLoad(page);

    await page.locator('#email').fill('test@example.com');

    await page.route('**/api/v1/auth/passwordless/start', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ message: 'Code sent' }),
      });
    });

    await page.getByRole('button', { name: /continue/i }).click();

    // After sending, resend button should show countdown
    await expect(page.getByText(/resend.*\d+s/i)).toBeVisible({ timeout: 5000 });

    await screenshot(page, 'passwordless-06-resend-countdown');
  });
});
