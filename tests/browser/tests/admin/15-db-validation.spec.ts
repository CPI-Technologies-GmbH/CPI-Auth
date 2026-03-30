import { test, expect } from '@playwright/test';
import { dbQuery, apiCall, goToAdmin, screenshot, waitForLoad, assertNoErrors, API_URL, ADMIN_EMAIL, ADMIN_PASSWORD } from './helpers';

let accessToken: string;

test.describe.serial('Database Validation', () => {
  test.beforeAll(async () => {
    const res = await fetch(`${API_URL}/admin/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email: ADMIN_EMAIL, password: ADMIN_PASSWORD }),
    });
    accessToken = (await res.json()).access_token;
  });

  test('should have admin user in database', async ({ page }) => {
    const rows = await dbQuery('SELECT email, status FROM users WHERE email = $1', [ADMIN_EMAIL]);
    expect(rows.length).toBe(1);
    expect(rows[0].email).toBe(ADMIN_EMAIL);
    expect(rows[0].status).toBe('active');

    await goToAdmin(page, '/');
    await waitForLoad(page);
    await assertNoErrors(page);

    await screenshot(page, 'db-01-admin-exists');
  });

  test('should verify user CRUD lifecycle in database', async ({ page }) => {
    const ts = Date.now();
    const email = `db-user-${ts}@cpi-auth.local`;

    // Create user via API
    const user = await apiCall(accessToken, 'POST', '/admin/users', {
      name: `DB User ${ts}`,
      email,
      password: 'Xk9mQ2vL8nR4!pw',
    });
    expect(user.id).toBeTruthy();

    // Verify creation in DB
    const createRows = await dbQuery('SELECT email, name, status FROM users WHERE id = $1', [user.id]);
    expect(createRows.length).toBe(1);
    expect(createRows[0].email).toBe(email);
    expect(createRows[0].name).toBe(`DB User ${ts}`);
    expect(createRows[0].status).toBe('active');

    // Update name via API
    await apiCall(accessToken, 'PATCH', `/admin/users/${user.id}`, {
      name: `Updated DB User ${ts}`,
    });

    // Verify update in DB
    const updateRows = await dbQuery('SELECT name FROM users WHERE id = $1', [user.id]);
    expect(updateRows.length).toBe(1);
    expect(updateRows[0].name).toBe(`Updated DB User ${ts}`);

    // Delete via API
    await apiCall(accessToken, 'DELETE', `/admin/users/${user.id}`);

    // Verify deletion in DB
    const deleteRows = await dbQuery('SELECT id FROM users WHERE id = $1', [user.id]);
    expect(deleteRows.length).toBe(0);
  });

  test('should verify application CRUD lifecycle in database', async ({ page }) => {
    const ts = Date.now();

    // Create application via API
    const app = await apiCall(accessToken, 'POST', '/admin/applications', {
      name: `DB App ${ts}`,
      type: 'spa',
      redirect_uris: ['http://localhost:3000/callback'],
      allowed_origins: ['http://localhost:3000'],
    });
    expect(app.id).toBeTruthy();

    // Verify creation in DB
    const createRows = await dbQuery('SELECT name, type, client_id FROM applications WHERE id = $1', [app.id]);
    expect(createRows.length).toBe(1);
    expect(createRows[0].name).toBe(`DB App ${ts}`);
    expect(createRows[0].type).toBe('spa');
    expect(createRows[0].client_id).toBeTruthy();

    // Delete via API
    await apiCall(accessToken, 'DELETE', `/admin/applications/${app.id}`);

    // Verify deletion in DB
    const deleteRows = await dbQuery('SELECT id FROM applications WHERE id = $1', [app.id]);
    expect(deleteRows.length).toBe(0);
  });

  test('should verify full multi-entity CRUD consistency', async ({ page }) => {
    const ts = Date.now();

    // Create user, application, and role
    const user = await apiCall(accessToken, 'POST', '/admin/users', {
      name: `Multi User ${ts}`,
      email: `multi-${ts}@cpi-auth.local`,
      password: 'Xk9mQ2vL8nR4!pw',
    });
    const app = await apiCall(accessToken, 'POST', '/admin/applications', {
      name: `Multi App ${ts}`,
      type: 'spa',
      redirect_uris: ['http://localhost:3000/callback'],
      allowed_origins: ['http://localhost:3000'],
    });
    const role = await apiCall(accessToken, 'POST', '/admin/roles', {
      name: `multi-role-${ts}`,
      permissions: ['users:read'],
    });

    expect(user.id).toBeTruthy();
    expect(app.id).toBeTruthy();
    expect(role.id).toBeTruthy();

    // Verify all 3 exist in DB
    const userRows = await dbQuery('SELECT id FROM users WHERE id = $1', [user.id]);
    const appRows = await dbQuery('SELECT id FROM applications WHERE id = $1', [app.id]);
    const roleRows = await dbQuery('SELECT id FROM roles WHERE id = $1', [role.id]);
    expect(userRows.length).toBe(1);
    expect(appRows.length).toBe(1);
    expect(roleRows.length).toBe(1);

    // Delete all 3
    await apiCall(accessToken, 'DELETE', `/admin/users/${user.id}`);
    await apiCall(accessToken, 'DELETE', `/admin/applications/${app.id}`);
    await apiCall(accessToken, 'DELETE', `/admin/roles/${role.id}`);

    // Verify all 3 are gone from DB
    const userRows2 = await dbQuery('SELECT id FROM users WHERE id = $1', [user.id]);
    const appRows2 = await dbQuery('SELECT id FROM applications WHERE id = $1', [app.id]);
    const roleRows2 = await dbQuery('SELECT id FROM roles WHERE id = $1', [role.id]);
    expect(userRows2.length).toBe(0);
    expect(appRows2.length).toBe(0);
    expect(roleRows2.length).toBe(0);
  });
});
