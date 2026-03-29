import { test, expect } from '@playwright/test';
import { goToAdmin, screenshot, waitForLoad, assertNoErrors, apiCall, API_URL, ADMIN_EMAIL, ADMIN_PASSWORD } from './helpers';

let accessToken: string;

test.describe.serial('Custom Domains with DNS Verification', () => {
  let verificationId: string;
  const testDomain = `e2e-test-${Date.now()}.example.com`;

  test.beforeAll(async () => {
    const res = await fetch(`${API_URL}/admin/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email: ADMIN_EMAIL, password: ADMIN_PASSWORD }),
    });
    accessToken = (await res.json()).access_token;
  });

  test.afterAll(async () => {
    if (verificationId) {
      await apiCall(accessToken, 'DELETE', `/admin/domains/verification/${verificationId}`).catch(() => {});
    }
  });

  // --- API Tests ---

  test('should return empty state when no domain configured', async () => {
    const result = await apiCall(accessToken, 'GET', '/admin/domains/verification');
    // Backend returns either {"status":"none"} or a not_found error when no domain is configured
    expect(result.status === 'none' || result.error === 'not_found').toBe(true);
  });

  test('should initiate domain verification', async () => {
    const result = await apiCall(accessToken, 'POST', '/admin/domains/verification', {
      domain: testDomain,
    });
    expect(result.id).toBeTruthy();
    expect(result.domain).toBe(testDomain);
    expect(result.is_verified).toBe(false);
    expect(result.verification_method).toBe('TXT');
    expect(result.dns_record).toBeDefined();
    expect(result.dns_record.record_type).toBe('TXT');
    expect(result.dns_record.host).toContain('_cpi-auth-verification');
    expect(result.dns_record.value).toContain('cpi-auth-verify=');
    verificationId = result.id;
  });

  test('should get existing verification status', async () => {
    const result = await apiCall(accessToken, 'GET', '/admin/domains/verification');
    expect(result.status).toBe('pending');
    expect(result.domain).toBe(testDomain);
    expect(result.is_verified).toBe(false);
  });

  test('should return same verification for duplicate domain initiation', async () => {
    const result = await apiCall(accessToken, 'POST', '/admin/domains/verification', {
      domain: testDomain,
    });
    expect(result.id).toBe(verificationId); // Same record
  });

  test('should check DNS and report pending (no real DNS record)', async () => {
    const result = await apiCall(accessToken, 'POST', `/admin/domains/verification/${verificationId}/check`);
    expect(result.is_verified).toBe(false);
    expect(result.status).toBe('pending');
  });

  test('should delete domain verification', async () => {
    const result = await apiCall(accessToken, 'DELETE', `/admin/domains/verification/${verificationId}`);
    expect(result).toBeNull(); // 204

    // Verify it's gone
    const check = await apiCall(accessToken, 'GET', '/admin/domains/verification');
    // Backend returns either {"status":"none"} or a not_found error when no domain is configured
    expect(check.status === 'none' || check.error === 'not_found').toBe(true);
    verificationId = ''; // Prevent afterAll cleanup
  });

  // --- UI Tests ---

  test('should show custom domain tab on settings page', async ({ page }) => {
    await goToAdmin(page, '/settings');
    await waitForLoad(page);
    await screenshot(page, '24-settings-page');
    await assertNoErrors(page);

    // Click the Custom Domain tab
    const domainTab = page.locator('button:has-text("Custom Domain")');
    await expect(domainTab).toBeVisible({ timeout: 5000 });
    await domainTab.click();
    await page.waitForTimeout(500);
    await screenshot(page, '24-custom-domain-tab-empty');
  });

  test('should initiate verification from UI', async ({ page }) => {
    await goToAdmin(page, '/settings');
    await waitForLoad(page);

    // Click Custom Domain tab
    const domainTab = page.locator('button:has-text("Custom Domain")');
    await domainTab.click();
    await page.waitForTimeout(500);

    // Enter domain
    const input = page.locator('input[placeholder="auth.yourcompany.com"]');
    await input.fill('ui-test.example.com');
    await screenshot(page, '24-custom-domain-entered');

    // Click verify
    const verifyBtn = page.locator('button:has-text("Verify")');
    await verifyBtn.click();
    await page.waitForTimeout(1000);
    await screenshot(page, '24-custom-domain-pending');
    await assertNoErrors(page);

    // Clean up via API
    const dv = await apiCall(accessToken, 'GET', '/admin/domains/verification');
    if (dv.id) {
      await apiCall(accessToken, 'DELETE', `/admin/domains/verification/${dv.id}`).catch(() => {});
    }
  });
});
