<script lang="ts">
	import { page } from '$app/stores';
	import { t } from '$lib/i18n';
	import { api, ApiClientError } from '$lib/api';
	import { validatePassword, validatePasswordMatch } from '$lib/utils/validation';
	import { getErrorMessage } from '$lib/utils';
	import AuthLayout from '$lib/components/AuthLayout.svelte';
	import PasswordInput from '$lib/components/PasswordInput.svelte';
	import Alert from '$lib/components/Alert.svelte';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	let password = $state('');
	let confirmPassword = $state('');
	let loading = $state(false);
	let error = $state('');
	let passwordError = $state('');
	let confirmError = $state('');
	let success = $state(false);

	let token = $derived($page.url.searchParams.get('token') || '');

	$effect(() => {
		if (!token) {
			error = 'reset.error.token';
		}
	});

	function validate(): boolean {
		let valid = true;

		const passResult = validatePassword(password);
		passwordError = passResult.valid ? '' : (passResult.error || '');
		if (!passResult.valid) valid = false;

		const confirmResult = validatePasswordMatch(password, confirmPassword);
		confirmError = confirmResult.valid ? '' : (confirmResult.error || '');
		if (!confirmResult.valid) valid = false;

		return valid;
	}

	async function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		error = '';

		if (!token) {
			error = 'reset.error.token';
			return;
		}

		if (!validate()) return;

		loading = true;

		try {
			await api.resetPassword({
				token,
				password
			});
			success = true;
		} catch (err) {
			if (err instanceof ApiClientError) {
				if (err.code === 'invalid_token' || err.code === 'token_expired') {
					error = 'reset.error.token';
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
</script>

<svelte:head>
	<title>{$t('reset.title')}</title>
</svelte:head>

<AuthLayout>
	<div class="space-y-6">
		<div class="text-center">
			<h1 class="text-2xl font-bold" style="color: var(--af-color-text)">
				{$t('reset.title')}
			</h1>
			{#if !success && !error}
				<p class="mt-2 text-sm" style="color: var(--af-color-text-muted)">
					{$t('reset.description')}
				</p>
			{/if}
		</div>

		{#if success}
			<Alert type="success" message={$t('reset.success')} />
			<div class="text-center">
				<a href="/login" class="af-btn af-btn-primary inline-flex">
					{$t('reset.login_link')}
				</a>
			</div>
		{:else}
			{#if error}
				<Alert type="error" message={error.startsWith('reset.') ? $t(error) : error} dismissible />
			{/if}

			{#if token}
				<form onsubmit={handleSubmit} class="space-y-4" novalidate>
					<PasswordInput
						bind:value={password}
						name="password"
						label={$t('reset.password')}
						placeholder={$t('reset.password.placeholder')}
						error={passwordError}
						showStrength
						disabled={loading}
						autocomplete="new-password"
					/>

					<PasswordInput
						bind:value={confirmPassword}
						name="confirmPassword"
						id="confirmPassword"
						label={$t('reset.confirm')}
						placeholder={$t('reset.confirm.placeholder')}
						error={confirmError}
						disabled={loading}
						autocomplete="new-password"
					/>

					<button type="submit" class="af-btn af-btn-primary w-full" disabled={loading}>
						{#if loading}
							<LoadingSpinner size="sm" />
						{/if}
						{$t('reset.submit')}
					</button>
				</form>
			{/if}

			<div class="text-center">
				<a href="/login" class="af-link text-sm">{$t('forgot.back')}</a>
			</div>
		{/if}
	</div>
</AuthLayout>
