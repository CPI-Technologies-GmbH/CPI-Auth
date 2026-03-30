<script lang="ts">
	import { page } from '$app/stores';
	import { t } from '$lib/i18n';
	import AuthLayout from '$lib/components/AuthLayout.svelte';
	import Alert from '$lib/components/Alert.svelte';

	let errorCode = $derived($page.url.searchParams.get('error') || '');
	let errorDescription = $derived(
		$page.url.searchParams.get('error_description') || $t('error.generic')
	);
</script>

<svelte:head>
	<title>{$t('error.title')}</title>
</svelte:head>

<AuthLayout>
	<div class="space-y-6 text-center">
		<div>
			<div
				class="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full"
				style="background-color: var(--af-color-error-light)"
			>
				<svg
					class="h-8 w-8"
					style="color: var(--af-color-error)"
					fill="none"
					viewBox="0 0 24 24"
					stroke="currentColor"
					stroke-width="2"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
					/>
				</svg>
			</div>

			<h1 class="text-2xl font-bold" style="color: var(--af-color-text)">
				{$t('error.title')}
			</h1>
		</div>

		<Alert type="error" message={errorDescription} />

		{#if errorCode}
			<p class="text-xs font-mono" style="color: var(--af-color-text-muted)">
				Error code: {errorCode}
			</p>
		{/if}

		<a href="/login" class="af-btn af-btn-primary inline-flex">
			{$t('error.back')}
		</a>
	</div>
</AuthLayout>
