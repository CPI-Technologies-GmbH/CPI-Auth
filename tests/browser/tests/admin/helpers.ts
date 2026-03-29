import { type Page, expect } from '@playwright/test';
import { Client } from 'pg';

const API_URL = process.env.E2E_API_URL || 'http://localhost:5050';
const ADMIN_EMAIL = 'admin@cpi-auth.local';
const ADMIN_PASSWORD = 'admin123!';

export { API_URL, ADMIN_EMAIL, ADMIN_PASSWORD };

/** Login via API and store tokens in localStorage so the admin-ui picks them up. */
export async function adminLogin(page: Page): Promise<string> {
  const res = await fetch(`${API_URL}/admin/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email: ADMIN_EMAIL, password: ADMIN_PASSWORD }),
  });
  const body = await res.json();
  if (!body.access_token) throw new Error('Admin login failed: ' + JSON.stringify(body));

  // Navigate first so we have a page context, then set localStorage
  await page.goto('/login', { waitUntil: 'domcontentloaded' });
  await page.evaluate((tokens) => {
    localStorage.setItem('access_token', tokens.access_token);
    localStorage.setItem('refresh_token', tokens.refresh_token || '');
    localStorage.setItem('token_expires_at', String(Date.now() + (tokens.expires_in || 3600) * 1000));
  }, body);

  return body.access_token;
}

/** Navigate to admin page after login. */
export async function goToAdmin(page: Page, path = '/') {
  await adminLogin(page);
  await page.goto(path, { waitUntil: 'domcontentloaded' });
  await waitForLoad(page);
}

/** Take a named screenshot and return the path. */
export async function screenshot(page: Page, name: string) {
  await page.screenshot({ path: `test-results/screenshots/${name}.png`, fullPage: true, timeout: 10000 }).catch(() => {});
}

/** Wait for the page to be ready (no loading spinners). */
export async function waitForLoad(page: Page, timeout = 5000) {
  await page.waitForLoadState('networkidle', { timeout }).catch(() => {});
}

/** Assert the page rendered without error boundaries or crash screens. */
export async function assertNoErrors(page: Page) {
  // Check no error boundary is visible
  const body = await page.textContent('body');
  const errorPatterns = [
    'is not a function',
    'Cannot read properties of undefined',
    'Cannot read properties of null',
    'Something went wrong',
    'Unhandled Runtime Error',
    'Application error',
  ];
  for (const pattern of errorPatterns) {
    expect(body, `Page should not contain error: "${pattern}"`).not.toContain(pattern);
  }
}

/** Create a PostgreSQL client for DB validation. */
export function getDbClient(): Client {
  return new Client({
    host: 'localhost',
    port: 5052,
    user: 'authforge',
    password: 'authforge_secret',
    database: 'authforge',
  });
}

/** Query the database and return rows. */
export async function dbQuery(sql: string, params: any[] = []): Promise<any[]> {
  const client = getDbClient();
  await client.connect();
  try {
    const result = await client.query(sql, params);
    return result.rows;
  } finally {
    await client.end();
  }
}

/** Make an authenticated API call. */
export async function apiCall(token: string, method: string, path: string, body?: any) {
  const opts: RequestInit = {
    method,
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
  };
  if (body) opts.body = JSON.stringify(body);
  opts.signal = AbortSignal.timeout(15000);
  const res = await fetch(`${API_URL}${path}`, opts);
  if (res.status === 204) return null;
  const text = await res.text();
  try {
    return JSON.parse(text);
  } catch {
    throw new Error(`API ${method} ${path} returned ${res.status}: ${text.slice(0, 200)}`);
  }
}
