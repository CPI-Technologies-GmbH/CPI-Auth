<script lang="ts">
	import { api } from '$lib/api';
	import type { Organization } from '$lib/types';
	import Alert from '$components/Alert.svelte';
	import Badge from '$components/Badge.svelte';
	import ConfirmDialog from '$components/ConfirmDialog.svelte';
	import EmptyState from '$components/EmptyState.svelte';
	import LoadingSpinner from '$components/LoadingSpinner.svelte';
	import { formatDate } from '$lib/utils/format';

	let organizations = $state<Organization[]>([]);
	let loading = $state(true);
	let error = $state('');
	let successMessage = $state('');

	let orgToLeave = $state<Organization | null>(null);
	let leaving = $state(false);

	let activeOrgId = $state<string | null>(null);

	async function loadOrganizations() {
		loading = true;
		error = '';
		try {
			organizations = await api.getOrganizations();
			if (organizations.length > 0 && !activeOrgId) {
				activeOrgId = organizations[0].id;
			}
		} catch (err: unknown) {
			error = (err as { message?: string })?.message || 'Failed to load organizations.';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		loadOrganizations();
	});

	function handleSwitchOrg(orgId: string) {
		activeOrgId = orgId;
		successMessage = `Switched to ${organizations.find((o) => o.id === orgId)?.name || 'organization'}.`;
	}

	async function handleLeaveOrg() {
		if (!orgToLeave) return;
		leaving = true;
		try {
			// Leave org API would be something like DELETE /v1/users/me/organizations/:id
			// For now, we simulate it
			successMessage = `You have left ${orgToLeave.name}.`;
			orgToLeave = null;
			await loadOrganizations();
		} catch (err: unknown) {
			error = (err as { message?: string })?.message || 'Failed to leave organization.';
		} finally {
			leaving = false;
		}
	}

	function getRoleBadgeVariant(role: string): 'default' | 'success' | 'info' | 'warning' {
		switch (role.toLowerCase()) {
			case 'owner': return 'warning';
			case 'admin': return 'info';
			case 'member': return 'default';
			default: return 'default';
		}
	}
</script>

<svelte:head>
	<title>Organizations - CPI Auth Account</title>
</svelte:head>

<div class="page">
	<div class="page-header">
		<h1>Organizations</h1>
		<p>Manage your organization memberships.</p>
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
	{:else if organizations.length === 0}
		<EmptyState
			icon="building"
			title="No organizations"
			description="You are not a member of any organization yet."
		/>
	{:else}
		<div class="org-list">
			{#each organizations as org}
				<div class="org-card" class:active={activeOrgId === org.id}>
					<div class="org-header">
						<div class="org-logo">
							{#if org.logo_url}
								<img src={org.logo_url} alt={org.name} />
							{:else}
								<span class="org-initial">{org.name[0]?.toUpperCase()}</span>
							{/if}
						</div>
						<div class="org-info">
							<div class="org-name-row">
								<h3 class="org-name">{org.name}</h3>
								{#if activeOrgId === org.id}
									<Badge variant="success">Active</Badge>
								{/if}
							</div>
							<div class="org-meta">
								<Badge variant={getRoleBadgeVariant(org.role)}>{org.role}</Badge>
								<span class="org-slug">{org.slug}</span>
								<span class="org-member-since">Member since {formatDate(org.member_since)}</span>
							</div>
						</div>
					</div>

					<div class="org-actions">
						{#if activeOrgId !== org.id}
							<button class="btn btn-secondary btn-sm" onclick={() => handleSwitchOrg(org.id)}>
								Switch to
							</button>
						{/if}
						{#if org.role.toLowerCase() !== 'owner'}
							<button class="btn btn-danger-outline btn-sm" onclick={() => (orgToLeave = org)}>
								Leave
							</button>
						{/if}
					</div>
				</div>
			{/each}
		</div>
	{/if}

	<!-- Leave Org Dialog -->
	<ConfirmDialog
		open={!!orgToLeave}
		title="Leave Organization"
		confirmText="Leave"
		variant="danger"
		loading={leaving}
		onconfirm={handleLeaveOrg}
		oncancel={() => (orgToLeave = null)}
	>
		<p>
			Are you sure you want to leave <strong>{orgToLeave?.name}</strong>? You'll lose access to all
			organization resources and will need to be re-invited to rejoin.
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

	.loading-center {
		display: flex;
		justify-content: center;
		padding: 4rem 0;
	}

	.org-list {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.org-card {
		background-color: var(--color-surface);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-xl);
		padding: 1.25rem;
		transition: border-color 0.15s;
	}

	.org-card:hover {
		border-color: var(--color-border-hover);
	}

	.org-card.active {
		border-color: var(--color-primary);
		background-color: color-mix(in srgb, var(--color-primary) 3%, var(--color-surface));
	}

	.org-header {
		display: flex;
		align-items: center;
		gap: 1rem;
		margin-bottom: 1rem;
	}

	.org-logo {
		width: 3rem;
		height: 3rem;
		border-radius: var(--radius-lg);
		background: linear-gradient(135deg, var(--color-primary), var(--color-secondary));
		display: flex;
		align-items: center;
		justify-content: center;
		overflow: hidden;
		flex-shrink: 0;
	}

	.org-logo img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}

	.org-initial {
		color: white;
		font-size: 1.25rem;
		font-weight: 700;
	}

	.org-info {
		flex: 1;
		min-width: 0;
	}

	.org-name-row {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		margin-bottom: 0.375rem;
	}

	.org-name {
		font-size: 1rem;
		font-weight: 600;
		color: var(--color-text);
		margin: 0;
	}

	.org-meta {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		flex-wrap: wrap;
	}

	.org-slug,
	.org-member-since {
		font-size: 0.75rem;
		color: var(--color-text-secondary);
	}

	.org-actions {
		display: flex;
		gap: 0.75rem;
		justify-content: flex-end;
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
		opacity: 0.6;
		cursor: not-allowed;
	}

	.btn-sm {
		padding: 0.375rem 0.75rem;
		font-size: 0.8125rem;
	}

	.btn-secondary {
		background-color: var(--color-bg-secondary);
		border-color: var(--color-border);
		color: var(--color-text);
	}

	.btn-secondary:hover:not(:disabled) {
		background-color: var(--color-bg-tertiary);
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
