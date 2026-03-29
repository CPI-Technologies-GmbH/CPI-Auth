import { test, expect } from '@playwright/test';
import {
  LOGIN_UI,
  screenshot,
  waitForLoad,
  assertNoErrors,
  waitForApiReady,
} from './helpers';

test.describe.serial('MFA Enrollment Page', () => {
  test.beforeAll(async () => {
    await waitForApiReady();
  });

  test('should redirect to /login when no mfa_token is present', async ({ page }) => {
    await page.goto(`${LOGIN_UI}/mfa/enroll`);
    await page.waitForURL(/\/login/, { timeout: 10000 });

    await screenshot(page, 'mfa-enroll-01-redirect-no-token');
  });

  test('should show loading state initially', async ({ page }) => {
    // Delay the API response to observe loading state
    await page.route('**/api/v1/auth/mfa/enroll', (route) => {
      setTimeout(() => {
        route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            secret: 'JBSWY3DPEHPK3PXP',
            qr_code: 'data:image/png;base64,iVBOR',
            otpauth_url: 'otpauth://totp/CPI Auth:test?secret=JBSWY3DPEHPK3PXP',
            recovery_codes: ['code1', 'code2', 'code3', 'code4'],
          }),
        });
      }, 500);
    });

    await page.goto(`${LOGIN_UI}/mfa/enroll?mfa_token=fake-enroll-token`);

    // Should see loading spinner initially
    await expect(page.locator('[role="status"], .animate-spin').first()).toBeVisible({ timeout: 3000 });

    await screenshot(page, 'mfa-enroll-02-loading');
  });

  test('should display QR code and manual entry toggle', async ({ page }) => {
    await page.route('**/api/v1/auth/mfa/enroll', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          secret: 'JBSWY3DPEHPK3PXP',
          qr_code: 'data:image/png;base64,iVBOR',
          otpauth_url: 'otpauth://totp/CPI Auth:test?secret=JBSWY3DPEHPK3PXP',
          recovery_codes: ['aaaa-bbbb', 'cccc-dddd', 'eeee-ffff', 'gggg-hhhh'],
        }),
      });
    });

    await page.goto(`${LOGIN_UI}/mfa/enroll?mfa_token=fake-enroll-token`);
    await waitForLoad(page);

    await expect(page.locator('h1')).toContainText('Set up two-factor authentication');
    await expect(page.locator('img[alt*="QR code"]')).toBeVisible({ timeout: 5000 });

    // Click manual entry toggle
    await page.getByText(/can't scan/i).click();
    await expect(page.getByText('JBSWY3DPEHPK3PXP')).toBeVisible();

    // OTP verification input should be present
    await expect(page.locator('#otp-0')).toBeVisible();
    await expect(page.getByRole('button', { name: /verify and enable/i })).toBeVisible();

    await screenshot(page, 'mfa-enroll-03-qr-manual');
  });

  test('should show validation error for empty OTP', async ({ page }) => {
    await page.route('**/api/v1/auth/mfa/enroll', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          secret: 'JBSWY3DPEHPK3PXP',
          qr_code: 'data:image/png;base64,iVBOR',
          otpauth_url: 'otpauth://totp/CPI Auth:test?secret=JBSWY3DPEHPK3PXP',
          recovery_codes: ['aaaa-bbbb', 'cccc-dddd'],
        }),
      });
    });

    await page.goto(`${LOGIN_UI}/mfa/enroll?mfa_token=fake-enroll-token`);
    await waitForLoad(page);
    await expect(page.locator('#otp-0')).toBeVisible({ timeout: 5000 });

    await page.getByRole('button', { name: /verify and enable/i }).click();

    await expect(page.locator('[role="alert"]').first()).toBeVisible();

    await screenshot(page, 'mfa-enroll-04-empty-otp');
  });

  test('should display recovery codes after verification', async ({ page }) => {
    const recoveryCodes = ['aaaa-1111', 'bbbb-2222', 'cccc-3333', 'dddd-4444'];

    await page.route('**/api/v1/auth/mfa/enroll', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          secret: 'JBSWY3DPEHPK3PXP',
          qr_code: 'data:image/png;base64,iVBOR',
          otpauth_url: 'otpauth://totp/CPI Auth:test?secret=JBSWY3DPEHPK3PXP',
          recovery_codes: recoveryCodes,
        }),
      });
    });

    // mfaEnrollVerify actually calls /api/v1/auth/mfa/verify (same as mfaVerify)
    await page.route('**/api/v1/auth/mfa/verify', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ success: true }),
      });
    });

    await page.goto(`${LOGIN_UI}/mfa/enroll?mfa_token=fake-enroll-token`);
    await waitForLoad(page);
    await expect(page.locator('#otp-0')).toBeVisible({ timeout: 5000 });

    // Fill valid OTP
    await page.locator('#otp-0').fill('1');
    await page.locator('#otp-1').fill('2');
    await page.locator('#otp-2').fill('3');
    await page.locator('#otp-3').fill('4');
    await page.locator('#otp-4').fill('5');
    await page.locator('#otp-5').fill('6');

    await page.getByRole('button', { name: /verify and enable/i }).click();

    // Should show recovery codes
    await expect(page.getByText('Save your recovery codes')).toBeVisible({ timeout: 5000 });
    for (const code of recoveryCodes) {
      await expect(page.getByText(code)).toBeVisible();
    }

    await screenshot(page, 'mfa-enroll-05-recovery-codes');
  });

  test('should navigate to /login when done button is clicked', async ({ page }) => {
    const recoveryCodes = ['aaaa-1111', 'bbbb-2222', 'cccc-3333', 'dddd-4444'];

    await page.route('**/api/v1/auth/mfa/enroll', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          secret: 'JBSWY3DPEHPK3PXP',
          qr_code: 'data:image/png;base64,iVBOR',
          otpauth_url: 'otpauth://totp/CPI Auth:test?secret=JBSWY3DPEHPK3PXP',
          recovery_codes: recoveryCodes,
        }),
      });
    });

    await page.route('**/api/v1/auth/mfa/verify', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ success: true }),
      });
    });

    await page.goto(`${LOGIN_UI}/mfa/enroll?mfa_token=fake-enroll-token`);
    await waitForLoad(page);
    await expect(page.locator('#otp-0')).toBeVisible({ timeout: 5000 });

    // Fill and verify OTP
    await page.locator('#otp-0').fill('1');
    await page.locator('#otp-1').fill('2');
    await page.locator('#otp-2').fill('3');
    await page.locator('#otp-3').fill('4');
    await page.locator('#otp-4').fill('5');
    await page.locator('#otp-5').fill('6');
    await page.getByRole('button', { name: /verify and enable/i }).click();

    await expect(page.getByText('Save your recovery codes')).toBeVisible({ timeout: 5000 });

    // Click Done
    await page.getByRole('button', { name: /done/i }).click();
    await page.waitForURL(/\/login/, { timeout: 10000 });

    await screenshot(page, 'mfa-enroll-06-done-redirect');
  });
});
