import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

// Mock $env/dynamic/public before importing
vi.mock('$env/dynamic/public', () => ({
	env: { PUBLIC_API_URL: 'http://test-api.example.com' }
}));

import { api, ApiClientError } from './client';

describe('CPIAuthApiClient', () => {
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

	function mockNonJsonErrorResponse(status: number) {
		return {
			ok: false,
			status,
			headers: new Headers({ 'content-type': 'text/html' }),
			json: () => Promise.reject(new Error('not json'))
		};
	}

	// ── login ──

	describe('login', () => {
		it('should send POST /api/v1/auth/login with credentials', async () => {
			const loginResponse = { access_token: 'tok123', redirect_url: '/dashboard' };
			mockFetch.mockResolvedValueOnce(mockJsonResponse(loginResponse));

			const result = await api.login({ email: 'user@test.com', password: 'Pass1234' });

			expect(mockFetch).toHaveBeenCalledOnce();
			const [url, options] = mockFetch.mock.calls[0];
			expect(url).toBe('http://test-api.example.com/api/v1/auth/login');
			expect(options.method).toBe('POST');
			expect(JSON.parse(options.body)).toEqual({ email: 'user@test.com', password: 'Pass1234' });
			expect(options.credentials).toBe('include');
			expect(result).toEqual(loginResponse);
		});

		it('should send optional OAuth fields', async () => {
			mockFetch.mockResolvedValueOnce(mockJsonResponse({ redirect_url: '/callback' }));

			await api.login({
				email: 'user@test.com',
				password: 'Pass1234',
				client_id: 'app1',
				redirect_uri: 'http://localhost/callback',
				remember_me: true
			});

			const body = JSON.parse(mockFetch.mock.calls[0][1].body);
			expect(body.client_id).toBe('app1');
			expect(body.redirect_uri).toBe('http://localhost/callback');
			expect(body.remember_me).toBe(true);
		});

		it('should throw ApiClientError on 401', async () => {
			mockFetch.mockResolvedValueOnce(
				mockErrorResponse(
					{ error: 'invalid_credentials', error_description: 'Invalid email or password' },
					401
				)
			);

			await expect(api.login({ email: 'user@test.com', password: 'wrong' })).rejects.toThrow(
				ApiClientError
			);

			try {
				mockFetch.mockResolvedValueOnce(
					mockErrorResponse(
						{ error: 'invalid_credentials', error_description: 'Invalid email or password' },
						401
					)
				);
				await api.login({ email: 'user@test.com', password: 'wrong' });
			} catch (e) {
				const err = e as ApiClientError;
				expect(err.code).toBe('invalid_credentials');
				expect(err.statusCode).toBe(401);
				expect(err.message).toBe('Invalid email or password');
			}
		});
	});

	// ── register ──

	describe('register', () => {
		it('should send POST /api/v1/auth/register', async () => {
			const registerResponse = { access_token: 'tok456' };
			mockFetch.mockResolvedValueOnce(mockJsonResponse(registerResponse));

			const result = await api.register({
				email: 'new@test.com',
				password: 'StrongPass1!',
				name: 'New User'
			});

			const [url, options] = mockFetch.mock.calls[0];
			expect(url).toBe('http://test-api.example.com/api/v1/auth/register');
			expect(options.method).toBe('POST');
			expect(JSON.parse(options.body)).toEqual({
				email: 'new@test.com',
				password: 'StrongPass1!',
				name: 'New User'
			});
			expect(result).toEqual(registerResponse);
		});

		it('should handle registration with custom fields', async () => {
			mockFetch.mockResolvedValueOnce(mockJsonResponse({ access_token: 'tok789' }));

			await api.register({
				email: 'new@test.com',
				password: 'StrongPass1!',
				name: 'New User',
				custom_fields: { company: 'Acme Inc' }
			});

			const body = JSON.parse(mockFetch.mock.calls[0][1].body);
			expect(body.custom_fields).toEqual({ company: 'Acme Inc' });
		});
	});

	// ── forgotPassword ──

	describe('forgotPassword', () => {
		it('should send POST /api/v1/auth/forgot-password', async () => {
			const response = { message: 'Password reset email sent' };
			mockFetch.mockResolvedValueOnce(mockJsonResponse(response));

			const result = await api.forgotPassword({ email: 'user@test.com' });

			const [url, options] = mockFetch.mock.calls[0];
			expect(url).toBe('http://test-api.example.com/api/v1/auth/forgot-password');
			expect(options.method).toBe('POST');
			expect(result).toEqual(response);
		});

		it('should include client_id if provided', async () => {
			mockFetch.mockResolvedValueOnce(mockJsonResponse({ message: 'sent' }));

			await api.forgotPassword({ email: 'user@test.com', client_id: 'app1' });

			const body = JSON.parse(mockFetch.mock.calls[0][1].body);
			expect(body.client_id).toBe('app1');
		});
	});

	// ── resetPassword ──

	describe('resetPassword', () => {
		it('should send POST /api/v1/auth/reset-password', async () => {
			const response = { message: 'Password reset successful' };
			mockFetch.mockResolvedValueOnce(mockJsonResponse(response));

			const result = await api.resetPassword({ token: 'reset-tok', password: 'NewPass1!' });

			const [url, options] = mockFetch.mock.calls[0];
			expect(url).toBe('http://test-api.example.com/api/v1/auth/reset-password');
			expect(options.method).toBe('POST');
			expect(JSON.parse(options.body)).toEqual({ token: 'reset-tok', password: 'NewPass1!' });
			expect(result).toEqual(response);
		});
	});

	// ── getBranding ──

	describe('getBranding', () => {
		it('should GET /api/v1/branding without params', async () => {
			const branding = { primary_color: '#000', app_name: 'MyApp' };
			mockFetch.mockResolvedValueOnce(mockJsonResponse(branding));

			const result = await api.getBranding();

			expect(mockFetch.mock.calls[0][0]).toBe('http://test-api.example.com/api/v1/branding');
			expect(result).toEqual(branding);
		});

		it('should GET /api/v1/branding with client_id query param', async () => {
			mockFetch.mockResolvedValueOnce(mockJsonResponse({ primary_color: '#fff' }));

			await api.getBranding('my-client');

			expect(mockFetch.mock.calls[0][0]).toBe(
				'http://test-api.example.com/api/v1/branding?client_id=my-client'
			);
		});

		it('should encode special characters in client_id', async () => {
			mockFetch.mockResolvedValueOnce(mockJsonResponse({}));

			await api.getBranding('client with spaces');

			expect(mockFetch.mock.calls[0][0]).toContain('client_id=client%20with%20spaces');
		});
	});

	// ── verifyEmail ──

	describe('verifyEmail', () => {
		it('should send POST /api/v1/auth/verify-email', async () => {
			mockFetch.mockResolvedValueOnce(mockJsonResponse({ message: 'Verified' }));

			const result = await api.verifyEmail({ token: 'verify-tok' });

			expect(mockFetch.mock.calls[0][0]).toBe(
				'http://test-api.example.com/api/v1/auth/verify-email'
			);
			expect(result).toEqual({ message: 'Verified' });
		});
	});

	// ── MFA ──

	describe('MFA', () => {
		it('mfaChallenge should POST to /api/v1/auth/mfa/challenge', async () => {
			const response = { challenge_id: 'ch1', method: 'totp', expires_at: '2026-01-01' };
			mockFetch.mockResolvedValueOnce(mockJsonResponse(response));

			const result = await api.mfaChallenge({ mfa_token: 'mfa-tok' });

			expect(mockFetch.mock.calls[0][0]).toBe(
				'http://test-api.example.com/api/v1/auth/mfa/challenge'
			);
			expect(result).toEqual(response);
		});

		it('mfaVerify should POST to /api/v1/auth/mfa/verify', async () => {
			const response = { access_token: 'verified-tok' };
			mockFetch.mockResolvedValueOnce(mockJsonResponse(response));

			const result = await api.mfaVerify({
				mfa_token: 'mfa-tok',
				code: '123456',
				method: 'totp'
			});

			expect(result).toEqual(response);
		});

		it('mfaEnroll should POST to /api/v1/auth/mfa/enroll', async () => {
			const response = {
				secret: 'JBSWY3DPEHPK3PXP',
				qr_code: 'data:image/png;base64,...',
				recovery_codes: ['code1', 'code2'],
				otpauth_url: 'otpauth://totp/...'
			};
			mockFetch.mockResolvedValueOnce(mockJsonResponse(response));

			const result = await api.mfaEnroll({ mfa_token: 'mfa-tok', method: 'totp' });

			expect(result).toEqual(response);
		});
	});

	// ── Passwordless ──

	describe('Passwordless', () => {
		it('passwordlessStart should POST to /api/v1/auth/passwordless/start', async () => {
			const response = { message: 'Code sent', method: 'email_otp', expires_at: '2026-01-01' };
			mockFetch.mockResolvedValueOnce(mockJsonResponse(response));

			const result = await api.passwordlessStart({
				email: 'user@test.com',
				method: 'email_otp'
			});

			expect(mockFetch.mock.calls[0][0]).toBe(
				'http://test-api.example.com/api/v1/auth/passwordless/start'
			);
			expect(result).toEqual(response);
		});

		it('passwordlessVerify should POST to /api/v1/auth/passwordless/verify', async () => {
			mockFetch.mockResolvedValueOnce(mockJsonResponse({ access_token: 'pl-tok' }));

			const result = await api.passwordlessVerify({ token: 'pl-token', code: '123456' });

			expect(result.access_token).toBe('pl-tok');
		});
	});

	// ── Consent ──

	describe('Consent', () => {
		it('getConsentInfo should GET with consent_challenge param', async () => {
			const consentInfo = {
				client_name: 'TestApp',
				client_logo: '',
				client_uri: '',
				requested_scopes: [],
				subject: 'user1'
			};
			mockFetch.mockResolvedValueOnce(mockJsonResponse(consentInfo));

			const result = await api.getConsentInfo('challenge123');

			expect(mockFetch.mock.calls[0][0]).toContain('consent_challenge=challenge123');
			expect(result).toEqual(consentInfo);
		});

		it('submitConsent should POST consent decision', async () => {
			mockFetch.mockResolvedValueOnce(mockJsonResponse({ redirect_url: '/done' }));

			await api.submitConsent({
				consent_challenge: 'ch1',
				grant: true,
				remember: true,
				scopes: ['openid', 'email']
			});

			const body = JSON.parse(mockFetch.mock.calls[0][1].body);
			expect(body.grant).toBe(true);
			expect(body.remember).toBe(true);
			expect(body.scopes).toEqual(['openid', 'email']);
		});
	});

	// ── Error handling ──

	describe('Error handling', () => {
		it('should throw ApiClientError with error code from JSON error response', async () => {
			mockFetch.mockResolvedValueOnce(
				mockErrorResponse(
					{
						error: 'account_locked',
						error_description: 'Account is locked',
						status_code: 403
					},
					403
				)
			);

			try {
				await api.login({ email: 'user@test.com', password: 'pass' });
				expect.unreachable('Should have thrown');
			} catch (e) {
				const err = e as ApiClientError;
				expect(err).toBeInstanceOf(ApiClientError);
				expect(err.code).toBe('account_locked');
				expect(err.message).toBe('Account is locked');
				expect(err.statusCode).toBe(403);
			}
		});

		it('should handle non-JSON error responses', async () => {
			mockFetch.mockResolvedValueOnce(mockNonJsonErrorResponse(500));

			try {
				await api.login({ email: 'user@test.com', password: 'pass' });
				expect.unreachable('Should have thrown');
			} catch (e) {
				const err = e as ApiClientError;
				expect(err).toBeInstanceOf(ApiClientError);
				expect(err.code).toBe('request_failed');
				expect(err.statusCode).toBe(500);
				expect(err.message).toContain('500');
			}
		});

		it('should handle network errors', async () => {
			mockFetch.mockRejectedValueOnce(new TypeError('Failed to fetch'));

			await expect(
				api.login({ email: 'user@test.com', password: 'pass' })
			).rejects.toThrow('Failed to fetch');
		});

		it('should handle empty response body for non-JSON content type', async () => {
			mockFetch.mockResolvedValueOnce({
				ok: true,
				status: 200,
				headers: new Headers({ 'content-type': 'text/plain' }),
				json: () => Promise.resolve({})
			});

			const result = await api.getBranding();
			expect(result).toEqual({});
		});

		it('should use error field as message when error_description is missing', async () => {
			mockFetch.mockResolvedValueOnce(
				mockErrorResponse({ error: 'server_error' }, 500)
			);

			try {
				await api.login({ email: 'user@test.com', password: 'pass' });
				expect.unreachable('Should have thrown');
			} catch (e) {
				const err = e as ApiClientError;
				expect(err.message).toBe('server_error');
			}
		});
	});

	// ── ApiClientError class ──

	describe('ApiClientError', () => {
		it('should have correct name', () => {
			const error = new ApiClientError('test error', 'test_code', 400);
			expect(error.name).toBe('ApiClientError');
		});

		it('should be instanceof Error', () => {
			const error = new ApiClientError('msg', 'code', 500);
			expect(error).toBeInstanceOf(Error);
		});

		it('should store code and statusCode', () => {
			const error = new ApiClientError('msg', 'my_code', 422);
			expect(error.code).toBe('my_code');
			expect(error.statusCode).toBe(422);
			expect(error.message).toBe('msg');
		});
	});
});
