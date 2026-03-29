<script lang="ts">
	import { page } from '$app/stores';
	import { t } from '$lib/i18n';
	import { api, ApiClientError } from '$lib/api';
	import { validateEmail } from '$lib/utils/validation';
	import { getErrorMessage } from '$lib/utils';
	import AuthLayout from '$lib/components/AuthLayout.svelte';
	import Alert from '$lib/components/Alert.svelte';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	let email = $state('');
	let loading = $state(false);
	let error = $state('');
	let emailError = $state('');
	let success = $state(false);

	let clientId = $derived($page.url.searchParams.get('client_id') || undefined);

	function validate(): boolean {
		const result = validateEmail(email);
		emailError = result.valid ? '' : (result.error || '');
		return result.valid;
	}

	async function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		error = '';

		if (!validate()) return;

		loading = true;

		try {
			await api.forgotPassword({
				email,
				client_id: clientId
			});
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
	}
</script>

<svelte:head>
	<title>{$t('forgot.title')}</title>
</svelte:head>

<AuthLayout>
	<div class="space-y-6">
		<div class="text-center">
			<h1 class="text-2xl font-bold" style="color: var(--af-color-text)">
				{$t('forgot.title')}
			</h1>
			{#if !success}
				<p class="mt-2 text-sm" style="color: var(--af-color-text-muted)">
					{$t('forgot.description')}
				</p>
			{/if}
		</div>

		{#if success}
			<Alert type="success" message={$t('forgot.success')} />
			<div class="text-center">
				<a href="/login" class="af-link text-sm">
					{$t('forgot.back')}
				</a>
			</div>
		{:else}
			{#if error}
				<Alert type="error" message={error} dismissible />
			{/if}

			<form onsubmit={handleSubmit} class="space-y-4" novalidate>
				<div>
					<label for="email" class="af-label">{$t('forgot.email')}</label>
					<input
						id="email"
						name="email"
						type="email"
						bind:value={email}
						placeholder={$t('forgot.email.placeholder')}
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

				<button type="submit" class="af-btn af-btn-primary w-full" disabled={loading}>
					{#if loading}
						<LoadingSpinner size="sm" />
					{/if}
					{$t('forgot.submit')}
				</button>
			</form>

			<div class="text-center">
				<a href="/login" class="af-link text-sm">
					{$t('forgot.back')}
				</a>
			</div>
		{/if}
	</div>
</AuthLayout>
