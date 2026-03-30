<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { t } from '$lib/i18n';
	import { api, ApiClientError } from '$lib/api';
	import type { MfaEnrollResponse } from '$lib/api/types';
	import { validateOTP } from '$lib/utils/validation';
	import { redirectTo, getErrorMessage } from '$lib/utils';
	import AuthLayout from '$lib/components/AuthLayout.svelte';
	import OTPInput from '$lib/components/OTPInput.svelte';
	import Alert from '$lib/components/Alert.svelte';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	let code = $state('');
	let loading = $state(true);
	let verifying = $state(false);
	let error = $state('');
	let codeError = $state('');
	let enrollData = $state<MfaEnrollResponse | null>(null);
	let showManualCode = $state(false);
	let showRecoveryCodes = $state(false);
	let recoveryCodes = $state<string[]>([]);
	let copied = $state(false);

	let mfaToken = $derived($page.url.searchParams.get('mfa_token') || '');

	$effect(() => {
		if (!mfaToken) {
			goto('/login');
			return;
		}
		enrollMfa();
	});

	async function enrollMfa() {
		loading = true;
		error = '';

		try {
			enrollData = await api.mfaEnroll({
				mfa_token: mfaToken,
				method: 'totp'
			});
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

	function validate(): boolean {
		const result = validateOTP(code);
		codeError = result.valid ? '' : (result.error || '');
		return result.valid;
	}

	async function handleVerify(e: SubmitEvent) {
		e.preventDefault();
		error = '';

		if (!validate()) return;

		verifying = true;

		try {
			const result = await api.mfaEnrollVerify({
				mfa_token: mfaToken,
				code
			});

			// Show recovery codes before completing
			if (enrollData?.recovery_codes) {
				recoveryCodes = enrollData.recovery_codes;
				showRecoveryCodes = true;
			} else if (result.redirect_url) {
				redirectTo(result.redirect_url);
			}
		} catch (err) {
			if (err instanceof ApiClientError) {
				error = err.message;
			} else {
				error = getErrorMessage(err);
			}
		} finally {
			verifying = false;
		}
	}

	async function copyRecoveryCodes() {
		try {
			await navigator.clipboard.writeText(recoveryCodes.join('\n'));
			copied = true;
			setTimeout(() => (copied = false), 2000);
		} catch {
			// Fallback: select text
		}
	}

	function handleDone() {
		goto('/login');
	}
</script>

<svelte:head>
	<title>{$t('mfa_enroll.title')}</title>
</svelte:head>

<AuthLayout maxWidth="md">
	<div class="space-y-6">
		<div class="text-center">
			<h1 class="text-2xl font-bold" style="color: var(--af-color-text)">
				{#if showRecoveryCodes}
					{$t('mfa_enroll.recovery_title')}
				{:else}
					{$t('mfa_enroll.title')}
				{/if}
			</h1>
		</div>

		{#if error}
			<Alert type="error" message={error} dismissible />
		{/if}

		{#if loading}
			<div class="flex justify-center py-12">
				<LoadingSpinner size="lg" color="var(--af-color-primary)" />
			</div>
		{:else if showRecoveryCodes}
			<div class="space-y-4">
				<p class="text-sm" style="color: var(--af-color-text-muted)">
					{$t('mfa_enroll.recovery_description')}
				</p>

				<div class="rounded-lg p-4" style="background-color: var(--af-color-surface); border: 1px solid var(--af-color-border)">
					<div class="grid grid-cols-2 gap-2">
						{#each recoveryCodes as recoveryCode}
							<code class="rounded px-2 py-1 text-center text-sm font-mono" style="background-color: var(--af-color-background)">
								{recoveryCode}
							</code>
						{/each}
					</div>
				</div>

				<button
					type="button"
					class="af-btn af-btn-secondary w-full"
					onclick={copyRecoveryCodes}
				>
					{#if copied}
						<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
							<path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
						</svg>
						{$t('mfa_enroll.recovery_copied')}
					{:else}
						<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
							<path stroke-linecap="round" stroke-linejoin="round" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
						</svg>
						{$t('mfa_enroll.copy')}
					{/if}
				</button>

				<button
					type="button"
					class="af-btn af-btn-primary w-full"
					onclick={handleDone}
				>
					{$t('mfa_enroll.done')}
				</button>
			</div>
		{:else if enrollData}
			<div class="space-y-6">
				<!-- QR Code -->
				<div class="text-center">
					<p class="mb-4 text-sm" style="color: var(--af-color-text-muted)">
						{$t('mfa_enroll.scan')}
					</p>
					<div class="inline-block rounded-lg bg-white p-4">
						{#if enrollData.qr_code}
							<img
								src={enrollData.qr_code}
								alt="QR code for authenticator app"
								class="h-48 w-48"
							/>
						{:else if enrollData.otpauth_url}
							<!-- Fallback: show otpauth URL as a QR code image -->
							<img
								src="https://api.qrserver.com/v1/create-qr-code/?size=192x192&data={encodeURIComponent(enrollData.otpauth_url)}"
								alt="QR code for authenticator app"
								class="h-48 w-48"
							/>
						{/if}
					</div>
				</div>

				<!-- Manual Code -->
				<div class="text-center">
					<button
						type="button"
						class="af-link text-sm"
						onclick={() => (showManualCode = !showManualCode)}
					>
						{$t('mfa_enroll.manual')}
					</button>

					{#if showManualCode}
						<div class="mt-2 rounded-lg p-3" style="background-color: var(--af-color-surface); border: 1px solid var(--af-color-border)">
							<code class="text-sm font-mono font-semibold" style="color: var(--af-color-text)">
								{enrollData.secret}
							</code>
						</div>
					{/if}
				</div>

				<!-- Verification -->
				<div>
					<p class="mb-3 text-center text-sm" style="color: var(--af-color-text-muted)">
						{$t('mfa_enroll.verify')}
					</p>

					<form onsubmit={handleVerify} class="space-y-4" novalidate>
						<OTPInput
							bind:value={code}
							error={codeError}
							disabled={verifying}
						/>

						<button type="submit" class="af-btn af-btn-primary w-full" disabled={verifying}>
							{#if verifying}
								<LoadingSpinner size="sm" />
							{/if}
							{$t('mfa_enroll.submit')}
						</button>
					</form>
				</div>
			</div>
		{/if}
	</div>
</AuthLayout>
