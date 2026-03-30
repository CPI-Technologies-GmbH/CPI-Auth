<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { t } from '$lib/i18n';
	import { api, ApiClientError } from '$lib/api';
	import { extractOAuthParams } from '$lib/stores';
	import { validateOTP } from '$lib/utils/validation';
	import { redirectTo, getErrorMessage } from '$lib/utils';
	import AuthLayout from '$lib/components/AuthLayout.svelte';
	import OTPInput from '$lib/components/OTPInput.svelte';
	import Alert from '$lib/components/Alert.svelte';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	type MfaMode = 'totp' | 'sms' | 'email' | 'recovery';

	let code = $state('');
	let recoveryCode = $state('');
	let loading = $state(false);
	let sendingChallenge = $state(false);
	let error = $state('');
	let codeError = $state('');
	let mode = $state<MfaMode>('totp');
	let challengeSent = $state(false);

	let mfaToken = $derived($page.url.searchParams.get('mfa_token') || '');
	let methods = $derived(($page.url.searchParams.get('methods') || 'totp').split(',') as MfaMode[]);
	let oauth = $derived(extractOAuthParams(new URL($page.url)));

	$effect(() => {
		if (!mfaToken) {
			goto('/login');
		}
	});

	function validate(): boolean {
		if (mode === 'recovery') {
			if (!recoveryCode.trim()) {
				codeError = 'validation.required';
				return false;
			}
			codeError = '';
			return true;
		}

		const result = validateOTP(code);
		codeError = result.valid ? '' : (result.error || '');
		return result.valid;
	}

	async function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		error = '';

		if (!validate()) return;

		loading = true;

		try {
			const result = await api.mfaVerify({
				mfa_token: mfaToken,
				code: mode === 'recovery' ? recoveryCode : code,
				method: mode
			});

			if (result.redirect_url) {
				redirectTo(result.redirect_url);
			} else if (oauth?.redirect_uri) {
				redirectTo(oauth.redirect_uri);
			}
		} catch (err) {
			if (err instanceof ApiClientError) {
				error = err.message;
			} else {
				error = getErrorMessage(err);
			}
		} finally {
			loading = false;
		}
	}

	async function sendChallenge(method: 'sms' | 'email') {
		sendingChallenge = true;
		error = '';

		try {
			await api.mfaChallenge({
				mfa_token: mfaToken,
				method
			});
			mode = method;
			challengeSent = true;
			code = '';
			codeError = '';
		} catch (err) {
			if (err instanceof ApiClientError) {
				error = err.message;
			} else {
				error = getErrorMessage(err);
			}
		} finally {
			sendingChallenge = false;
		}
	}

	function switchToRecovery() {
		mode = 'recovery';
		error = '';
		codeError = '';
		code = '';
	}

	function switchToTotp() {
		mode = 'totp';
		error = '';
		codeError = '';
		recoveryCode = '';
		challengeSent = false;
	}
</script>

<svelte:head>
	<title>{$t('mfa.title')}</title>
</svelte:head>

<AuthLayout>
	<div class="space-y-6">
		<div class="text-center">
			<div class="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-full" style="background-color: var(--af-color-primary-light)">
				<svg class="h-7 w-7" style="color: var(--af-color-primary)" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
					<path stroke-linecap="round" stroke-linejoin="round" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
				</svg>
			</div>

			<h1 class="text-2xl font-bold" style="color: var(--af-color-text)">
				{#if mode === 'recovery'}
					{$t('mfa.recovery.title')}
				{:else}
					{$t('mfa.title')}
				{/if}
			</h1>
			<p class="mt-2 text-sm" style="color: var(--af-color-text-muted)">
				{#if mode === 'recovery'}
					{$t('mfa.recovery.description')}
				{:else if mode === 'sms' && challengeSent}
					A code has been sent to your phone number.
				{:else if mode === 'email' && challengeSent}
					A code has been sent to your email address.
				{:else}
					{$t('mfa.description')}
				{/if}
			</p>
		</div>

		{#if error}
			<Alert type="error" message={error} dismissible />
		{/if}

		<form onsubmit={handleSubmit} class="space-y-4" novalidate>
			{#if mode === 'recovery'}
				<div>
					<label for="recovery" class="af-label">{$t('mfa.recovery.title')}</label>
					<input
						id="recovery"
						type="text"
						bind:value={recoveryCode}
						placeholder={$t('mfa.recovery.placeholder')}
						disabled={loading}
						class="af-input text-center font-mono"
						class:error={!!codeError}
						autocomplete="off"
					/>
					{#if codeError}
						<p class="mt-1 text-sm" style="color: var(--af-color-error)" role="alert">
							{$t(codeError)}
						</p>
					{/if}
				</div>
			{:else}
				<OTPInput
					bind:value={code}
					label={$t('mfa.code')}
					error={codeError}
					disabled={loading}
				/>
			{/if}

			<button type="submit" class="af-btn af-btn-primary w-full" disabled={loading}>
				{#if loading}
					<LoadingSpinner size="sm" />
				{/if}
				{$t('mfa.submit')}
			</button>
		</form>

		<div class="space-y-2">
			{#if mode !== 'recovery'}
				<button
					type="button"
					class="af-btn af-btn-ghost w-full text-sm"
					onclick={switchToRecovery}
					disabled={loading || sendingChallenge}
				>
					{$t('mfa.recovery')}
				</button>
			{:else}
				<button
					type="button"
					class="af-btn af-btn-ghost w-full text-sm"
					onclick={switchToTotp}
					disabled={loading || sendingChallenge}
				>
					Use authenticator app
				</button>
			{/if}

			{#if methods.includes('sms') && mode !== 'sms'}
				<button
					type="button"
					class="af-btn af-btn-ghost w-full text-sm"
					onclick={() => sendChallenge('sms')}
					disabled={loading || sendingChallenge}
				>
					{#if sendingChallenge}
						<LoadingSpinner size="sm" />
					{/if}
					{$t('mfa.sms')}
				</button>
			{/if}

			{#if methods.includes('email') && mode !== 'email'}
				<button
					type="button"
					class="af-btn af-btn-ghost w-full text-sm"
					onclick={() => sendChallenge('email')}
					disabled={loading || sendingChallenge}
				>
					{#if sendingChallenge}
						<LoadingSpinner size="sm" />
					{/if}
					{$t('mfa.email')}
				</button>
			{/if}
		</div>
	</div>
</AuthLayout>
