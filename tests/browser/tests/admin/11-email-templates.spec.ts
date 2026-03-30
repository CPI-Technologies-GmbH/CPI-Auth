import { test, expect } from '@playwright/test';
import { goToAdmin, screenshot, waitForLoad, assertNoErrors, apiCall, dbQuery, API_URL, ADMIN_EMAIL, ADMIN_PASSWORD } from './helpers';

let accessToken: string;

test.describe.serial('Admin Email Templates', () => {
  test.beforeAll(async () => {
    const res = await fetch(`${API_URL}/admin/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email: ADMIN_EMAIL, password: ADMIN_PASSWORD }),
    });
    accessToken = (await res.json()).access_token;
  });

  test('should display email templates page correctly', async ({ page }) => {
    await goToAdmin(page, '/email-templates');
    await waitForLoad(page);
    await assertNoErrors(page);

    await expect(page.getByRole('heading', { name: 'Email Templates' })).toBeVisible({ timeout: 15000 });
    await expect(page.getByText('Customize email templates')).toBeVisible();

    await screenshot(page, 'email-01-page-display');
  });

  test('should create template via API and verify it appears in UI', async ({ page }) => {
    // Clean up any existing magic_link/de template to avoid unique constraint
    const existing = await dbQuery("SELECT id FROM email_templates WHERE type = 'magic_link' AND locale = 'de'");
    for (const row of existing) {
      await apiCall(accessToken, 'DELETE', `/admin/email-templates/${row.id}`);
    }

    const template = await apiCall(accessToken, 'POST', '/admin/email-templates', {
      type: 'magic_link',
      locale: 'de',
      subject: 'Dein Magic Link, {{user.name}}!',
      body_html: '<h1>Hallo, {{user.name}}!</h1><p>Klicke hier zum Einloggen.</p>',
    });
    expect(template?.id).toBeTruthy();

    // Navigate to page and verify template appears in sidebar
    await goToAdmin(page, '/email-templates');
    await waitForLoad(page);
    await assertNoErrors(page);

    await expect(page.getByText('magic_link').first()).toBeVisible();

    // Verify DB row exists
    const rows = await dbQuery('SELECT type, locale, subject FROM email_templates WHERE id = $1', [template.id]);
    expect(rows.length).toBe(1);
    expect(rows[0].type).toBe('magic_link');
    expect(rows[0].locale).toBe('de');

    // Cleanup
    await apiCall(accessToken, 'DELETE', `/admin/email-templates/${template.id}`);

    await screenshot(page, 'email-02-template-in-ui');
  });

  test('should list, update, and delete template via API', async ({ page }) => {
    // Clean up any existing magic_link/de template
    const existing = await dbQuery("SELECT id FROM email_templates WHERE type = 'magic_link' AND locale = 'de'");
    for (const row of existing) {
      await apiCall(accessToken, 'DELETE', `/admin/email-templates/${row.id}`);
    }

    // Create template
    const template = await apiCall(accessToken, 'POST', '/admin/email-templates', {
      type: 'magic_link',
      locale: 'de',
      subject: 'Original Subject',
      body_html: '<p>Original body</p>',
    });
    expect(template?.id).toBeTruthy();

    // GET list should contain the template
    const templates = await apiCall(accessToken, 'GET', '/admin/email-templates');
    const list = Array.isArray(templates) ? templates : templates?.data || [];
    expect(Array.isArray(list)).toBe(true);
    const found = list.find((t: any) => t.id === template.id);
    expect(found).toBeTruthy();

    // PATCH to update subject
    const updated = await apiCall(accessToken, 'PATCH', `/admin/email-templates/${template.id}`, {
      subject: 'Updated Magic Link Subject',
    });
    expect(updated).toBeTruthy();

    // Verify DB has updated subject
    const rows = await dbQuery('SELECT subject FROM email_templates WHERE id = $1', [template.id]);
    expect(rows.length).toBe(1);
    expect(rows[0].subject).toBe('Updated Magic Link Subject');

    // DELETE the template
    await apiCall(accessToken, 'DELETE', `/admin/email-templates/${template.id}`);

    // Verify DB row is gone
    const afterDelete = await dbQuery('SELECT id FROM email_templates WHERE id = $1', [template.id]);
    expect(afterDelete.length).toBe(0);

    await goToAdmin(page, '/email-templates');
    await waitForLoad(page);
    await assertNoErrors(page);
    await screenshot(page, 'email-03-api-crud');
  });
});
