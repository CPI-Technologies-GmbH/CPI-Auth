import { type Page, expect } from '@playwright/test';
import { Client } from 'pg';

export const API_URL = process.env.E2E_API_URL || 'http://localhost:5050';
export const LOGIN_UI = 'http://localhost:5053';
export const MAILHOG_API = 'http://localhost:8025/api';
export const ADMIN_EMAIL = 'admin@cpi-auth.local';
export const ADMIN_PASSWORD = 'admin123!';
export const TEST_PASSWORD = 'Xk9mQ2vL8nR4!pw';

export function uniqueEmail(prefix = 'e2e'): string {
  return `${prefix}-${Date.now()}-${Math.random().toString(36).slice(2, 7)}@cpi-auth.local`;
}

/** Unauthenticated API call. */
export async function publicApiCall(method: string, path: string, body?: any) {
  const opts: RequestInit = {
    method,
    headers: { 'Content-Type': 'application/json' },
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

/** Get admin bearer token. */
export async function getAdminToken(): Promise<string> {
  const res = await fetch(`${API_URL}/admin/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email: ADMIN_EMAIL, password: ADMIN_PASSWORD }),
  });
  const body = await res.json();
  if (!body.access_token) throw new Error('Admin login failed: ' + JSON.stringify(body));
  return body.access_token;
}

/** Authenticated admin API call. */
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

// ── MailHog helpers ──────────────────────────────────────────────────

export async function clearMailHog() {
  try {
    await fetch(`${MAILHOG_API}/v1/messages`, { method: 'DELETE', signal: AbortSignal.timeout(5000) });
  } catch { /* MailHog may not be running */ }
}

export async function getMailHogMessages(): Promise<any[]> {
  try {
    const res = await fetch(`${MAILHOG_API}/v2/messages`, { signal: AbortSignal.timeout(5000) });
    const data = await res.json();
    return data.items || [];
  } catch {
    return [];
  }
}

export async function waitForEmail(to: string, maxWait = 10000): Promise<any | null> {
  const start = Date.now();
  while (Date.now() - start < maxWait) {
    const messages = await getMailHogMessages();
    const match = messages.find(
      (m: any) =>
        m.Raw?.To?.some((addr: string) => addr.includes(to)) ||
        m.Content?.Headers?.To?.some((addr: string) => addr.includes(to))
    );
    if (match) return match;
    await new Promise((r) => setTimeout(r, 500));
  }
  return null;
}

export async function extractTokenFromEmail(email: any): Promise<string | null> {
  const body = email?.Content?.Body || email?.Raw?.Data || '';
  const tokenMatch = body.match(/token=([a-zA-Z0-9_-]+)/);
  return tokenMatch ? tokenMatch[1] : null;
}

// ── User management ──────────────────────────────────────────────────

export async function registerTestUser(email: string, password = TEST_PASSWORD, name = 'E2E Test User') {
  return publicApiCall('POST', '/api/v1/auth/register', { email, password, name });
}

export async function deleteTestUserByEmail(email: string) {
  try {
    const token = await getAdminToken();
    const users = await apiCall(token, 'GET', `/admin/users?email=${encodeURIComponent(email)}`);
    const list = Array.isArray(users) ? users : users?.data ?? [];
    const user = list.find((u: any) => u.email === email);
    if (user) {
      await apiCall(token, 'DELETE', `/admin/users/${user.id}`);
    }
  } catch { /* user may not exist */ }
}

// ── Utilities ────────────────────────────────────────────────────────

export async function waitForApiReady(maxRetries = 30) {
  for (let i = 0; i < maxRetries; i++) {
    try {
      const res = await fetch(`${API_URL}/health`);
      if (res.ok) return;
    } catch { /* server not ready */ }
    await new Promise((r) => setTimeout(r, 1000));
  }
  throw new Error(`API at ${API_URL} not ready after ${maxRetries}s`);
}

export async function screenshot(page: Page, name: string) {
  await page
    .screenshot({ path: `test-results/screenshots/${name}.png`, fullPage: true, timeout: 10000 })
    .catch(() => {});
}

export async function waitForLoad(page: Page, timeout = 5000) {
  await page.waitForLoadState('networkidle', { timeout }).catch(() => {});
}

export async function assertNoErrors(page: Page) {
  const body = await page.textContent('body');
  const errorPatterns = [
    'is not a function',
    'Cannot read properties of undefined',
    'Cannot read properties of null',
    'Unhandled Runtime Error',
  ];
  for (const pattern of errorPatterns) {
    expect(body, `Page should not contain error: "${pattern}"`).not.toContain(pattern);
  }
}

export function getDbClient(): Client {
  return new Client({
    host: 'localhost',
    port: 5052,
    user: 'cpi-auth',
    password: 'cpi-auth_secret',
    database: 'cpi-auth',
  });
}

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
