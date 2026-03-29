import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

// Mock the env module
vi.mock('$env/static/public', () => ({
	PUBLIC_API_URL: 'http://test-api.example.com'
}));

import { api, createServerApi } from './api';

describe('Account UI ApiClient', () => {
	const mockFetch = vi.fn();

	beforeEach(() => {
		vi.stubGlobal('fetch', mockFetch);
		mockFetch.mockReset();
	});

	afterEach(() => {
		vi.restoreAllMocks();
	});

	function mockJsonResponse(data: unknown, status = 200) {
		return {
			ok: status >= 200 && status < 300,
			status,
			headers: new Headers({ 'content-type': 'application/json' }),
			json: () => Promise.resolve(data)
		};
	}

	function mockErrorResponse(data: unknown, status: number) {
		return {
			ok: false,
			status,
			headers: new Headers({ 'content-type': 'application/json' }),
			json: () => Promise.resolve(data)
		};
	}

	// ── User Profile ──

	describe('getMe', () => {
		it('should GET /v1/users/me', async () => {
			const user = { id: 'u1', email: 'user@test.com', name: 'Test User' };
			mockFetch.mockResolvedValueOnce(mockJsonResponse(user));

			const result = await api.getMe();

			expect(mockFetch.mock.calls[0][0]).toBe('http://test-api.example.com/v1/users/me');
			const options = mockFetch.mock.calls[0][1];
			expect(options.method).toBe('GET');
			expect(result).toEqual(user);
		});
	});

	describe('updateMe', () => {
		it('should PATCH /v1/users/me', async () => {
			const updated = { id: 'u1', email: 'user@test.com', name: 'Updated' };
			mockFetch.mockResolvedValueOnce(mockJsonResponse(updated));

			const result = await api.updateMe({ name: 'Updated' });

			const [url, options] = mockFetch.mock.calls[0];
			expect(url).toBe('http://test-api.example.com/v1/users/me');
			expect(options.method).toBe('PATCH');
			expect(JSON.parse(options.body)).toEqual({ name: 'Updated' });
			expect(result).toEqual(updated);
		});
	});

	describe('changePassword', () => {
		it('should POST /v1/users/me/change-password', async () => {
			mockFetch.mockResolvedValueOnce(mockJsonResponse(undefined, 204));

			await api.changePassword({
				current_password: 'old',
				new_password: 'newStrongPass1!'
			});

			const [url, options] = mockFetch.mock.calls[0];
			expect(url).toBe('http://test-api.example.com/v1/users/me/change-password');
			expect(options.method).toBe('POST');
			expect(JSON.parse(options.body)).toEqual({
				current_password: 'old',
				new_password: 'newStrongPass1!'
			});
		});
	});

	describe('deleteAccount', () => {
		it('should DELETE /v1/users/me with password', async () => {
			mockFetch.mockResolvedValueOnce(mockJsonResponse(undefined, 204));

			await api.deleteAccount('mypassword');

			const [url, options] = mockFetch.mock.calls[0];
			expect(url).toBe('http://test-api.example.com/v1/users/me');
			expect(options.method).toBe('DELETE');
			expect(JSON.parse(options.body)).toEqual({ password: 'mypassword' });
		});

		it('should DELETE /v1/users/me without password', async () => {
			mockFetch.mockResolvedValueOnce(mockJsonResponse(undefined, 204));

			await api.deleteAccount();

			const [url, options] = mockFetch.mock.calls[0];
			expect(url).toBe('http://test-api.example.com/v1/users/me');
			expect(options.method).toBe('DELETE');
			expect(options.body).toBeUndefined();
		});
	});

	// ── Sessions ──

	describe('sessions', () => {
		it('getSessions should GET /v1/users/me/sessions', async () => {
			const sessions = [{ id: 's1', user_id: 'u1', ip_address: '1.2.3.4' }];
			mockFetch.mockResolvedValueOnce(mockJsonResponse(sessions));

			const result = await api.getSessions();

			expect(mockFetch.mock.calls[0][0]).toBe(
				'http://test-api.example.com/v1/users/me/sessions'
			);
			expect(result).toEqual(sessions);
		});

		it('revokeSession should DELETE /v1/users/me/sessions/:id', async () => {
			mockFetch.mockResolvedValueOnce(mockJsonResponse(undefined, 204));

			await api.revokeSession('s1');

			const [url, options] = mockFetch.mock.calls[0];
			expect(url).toBe('http://test-api.example.com/v1/users/me/sessions/s1');
			expect(options.method).toBe('DELETE');
		});

		it('revokeAllOtherSessions should DELETE /v1/users/me/sessions', async () => {
			mockFetch.mockResolvedValueOnce(mockJsonResponse(undefined, 204));

			await api.revokeAllOtherSessions();

			const [url, options] = mockFetch.mock.calls[0];
			expect(url).toBe('http://test-api.example.com/v1/users/me/sessions');
			expect(options.method).toBe('DELETE');
		});
	});

	// ── MFA ──

	describe('MFA', () => {
		it('getMFAMethods should GET /v1/users/me/mfa', async () => {
			const methods = [{ id: 'm1', type: 'totp', verified: true }];
			mockFetch.mockResolvedValueOnce(mockJsonResponse(methods));

			const result = await api.getMFAMethods();

			expect(mockFetch.mock.calls[0][0]).toBe('http://test-api.example.com/v1/users/me/mfa');
			expect(result).toEqual(methods);
		});

		it('enrollTOTP should POST /v1/users/me/mfa/totp/enroll', async () => {
			const enrollment = { id: 'e1', secret: 'SECRET', uri: 'otpauth://...' };
			mockFetch.mockResolvedValueOnce(mockJsonResponse(enrollment));

			const result = await api.enrollTOTP();

			expect(mockFetch.mock.calls[0][0]).toBe(
				'http://test-api.example.com/v1/users/me/mfa/totp/enroll'
			);
			expect(result).toEqual(enrollment);
		});

		it('verifyTOTP should POST /v1/users/me/mfa/totp/verify', async () => {
			mockFetch.mockResolvedValueOnce(mockJsonResponse(undefined, 204));

			await api.verifyTOTP('e1', '123456');

			const [url, options] = mockFetch.mock.calls[0];
			expect(url).toBe('http://test-api.example.com/v1/users/me/mfa/totp/verify');
			expect(JSON.parse(options.body)).toEqual({ id: 'e1', code: '123456' });
		});

		it('removeMFA should DELETE /v1/users/me/mfa/:id', async () => {
			mockFetch.mockResolvedValueOnce(mockJsonResponse(undefined, 204));

			await api.removeMFA('m1');

			const [url, options] = mockFetch.mock.calls[0];
			expect(url).toBe('http://test-api.example.com/v1/users/me/mfa/m1');
			expect(options.method).toBe('DELETE');
		});

		it('getRecoveryCodes should GET /v1/users/me/mfa/recovery-codes', async () => {
			const codes = { codes: ['code1', 'code2'], generated_at: '2024-01-01' };
			mockFetch.mockResolvedValueOnce(mockJsonResponse(codes));

			const result = await api.getRecoveryCodes();

			expect(result).toEqual(codes);
		});

		it('regenerateRecoveryCodes should POST', async () => {
			const codes = { codes: ['new1', 'new2'], generated_at: '2024-01-01' };
			mockFetch.mockResolvedValueOnce(mockJsonResponse(codes));

			const result = await api.regenerateRecoveryCodes();

			expect(mockFetch.mock.calls[0][1].method).toBe('POST');
			expect(result).toEqual(codes);
		});
	});

	// ── Passkeys ──

	describe('passkeys', () => {
		it('getPasskeys should GET /v1/users/me/passkeys', async () => {
			mockFetch.mockResolvedValueOnce(mockJsonResponse([]));
			const result = await api.getPasskeys();
			expect(result).toEqual([]);
		});

		it('deletePasskey should DELETE /v1/users/me/passkeys/:id', async () => {
			mockFetch.mockResolvedValueOnce(mockJsonResponse(undefined, 204));

			await api.deletePasskey('pk1');

			const [url, options] = mockFetch.mock.calls[0];
			expect(url).toBe('http://test-api.example.com/v1/users/me/passkeys/pk1');
			expect(options.method).toBe('DELETE');
		});

		it('renamePasskey should PATCH /v1/users/me/passkeys/:id', async () => {
			const passkey = { id: 'pk1', name: 'New Name', credential_id: 'cred1' };
			mockFetch.mockResolvedValueOnce(mockJsonResponse(passkey));

			const result = await api.renamePasskey('pk1', 'New Name');

			const [url, options] = mockFetch.mock.calls[0];
			expect(url).toBe('http://test-api.example.com/v1/users/me/passkeys/pk1');
			expect(options.method).toBe('PATCH');
			expect(JSON.parse(options.body)).toEqual({ name: 'New Name' });
			expect(result).toEqual(passkey);
		});
	});

	// ── Identities ──

	describe('identities', () => {
		it('getIdentities should GET /v1/users/me/identities', async () => {
			const identities = [{ id: 'i1', provider: 'google', email: 'user@gmail.com' }];
			mockFetch.mockResolvedValueOnce(mockJsonResponse(identities));

			const result = await api.getIdentities();
			expect(result).toEqual(identities);
		});

		it('unlinkIdentity should DELETE /v1/users/me/identities/:id', async () => {
			mockFetch.mockResolvedValueOnce(mockJsonResponse(undefined, 204));

			await api.unlinkIdentity('i1');

			expect(mockFetch.mock.calls[0][1].method).toBe('DELETE');
		});
	});

	// ── Organizations ──

	describe('organizations', () => {
		it('getOrganizations should GET /v1/users/me/organizations', async () => {
			mockFetch.mockResolvedValueOnce(mockJsonResponse([]));
			const result = await api.getOrganizations();
			expect(result).toEqual([]);
		});
	});

	// ── Activity ──

	describe('activity', () => {
		it('getActivity should GET /v1/users/me/activity', async () => {
			mockFetch.mockResolvedValueOnce(mockJsonResponse([]));
			const result = await api.getActivity();

			expect(mockFetch.mock.calls[0][0]).toBe(
				'http://test-api.example.com/v1/users/me/activity'
			);
			expect(result).toEqual([]);
		});

		it('getActivity should include query params', async () => {
			mockFetch.mockResolvedValueOnce(mockJsonResponse([]));

			await api.getActivity({ type: 'login', limit: 10, offset: 20 });

			const url = mockFetch.mock.calls[0][0] as string;
			expect(url).toContain('type=login');
			expect(url).toContain('limit=10');
			expect(url).toContain('offset=20');
		});

		it('getActivity should omit undefined params', async () => {
			mockFetch.mockResolvedValueOnce(mockJsonResponse([]));

			await api.getActivity({ type: 'login' });

			const url = mockFetch.mock.calls[0][0] as string;
			expect(url).toContain('type=login');
			expect(url).not.toContain('limit');
			expect(url).not.toContain('offset');
		});
	});

	// ── Consents ──

	describe('consents', () => {
		it('getConsents should GET /v1/users/me/consents', async () => {
			mockFetch.mockResolvedValueOnce(mockJsonResponse([]));
			const result = await api.getConsents();
			expect(result).toEqual([]);
		});

		it('revokeConsent should DELETE /v1/users/me/consents/:id', async () => {
			mockFetch.mockResolvedValueOnce(mockJsonResponse(undefined, 204));

			await api.revokeConsent('c1');

			const [url, options] = mockFetch.mock.calls[0];
			expect(url).toBe('http://test-api.example.com/v1/users/me/consents/c1');
			expect(options.method).toBe('DELETE');
		});
	});

	// ── Branding ──

	describe('branding', () => {
		it('getBranding should GET /v1/branding', async () => {
			const branding = { primary_color: '#000', company_name: 'Test' };
			mockFetch.mockResolvedValueOnce(mockJsonResponse(branding));

			const result = await api.getBranding();
			expect(result).toEqual(branding);
		});
	});

	// ── Data Export ──

	describe('data export', () => {
		it('requestDataExport should POST /v1/users/me/export', async () => {
			const exportReq = { id: 'exp1', status: 'pending', requested_at: '2024-01-01' };
			mockFetch.mockResolvedValueOnce(mockJsonResponse(exportReq));

			const result = await api.requestDataExport();

			expect(mockFetch.mock.calls[0][1].method).toBe('POST');
			expect(result).toEqual(exportReq);
		});
	});

	// ── Session token passing ──

	describe('session token passing', () => {
		it('should include Authorization header when token is set', async () => {
			api.setToken('my-session-token');
			mockFetch.mockResolvedValueOnce(mockJsonResponse({ id: 'u1' }));

			await api.getMe();

			const options = mockFetch.mock.calls[0][1];
			expect(options.headers['Authorization']).toBe('Bearer my-session-token');

			// Clean up
			api.setToken(null);
		});

		it('should not include Authorization header when no token is set', async () => {
			api.setToken(null);
			mockFetch.mockResolvedValueOnce(mockJsonResponse({ id: 'u1' }));

			await api.getMe();

			const options = mockFetch.mock.calls[0][1];
			expect(options.headers['Authorization']).toBeUndefined();
		});
	});

	// ── createServerApi ──

	describe('createServerApi', () => {
		it('should create a new ApiClient with token set', async () => {
			const serverApi = createServerApi('server-token');
			mockFetch.mockResolvedValueOnce(mockJsonResponse({ id: 'u1' }));

			await serverApi.getMe();

			const options = mockFetch.mock.calls[0][1];
			expect(options.headers['Authorization']).toBe('Bearer server-token');
		});
	});

	// ── Error handling ──

	describe('error handling', () => {
		it('should throw API error object on non-ok response', async () => {
			mockFetch.mockResolvedValueOnce(
				mockErrorResponse(
					{ error: 'unauthorized', message: 'Invalid session', status: 401 },
					401
				)
			);

			await expect(api.getMe()).rejects.toMatchObject({
				error: 'unauthorized',
				message: 'Invalid session',
				status: 401
			});
		});

		it('should handle non-JSON error responses', async () => {
			mockFetch.mockResolvedValueOnce({
				ok: false,
				status: 500,
				json: () => Promise.reject(new Error('not json'))
			});

			await expect(api.getMe()).rejects.toMatchObject({
				error: 'unknown_error',
				status: 500
			});
		});

		it('should handle 204 No Content responses', async () => {
			mockFetch.mockResolvedValueOnce({
				ok: true,
				status: 204,
				json: () => Promise.reject(new Error('no body'))
			});

			const result = await api.revokeSession('s1');
			expect(result).toBeUndefined();
		});

		it('should include credentials: include in requests', async () => {
			mockFetch.mockResolvedValueOnce(mockJsonResponse({ id: 'u1' }));

			await api.getMe();

			const options = mockFetch.mock.calls[0][1];
			expect(options.credentials).toBe('include');
		});
	});
});
