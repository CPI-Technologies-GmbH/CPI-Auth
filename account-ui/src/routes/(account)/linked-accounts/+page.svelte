<script lang="ts">
	import { api } from '$lib/api';
	import type { LinkedIdentity } from '$lib/types';
	import ProviderIcon from '$components/ProviderIcon.svelte';
	import Alert from '$components/Alert.svelte';
	import Badge from '$components/Badge.svelte';
	import ConfirmDialog from '$components/ConfirmDialog.svelte';
	import EmptyState from '$components/EmptyState.svelte';
	import LoadingSpinner from '$components/LoadingSpinner.svelte';
	import { formatDate } from '$lib/utils/format';

	let identities = $state<LinkedIdentity[]>([]);
	let loading = $state(true);
	let error = $state('');
	let successMessage = $state('');

	let identityToUnlink = $state<LinkedIdentity | null>(null);
	let unlinking = $state(false);

	const availableProviders = [
		{ id: 'google', name: 'Google' },
		{ id: 'github', name: 'GitHub' },
		{ id: 'microsoft', name: 'Microsoft' },
		{ id: 'apple', name: 'Apple' },
		{ id: 'facebook', name: 'Facebook' },
		{ id: 'twitter', name: 'Twitter/X' },
		{ id: 'discord', name: 'Discord' }
	];

	async function loadIdentities() {
		loading = true;
		error = '';
		try {
			identities = await api.getIdentities();
		} catch (err: unknown) {
			error = (err as { message?: string })?.message || 'Failed to load linked accounts.';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		loadIdentities();
	});

	function getProviderName(providerId: string): string {
		const provider = availableProviders.find((p) => p.id === providerId);
		return provider?.name || providerId;
	}

	function isLinked(providerId: string): boolean {
		return identities.some((i) => i.provider === providerId);
	}

	function handleLinkProvider(providerId: string) {
		// Redirect to the social auth flow
		const returnUrl = encodeURIComponent(window.location.href);
		window.location.href = `/api/auth/link/${providerId}?return_to=${returnUrl}`;
	}

	async function handleUnlinkIdentity() {
		if (!identityToUnlink) return;

		// Safety check: don't allow unlinking if it's the only auth method
		if (identities.length <= 1) {
			error = 'Cannot unlink the last authentication method. You must have at least one way to sign in.';
			identityToUnlink = null;
			return;
		}

		unlinking = true;
		try {
			await api.unlinkIdentity(identityToUnlink.id);
			successMessage = `${getProviderName(identityToUnlink.provider)} account has been unlinked.`;
			identityToUnlink = null;
			await loadIdentities();
		} catch (err: unknown) {
			error = (err as { message?: string })?.message || 'Failed to unlink account.';
		} finally {
			unlinking = false;
		}
	}

	const unlinkedProviders = $derived(
		availableProviders.filter((p) => !identities.some((i) => i.provider === p.id))
	);
</script>

<svelte:head>
	<title>Linked Accounts - CPI Auth Account</title>
</svelte:head>

<div class="page">
	<div class="page-header">
		<h1>Linked Accounts</h1>
		<p>Connect third-party accounts for easier sign-in.</p>
	</div>

	{#if successMessage}
		<Alert variant="success" dismissible ondismiss={() => (successMessage = '')}>{successMessage}</Alert>
	{/if}
	{#if error}
		<Alert variant="danger" dismissible ondismiss={() => (error = '')}>{error}</Alert>
	{/if}

	{#if loading}
		<div class="loading-center">
			<LoadingSpinner size={40} />
		</div>
	{:else}
		<!-- Connected accounts -->
		{#if identities.length > 0}
			<section class="card">
				<h2 class="card-title">Connected Accounts</h2>
				<div class="identities-list">
					{#each identities as identity}
						<div class="identity-item">
							<ProviderIcon provider={identity.provider} size={32} />
							<div class="identity-info">
								<div class="identity-header">
									<span class="identity-provider">{getProviderName(identity.provider)}</span>
									<Badge variant="success">Connected</Badge>
								</div>
								<div class="identity-details">
									{#if identity.email}
										<span class="identity-detail">{identity.email}</span>
									{/if}
									{#if identity.username}
										<span class="identity-detail">@{identity.username}</span>
									{/if}
									<span class="identity-detail">Linked {formatDate(identity.linked_at)}</span>
								</div>
							</div>
							<button
								class="btn btn-danger-outline btn-sm"
								onclick={() => (identityToUnlink = identity)}
								disabled={identities.length <= 1}
								title={identities.length <= 1 ? 'Cannot unlink the last authentication method' : 'Unlink account'}
							>
								Unlink
							</button>
						</div>
					{/each}
				</div>
			</section>
		{/if}

		<!-- Available providers -->
		{#if unlinkedProviders.length > 0}
			<section class="card">
				<h2 class="card-title">Available Providers</h2>
				<p class="card-description">Link additional accounts for easier sign-in.</p>
				<div class="providers-grid">
					{#each unlinkedProviders as provider}
						<button class="provider-btn" onclick={() => handleLinkProvider(provider.id)}>
							<ProviderIcon provider={provider.id} size={24} />
							<span>Connect {provider.name}</span>
						</button>
					{/each}
				</div>
			</section>
		{/if}

		{#if identities.length === 0 && unlinkedProviders.length === 0}
			<EmptyState
				icon="link"
				title="No linked accounts"
				description="Connect a third-party account to enable social sign-in."
			/>
		{/if}
	{/if}

	<!-- Unlink Dialog -->
	<ConfirmDialog
		open={!!identityToUnlink}
		title="Unlink Account"
		confirmText="Unlink"
		variant="danger"
		loading={unlinking}
		onconfirm={handleUnlinkIdentity}
		oncancel={() => (identityToUnlink = null)}
	>
		<p>
			Are you sure you want to unlink your {getProviderName(identityToUnlink?.provider || '')} account?
			You won't be able to sign in with it unless you link it again.
		</p>
	</ConfirmDialog>
</div>

<style>
	.page {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.page-header {
		margin-bottom: 0.5rem;
	}

	.page-header h1 {
		font-size: 1.5rem;
		font-weight: 700;
		color: var(--color-text);
		margin: 0 0 0.25rem;
	}

	.page-header p {
		font-size: 0.875rem;
		color: var(--color-text-secondary);
		margin: 0;
	}

	.card {
		background-color: var(--color-surface);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-xl);
		padding: 1.5rem;
	}

	.card-title {
		font-size: 1rem;
		font-weight: 600;
		color: var(--color-text);
		margin: 0 0 1rem;
	}

	.card-description {
		font-size: 0.8125rem;
		color: var(--color-text-secondary);
		margin: -0.5rem 0 1rem;
	}

	.loading-center {
		display: flex;
		justify-content: center;
		padding: 4rem 0;
	}

	.identities-list {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.identity-item {
		display: flex;
		align-items: center;
		gap: 1rem;
		padding: 1rem;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-lg);
		transition: border-color 0.15s;
	}

	.identity-item:hover {
		border-color: var(--color-border-hover);
	}

	.identity-info {
		flex: 1;
		min-width: 0;
	}

	.identity-header {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		margin-bottom: 0.25rem;
	}

	.identity-provider {
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--color-text);
	}

	.identity-details {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
	}

	.identity-detail {
		font-size: 0.75rem;
		color: var(--color-text-secondary);
	}

	.identity-detail + .identity-detail::before {
		content: '\00B7';
		margin-right: 0.5rem;
	}

	.providers-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(12rem, 1fr));
		gap: 0.75rem;
	}

	.provider-btn {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 0.875rem;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-lg);
		background-color: var(--color-bg);
		color: var(--color-text);
		font-size: 0.875rem;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.15s;
	}

	.provider-btn:hover {
		border-color: var(--color-primary);
		background-color: var(--color-primary-light);
	}

	.btn {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem 1rem;
		border-radius: var(--radius-md);
		font-size: 0.875rem;
		font-weight: 500;
		cursor: pointer;
		border: 1px solid transparent;
		transition: all 0.15s;
		white-space: nowrap;
	}

	.btn:disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}

	.btn-sm {
		padding: 0.375rem 0.75rem;
		font-size: 0.8125rem;
	}

	.btn-danger-outline {
		background: transparent;
		border-color: var(--color-danger);
		color: var(--color-danger);
	}

	.btn-danger-outline:hover:not(:disabled) {
		background-color: color-mix(in srgb, var(--color-danger) 10%, transparent);
	}
</style>
