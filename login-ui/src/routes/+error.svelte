<script lang="ts">
	import { page } from '$app/stores';
	import AuthLayout from '$lib/components/AuthLayout.svelte';
	import Alert from '$lib/components/Alert.svelte';
</script>

<svelte:head>
	<title>Error - {$page.status}</title>
</svelte:head>

<AuthLayout>
	<div class="space-y-6 text-center">
		<div>
			<div
				class="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full"
				style="background-color: var(--af-color-error-light)"
			>
				<span class="text-2xl font-bold" style="color: var(--af-color-error)">
					{$page.status}
				</span>
			</div>

			<h1 class="text-2xl font-bold" style="color: var(--af-color-text)">
				{#if $page.status === 404}
					Page Not Found
				{:else}
					Something Went Wrong
				{/if}
			</h1>
		</div>

		<Alert type="error" message={$page.error?.message || 'An unexpected error occurred.'} />

		<a href="/login" class="af-btn af-btn-primary inline-flex">
			Back to Sign In
		</a>
	</div>
</AuthLayout>
