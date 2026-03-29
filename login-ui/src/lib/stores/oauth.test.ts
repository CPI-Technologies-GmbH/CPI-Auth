import { describe, it, expect } from 'vitest';
import { get } from 'svelte/store';
import { oauthParams, extractOAuthParams, buildCallbackUrl } from './oauth';

describe('OAuth store', () => {
	describe('oauthParams store', () => {
		it('should initialize as null', () => {
			expect(get(oauthParams)).toBeNull();
		});

		it('should allow setting OAuth params', () => {
			oauthParams.set({
				client_id: 'app1',
				redirect_uri: 'http://localhost/callback',
				response_type: 'code',
				scope: 'openid',
				state: 'state123'
			});
			const val = get(oauthParams);
			expect(val?.client_id).toBe('app1');
			expect(val?.redirect_uri).toBe('http://localhost/callback');
			// Clean up
			oauthParams.set(null);
		});
	});

	describe('extractOAuthParams', () => {
		it('should extract all OAuth params from URL', () => {
			const url = new URL(
				'http://localhost/login?client_id=myapp&redirect_uri=http://localhost/cb&response_type=code&scope=openid+email&state=abc123'
			);
			const result = extractOAuthParams(url);

			expect(result).not.toBeNull();
			expect(result!.client_id).toBe('myapp');
			expect(result!.redirect_uri).toBe('http://localhost/cb');
			expect(result!.response_type).toBe('code');
			expect(result!.scope).toBe('openid email');
			expect(result!.state).toBe('abc123');
		});

		it('should return null when client_id is missing', () => {
			const url = new URL('http://localhost/login?redirect_uri=http://localhost/cb');
			expect(extractOAuthParams(url)).toBeNull();
		});

		it('should return null when redirect_uri is missing', () => {
			const url = new URL('http://localhost/login?client_id=myapp');
			expect(extractOAuthParams(url)).toBeNull();
		});

		it('should return null when both client_id and redirect_uri are missing', () => {
			const url = new URL('http://localhost/login');
			expect(extractOAuthParams(url)).toBeNull();
		});

		it('should default response_type to "code"', () => {
			const url = new URL(
				'http://localhost/login?client_id=myapp&redirect_uri=http://localhost/cb'
			);
			const result = extractOAuthParams(url);
			expect(result!.response_type).toBe('code');
		});

		it('should default scope to "openid"', () => {
			const url = new URL(
				'http://localhost/login?client_id=myapp&redirect_uri=http://localhost/cb'
			);
			const result = extractOAuthParams(url);
			expect(result!.scope).toBe('openid');
		});

		it('should default state to empty string', () => {
			const url = new URL(
				'http://localhost/login?client_id=myapp&redirect_uri=http://localhost/cb'
			);
			const result = extractOAuthParams(url);
			expect(result!.state).toBe('');
		});

		it('should extract optional PKCE params', () => {
			const url = new URL(
				'http://localhost/login?client_id=myapp&redirect_uri=http://localhost/cb&code_challenge=abc&code_challenge_method=S256'
			);
			const result = extractOAuthParams(url);
			expect(result!.code_challenge).toBe('abc');
			expect(result!.code_challenge_method).toBe('S256');
		});

		it('should extract nonce param', () => {
			const url = new URL(
				'http://localhost/login?client_id=myapp&redirect_uri=http://localhost/cb&nonce=n123'
			);
			const result = extractOAuthParams(url);
			expect(result!.nonce).toBe('n123');
		});

		it('should extract prompt param', () => {
			const url = new URL(
				'http://localhost/login?client_id=myapp&redirect_uri=http://localhost/cb&prompt=consent'
			);
			const result = extractOAuthParams(url);
			expect(result!.prompt).toBe('consent');
		});

		it('should extract login_hint param', () => {
			const url = new URL(
				'http://localhost/login?client_id=myapp&redirect_uri=http://localhost/cb&login_hint=user@example.com'
			);
			const result = extractOAuthParams(url);
			expect(result!.login_hint).toBe('user@example.com');
		});

		it('should set optional params to undefined when not present', () => {
			const url = new URL(
				'http://localhost/login?client_id=myapp&redirect_uri=http://localhost/cb'
			);
			const result = extractOAuthParams(url);
			expect(result!.code_challenge).toBeUndefined();
			expect(result!.code_challenge_method).toBeUndefined();
			expect(result!.nonce).toBeUndefined();
			expect(result!.prompt).toBeUndefined();
			expect(result!.login_hint).toBeUndefined();
		});
	});

	describe('buildCallbackUrl', () => {
		it('should add code and state to redirect URI', () => {
			const result = buildCallbackUrl('http://localhost/callback', 'authcode123', 'state456');
			const url = new URL(result);
			expect(url.searchParams.get('code')).toBe('authcode123');
			expect(url.searchParams.get('state')).toBe('state456');
		});

		it('should not add state param when state is empty', () => {
			const result = buildCallbackUrl('http://localhost/callback', 'authcode123', '');
			const url = new URL(result);
			expect(url.searchParams.get('code')).toBe('authcode123');
			expect(url.searchParams.has('state')).toBe(false);
		});

		it('should preserve existing query params on redirect URI', () => {
			const result = buildCallbackUrl(
				'http://localhost/callback?existing=true',
				'authcode123',
				'state456'
			);
			const url = new URL(result);
			expect(url.searchParams.get('existing')).toBe('true');
			expect(url.searchParams.get('code')).toBe('authcode123');
			expect(url.searchParams.get('state')).toBe('state456');
		});

		it('should handle redirect URI with port', () => {
			const result = buildCallbackUrl('http://localhost:3000/cb', 'code1', 'st1');
			const url = new URL(result);
			expect(url.port).toBe('3000');
			expect(url.searchParams.get('code')).toBe('code1');
		});

		it('should handle redirect URI with path segments', () => {
			const result = buildCallbackUrl(
				'https://app.example.com/auth/callback',
				'code1',
				'st1'
			);
			const url = new URL(result);
			expect(url.pathname).toBe('/auth/callback');
			expect(url.searchParams.get('code')).toBe('code1');
		});
	});
});
