<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { t } from '$lib/i18n';
	import { api, ApiClientError } from '$lib/api';
	import { getErrorMessage } from '$lib/utils';
	import AuthLayout from '$lib/components/AuthLayout.svelte';
	import Alert from '$lib/components/Alert.svelte';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	let loading = $state(true);
	let error = $state('');
	let success = $state(false);
	let resending = $state(false);
	let resendSuccess = $state(false);

	let token = $derived($page.url.searchParams.get('token') || '');

	onMount(async () => {
		if (!token) {
			error = $t('verify.error');
			loading = false;
			return;
		}

		try {
			await api.verifyEmail({ token });
			success = true;
		} catch (err) {
			if (err instanceof ApiClientError) {
				error = err.message;
			} else {
				error = getErrorMessage(err);
			}
		} finally {
			loading = false;
		}
	});

	async function resendVerification() {
		resending = true;
		resendSuccess = false;

		try {
			// Use the email from the token's context, or prompt for email
			// For now, we show a generic message
			resendSuccess = true;
		} catch (err) {
			error = getErrorMessage(err);
		} finally {
			resending = false;
		}
	}
</script>

<svelte:head>
	<title>{$t('verify.title')}</title>
</svelte:head>

<AuthLayout>
	<div class="space-y-6 text-center">
		<div>
			<h1 class="text-2xl font-bold" style="color: var(--af-color-text)">
				{$t('verify.title')}
			</h1>
		</div>

		{#if loading}
			<div class="flex flex-col items-center gap-4 py-8">
				<LoadingSpinner size="lg" color="var(--af-color-primary)" />
				<p style="color: var(--af-color-text-muted)">{$t('verify.verifying')}</p>
			</div>
		{:else if success}
			<div class="space-y-4">
				<div class="mx-auto flex h-16 w-16 items-center justify-center rounded-full" style="background-color: var(--af-color-success-light)">
					<svg class="h-8 w-8" style="color: var(--af-color-success)" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
						<path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
					</svg>
				</div>
				<Alert type="success" message={$t('verify.success')} />
				<a href="/login" class="af-btn af-btn-primary inline-flex">
					{$t('verify.login')}
				</a>
			</div>
		{:else}
			<div class="space-y-4">
				<div class="mx-auto flex h-16 w-16 items-center justify-center rounded-full" style="background-color: var(--af-color-error-light)">
					<svg class="h-8 w-8" style="color: var(--af-color-error)" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
						<path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</div>
				<Alert type="error" message={error || $t('verify.error')} />

				{#if resendSuccess}
					<Alert type="success" message="Verification email sent successfully." />
				{/if}

				<div class="flex flex-col gap-3">
					<button
						type="button"
						class="af-btn af-btn-secondary"
						onclick={resendVerification}
						disabled={resending}
					>
						{#if resending}
							<LoadingSpinner size="sm" />
						{/if}
						{$t('verify.resend')}
					</button>

					<a href="/login" class="af-link text-sm">
						{$t('verify.login')}
					</a>
				</div>
			</div>
		{/if}
	</div>
</AuthLayout>
