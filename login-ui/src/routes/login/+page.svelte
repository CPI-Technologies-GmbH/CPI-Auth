<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { t } from '$lib/i18n';
	import { api, ApiClientError } from '$lib/api';
	import { branding, extractOAuthParams, oauthParams } from '$lib/stores';
	import { validateEmail, validatePassword } from '$lib/utils/validation';
	import { redirectTo, getErrorMessage } from '$lib/utils';
	import AuthLayout from '$lib/components/AuthLayout.svelte';
	import PasswordInput from '$lib/components/PasswordInput.svelte';
	import SocialButtons from '$lib/components/SocialButtons.svelte';
	import Alert from '$lib/components/Alert.svelte';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	let email = $state('');
	let password = $state('');
	let rememberMe = $state(false);
	let loading = $state(false);
	let error = $state('');
	let emailError = $state('');
	let passwordError = $state('');

	let oauth = $derived(extractOAuthParams(new URL($page.url)));
	let loginHint = $derived($page.url.searchParams.get('login_hint') || '');
	let providers = $derived($branding?.social_providers || []);
	let passkeysEnabled = $derived($branding?.passkeys_enabled ?? false);
	let magicLinkEnabled = $derived($branding?.magic_link_enabled ?? false);

	$effect(() => {
		if (oauth) oauthParams.set(oauth);
		if (loginHint) email = loginHint;
	});

	function validate(): boolean {
		let valid = true;
		const emailResult = validateEmail(email);
		if (!emailResult.valid) {
			emailError = emailResult.error || '';
			valid = false;
		} else {
			emailError = '';
		}

		const passResult = validatePassword(password);
		if (!passResult.valid) {
			passwordError = passResult.error || '';
			valid = false;
		} else {
			passwordError = '';
		}

		return valid;
	}

	async function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		error = '';

		if (!validate()) return;

		loading = true;

		try {
			const result = await api.login({
				email,
				password,
				remember_me: rememberMe,
				...(oauth && {
					client_id: oauth.client_id,
					redirect_uri: oauth.redirect_uri,
					scope: oauth.scope,
					state: oauth.state,
					code_challenge: oauth.code_challenge,
					code_challenge_method: oauth.code_challenge_method,
					response_type: oauth.response_type
				})
			});

			if (result.mfa_required && result.mfa_token) {
				const params = new URLSearchParams({ mfa_token: result.mfa_token });
				if (result.mfa_methods) {
					params.set('methods', result.mfa_methods.join(','));
				}
				if (oauth) {
					params.set('client_id', oauth.client_id);
					params.set('redirect_uri', oauth.redirect_uri);
				}
				await goto(`/mfa?${params.toString()}`);
				return;
			}

			if (result.redirect_url) {
				redirectTo(result.redirect_url);
			} else if (oauth?.redirect_uri) {
				redirectTo(oauth.redirect_uri);
			} else {
				// No OAuth flow — redirect to admin dashboard
				redirectTo('/');
			}
		} catch (err) {
			if (err instanceof ApiClientError) {
				if (err.code === 'account_locked') {
					error = 'login.error.locked';
				} else if (err.code === 'email_unverified') {
					error = 'login.error.unverified';
				} else {
					error = err.message;
				}
			} else {
				error = getErrorMessage(err);
			}
		} finally {
			loading = false;
		}
	}

	async function handlePasskeyLogin() {
		loading = true;
		error = '';

		try {
			const beginResult = await api.webAuthnLoginBegin({
				...(oauth && {
					client_id: oauth.client_id,
					redirect_uri: oauth.redirect_uri
				})
			});

			// Use Web Authentication API
			const credential = await navigator.credentials.get({
				publicKey: beginResult.publicKey as PublicKeyCredentialRequestOptions
			});

			const result = await api.webAuthnLoginFinish({
				credential,
				...(oauth && {
					client_id: oauth.client_id,
					redirect_uri: oauth.redirect_uri
				})
			});

			if (result.redirect_url) {
				redirectTo(result.redirect_url);
			}
		} catch (err) {
			error = getErrorMessage(err);
		} finally {
			loading = false;
		}
	}

	function handleMagicLink() {
		const params = new URLSearchParams();
		if (email) params.set('email', email);
		if (oauth) {
			params.set('client_id', oauth.client_id);
			params.set('redirect_uri', oauth.redirect_uri);
			params.set('scope', oauth.scope);
			params.set('state', oauth.state);
		}
		goto(`/passwordless?${params.toString()}`);
	}

	function buildLink(path: string): string {
		const params = new URLSearchParams();
		if (oauth) {
			params.set('client_id', oauth.client_id);
			params.set('redirect_uri', oauth.redirect_uri);
			params.set('scope', oauth.scope);
			params.set('state', oauth.state);
		}
		const qs = params.toString();
		return qs ? `${path}?${qs}` : path;
	}
</script>

<svelte:head>
	<title>{$t('login.title')}</title>
</svelte:head>

<AuthLayout>
	<div class="space-y-6">
		<div class="text-center">
			<h1 class="text-2xl font-bold" style="color: var(--af-color-text)">
				{$t('login.title')}
			</h1>
		</div>

		{#if error}
			<Alert type="error" message={error.startsWith('login.') ? $t(error) : error} dismissible />
		{/if}

		<form onsubmit={handleSubmit} class="space-y-4" novalidate>
			<div>
				<label for="email" class="af-label">{$t('login.email')}</label>
				<input
					id="email"
					name="email"
					type="email"
					bind:value={email}
					placeholder={$t('login.email.placeholder')}
					autocomplete="email"
					required
					disabled={loading}
					class="af-input"
					class:error={!!emailError}
					aria-invalid={!!emailError}
					aria-describedby={emailError ? 'email-error' : undefined}
				/>
				{#if emailError}
					<p id="email-error" class="mt-1 text-sm" style="color: var(--af-color-error)" role="alert">
						{$t(emailError)}
					</p>
				{/if}
			</div>

			<PasswordInput
				bind:value={password}
				label={$t('login.password')}
				placeholder={$t('login.password.placeholder')}
				error={passwordError}
				disabled={loading}
				autocomplete="current-password"
			/>

			<div class="flex items-center justify-between">
				<label class="flex items-center gap-2 text-sm" style="color: var(--af-color-text)">
					<input type="checkbox" bind:checked={rememberMe} class="af-checkbox" disabled={loading} />
					{$t('login.remember')}
				</label>
				<a href={buildLink('/forgot-password')} class="af-link text-sm">
					{$t('login.forgot')}
				</a>
			</div>

			<button
				type="submit"
				class="af-btn af-btn-primary w-full"
				disabled={loading}
			>
				{#if loading}
					<LoadingSpinner size="sm" />
				{/if}
				{$t('login.submit')}
			</button>
		</form>

		{#if passkeysEnabled}
			<button
				type="button"
				class="af-btn af-btn-secondary w-full"
				onclick={handlePasskeyLogin}
				disabled={loading}
			>
				<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" aria-hidden="true">
					<path stroke-linecap="round" stroke-linejoin="round" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
				</svg>
				{$t('login.passkey')}
			</button>
		{/if}

		{#if magicLinkEnabled}
			<button
				type="button"
				class="af-btn af-btn-ghost w-full"
				onclick={handleMagicLink}
				disabled={loading}
			>
				<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" aria-hidden="true">
					<path stroke-linecap="round" stroke-linejoin="round" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
				</svg>
				{$t('login.magic_link')}
			</button>
		{/if}

		{#if providers.length > 0}
			<div class="af-divider">{$t('login.or')}</div>
			<SocialButtons {providers} oauthParams={oauth} mode="login" />
		{/if}

		<p class="text-center text-sm" style="color: var(--af-color-text-muted)">
			{$t('login.no_account')}
			<a href={buildLink('/register')} class="af-link">{$t('login.signup_link')}</a>
		</p>
	</div>
</AuthLayout>
