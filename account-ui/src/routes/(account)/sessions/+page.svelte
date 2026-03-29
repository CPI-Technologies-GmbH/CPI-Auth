<script lang="ts">
	import { api } from '$lib/api';
	import type { Session } from '$lib/types';
	import SessionCard from '$components/SessionCard.svelte';
	import Alert from '$components/Alert.svelte';
	import ConfirmDialog from '$components/ConfirmDialog.svelte';
	import EmptyState from '$components/EmptyState.svelte';
	import LoadingSpinner from '$components/LoadingSpinner.svelte';

	let sessions = $state<Session[]>([]);
	let loading = $state(true);
	let error = $state('');
	let successMessage = $state('');

	let sessionToRevoke = $state<string | null>(null);
	let revoking = $state(false);

	let showRevokeAll = $state(false);
	let revokingAll = $state(false);

	async function loadSessions() {
		loading = true;
		error = '';
		try {
			sessions = await api.getSessions();
		} catch (err: unknown) {
			error = (err as { message?: string })?.message || 'Failed to load sessions.';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		loadSessions();
	});

	async function handleRevokeSession() {
		if (!sessionToRevoke) return;
		revoking = true;
		try {
			await api.revokeSession(sessionToRevoke);
			successMessage = 'Session has been revoked.';
			sessionToRevoke = null;
			await loadSessions();
		} catch (err: unknown) {
			error = (err as { message?: string })?.message || 'Failed to revoke session.';
		} finally {
			revoking = false;
		}
	}

	async function handleRevokeAllOther() {
		revokingAll = true;
		try {
			await api.revokeAllOtherSessions();
			successMessage = 'All other sessions have been revoked.';
			showRevokeAll = false;
			await loadSessions();
		} catch (err: unknown) {
			error = (err as { message?: string })?.message || 'Failed to revoke sessions.';
		} finally {
			revokingAll = false;
		}
	}

	const otherSessions = $derived(sessions.filter((s) => !s.is_current));
	const currentSession = $derived(sessions.find((s) => s.is_current));
</script>

<svelte:head>
	<title>Sessions - CPI Auth Account</title>
</svelte:head>

<div class="page">
	<div class="page-header">
		<div class="page-header-row">
			<div>
				<h1>Active Sessions</h1>
				<p>Manage your active sessions across different devices.</p>
			</div>
			{#if otherSessions.length > 0}
				<button class="btn btn-danger-outline" onclick={() => (showRevokeAll = true)}>
					Sign Out All Others
				</button>
			{/if}
		</div>
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
	{:else if sessions.length === 0}
		<EmptyState
			icon="inbox"
			title="No active sessions"
			description="You don't have any active sessions."
		/>
	{:else}
		<!-- Current Session -->
		{#if currentSession}
			<section class="section">
				<h2 class="section-title">Current Session</h2>
				<SessionCard session={currentSession} onrevoke={() => {}} />
			</section>
		{/if}

		<!-- Other Sessions -->
		{#if otherSessions.length > 0}
			<section class="section">
				<h2 class="section-title">Other Sessions ({otherSessions.length})</h2>
				<div class="sessions-list">
					{#each otherSessions as session}
						<SessionCard
							{session}
							onrevoke={(id) => (sessionToRevoke = id)}
						/>
					{/each}
				</div>
			</section>
		{/if}
	{/if}

	<!-- Revoke Session Dialog -->
	<ConfirmDialog
		open={!!sessionToRevoke}
		title="Revoke Session"
		confirmText="Revoke"
		variant="danger"
		loading={revoking}
		onconfirm={handleRevokeSession}
		oncancel={() => (sessionToRevoke = null)}
	>
		<p>Are you sure you want to revoke this session? The device will be signed out immediately.</p>
	</ConfirmDialog>

	<!-- Revoke All Dialog -->
	<ConfirmDialog
		open={showRevokeAll}
		title="Sign Out All Other Sessions"
		confirmText="Sign Out All"
		variant="danger"
		loading={revokingAll}
		onconfirm={handleRevokeAllOther}
		oncancel={() => (showRevokeAll = false)}
	>
		<p>This will sign out all sessions except your current one. All other devices will need to sign in again.</p>
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

	.page-header-row {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		gap: 1rem;
		flex-wrap: wrap;
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

	.section {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.section-title {
		font-size: 0.875rem;
		font-weight: 600;
		color: var(--color-text-secondary);
		text-transform: uppercase;
		letter-spacing: 0.05em;
		margin: 0;
	}

	.sessions-list {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.loading-center {
		display: flex;
		justify-content: center;
		padding: 4rem 0;
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

	.btn-danger-outline {
		background: transparent;
		border-color: var(--color-danger);
		color: var(--color-danger);
	}

	.btn-danger-outline:hover:not(:disabled) {
		background-color: color-mix(in srgb, var(--color-danger) 10%, transparent);
	}
</style>
