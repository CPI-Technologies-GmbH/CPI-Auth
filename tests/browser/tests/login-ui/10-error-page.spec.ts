import { test, expect } from '@playwright/test';
import {
  LOGIN_UI,
  screenshot,
  waitForLoad,
  assertNoErrors,
  waitForApiReady,
} from './helpers';

test.describe.serial('Error Page', () => {
  test.beforeAll(async () => {
    await waitForApiReady();
  });

  test('should display error with code and description', async ({ page }) => {
    await page.goto(
      `${LOGIN_UI}/error?error=access_denied&error_description=${encodeURIComponent('The resource owner denied the request')}`
    );
    await waitForLoad(page);

    await expect(page.locator('h1')).toContainText('Something went wrong');
    await expect(page.getByText('The resource owner denied the request')).toBeVisible();
    await expect(page.getByText('access_denied')).toBeVisible();

    await screenshot(page, 'error-01-code-and-description');
  });

  test('should show generic description when only code is present', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/error?error=server_error`);
    await waitForLoad(page);

    await expect(page.locator('h1')).toContainText('Something went wrong');
    await expect(page.getByText('server_error')).toBeVisible();
    // Should show generic error message
    await expect(page.getByText(/unexpected error/i)).toBeVisible();

    await screenshot(page, 'error-02-code-only');
  });

  test('should show generic error when no params are present', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/error`);
    await waitForLoad(page);

    await expect(page.locator('h1')).toContainText('Something went wrong');
    await expect(page.getByText(/unexpected error/i)).toBeVisible();

    await screenshot(page, 'error-03-no-params');
  });

  test('should navigate back to login', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/error?error=test_error`);
    await waitForLoad(page);

    const backLink = page.getByRole('link', { name: /back to sign in/i });
    await expect(backLink).toBeVisible();
    await backLink.click();

    await page.waitForURL(/\/login/, { timeout: 5000 });

    await screenshot(page, 'error-04-back-to-login');
  });

  test('should handle various OAuth error types', async ({ page }) => {
    const errorTypes = [
      { code: 'invalid_request', desc: 'The request is missing a required parameter' },
      { code: 'unauthorized_client', desc: 'The client is not authorized' },
      { code: 'invalid_scope', desc: 'The requested scope is invalid' },
      { code: 'login_required', desc: 'The user must log in' },
      { code: 'consent_required', desc: 'User consent is required' },
    ];

    for (const { code, desc } of errorTypes) {
      await page.goto(
        `${LOGIN_UI}/error?error=${code}&error_description=${encodeURIComponent(desc)}`
      );
      await waitForLoad(page);

      await expect(page.locator('h1')).toContainText('Something went wrong');
      await expect(page.getByText(desc)).toBeVisible();
      await expect(page.getByText(code)).toBeVisible();
    }

    await screenshot(page, 'error-05-oauth-types');
  });
});
