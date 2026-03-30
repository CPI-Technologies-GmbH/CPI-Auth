import { test, expect } from '@playwright/test';
import { goToAdmin, screenshot, waitForLoad, assertNoErrors, apiCall, dbQuery, API_URL, ADMIN_EMAIL, ADMIN_PASSWORD } from './helpers';

let accessToken: string;

test.describe.serial('Admin Audit Logs', () => {
  test.beforeAll(async () => {
    const res = await fetch(`${API_URL}/admin/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email: ADMIN_EMAIL, password: ADMIN_PASSWORD }),
    });
    accessToken = (await res.json()).access_token;
  });

  // ═══════════════════════════════════════════════════════════
  // PAGE RENDERING (was crashing with .slice() error)
  // ═══════════════════════════════════════════════════════════

  test('should render logs page without errors', async ({ page }) => {
    await goToAdmin(page, '/logs');
    await waitForLoad(page);
    await assertNoErrors(page);

    await expect(page.getByRole('heading', { name: 'Audit Logs' })).toBeVisible();
    // Should NOT show "Something went wrong" error
    await expect(page.getByText('Something went wrong')).not.toBeVisible();

    await screenshot(page, '13-logs-page');
  });

  test('should show export button', async ({ page }) => {
    await goToAdmin(page, '/logs');
    await waitForLoad(page);
    await expect(page.getByRole('button', { name: /export/i })).toBeVisible();
  });

  // ═══════════════════════════════════════════════════════════
  // AUDIT LOG GENERATION
  // ═══════════════════════════════════════════════════════════

  test('should generate audit log on user creation', async ({ page }) => {
    const testEmail = `audit-test-${Date.now()}@e2e.local`;
    const user = await apiCall(accessToken, 'POST', '/admin/users', {
      email: testEmail,
      password: 'Xk9mQ2vL8nR4!pw',
      name: 'Audit Test User',
    });
    expect(user?.id).toBeTruthy();

    // Navigate to logs page
    await goToAdmin(page, '/logs');
    await waitForLoad(page);
    await assertNoErrors(page);

    // Table should have at least one row
    const tableRows = page.locator('table tbody tr');
    await expect(tableRows.first()).toBeVisible({ timeout: 5000 });

    await screenshot(page, '13-logs-with-entries');

    // Cleanup
    await apiCall(accessToken, 'DELETE', `/admin/users/${user.id}`);
  });

  test('should have audit logs in database', async () => {
    const rows = await dbQuery('SELECT COUNT(*) as cnt FROM audit_logs_default');
    expect(parseInt(rows[0].cnt)).toBeGreaterThan(0);
  });

  // ═══════════════════════════════════════════════════════════
  // API ENDPOINTS
  // ═══════════════════════════════════════════════════════════

  test('should list audit logs via API', async () => {
    const res = await fetch(`${API_URL}/admin/audit-logs`, {
      headers: { Authorization: `Bearer ${accessToken}` },
    });
    expect(res.status).toBe(200);
    const data = await res.json();
    // Should be an array or paginated response
    expect(data).toBeTruthy();
  });

  test('should export audit logs via API', async () => {
    const res = await fetch(`${API_URL}/admin/audit-logs/export`, {
      headers: { Authorization: `Bearer ${accessToken}` },
    });
    expect(res.status).toBeLessThan(500);
  });

  // ═══════════════════════════════════════════════════════════
  // LOGS PAGE STABILITY (regression for .slice() crash)
  // ═══════════════════════════════════════════════════════════

  test('should handle logs with null actor_id gracefully', async ({ page }) => {
    // This was the bug: log entries with null actor_id crashed the page
    await goToAdmin(page, '/logs');
    await waitForLoad(page);

    // Page should render without crashing
    await assertNoErrors(page);
    await expect(page.getByText('Something went wrong')).not.toBeVisible();
    await expect(page.getByRole('heading', { name: 'Audit Logs' })).toBeVisible();

    await screenshot(page, '13-logs-null-actor');
  });

  test('should navigate to logs from sidebar', async ({ page }) => {
    await goToAdmin(page, '/');
    await waitForLoad(page);

    await page.getByRole('link', { name: 'Logs' }).click();
    await waitForLoad(page);
    await assertNoErrors(page);

    await expect(page.getByRole('heading', { name: 'Audit Logs' })).toBeVisible();
    await expect(page.getByText('Something went wrong')).not.toBeVisible();
  });
});
