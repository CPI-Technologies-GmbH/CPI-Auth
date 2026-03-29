<script lang="ts">
	import { page } from '$app/stores';
	import { t } from '$lib/i18n';
	import { api, ApiClientError } from '$lib/api';
	import { extractOAuthParams } from '$lib/stores';
	import { validateEmail, validateOTP } from '$lib/utils/validation';
	import { redirectTo, getErrorMessage } from '$lib/utils';
	import AuthLayout from '$lib/components/AuthLayout.svelte';
	import OTPInput from '$lib/components/OTPInput.svelte';
	import Alert from '$lib/components/Alert.svelte';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	type Step = 'email' | 'otp' | 'link_sent';

	let email = $state($page.url.searchParams.get('email') || '');
	let code = $state('');
	let step = $state<Step>('email');
	let method = $state<'email_link' | 'email_otp'>('email_otp');
	let loading = $state(false);
	let error = $state('');
	let emailError = $state('');
	let codeError = $state('');
	let token = $state('');
	let resendCountdown = $state(0);

	let oauth = $derived(extractOAuthParams(new URL($page.url)));

	function startResendTimer() {
		resendCountdown = 60;
		const interval = setInterval(() => {
			resendCountdown--;
			if (resendCountdown <= 0) clearInterval(interval);
		}, 1000);
	}

	async function handleEmailSubmit(e: SubmitEvent) {
		e.preventDefault();
		error = '';

		const result = validateEmail(email);
		emailError = result.valid ? '' : (result.error || '');
		if (!result.valid) return;

		loading = true;

		try {
			const response = await api.passwordlessStart({
				email,
				method,
				...(oauth && {
					client_id: oauth.client_id,
					redirect_uri: oauth.redirect_uri,
					scope: oauth.scope,
					state: oauth.state
				})
			});

			if (method === 'email_link') {
				step = 'link_sent';
			} else {
				step = 'otp';
			}
			startResendTimer();
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

	async function handleOtpSubmit(e: SubmitEvent) {
		e.preventDefault();
		error = '';

		const result = validateOTP(code);
		codeError = result.valid ? '' : (result.error || '');
		if (!result.valid) return;

		loading = true;

		try {
			const response = await api.passwordlessVerify({
				token,
				code
			});

			if (response.redirect_url) {
				redirectTo(response.redirect_url);
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

	async function resendCode() {
		if (resendCountdown > 0) return;
		loading = true;
		error = '';

		try {
			await api.passwordlessStart({
				email,
				method,
				...(oauth && {
					client_id: oauth.client_id,
					redirect_uri: oauth.redirect_uri,
					scope: oauth.scope,
					state: oauth.state
				})
			});
			startResendTimer();
		} catch (err) {
			error = getErrorMessage(err);
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>{$t('passwordless.title')}</title>
</svelte:head>

<AuthLayout>
	<div class="space-y-6">
		{#if step === 'email'}
			<div class="text-center">
				<div class="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-full" style="background-color: var(--af-color-primary-light)">
					<svg class="h-7 w-7" style="color: var(--af-color-primary)" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
						<path stroke-linecap="round" stroke-linejoin="round" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
					</svg>
				</div>
				<h1 class="text-2xl font-bold" style="color: var(--af-color-text)">
					Passwordless Sign In
				</h1>
			</div>

			{#if error}
				<Alert type="error" message={error} dismissible />
			{/if}

			<form onsubmit={handleEmailSubmit} class="space-y-4" novalidate>
				<div>
					<label for="email" class="af-label">Email address</label>
					<input
						id="email"
						type="email"
						bind:value={email}
						placeholder="you@example.com"
						autocomplete="email"
						required
						disabled={loading}
						class="af-input"
						class:error={!!emailError}
					/>
					{#if emailError}
						<p class="mt-1 text-sm" style="color: var(--af-color-error)" role="alert">
							{$t(emailError)}
						</p>
					{/if}
				</div>

				<div class="flex gap-2">
					<label class="flex flex-1 cursor-pointer items-center justify-center gap-2 rounded-lg border p-3 text-sm transition-all"
						style="border-color: {method === 'email_otp' ? 'var(--af-color-primary)' : 'var(--af-color-border)'}; background-color: {method === 'email_otp' ? 'var(--af-color-primary-light)' : 'transparent'}"
					>
						<input type="radio" bind:group={method} value="email_otp" class="sr-only" />
						<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
							<path stroke-linecap="round" stroke-linejoin="round" d="M7 20l4-16m2 16l4-16M6 9h14M4 15h14" />
						</svg>
						Code
					</label>
					<label class="flex flex-1 cursor-pointer items-center justify-center gap-2 rounded-lg border p-3 text-sm transition-all"
						style="border-color: {method === 'email_link' ? 'var(--af-color-primary)' : 'var(--af-color-border)'}; background-color: {method === 'email_link' ? 'var(--af-color-primary-light)' : 'transparent'}"
					>
						<input type="radio" bind:group={method} value="email_link" class="sr-only" />
						<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
							<path stroke-linecap="round" stroke-linejoin="round" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
						</svg>
						Magic Link
					</label>
				</div>

				<button type="submit" class="af-btn af-btn-primary w-full" disabled={loading}>
					{#if loading}
						<LoadingSpinner size="sm" />
					{/if}
					Continue
				</button>
			</form>

			<div class="text-center">
				<a href="/login" class="af-link text-sm">{$t('forgot.back')}</a>
			</div>

		{:else if step === 'otp'}
			<div class="text-center">
				<div class="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-full" style="background-color: var(--af-color-primary-light)">
					<svg class="h-7 w-7" style="color: var(--af-color-primary)" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
						<path stroke-linecap="round" stroke-linejoin="round" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
					</svg>
				</div>
				<h1 class="text-2xl font-bold" style="color: var(--af-color-text)">
					{$t('passwordless.title')}
				</h1>
				<p class="mt-2 text-sm" style="color: var(--af-color-text-muted)">
					{$t('passwordless.email_sent')}
					<span class="font-medium" style="color: var(--af-color-text)">{email}</span>
				</p>
			</div>

			{#if error}
				<Alert type="error" message={error} dismissible />
			{/if}

			<form onsubmit={handleOtpSubmit} class="space-y-4" novalidate>
				<OTPInput
					bind:value={code}
					label={$t('passwordless.code')}
					error={codeError}
					disabled={loading}
				/>

				<button type="submit" class="af-btn af-btn-primary w-full" disabled={loading}>
					{#if loading}
						<LoadingSpinner size="sm" />
					{/if}
					{$t('passwordless.submit')}
				</button>
			</form>

			<div class="text-center">
				<button
					type="button"
					class="af-link text-sm"
					onclick={resendCode}
					disabled={resendCountdown > 0}
					class:opacity-50={resendCountdown > 0}
				>
					{#if resendCountdown > 0}
						{$t('passwordless.resend')} ({resendCountdown}s)
					{:else}
						{$t('passwordless.resend')}
					{/if}
				</button>
			</div>

		{:else if step === 'link_sent'}
			<div class="space-y-4 text-center">
				<div class="mx-auto flex h-16 w-16 items-center justify-center rounded-full" style="background-color: var(--af-color-success-light)">
					<svg class="h-8 w-8" style="color: var(--af-color-success)" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
						<path stroke-linecap="round" stroke-linejoin="round" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
					</svg>
				</div>

				<h1 class="text-2xl font-bold" style="color: var(--af-color-text)">
					{$t('passwordless.title')}
				</h1>
				<p class="text-sm" style="color: var(--af-color-text-muted)">
					{$t('passwordless.link_sent')}
					<span class="font-medium" style="color: var(--af-color-text)">{email}</span>
				</p>

				<Alert type="info" message="Click the link in your email to complete sign-in. You can close this page." />

				<div class="pt-2">
					<button
						type="button"
						class="af-link text-sm"
						onclick={resendCode}
						disabled={resendCountdown > 0}
						class:opacity-50={resendCountdown > 0}
					>
						{#if resendCountdown > 0}
							{$t('passwordless.resend')} ({resendCountdown}s)
						{:else}
							{$t('passwordless.resend')}
						{/if}
					</button>
				</div>

				<a href="/login" class="af-link text-sm">{$t('forgot.back')}</a>
			</div>
		{/if}
	</div>
</AuthLayout>
