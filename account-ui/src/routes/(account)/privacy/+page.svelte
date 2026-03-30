<script lang="ts">
	import { api } from '$lib/api';
	import type { Consent, User } from '$lib/types';
	import Alert from '$components/Alert.svelte';
	import Badge from '$components/Badge.svelte';
	import ConfirmDialog from '$components/ConfirmDialog.svelte';
	import EmptyState from '$components/EmptyState.svelte';
	import LoadingSpinner from '$components/LoadingSpinner.svelte';
	import PasswordInput from '$components/PasswordInput.svelte';
	import { formatDate, formatDateTime } from '$lib/utils/format';

	let { data }: { data: { user: User } } = $props();

	let successMessage = $state('');
	let error = $state('');

	// Data Export
	let exportLoading = $state(false);
	let exportRequested = $state(false);

	// Account Deletion
	let showDeleteConfirm = $state(false);
	let deletePassword = $state('');
	let deleting = $state(false);
	let deleteStep = $state(1);

	// Consents
	let consents = $state<Consent[]>([]);
	let consentsLoading = $state(true);
	let consentToRevoke = $state<Consent | null>(null);
	let revokingConsent = $state(false);

	async function loadConsents() {
		consentsLoading = true;
		try {
			consents = await api.getConsents();
		} catch (err: unknown) {
			error = (err as { message?: string })?.message || 'Failed to load consents.';
		} finally {
			consentsLoading = false;
		}
	}

	$effect(() => {
		loadConsents();
	});

	async function handleRequestExport() {
		exportLoading = true;
		error = '';
		try {
			await api.requestDataExport();
			exportRequested = true;
			successMessage = 'Data export requested. You will receive an email when it is ready for download.';
		} catch (err: unknown) {
			error = (err as { message?: string })?.message || 'Failed to request data export.';
		} finally {
			exportLoading = false;
		}
	}

	async function handleDeleteAccount() {
		deleting = true;
		error = '';
		try {
			await api.deleteAccount(deletePassword || undefined);
			// Redirect to login after deletion
			window.location.href = '/api/logout';
		} catch (err: unknown) {
			error = (err as { message?: string })?.message || 'Failed to delete account.';
		} finally {
			deleting = false;
		}
	}

	async function handleRevokeConsent() {
		if (!consentToRevoke) return;
		revokingConsent = true;
		try {
			await api.revokeConsent(consentToRevoke.id);
			successMessage = `Access for ${consentToRevoke.client_name} has been revoked.`;
			consentToRevoke = null;
			await loadConsents();
		} catch (err: unknown) {
			error = (err as { message?: string })?.message || 'Failed to revoke consent.';
		} finally {
			revokingConsent = false;
		}
	}

	function formatScopes(scopes: string[]): string {
		return scopes.map((s) => s.replace(/_/g, ' ').replace(/:/g, ' - ')).join(', ');
	}
</script>

<svelte:head>
	<title>Privacy & Data - CPI Auth Account</title>
</svelte:head>

<div class="page">
	<div class="page-header">
		<h1>Privacy & Data</h1>
		<p>Manage your data, consents, and account deletion.</p>
	</div>

	{#if successMessage}
		<Alert variant="success" dismissible ondismiss={() => (successMessage = '')}>{successMessage}</Alert>
	{/if}
	{#if error}
		<Alert variant="danger" dismissible ondismiss={() => (error = '')}>{error}</Alert>
	{/if}

	<!-- Data Export (GDPR Art. 20) -->
	<section class="card">
		<div class="card-icon-header">
			<div class="card-icon">
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
					<path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4M7 10l5 5 5-5M12 15V3" />
				</svg>
			</div>
			<div>
				<h2 class="card-title">Export Your Data</h2>
				<p class="card-description">Download a copy of all your personal data in JSON format (GDPR Art. 20).</p>
			</div>
		</div>

		<div class="export-info">
			<h3 class="info-subtitle">What's included:</h3>
			<ul class="info-list">
				<li>Profile information (name, email, phone)</li>
				<li>Account settings and preferences</li>
				<li>Active sessions and login history</li>
				<li>Linked identity providers</li>
				<li>Organization memberships</li>
				<li>Consent records</li>
				<li>Activity log</li>
			</ul>
		</div>

		<div class="card-actions">
			<button
				class="btn btn-primary"
				onclick={handleRequestExport}
				disabled={exportLoading || exportRequested}
			>
				{#if exportLoading}
					<LoadingSpinner size={16} color="white" />
				{/if}
				{exportRequested ? 'Export Requested' : 'Request Data Export'}
			</button>
		</div>
	</section>

	<!-- Consent Records -->
	<section class="card">
		<h2 class="card-title">Application Consents</h2>
		<p class="card-description">Applications you have authorized to access your data.</p>

		{#if consentsLoading}
			<div class="loading-center">
				<LoadingSpinner size={32} />
			</div>
		{:else if consents.length === 0}
			<EmptyState
				icon="shield"
				title="No active consents"
				description="You haven't authorized any third-party applications to access your data."
			/>
		{:else}
			<div class="consents-list">
				{#each consents as consent}
					<div class="consent-item">
						<div class="consent-logo">
							{#if consent.client_logo_url}
								<img src={consent.client_logo_url} alt={consent.client_name} />
							{:else}
								<span class="consent-initial">{consent.client_name[0]?.toUpperCase()}</span>
							{/if}
						</div>
						<div class="consent-info">
							<h3 class="consent-name">{consent.client_name}</h3>
							<div class="consent-scopes">
								{#each consent.scopes as scope}
									<Badge variant="default">{scope}</Badge>
								{/each}
							</div>
							<span class="consent-date">Granted {formatDateTime(consent.granted_at)}</span>
						</div>
						<button class="btn btn-danger-outline btn-sm" onclick={() => (consentToRevoke = consent)}>
							Revoke
						</button>
					</div>
				{/each}
			</div>
		{/if}
	</section>

	<!-- Account Deletion (GDPR Art. 17) -->
	<section class="card card-danger">
		<div class="card-icon-header">
			<div class="card-icon card-icon-danger">
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
					<path d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.34 16.5c-.77.833.192 2.5 1.732 2.5z" />
				</svg>
			</div>
			<div>
				<h2 class="card-title">Delete Account</h2>
				<p class="card-description">Permanently delete your account and all associated data (GDPR Art. 17).</p>
			</div>
		</div>

		<Alert variant="warning">
			This action is irreversible. All your data will be permanently deleted after a 30-day grace period.
			During the grace period, your account will be deactivated but can be recovered by contacting support.
		</Alert>

		<div class="delete-consequences">
			<h3 class="info-subtitle">What happens when you delete your account:</h3>
			<ul class="info-list danger-list">
				<li>Your profile and personal data will be permanently deleted</li>
				<li>All active sessions will be terminated</li>
				<li>All linked accounts will be disconnected</li>
				<li>Organization memberships will be removed</li>
				<li>Application consents will be revoked</li>
				<li>This action cannot be undone after the grace period</li>
			</ul>
		</div>

		<div class="card-actions">
			{#if !showDeleteConfirm}
				<button class="btn btn-danger" onclick={() => { showDeleteConfirm = true; deleteStep = 1; }}>
					Delete My Account
				</button>
			{:else}
				<div class="delete-confirm-section">
					{#if deleteStep === 1}
						<p class="delete-confirm-text">Are you absolutely sure? Type your password to confirm.</p>
						<PasswordInput
							bind:value={deletePassword}
							label="Confirm Password"
							name="delete-password"
							autocomplete="current-password"
							required
						/>
						<div class="delete-actions">
							<button class="btn btn-secondary" onclick={() => { showDeleteConfirm = false; deletePassword = ''; }}>
								Cancel
							</button>
							<button
								class="btn btn-danger"
								onclick={handleDeleteAccount}
								disabled={deleting || !deletePassword}
							>
								{#if deleting}
									<LoadingSpinner size={16} color="white" />
								{/if}
								Permanently Delete Account
							</button>
						</div>
					{/if}
				</div>
			{/if}
		</div>
	</section>

	<!-- Revoke Consent Dialog -->
	<ConfirmDialog
		open={!!consentToRevoke}
		title="Revoke Consent"
		confirmText="Revoke Access"
		variant="danger"
		loading={revokingConsent}
		onconfirm={handleRevokeConsent}
		oncancel={() => (consentToRevoke = null)}
	>
		<p>
			Are you sure you want to revoke <strong>{consentToRevoke?.client_name}</strong>'s access to your data?
			The application will no longer be able to access the following scopes:
		</p>
		{#if consentToRevoke}
			<div class="revoke-scopes">
				{#each consentToRevoke.scopes as scope}
					<Badge variant="default">{scope}</Badge>
				{/each}
			</div>
		{/if}
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

	.card-danger {
		border-color: color-mix(in srgb, var(--color-danger) 30%, var(--color-border));
	}

	.card-title {
		font-size: 1rem;
		font-weight: 600;
		color: var(--color-text);
		margin: 0 0 0.25rem;
	}

	.card-description {
		font-size: 0.8125rem;
		color: var(--color-text-secondary);
		margin: 0 0 1rem;
	}

	.card-icon-header {
		display: flex;
		align-items: flex-start;
		gap: 1rem;
		margin-bottom: 1rem;
	}

	.card-icon {
		width: 2.5rem;
		height: 2.5rem;
		padding: 0.5rem;
		border-radius: var(--radius-lg);
		background-color: color-mix(in srgb, var(--color-primary) 10%, transparent);
		color: var(--color-primary);
		flex-shrink: 0;
	}

	.card-icon svg {
		width: 100%;
		height: 100%;
	}

	.card-icon-danger {
		background-color: color-mix(in srgb, var(--color-danger) 10%, transparent);
		color: var(--color-danger);
	}

	.card-actions {
		margin-top: 1.25rem;
	}

	.loading-center {
		display: flex;
		justify-content: center;
		padding: 2rem 0;
	}

	.export-info,
	.delete-consequences {
		margin-bottom: 0.5rem;
	}

	.info-subtitle {
		font-size: 0.8125rem;
		font-weight: 600;
		color: var(--color-text);
		margin: 0 0 0.5rem;
	}

	.info-list {
		font-size: 0.8125rem;
		color: var(--color-text-secondary);
		padding-left: 1.25rem;
		margin: 0;
		line-height: 1.75;
	}

	.danger-list {
		color: var(--color-text-secondary);
	}

	/* Consents */
	.consents-list {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.consent-item {
		display: flex;
		align-items: center;
		gap: 1rem;
		padding: 1rem;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-lg);
	}

	.consent-logo {
		width: 2.5rem;
		height: 2.5rem;
		border-radius: var(--radius-md);
		background: linear-gradient(135deg, var(--color-bg-tertiary), var(--color-bg-secondary));
		display: flex;
		align-items: center;
		justify-content: center;
		overflow: hidden;
		flex-shrink: 0;
	}

	.consent-logo img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}

	.consent-initial {
		font-weight: 600;
		color: var(--color-text-secondary);
	}

	.consent-info {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 0.375rem;
	}

	.consent-name {
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--color-text);
		margin: 0;
	}

	.consent-scopes {
		display: flex;
		flex-wrap: wrap;
		gap: 0.375rem;
	}

	.consent-date {
		font-size: 0.75rem;
		color: var(--color-text-secondary);
	}

	.revoke-scopes {
		display: flex;
		flex-wrap: wrap;
		gap: 0.375rem;
		margin-top: 0.75rem;
	}

	/* Delete section */
	.delete-confirm-section {
		display: flex;
		flex-direction: column;
		gap: 1rem;
		padding: 1rem;
		background-color: color-mix(in srgb, var(--color-danger) 5%, var(--color-bg));
		border: 1px solid color-mix(in srgb, var(--color-danger) 20%, var(--color-border));
		border-radius: var(--radius-lg);
	}

	.delete-confirm-text {
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--color-text);
		margin: 0;
	}

	.delete-actions {
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

	.btn-primary {
		background-color: var(--color-primary);
		color: var(--color-text-on-primary);
	}

	.btn-primary:hover:not(:disabled) {
		background-color: var(--color-primary-hover);
	}

	.btn-secondary {
		background-color: var(--color-bg-secondary);
		border-color: var(--color-border);
		color: var(--color-text);
	}

	.btn-secondary:hover:not(:disabled) {
		background-color: var(--color-bg-tertiary);
	}

	.btn-danger {
		background-color: var(--color-danger);
		color: white;
	}

	.btn-danger:hover:not(:disabled) {
		background-color: var(--color-danger-hover);
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
