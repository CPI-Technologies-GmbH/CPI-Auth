<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { t } from '$lib/i18n';
	import { api, ApiClientError } from '$lib/api';
	import type { ConsentInfo } from '$lib/api/types';
	import { redirectTo, getErrorMessage, formatScope } from '$lib/utils';
	import AuthLayout from '$lib/components/AuthLayout.svelte';
	import Alert from '$lib/components/Alert.svelte';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	let loading = $state(true);
	let submitting = $state(false);
	let error = $state('');
	let consentInfo = $state<ConsentInfo | null>(null);
	let remember = $state(false);

	let consentChallenge = $derived($page.url.searchParams.get('consent_challenge') || '');

	onMount(async () => {
		if (!consentChallenge) {
			error = 'Missing consent challenge parameter.';
			loading = false;
			return;
		}

		try {
			consentInfo = await api.getConsentInfo(consentChallenge);
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

	async function handleDecision(grant: boolean) {
		submitting = true;
		error = '';

		try {
			const result = await api.submitConsent({
				consent_challenge: consentChallenge,
				grant,
				remember,
				scopes: grant ? consentInfo?.requested_scopes.map((s) => s.name) : undefined
			});

			if (result.redirect_url) {
				redirectTo(result.redirect_url);
			}
		} catch (err) {
			if (err instanceof ApiClientError) {
				error = err.message;
			} else {
				error = getErrorMessage(err);
			}
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head>
	<title>{$t('consent.title')}</title>
</svelte:head>

<AuthLayout maxWidth="md">
	<div class="space-y-6">
		{#if loading}
			<div class="flex justify-center py-12">
				<LoadingSpinner size="lg" color="var(--af-color-primary)" />
			</div>
		{:else if error}
			<Alert type="error" message={error} />
			<div class="text-center">
				<a href="/login" class="af-link text-sm">{$t('error.back')}</a>
			</div>
		{:else if consentInfo}
			<div class="text-center">
				{#if consentInfo.client_logo}
					<img
						src={consentInfo.client_logo}
						alt="{consentInfo.client_name} logo"
						class="mx-auto mb-4 h-16 w-16 rounded-xl object-contain"
					/>
				{:else}
					<div
						class="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-xl"
						style="background-color: var(--af-color-primary-light)"
					>
						<svg class="h-8 w-8" style="color: var(--af-color-primary)" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
							<path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
						</svg>
					</div>
				{/if}

				<h1 class="text-2xl font-bold" style="color: var(--af-color-text)">
					{$t('consent.title')}
				</h1>
				<p class="mt-1 text-lg font-medium" style="color: var(--af-color-text)">
					{consentInfo.client_name}
				</p>
				<p class="text-sm" style="color: var(--af-color-text-muted)">
					{$t('consent.description')}
				</p>
			</div>

			{#if consentInfo.requested_scopes.length > 0}
				<div>
					<h2 class="mb-3 text-sm font-medium" style="color: var(--af-color-text)">
						{$t('consent.scopes_title')}
					</h2>
					<div
						class="divide-y rounded-lg"
						style="background-color: var(--af-color-surface); border: 1px solid var(--af-color-border); divide-color: var(--af-color-border)"
					>
						{#each consentInfo.requested_scopes as scope}
							<div class="flex items-center gap-3 px-4 py-3">
								<svg class="h-5 w-5 shrink-0" style="color: var(--af-color-success)" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
									<path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
								</svg>
								<div>
									<p class="text-sm font-medium" style="color: var(--af-color-text)">
										{scope.description || formatScope(scope.name)}
									</p>
								</div>
							</div>
						{/each}
					</div>
				</div>
			{/if}

			<label class="flex items-center gap-2 text-sm" style="color: var(--af-color-text-muted)">
				<input type="checkbox" bind:checked={remember} class="af-checkbox" disabled={submitting} />
				{$t('consent.remember')}
			</label>

			<div class="flex gap-3">
				<button
					type="button"
					class="af-btn af-btn-secondary flex-1"
					onclick={() => handleDecision(false)}
					disabled={submitting}
				>
					{#if submitting}
						<LoadingSpinner size="sm" />
					{/if}
					{$t('consent.deny')}
				</button>
				<button
					type="button"
					class="af-btn af-btn-primary flex-1"
					onclick={() => handleDecision(true)}
					disabled={submitting}
				>
					{#if submitting}
						<LoadingSpinner size="sm" />
					{/if}
					{$t('consent.allow')}
				</button>
			</div>

			{#if consentInfo.client_uri}
				<p class="text-center text-xs" style="color: var(--af-color-text-muted)">
					<a href={consentInfo.client_uri} class="af-link" target="_blank" rel="noopener noreferrer">
						{consentInfo.client_uri}
					</a>
				</p>
			{/if}
		{/if}
	</div>
</AuthLayout>
