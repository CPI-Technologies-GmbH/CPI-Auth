import { test, expect } from '@playwright/test';
import {
  LOGIN_UI,
  screenshot,
  waitForLoad,
  assertNoErrors,
  waitForApiReady,
} from './helpers';

test.describe.serial('Consent Page', () => {
  test.beforeAll(async () => {
    await waitForApiReady();
  });

  test('should show error when no consent_challenge is provided', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/consent`);
    await waitForLoad(page);

    await expect(page.getByText(/missing consent challenge/i)).toBeVisible({ timeout: 5000 });

    await screenshot(page, 'consent-01-no-challenge');
  });

  test('should show error for invalid consent challenge', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/consent?consent_challenge=invalid-challenge`);
    await waitForLoad(page);

    // API should return error for invalid challenge
    await expect(page.locator('[role="alert"], [data-alert="error"]').first()).toBeVisible({ timeout: 10000 });

    await screenshot(page, 'consent-02-invalid-challenge');
  });

  test('should display full consent screen with scopes', async ({ page }) => {
    // getConsentInfo uses GET /api/v1/auth/consent?consent_challenge=...
    await page.route('**/api/v1/auth/consent?**', (route) => {
      if (route.request().method() === 'GET') {
        route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            client_name: 'Test Application',
            client_uri: 'https://testapp.example.com',
            client_logo: null,
            requested_scopes: [
              { name: 'openid', description: 'Access your identity' },
              { name: 'profile', description: 'Access your profile information' },
              { name: 'email', description: 'Access your email address' },
            ],
          }),
        });
      } else {
        route.continue();
      }
    });

    await page.goto(`${LOGIN_UI}/consent?consent_challenge=valid-challenge`);
    await waitForLoad(page);
    await assertNoErrors(page);

    await expect(page.locator('h1')).toContainText('Authorize');
    await expect(page.getByText('Test Application')).toBeVisible();
    await expect(page.getByText('Access your identity')).toBeVisible();
    await expect(page.getByText('Access your profile information')).toBeVisible();
    await expect(page.getByText('Access your email address')).toBeVisible();
    await expect(page.getByText('Remember this decision')).toBeVisible();
    await expect(page.getByRole('button', { name: /allow/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /deny/i })).toBeVisible();

    await screenshot(page, 'consent-03-full-display');
  });

  test('should send grant=true when allow is clicked', async ({ page }) => {
    await page.route('**/api/v1/auth/consent?**', (route) => {
      if (route.request().method() === 'GET') {
        route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            client_name: 'Test App',
            client_uri: null,
            client_logo: null,
            requested_scopes: [{ name: 'openid', description: 'OpenID' }],
          }),
        });
      } else {
        route.continue();
      }
    });

    let capturedBody: any = null;
    await page.route('**/api/v1/auth/consent', (route) => {
      if (route.request().method() === 'POST') {
        capturedBody = JSON.parse(route.request().postData() || '{}');
        route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ redirect_url: 'https://example.com/callback' }),
        });
      } else {
        route.continue();
      }
    });

    await page.goto(`${LOGIN_UI}/consent?consent_challenge=allow-challenge`);
    await waitForLoad(page);
    await expect(page.getByRole('button', { name: /allow/i })).toBeVisible({ timeout: 5000 });

    await page.getByRole('button', { name: /allow/i }).click();
    await page.waitForTimeout(1000);
    expect(capturedBody).toBeTruthy();
    expect(capturedBody.grant).toBe(true);
    expect(capturedBody.consent_challenge).toBe('allow-challenge');

    await screenshot(page, 'consent-04-allow');
  });

  test('should send grant=false when deny is clicked', async ({ page }) => {
    await page.route('**/api/v1/auth/consent?**', (route) => {
      if (route.request().method() === 'GET') {
        route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            client_name: 'Test App',
            client_uri: null,
            client_logo: null,
            requested_scopes: [{ name: 'openid', description: 'OpenID' }],
          }),
        });
      } else {
        route.continue();
      }
    });

    let capturedBody: any = null;
    await page.route('**/api/v1/auth/consent', (route) => {
      if (route.request().method() === 'POST') {
        capturedBody = JSON.parse(route.request().postData() || '{}');
        route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ redirect_url: 'https://example.com/callback' }),
        });
      } else {
        route.continue();
      }
    });

    await page.goto(`${LOGIN_UI}/consent?consent_challenge=deny-challenge`);
    await waitForLoad(page);
    await expect(page.getByRole('button', { name: /deny/i })).toBeVisible({ timeout: 5000 });

    await page.getByRole('button', { name: /deny/i }).click();

    await page.waitForTimeout(1000);
    expect(capturedBody).toBeTruthy();
    expect(capturedBody.grant).toBe(false);

    await screenshot(page, 'consent-05-deny');
  });

  test('should include remember=true when checkbox is checked', async ({ page }) => {
    await page.route('**/api/v1/auth/consent?**', (route) => {
      if (route.request().method() === 'GET') {
        route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            client_name: 'Test App',
            client_uri: null,
            client_logo: null,
            requested_scopes: [{ name: 'openid', description: 'OpenID' }],
          }),
        });
      } else {
        route.continue();
      }
    });

    let capturedBody: any = null;
    await page.route('**/api/v1/auth/consent', (route) => {
      if (route.request().method() === 'POST') {
        capturedBody = JSON.parse(route.request().postData() || '{}');
        route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ redirect_url: 'https://example.com/callback' }),
        });
      } else {
        route.continue();
      }
    });

    await page.goto(`${LOGIN_UI}/consent?consent_challenge=remember-challenge`);
    await waitForLoad(page);
    await expect(page.getByText('Remember this decision')).toBeVisible({ timeout: 5000 });

    await page.getByText('Remember this decision').click();
    await page.getByRole('button', { name: /allow/i }).click();

    await page.waitForTimeout(1000);
    expect(capturedBody).toBeTruthy();
    expect(capturedBody.remember).toBe(true);

    await screenshot(page, 'consent-06-remember');
  });
});
