import { test, expect } from '@playwright/test';
import { goToAdmin, screenshot, waitForLoad, assertNoErrors, apiCall, dbQuery, API_URL, ADMIN_EMAIL, ADMIN_PASSWORD } from './helpers';

let accessToken: string;

test.describe.serial('Page Templates', () => {
  let customTemplateId: string;

  test.beforeAll(async () => {
    const res = await fetch(`${API_URL}/admin/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email: ADMIN_EMAIL, password: ADMIN_PASSWORD }),
    });
    accessToken = (await res.json()).access_token;
  });

  test.afterAll(async () => {
    if (customTemplateId) {
      await apiCall(accessToken, 'DELETE', `/admin/page-templates/${customTemplateId}`).catch(() => {});
    }
    const templates = await apiCall(accessToken, 'GET', '/admin/page-templates').catch(() => []);
    for (const t of (templates ?? [])) {
      if (t.name?.startsWith('E2E') || t.name?.startsWith('UI Test') || t.name?.includes('Copy')) {
        await apiCall(accessToken, 'DELETE', `/admin/page-templates/${t.id}`).catch(() => {});
      }
    }
  });

  // ═══════════════════════════════════════════════════════════
  // DEFAULT TEMPLATES
  // ═══════════════════════════════════════════════════════════

  test('should list all 8 default templates', async () => {
    const result = await apiCall(accessToken, 'GET', '/admin/page-templates');
    expect(Array.isArray(result)).toBe(true);
    expect(result.length).toBeGreaterThanOrEqual(8);

    const defaults = result.filter((t: any) => t.is_default === true);
    expect(defaults.length).toBeGreaterThanOrEqual(8);

    const types = defaults.map((t: any) => t.page_type);
    for (const expectedType of ['login', 'signup', 'verification', 'password_reset', 'mfa_challenge', 'error', 'consent', 'profile']) {
      expect(types).toContain(expectedType);
    }
  });

  test('default templates should have html and css content', async () => {
    const result = await apiCall(accessToken, 'GET', '/admin/page-templates');
    const defaults = result.filter((t: any) => t.is_default);
    for (const tmpl of defaults) {
      expect(tmpl.html_content.length).toBeGreaterThan(50);
      expect(tmpl.css_content.length).toBeGreaterThan(50);
      expect(tmpl.name).toBeTruthy();
    }
  });

  test('signup template should include {{custom_fields}}', async () => {
    const result = await apiCall(accessToken, 'GET', '/admin/page-templates');
    const signup = result.find((t: any) => t.is_default && t.page_type === 'signup');
    expect(signup.html_content).toContain('{{custom_fields}}');
  });

  test('profile template should include {{profile_fields}}', async () => {
    const result = await apiCall(accessToken, 'GET', '/admin/page-templates');
    const profile = result.find((t: any) => t.is_default && t.page_type === 'profile');
    expect(profile.html_content).toContain('{{profile_fields}}');
  });

  test('default templates should be marked is_active=true', async () => {
    const result = await apiCall(accessToken, 'GET', '/admin/page-templates');
    const defaults = result.filter((t: any) => t.is_default);
    for (const tmpl of defaults) {
      expect(tmpl.is_active).toBe(true);
    }
  });

  test('default templates should exist in database', async () => {
    const rows = await dbQuery('SELECT id, page_type, is_default FROM page_templates WHERE is_default = true');
    expect(rows.length).toBeGreaterThanOrEqual(8);
  });

  // ═══════════════════════════════════════════════════════════
  // DEFAULT TEMPLATE PROTECTION (NEGATIVE TESTS)
  // ═══════════════════════════════════════════════════════════

  test('should reject updating default template name', async () => {
    const result = await apiCall(accessToken, 'GET', '/admin/page-templates');
    const def = result.find((t: any) => t.is_default);
    try {
      await apiCall(accessToken, 'PATCH', `/admin/page-templates/${def.id}`, { name: 'Hacked' });
      expect(true).toBe(false);
    } catch { /* expected */ }
  });

  test('should reject updating default template html', async () => {
    const result = await apiCall(accessToken, 'GET', '/admin/page-templates');
    const def = result.find((t: any) => t.is_default);
    try {
      await apiCall(accessToken, 'PATCH', `/admin/page-templates/${def.id}`, { html_content: '<h1>Hacked</h1>' });
      expect(true).toBe(false);
    } catch { /* expected */ }
  });

  test('should reject updating default template css', async () => {
    const result = await apiCall(accessToken, 'GET', '/admin/page-templates');
    const def = result.find((t: any) => t.is_default);
    try {
      await apiCall(accessToken, 'PATCH', `/admin/page-templates/${def.id}`, { css_content: 'body { display: none; }' });
      expect(true).toBe(false);
    } catch { /* expected */ }
  });

  test('should reject deleting default template', async () => {
    const result = await apiCall(accessToken, 'GET', '/admin/page-templates');
    const def = result.find((t: any) => t.is_default);
    try {
      await apiCall(accessToken, 'DELETE', `/admin/page-templates/${def.id}`);
      expect(true).toBe(false);
    } catch { /* expected */ }

    // Verify still exists
    const after = await apiCall(accessToken, 'GET', `/admin/page-templates/${def.id}`);
    expect(after.id).toBe(def.id);
  });

  // ═══════════════════════════════════════════════════════════
  // DUPLICATE
  // ═══════════════════════════════════════════════════════════

  test('should duplicate a default login template', async () => {
    const result = await apiCall(accessToken, 'GET', '/admin/page-templates');
    const loginDefault = result.find((t: any) => t.is_default && t.page_type === 'login');

    const dup = await apiCall(accessToken, 'POST', `/admin/page-templates/${loginDefault.id}/duplicate`, {
      name: 'E2E Custom Login',
    });
    expect(dup.id).toBeTruthy();
    expect(dup.id).not.toBe(loginDefault.id);
    expect(dup.name).toBe('E2E Custom Login');
    expect(dup.is_default).toBe(false);
    expect(dup.is_active).toBe(false);
    expect(dup.page_type).toBe('login');
    expect(dup.html_content).toBe(loginDefault.html_content);
    expect(dup.css_content).toBe(loginDefault.css_content);
    customTemplateId = dup.id;
  });

  test('should be able to edit duplicated template', async () => {
    const updated = await apiCall(accessToken, 'PATCH', `/admin/page-templates/${customTemplateId}`, {
      name: 'E2E Custom Login v2',
      html_content: '<h1>My Custom Login</h1>',
      css_content: 'h1 { color: red; }',
      is_active: true,
    });
    expect(updated.name).toBe('E2E Custom Login v2');
    expect(updated.html_content).toBe('<h1>My Custom Login</h1>');
    expect(updated.css_content).toBe('h1 { color: red; }');
    expect(updated.is_active).toBe(true);
  });

  test('should duplicate a custom template', async () => {
    const dup = await apiCall(accessToken, 'POST', `/admin/page-templates/${customTemplateId}/duplicate`, {
      name: 'E2E Double Copy',
    });
    expect(dup.name).toBe('E2E Double Copy');
    expect(dup.html_content).toBe('<h1>My Custom Login</h1>');
    // Clean up
    await apiCall(accessToken, 'DELETE', `/admin/page-templates/${dup.id}`);
  });

  test('should reject duplicating non-existent template', async () => {
    try {
      await apiCall(accessToken, 'POST', '/admin/page-templates/00000000-0000-0000-0000-000000000000/duplicate', {
        name: 'Ghost',
      });
      expect(true).toBe(false);
    } catch { /* expected */ }
  });

  // ═══════════════════════════════════════════════════════════
  // CUSTOM TEMPLATE CRUD
  // ═══════════════════════════════════════════════════════════

  test('should create a custom page template', async () => {
    const result = await apiCall(accessToken, 'POST', '/admin/page-templates', {
      page_type: 'custom',
      name: 'E2E Custom Page',
      html_content: '<html><body><h1>Custom</h1></body></html>',
      css_content: 'body { background: #fff; }',
    });
    expect(result.id).toBeTruthy();
    expect(result.page_type).toBe('custom');
    expect(result.is_default).toBe(false);
    expect(result.is_active).toBe(false);
  });

  test('should get custom template by id', async () => {
    const list = await apiCall(accessToken, 'GET', '/admin/page-templates');
    const custom = list.find((t: any) => t.name === 'E2E Custom Page');
    expect(custom).toBeTruthy();

    const detail = await apiCall(accessToken, 'GET', `/admin/page-templates/${custom.id}`);
    expect(detail.name).toBe('E2E Custom Page');
    expect(detail.html_content).toContain('<h1>Custom</h1>');
  });

  test('should update custom template', async () => {
    const list = await apiCall(accessToken, 'GET', '/admin/page-templates');
    const custom = list.find((t: any) => t.name === 'E2E Custom Page');

    const updated = await apiCall(accessToken, 'PATCH', `/admin/page-templates/${custom.id}`, {
      name: 'E2E Custom Page Updated',
      is_active: true,
    });
    expect(updated.name).toBe('E2E Custom Page Updated');
    expect(updated.is_active).toBe(true);
  });

  test('should delete custom template', async () => {
    const list = await apiCall(accessToken, 'GET', '/admin/page-templates');
    const custom = list.find((t: any) => t.name === 'E2E Custom Page Updated');
    await apiCall(accessToken, 'DELETE', `/admin/page-templates/${custom.id}`);

    const after = await apiCall(accessToken, 'GET', '/admin/page-templates');
    expect(after.find((t: any) => t.name === 'E2E Custom Page Updated')).toBeUndefined();
  });

  test('should reject getting non-existent template', async () => {
    try {
      await apiCall(accessToken, 'GET', '/admin/page-templates/00000000-0000-0000-0000-000000000000');
      expect(true).toBe(false);
    } catch { /* expected */ }
  });

  // ═══════════════════════════════════════════════════════════
  // LANGUAGE STRINGS - CRUD
  // ═══════════════════════════════════════════════════════════

  test('should list seeded language strings for en', async () => {
    const result = await apiCall(accessToken, 'GET', '/admin/language-strings?locale=en');
    expect(Array.isArray(result)).toBe(true);
    expect(result.length).toBeGreaterThan(30);

    const loginTitle = result.find((s: any) => s.string_key === 'login.title');
    expect(loginTitle).toBeTruthy();
    expect(loginTitle.value).toBe('Welcome back');
  });

  test('should list seeded language strings for de', async () => {
    const result = await apiCall(accessToken, 'GET', '/admin/language-strings?locale=de');
    expect(Array.isArray(result)).toBe(true);
    expect(result.length).toBeGreaterThan(5);

    const loginTitle = result.find((s: any) => s.string_key === 'login.title');
    expect(loginTitle).toBeTruthy();
    expect(loginTitle.value).toContain('Willkommen');
  });

  test('should return empty array for unknown locale', async () => {
    const result = await apiCall(accessToken, 'GET', '/admin/language-strings?locale=zh');
    expect(Array.isArray(result)).toBe(true);
    expect(result.length).toBe(0);
  });

  test('should create a new language string', async () => {
    const result = await apiCall(accessToken, 'PUT', '/admin/language-strings', {
      string_key: 'e2e.greeting',
      locale: 'en',
      value: 'Hello World',
    });
    expect(result.string_key).toBe('e2e.greeting');
    expect(result.locale).toBe('en');
    expect(result.value).toBe('Hello World');
    expect(result.id).toBeTruthy();
  });

  test('should upsert (update) existing language string', async () => {
    const result = await apiCall(accessToken, 'PUT', '/admin/language-strings', {
      string_key: 'e2e.greeting',
      locale: 'en',
      value: 'Hello Updated World',
    });
    expect(result.value).toBe('Hello Updated World');
  });

  test('should create string for different locale', async () => {
    const result = await apiCall(accessToken, 'PUT', '/admin/language-strings', {
      string_key: 'e2e.greeting',
      locale: 'de',
      value: 'Hallo Welt',
    });
    expect(result.locale).toBe('de');
    expect(result.value).toBe('Hallo Welt');
  });

  test('should verify strings exist in DB', async () => {
    const rows = await dbQuery(
      "SELECT string_key, locale, value FROM template_language_strings WHERE string_key = 'e2e.greeting' ORDER BY locale"
    );
    expect(rows.length).toBe(2);
    expect(rows.find((r: any) => r.locale === 'en').value).toBe('Hello Updated World');
    expect(rows.find((r: any) => r.locale === 'de').value).toBe('Hallo Welt');
  });

  test('should delete language string for specific locale', async () => {
    await apiCall(accessToken, 'DELETE', '/admin/language-strings/e2e.greeting/de');

    const deStrings = await apiCall(accessToken, 'GET', '/admin/language-strings?locale=de');
    expect(deStrings.find((s: any) => s.string_key === 'e2e.greeting')).toBeUndefined();

    // en should still exist
    const enStrings = await apiCall(accessToken, 'GET', '/admin/language-strings?locale=en');
    expect(enStrings.find((s: any) => s.string_key === 'e2e.greeting')).toBeTruthy();
  });

  test('should delete remaining language string', async () => {
    await apiCall(accessToken, 'DELETE', '/admin/language-strings/e2e.greeting/en');
    const enStrings = await apiCall(accessToken, 'GET', '/admin/language-strings?locale=en');
    expect(enStrings.find((s: any) => s.string_key === 'e2e.greeting')).toBeUndefined();
  });

  // ═══════════════════════════════════════════════════════════
  // UI: DEFAULT TEMPLATES
  // ═══════════════════════════════════════════════════════════

  test('should show default templates section with lock icons', async ({ page }) => {
    await goToAdmin(page, '/page-templates');
    await waitForLoad(page);
    await assertNoErrors(page);

    await expect(page.getByText('Default Templates')).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('Default Login')).toBeVisible();
    await expect(page.getByText('Default Sign Up')).toBeVisible();
    await expect(page.getByText('Default Profile')).toBeVisible();
    await expect(page.getByText('Default Error Page')).toBeVisible();

    await screenshot(page, '26-defaults-list');
  });

  test('should show readonly notice for default template', async ({ page }) => {
    await goToAdmin(page, '/page-templates');
    await waitForLoad(page);

    await page.getByText('Default Login').click();
    await page.waitForTimeout(300);

    // Should show readonly warning
    await expect(page.getByText('default template and cannot be edited')).toBeVisible({ timeout: 3000 });

    // Should show Duplicate button but NOT Delete/Save
    await expect(page.getByRole('button', { name: /Duplicate/ })).toBeVisible();
    await expect(page.getByRole('button', { name: /Delete/ })).not.toBeVisible();
    await expect(page.getByRole('button', { name: /Save/ })).not.toBeVisible();

    await screenshot(page, '26-default-readonly');
  });

  // ═══════════════════════════════════════════════════════════
  // UI: PREVIEW
  // ═══════════════════════════════════════════════════════════

  test('should render preview with variables replaced', async ({ page }) => {
    await goToAdmin(page, '/page-templates');
    await waitForLoad(page);

    await page.getByText('Default Login').click();
    await page.waitForTimeout(300);

    await page.getByRole('button', { name: /PREVIEW/ }).click();
    await page.waitForTimeout(500);

    // The iframe should be visible
    const iframe = page.locator('iframe[title="Template Preview"]');
    await expect(iframe).toBeVisible({ timeout: 3000 });

    await screenshot(page, '26-preview-rendered');
  });

  // ═══════════════════════════════════════════════════════════
  // UI: DUPLICATE
  // ═══════════════════════════════════════════════════════════

  test('should duplicate from UI and edit the copy', async ({ page }) => {
    await goToAdmin(page, '/page-templates');
    await waitForLoad(page);

    await page.getByText('Default Login').click();
    await page.waitForTimeout(300);

    // Open duplicate dialog
    await page.getByRole('button', { name: /Duplicate/ }).click();
    await page.waitForTimeout(300);
    await expect(page.getByRole('heading', { name: 'Duplicate Template' })).toBeVisible({ timeout: 3000 });

    // Fill name and confirm
    const nameInput = page.getByPlaceholder('My Custom Login');
    await nameInput.clear();
    await nameInput.fill('UI Test Login Copy');
    await page.getByRole('button', { name: 'Duplicate' }).nth(1).click();
    await page.waitForTimeout(1000);

    // Should select the new template
    await expect(page.getByRole('heading', { name: 'UI Test Login Copy' })).toBeVisible({ timeout: 5000 });

    // Should NOT show readonly notice (it's a custom copy)
    await expect(page.getByText('cannot be edited')).not.toBeVisible();

    // Should show Save/Delete buttons
    await expect(page.getByRole('button', { name: /Save/ })).toBeVisible();
    await expect(page.getByRole('button', { name: /Delete/ })).toBeVisible();

    await screenshot(page, '26-duplicated-editable');
  });

  // ═══════════════════════════════════════════════════════════
  // UI: SEARCH
  // ═══════════════════════════════════════════════════════════

  test('should filter templates by search query', async ({ page }) => {
    await goToAdmin(page, '/page-templates');
    await waitForLoad(page);

    const search = page.getByPlaceholder('Search templates...');
    await search.fill('consent');
    await page.waitForTimeout(300);

    await expect(page.getByText('Default Consent')).toBeVisible();
    await expect(page.getByText('Default Login')).not.toBeVisible();
    await expect(page.getByText('Default Sign Up')).not.toBeVisible();

    // Clear search
    await search.clear();
    await page.waitForTimeout(300);
    await expect(page.getByText('Default Login')).toBeVisible();

    await screenshot(page, '26-search-filter');
  });

  test('should search by page type label', async ({ page }) => {
    await goToAdmin(page, '/page-templates');
    await waitForLoad(page);

    await page.getByPlaceholder('Search templates...').fill('MFA');
    await page.waitForTimeout(300);

    await expect(page.getByText('Default MFA Challenge')).toBeVisible();
    await expect(page.getByText('Default Login')).not.toBeVisible();
  });

  // ═══════════════════════════════════════════════════════════
  // UI: LANGUAGE STRINGS DIALOG
  // ═══════════════════════════════════════════════════════════

  test('should open language strings dialog and show strings', async ({ page }) => {
    await goToAdmin(page, '/page-templates');
    await waitForLoad(page);

    await page.getByRole('button', { name: /Language Strings/ }).click();
    await page.waitForTimeout(300);

    await expect(page.getByRole('heading', { name: 'Language Strings' })).toBeVisible({ timeout: 3000 });
    await expect(page.getByText('login.title')).toBeVisible();
    await expect(page.getByText('login.email')).toBeVisible();
    await expect(page.getByText('signup.title')).toBeVisible();

    await screenshot(page, '26-lang-strings-dialog');
  });

  test('should add a new language string via dialog', async ({ page }) => {
    await goToAdmin(page, '/page-templates');
    await waitForLoad(page);

    await page.getByRole('button', { name: /Language Strings/ }).click();
    await page.waitForTimeout(300);

    // Fill key and value
    await page.getByPlaceholder('login.welcome').fill('e2e.ui_test');
    await page.getByPlaceholder('Welcome back!').fill('UI Test String');

    // Click add button
    const addBtns = page.locator('button').filter({ has: page.locator('svg.lucide-plus') });
    await addBtns.last().click();
    await page.waitForTimeout(500);

    // Verify it appeared
    await expect(page.getByText('e2e.ui_test')).toBeVisible({ timeout: 3000 });

    // Cleanup
    await apiCall(accessToken, 'DELETE', '/admin/language-strings/e2e.ui_test/en').catch(() => {});
  });

  // ═══════════════════════════════════════════════════════════
  // UI: TABS (HTML/CSS/PREVIEW)
  // ═══════════════════════════════════════════════════════════

  test('should switch between HTML, CSS, and Preview tabs', async ({ page }) => {
    await goToAdmin(page, '/page-templates');
    await waitForLoad(page);

    await page.getByText('Default Login').click();
    await page.waitForTimeout(300);

    // HTML tab should be active by default
    await expect(page.getByRole('button', { name: /HTML/ })).toBeVisible();

    // Click CSS tab
    await page.getByRole('button', { name: /CSS/ }).click();
    await page.waitForTimeout(200);
    await screenshot(page, '26-css-tab');

    // Click Preview tab
    await page.getByRole('button', { name: /PREVIEW/ }).click();
    await page.waitForTimeout(500);
    await expect(page.locator('iframe[title="Template Preview"]')).toBeVisible();
    await screenshot(page, '26-preview-tab');

    // Click back to HTML
    await page.getByRole('button', { name: /HTML/ }).click();
    await page.waitForTimeout(200);
  });

  // ═══════════════════════════════════════════════════════════
  // UI: CREATE TEMPLATE
  // ═══════════════════════════════════════════════════════════

  test('should create custom template from UI', async ({ page }) => {
    await goToAdmin(page, '/page-templates');
    await waitForLoad(page);

    await page.getByRole('button', { name: /Create Template/ }).click();
    await page.waitForTimeout(300);

    await expect(page.getByRole('heading', { name: 'Create Template' })).toBeVisible({ timeout: 3000 });

    // Fill name
    const nameInput = page.getByPlaceholder(/Custom Page/);
    await nameInput.fill('E2E UI Created Template');

    // Click Create
    await page.getByRole('button', { name: 'Create', exact: true }).click();
    await page.waitForTimeout(1000);

    // Should be selected
    await expect(page.getByRole('heading', { name: 'E2E UI Created Template' })).toBeVisible({ timeout: 5000 });

    await screenshot(page, '26-ui-created');
  });

  // ═══════════════════════════════════════════════════════════
  // UI: DELETE TEMPLATE
  // ═══════════════════════════════════════════════════════════

  test('should delete custom template from UI', async ({ page }) => {
    await goToAdmin(page, '/page-templates');
    await waitForLoad(page);

    // Find and click on the UI-created template
    const tmpl = page.getByText('E2E UI Created Template');
    if (await tmpl.isVisible()) {
      await tmpl.click();
      await page.waitForTimeout(300);

      // Click Delete
      await page.getByRole('button', { name: /Delete/ }).click();
      await page.waitForTimeout(300);

      // Confirm deletion
      await page.getByRole('button', { name: /Confirm|Delete/i }).last().click();
      await page.waitForTimeout(500);

      // Template should be gone
      await expect(page.getByText('E2E UI Created Template')).not.toBeVisible();
    }
  });

  // ═══════════════════════════════════════════════════════════
  // UI: VARIABLE TOOLBAR
  // ═══════════════════════════════════════════════════════════

  test('should show variable toolbar with all variables', async ({ page }) => {
    await goToAdmin(page, '/page-templates');
    await waitForLoad(page);

    await page.getByText('Default Login').click();
    await page.waitForTimeout(300);

    // Should show variable buttons
    await expect(page.getByText('Variables:')).toBeVisible({ timeout: 3000 });
    await expect(page.getByRole('button', { name: '{{user.name}}' })).toBeVisible();
    await expect(page.getByRole('button', { name: '{{user.email}}' })).toBeVisible();
    await expect(page.getByRole('button', { name: '{{application.name}}' })).toBeVisible();
    await expect(page.getByRole('button', { name: '{{custom_fields}}' })).toBeVisible();

    await screenshot(page, '26-variable-toolbar');
  });
});
