import { writable } from 'svelte/store';

export interface OAuthParams {
	client_id: string;
	redirect_uri: string;
	response_type: string;
	scope: string;
	state: string;
	code_challenge?: string;
	code_challenge_method?: string;
	nonce?: string;
	prompt?: string;
	login_hint?: string;
}

export const oauthParams = writable<OAuthParams | null>(null);

export function extractOAuthParams(url: URL): OAuthParams | null {
	const client_id = url.searchParams.get('client_id');
	const redirect_uri = url.searchParams.get('redirect_uri');
	const response_type = url.searchParams.get('response_type');
	const scope = url.searchParams.get('scope');
	const state = url.searchParams.get('state');

	if (!client_id || !redirect_uri) return null;

	return {
		client_id,
		redirect_uri,
		response_type: response_type || 'code',
		scope: scope || 'openid',
		state: state || '',
		code_challenge: url.searchParams.get('code_challenge') || undefined,
		code_challenge_method: url.searchParams.get('code_challenge_method') || undefined,
		nonce: url.searchParams.get('nonce') || undefined,
		prompt: url.searchParams.get('prompt') || undefined,
		login_hint: url.searchParams.get('login_hint') || undefined
	};
}

export function buildCallbackUrl(redirectUri: string, code: string, state: string): string {
	const url = new URL(redirectUri);
	url.searchParams.set('code', code);
	if (state) url.searchParams.set('state', state);
	return url.toString();
}
