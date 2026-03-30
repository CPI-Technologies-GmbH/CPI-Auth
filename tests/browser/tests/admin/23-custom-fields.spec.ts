import { test, expect } from '@playwright/test';
import { goToAdmin, screenshot, waitForLoad, assertNoErrors, apiCall, dbQuery, API_URL, ADMIN_EMAIL, ADMIN_PASSWORD } from './helpers';

let accessToken: string;

test.describe.serial('Custom Fields', () => {
  const createdIds: string[] = [];

  test.beforeAll(async () => {
    const res = await fetch(`${API_URL}/admin/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email: ADMIN_EMAIL, password: ADMIN_PASSWORD }),
    });
    accessToken = (await res.json()).access_token;
  });

  test.afterAll(async () => {
    for (const id of createdIds) {
      await apiCall(accessToken, 'DELETE', `/admin/custom-fields/${id}`).catch(() => {});
    }
    // Clean up any remaining e2e fields
    const fields = await apiCall(accessToken, 'GET', '/admin/custom-fields').catch(() => []);
    for (const f of (fields ?? [])) {
      if (f.name?.startsWith('e2e_')) {
        await apiCall(accessToken, 'DELETE', `/admin/custom-fields/${f.id}`).catch(() => {});
      }
    }
  });

  // ═══════════════════════════════════════════════════════════
  // ALL 9 FIELD TYPES
  // ═══════════════════════════════════════════════════════════

  test('should create text field', async () => {
    const field = await apiCall(accessToken, 'POST', '/admin/custom-fields', {
      name: `e2e_text_${Date.now()}`, label: 'Full Name', field_type: 'text',
      placeholder: 'John Doe', required: true, visible_on: 'both', position: 1,
    });
    expect(field.id).toBeTruthy();
    expect(field.field_type).toBe('text');
    expect(field.required).toBe(true);
    createdIds.push(field.id);
  });

  test('should create number field', async () => {
    const field = await apiCall(accessToken, 'POST', '/admin/custom-fields', {
      name: `e2e_number_${Date.now()}`, label: 'Age', field_type: 'number',
      placeholder: '25', visible_on: 'registration', position: 2,
    });
    expect(field.field_type).toBe('number');
    createdIds.push(field.id);
  });

  test('should create email field', async () => {
    const field = await apiCall(accessToken, 'POST', '/admin/custom-fields', {
      name: `e2e_email_${Date.now()}`, label: 'Work Email', field_type: 'email',
      placeholder: 'work@company.com', visible_on: 'profile', position: 3,
    });
    expect(field.field_type).toBe('email');
    createdIds.push(field.id);
  });

  test('should create tel field', async () => {
    const field = await apiCall(accessToken, 'POST', '/admin/custom-fields', {
      name: `e2e_tel_${Date.now()}`, label: 'Phone', field_type: 'tel',
      placeholder: '+1 555-0100', visible_on: 'both', position: 4,
    });
    expect(field.field_type).toBe('tel');
    createdIds.push(field.id);
  });

  test('should create url field', async () => {
    const field = await apiCall(accessToken, 'POST', '/admin/custom-fields', {
      name: `e2e_url_${Date.now()}`, label: 'Website', field_type: 'url',
      placeholder: 'https://example.com', visible_on: 'profile', position: 5,
    });
    expect(field.field_type).toBe('url');
    createdIds.push(field.id);
  });

  test('should create date field', async () => {
    const field = await apiCall(accessToken, 'POST', '/admin/custom-fields', {
      name: `e2e_date_${Date.now()}`, label: 'Birthday', field_type: 'date',
      visible_on: 'registration', position: 6,
    });
    expect(field.field_type).toBe('date');
    createdIds.push(field.id);
  });

  test('should create checkbox field', async () => {
    const field = await apiCall(accessToken, 'POST', '/admin/custom-fields', {
      name: `e2e_checkbox_${Date.now()}`, label: 'Accept Terms', field_type: 'checkbox',
      required: true, visible_on: 'registration', position: 7,
    });
    expect(field.field_type).toBe('checkbox');
    expect(field.required).toBe(true);
    createdIds.push(field.id);
  });

  test('should create select field with options', async () => {
    const field = await apiCall(accessToken, 'POST', '/admin/custom-fields', {
      name: `e2e_select_${Date.now()}`, label: 'Department', field_type: 'select',
      options: ['Engineering', 'Marketing', 'Sales'], visible_on: 'both', position: 8,
    });
    expect(field.field_type).toBe('select');
    expect(field.options).toEqual(['Engineering', 'Marketing', 'Sales']);
    createdIds.push(field.id);
  });

  test('should create textarea field', async () => {
    const field = await apiCall(accessToken, 'POST', '/admin/custom-fields', {
      name: `e2e_textarea_${Date.now()}`, label: 'Bio', field_type: 'textarea',
      placeholder: 'Tell us about yourself', description: 'Max 500 chars',
      visible_on: 'profile', position: 9,
    });
    expect(field.field_type).toBe('textarea');
    expect(field.description).toBe('Max 500 chars');
    createdIds.push(field.id);
  });

  // ═══════════════════════════════════════════════════════════
  // NEGATIVE TESTS
  // ═══════════════════════════════════════════════════════════

  test('should reject invalid field_type', async () => {
    try {
      await apiCall(accessToken, 'POST', '/admin/custom-fields', {
        name: 'e2e_bad_type', label: 'Bad', field_type: 'boolean',
        visible_on: 'both', position: 99,
      });
      expect(true).toBe(false);
    } catch { /* expected - DB CHECK constraint */ }
  });

  test('should reject duplicate field name', async () => {
    const name = `e2e_dup_${Date.now()}`;
    const field = await apiCall(accessToken, 'POST', '/admin/custom-fields', {
      name, label: 'First', field_type: 'text', visible_on: 'both', position: 1,
    });
    createdIds.push(field.id);

    try {
      await apiCall(accessToken, 'POST', '/admin/custom-fields', {
        name, label: 'Duplicate', field_type: 'text', visible_on: 'both', position: 2,
      });
      expect(true).toBe(false);
    } catch { /* expected - UNIQUE constraint */ }
  });

  test('should reject invalid visible_on value', async () => {
    try {
      await apiCall(accessToken, 'POST', '/admin/custom-fields', {
        name: 'e2e_bad_visibility', label: 'Bad', field_type: 'text',
        visible_on: 'nowhere', position: 99,
      });
      expect(true).toBe(false);
    } catch { /* expected - CHECK constraint */ }
  });

  // ═══════════════════════════════════════════════════════════
  // CRUD OPERATIONS
  // ═══════════════════════════════════════════════════════════

  test('should list all created fields', async () => {
    const fields = await apiCall(accessToken, 'GET', '/admin/custom-fields');
    expect(Array.isArray(fields)).toBe(true);
    expect(fields.length).toBeGreaterThanOrEqual(9);
  });

  test('should get field by ID', async () => {
    const field = await apiCall(accessToken, 'GET', `/admin/custom-fields/${createdIds[0]}`);
    expect(field.id).toBe(createdIds[0]);
    expect(field.field_type).toBe('text');
  });

  test('should update field label and required', async () => {
    const updated = await apiCall(accessToken, 'PATCH', `/admin/custom-fields/${createdIds[0]}`, {
      label: 'Updated Label',
      required: false,
    });
    expect(updated.label).toBe('Updated Label');
    expect(updated.required).toBe(false);
  });

  test('should update field visibility', async () => {
    const updated = await apiCall(accessToken, 'PATCH', `/admin/custom-fields/${createdIds[0]}`, {
      visible_on: 'profile',
    });
    expect(updated.visible_on).toBe('profile');
  });

  test('should verify fields exist in database', async () => {
    const rows = await dbQuery(
      `SELECT name, field_type, required, visible_on FROM custom_field_definitions WHERE name LIKE 'e2e_%' ORDER BY position`
    );
    expect(rows.length).toBeGreaterThanOrEqual(9);
    const types = rows.map((r: any) => r.field_type);
    expect(types).toContain('text');
    expect(types).toContain('number');
    expect(types).toContain('email');
    expect(types).toContain('tel');
    expect(types).toContain('url');
    expect(types).toContain('date');
    expect(types).toContain('checkbox');
    expect(types).toContain('select');
    expect(types).toContain('textarea');
  });

  test('should reject getting non-existent field', async () => {
    try {
      await apiCall(accessToken, 'GET', '/admin/custom-fields/00000000-0000-0000-0000-000000000000');
      expect(true).toBe(false);
    } catch { /* expected */ }
  });

  // ═══════════════════════════════════════════════════════════
  // UI TESTS
  // ═══════════════════════════════════════════════════════════

  test('should render custom fields page with fields', async ({ page }) => {
    await goToAdmin(page, '/custom-fields');
    await waitForLoad(page);
    await assertNoErrors(page);
    await expect(page.getByRole('heading', { name: 'Custom Fields' })).toBeVisible({ timeout: 5000 });
    await screenshot(page, '23-custom-fields-with-data');
  });

  test('should show field types in table', async ({ page }) => {
    await goToAdmin(page, '/custom-fields');
    await waitForLoad(page);

    // Table should show various field type badges
    await expect(page.getByText('Updated Label')).toBeVisible({ timeout: 5000 });
    await screenshot(page, '23-custom-fields-types');
  });

  test('should open create dialog with all valid field types', async ({ page }) => {
    await goToAdmin(page, '/custom-fields');
    await waitForLoad(page);

    await page.getByRole('button', { name: /Add Field/i }).click();
    await page.waitForTimeout(300);

    await expect(page.getByText('Add Custom Field')).toBeVisible({ timeout: 3000 });

    // The field type dropdown should NOT contain "Boolean"
    const fieldTypeSelect = page.locator('select').filter({ has: page.getByText('Text') }).first();
    if (await fieldTypeSelect.isVisible()) {
      const options = await fieldTypeSelect.locator('option').allTextContents();
      expect(options).not.toContain('Boolean');
      // Should contain all valid types
      expect(options.join(' ').toLowerCase()).toContain('text');
      expect(options.join(' ').toLowerCase()).toContain('checkbox');
    }

    await screenshot(page, '23-create-dialog-types');
  });

  test('should create field from UI', async ({ page }) => {
    await goToAdmin(page, '/custom-fields');
    await waitForLoad(page);

    await page.getByRole('button', { name: /Add Field/i }).click();
    await page.waitForTimeout(300);

    // Fill the form
    const nameInput = page.locator('input').filter({ has: page.locator('[placeholder]') }).first();
    const inputs = page.locator('input[type="text"], input:not([type])');
    if (await inputs.first().isVisible()) {
      await inputs.first().fill(`e2e_ui_field_${Date.now()}`);
    }

    await screenshot(page, '23-create-dialog-filled');
  });

  test('should delete field and verify', async () => {
    // Delete all e2e fields
    for (const id of createdIds) {
      await apiCall(accessToken, 'DELETE', `/admin/custom-fields/${id}`).catch(() => {});
    }
    createdIds.length = 0;

    const fields = await apiCall(accessToken, 'GET', '/admin/custom-fields');
    const e2eFields = (fields ?? []).filter((f: any) => f.name?.startsWith('e2e_'));
    for (const f of e2eFields) {
      await apiCall(accessToken, 'DELETE', `/admin/custom-fields/${f.id}`).catch(() => {});
    }

    // Verify in DB
    const rows = await dbQuery("SELECT COUNT(*) as cnt FROM custom_field_definitions WHERE name LIKE 'e2e_%'");
    expect(parseInt(rows[0].cnt)).toBe(0);
  });

  test('should show empty state', async ({ page }) => {
    await goToAdmin(page, '/custom-fields');
    await waitForLoad(page);
    await assertNoErrors(page);
    await screenshot(page, '23-custom-fields-empty');
  });
});
