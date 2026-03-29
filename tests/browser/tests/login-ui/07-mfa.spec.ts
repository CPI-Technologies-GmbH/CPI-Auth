import { test, expect } from '@playwright/test';
import {
  LOGIN_UI,
  screenshot,
  waitForLoad,
  assertNoErrors,
  waitForApiReady,
} from './helpers';

test.describe.serial('MFA Page', () => {
  test.beforeAll(async () => {
    await waitForApiReady();
  });

  test('should redirect to /login when no mfa_token is present', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/mfa`);
    await page.waitForURL(/\/login/, { timeout: 10000 });

    await screenshot(page, 'mfa-01-redirect-no-token');
  });

  test('should display TOTP verification UI with mfa_token', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/mfa?mfa_token=fake-mfa-token-123`);
    await waitForLoad(page);
    await assertNoErrors(page);

    await expect(page.locator('h1')).toContainText('Two-factor authentication');
    await expect(page.getByText(/enter the 6-digit code/i)).toBeVisible();
    await expect(page.locator('#otp-0')).toBeVisible();
    await expect(page.locator('#otp-5')).toBeVisible();
    await expect(page.getByRole('button', { name: /verify/i })).toBeVisible();

    await screenshot(page, 'mfa-02-totp-display');
  });

  test('should show validation error for empty OTP', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/mfa?mfa_token=fake-mfa-token-123`);
    await waitForLoad(page);

    await page.getByRole('button', { name: /verify/i }).click();

    await expect(page.locator('[role="alert"]').first()).toBeVisible();

    await screenshot(page, 'mfa-03-empty-otp');
  });

  test('should submit OTP with correct payload', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/mfa?mfa_token=fake-mfa-token-123`);
    await waitForLoad(page);

    // Fill OTP digits one by one
    await page.locator('#otp-0').fill('1');
    await page.locator('#otp-1').fill('2');
    await page.locator('#otp-2').fill('3');
    await page.locator('#otp-3').fill('4');
    await page.locator('#otp-4').fill('5');
    await page.locator('#otp-5').fill('6');

    const requestPromise = page.waitForRequest(
      (req) => req.url().includes('/api/v1/auth/mfa/verify') && req.method() === 'POST'
    );

    await page.getByRole('button', { name: /verify/i }).click();

    const request = await requestPromise;
    const postData = JSON.parse(request.postData() || '{}');
    expect(postData.mfa_token).toBe('fake-mfa-token-123');
    expect(postData.code).toBe('123456');
    expect(postData.method).toBe('totp');

    await screenshot(page, 'mfa-04-otp-submit');
  });

  test('should switch to recovery code mode', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/mfa?mfa_token=fake-mfa-token-123`);
    await waitForLoad(page);

    await page.getByRole('button', { name: /use a recovery code/i }).click();

    await expect(page.locator('h1')).toContainText('Enter recovery code');
    await expect(page.locator('#recovery')).toBeVisible();

    await screenshot(page, 'mfa-05-recovery-mode');
  });

  test('should submit recovery code with method=recovery', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/mfa?mfa_token=fake-mfa-token-123`);
    await waitForLoad(page);

    await page.getByRole('button', { name: /use a recovery code/i }).click();
    await page.locator('#recovery').fill('abcd-efgh-ijkl');

    const requestPromise = page.waitForRequest(
      (req) => req.url().includes('/api/v1/auth/mfa/verify') && req.method() === 'POST'
    );

    await page.getByRole('button', { name: /verify/i }).click();

    const request = await requestPromise;
    const postData = JSON.parse(request.postData() || '{}');
    expect(postData.method).toBe('recovery');
    expect(postData.code).toBe('abcd-efgh-ijkl');

    await screenshot(page, 'mfa-06-recovery-submit');
  });

  test('should switch back to TOTP from recovery', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/mfa?mfa_token=fake-mfa-token-123`);
    await waitForLoad(page);

    // Switch to recovery
    await page.getByRole('button', { name: /use a recovery code/i }).click();
    await expect(page.locator('#recovery')).toBeVisible();

    // Switch back to TOTP
    await page.getByRole('button', { name: /use authenticator app/i }).click();
    await expect(page.locator('#otp-0')).toBeVisible();
    await expect(page.locator('h1')).toContainText('Two-factor authentication');

    await screenshot(page, 'mfa-07-switch-back-totp');
  });

  test('should show SMS and email method buttons when available', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/mfa?mfa_token=fake-token&methods=totp,sms,email`);
    await waitForLoad(page);

    await expect(page.getByRole('button', { name: /send code via sms/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /send code via email/i })).toBeVisible();

    await screenshot(page, 'mfa-08-sms-email-buttons');
  });

  test('should handle OTP paste across inputs', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/mfa?mfa_token=fake-mfa-token-123`);
    await waitForLoad(page);

    // Focus first OTP input and paste
    await page.locator('#otp-0').focus();
    await page.locator('#otp-0').evaluate((el) => {
      const clipboardData = new DataTransfer();
      clipboardData.setData('text', '654321');
      const event = new ClipboardEvent('paste', { clipboardData, bubbles: true });
      el.dispatchEvent(event);
    });

    // All digits should be filled
    await expect(page.locator('#otp-0')).toHaveValue('6');
    await expect(page.locator('#otp-1')).toHaveValue('5');
    await expect(page.locator('#otp-2')).toHaveValue('4');
    await expect(page.locator('#otp-3')).toHaveValue('3');
    await expect(page.locator('#otp-4')).toHaveValue('2');
    await expect(page.locator('#otp-5')).toHaveValue('1');

    await screenshot(page, 'mfa-09-otp-paste');
  });
});
