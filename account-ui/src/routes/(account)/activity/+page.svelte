<script lang="ts">
	import { api } from '$lib/api';
	import type { ActivityEvent, ActivityType } from '$lib/types';
	import Alert from '$components/Alert.svelte';
	import Badge from '$components/Badge.svelte';
	import EmptyState from '$components/EmptyState.svelte';
	import LoadingSpinner from '$components/LoadingSpinner.svelte';
	import { formatDateTime, formatRelativeTime, getLocationString } from '$lib/utils/format';

	let events = $state<ActivityEvent[]>([]);
	let loading = $state(true);
	let loadingMore = $state(false);
	let error = $state('');
	let hasMore = $state(true);

	let filterType = $state('');
	let filterFrom = $state('');
	let filterTo = $state('');

	const LIMIT = 20;
	let offset = $state(0);

	const activityTypes = [
		{ value: '', label: 'All Activity' },
		{ value: 'login', label: 'Logins' },
		{ value: 'logout', label: 'Logouts' },
		{ value: 'password_change', label: 'Password Changes' },
		{ value: 'mfa_enable', label: 'MFA Enabled' },
		{ value: 'mfa_disable', label: 'MFA Disabled' },
		{ value: 'profile_update', label: 'Profile Updates' },
		{ value: 'session_revoke', label: 'Session Revocations' },
		{ value: 'email_change', label: 'Email Changes' },
		{ value: 'phone_change', label: 'Phone Changes' },
		{ value: 'passkey_register', label: 'Passkey Registered' },
		{ value: 'passkey_delete', label: 'Passkey Deleted' },
		{ value: 'identity_link', label: 'Account Linked' },
		{ value: 'identity_unlink', label: 'Account Unlinked' },
		{ value: 'consent_grant', label: 'Consent Granted' },
		{ value: 'consent_revoke', label: 'Consent Revoked' }
	];

	async function loadActivity(append = false) {
		if (!append) {
			loading = true;
			offset = 0;
		} else {
			loadingMore = true;
		}
		error = '';

		try {
			const result = await api.getActivity({
				type: filterType || undefined,
				from: filterFrom || undefined,
				to: filterTo || undefined,
				limit: LIMIT,
				offset: append ? offset : 0
			});

			if (append) {
				events = [...events, ...result];
			} else {
				events = result;
			}

			hasMore = result.length === LIMIT;
			offset = events.length;
		} catch (err: unknown) {
			error = (err as { message?: string })?.message || 'Failed to load activity.';
		} finally {
			loading = false;
			loadingMore = false;
		}
	}

	$effect(() => {
		loadActivity();
	});

	function handleFilterChange() {
		loadActivity(false);
	}

	function handleLoadMore() {
		loadActivity(true);
	}

	function getEventIcon(type: string): string {
		switch (type) {
			case 'login': return 'M15 3h4a2 2 0 012 2v14a2 2 0 01-2 2h-4M10 17l5-5-5-5M15 12H3';
			case 'logout': return 'M9 21H5a2 2 0 01-2-2V5a2 2 0 012-2h4M16 17l5-5-5-5M21 12H9';
			case 'password_change': return 'M21 2l-2 2m-7.61 7.61a5.5 5.5 0 11-7.778 7.778 5.5 5.5 0 017.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4';
			case 'mfa_enable':
			case 'mfa_disable': return 'M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z';
			case 'profile_update': return 'M20 21v-2a4 4 0 00-4-4H8a4 4 0 00-4 4v2M12 3a4 4 0 100 8 4 4 0 000-8z';
			case 'session_revoke': return 'M20 3H4a1 1 0 00-1 1v12a1 1 0 001 1h16a1 1 0 001-1V4a1 1 0 00-1-1zM8 21h8M12 17v4';
			case 'email_change': return 'M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2zM22 6l-10 7L2 6';
			case 'phone_change': return 'M22 16.92v3a2 2 0 01-2.18 2 19.79 19.79 0 01-8.63-3.07 19.5 19.5 0 01-6-6 19.79 19.79 0 01-3.07-8.67A2 2 0 014.11 2h3a2 2 0 012 1.72c.127.96.361 1.903.7 2.81a2 2 0 01-.45 2.11L8.09 9.91a16 16 0 006 6l1.27-1.27a2 2 0 012.11-.45c.907.339 1.85.573 2.81.7A2 2 0 0122 16.92z';
			case 'passkey_register':
			case 'passkey_delete': return 'M21 2l-2 2m-7.61 7.61a5.5 5.5 0 11-7.778 7.778 5.5 5.5 0 017.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4';
			case 'identity_link':
			case 'identity_unlink': return 'M10 13a5 5 0 007.54.54l3-3a5 5 0 00-7.07-7.07l-1.72 1.71M14 11a5 5 0 00-7.54-.54l-3 3a5 5 0 007.07 7.07l1.71-1.71';
			case 'consent_grant':
			case 'consent_revoke': return 'M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z';
			default: return 'M22 12h-4l-3 9L9 3l-3 9H2';
		}
	}

	function getEventBadgeVariant(type: string): 'default' | 'success' | 'warning' | 'danger' | 'info' {
		if (type.includes('delete') || type.includes('revoke') || type.includes('disable') || type === 'logout') return 'danger';
		if (type.includes('enable') || type.includes('register') || type.includes('link') || type.includes('grant') || type === 'login') return 'success';
		if (type.includes('change') || type.includes('update')) return 'info';
		return 'default';
	}

	function getEventTypeLabel(type: string): string {
		return type.replace(/_/g, ' ').replace(/\b\w/g, (l) => l.toUpperCase());
	}
</script>

<svelte:head>
	<title>Activity Log - CPI Auth Account</title>
</svelte:head>

<div class="page">
	<div class="page-header">
		<h1>Activity Log</h1>
		<p>Review recent security events and account activity.</p>
	</div>

	{#if error}
		<Alert variant="danger" dismissible ondismiss={() => (error = '')}>{error}</Alert>
	{/if}

	<!-- Filters -->
	<div class="filters">
		<div class="filter-field">
			<label for="filter-type" class="filter-label">Event Type</label>
			<select id="filter-type" class="filter-input" bind:value={filterType} onchange={handleFilterChange}>
				{#each activityTypes as type}
					<option value={type.value}>{type.label}</option>
				{/each}
			</select>
		</div>
		<div class="filter-field">
			<label for="filter-from" class="filter-label">From</label>
			<input id="filter-from" type="date" class="filter-input" bind:value={filterFrom} onchange={handleFilterChange} />
		</div>
		<div class="filter-field">
			<label for="filter-to" class="filter-label">To</label>
			<input id="filter-to" type="date" class="filter-input" bind:value={filterTo} onchange={handleFilterChange} />
		</div>
	</div>

	{#if loading}
		<div class="loading-center">
			<LoadingSpinner size={40} />
		</div>
	{:else if events.length === 0}
		<EmptyState
			icon="activity"
			title="No activity found"
			description={filterType ? 'No events match your current filters.' : 'Your activity log is empty.'}
		/>
	{:else}
		<div class="activity-timeline">
			{#each events as event, idx}
				{@const showDate = idx === 0 || new Date(event.created_at).toDateString() !== new Date(events[idx - 1].created_at).toDateString()}
				{#if showDate}
					<div class="timeline-date">
						{new Date(event.created_at).toLocaleDateString('en-US', { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' })}
					</div>
				{/if}
				<div class="timeline-event">
					<div class="event-dot">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
							<path d={getEventIcon(event.type)} />
						</svg>
					</div>
					<div class="event-content">
						<div class="event-header">
							<span class="event-description">{event.description}</span>
							<Badge variant={getEventBadgeVariant(event.type)}>
								{getEventTypeLabel(event.type)}
							</Badge>
						</div>
						<div class="event-meta">
							<span class="event-time">{formatDateTime(event.created_at)}</span>
							{#if event.ip_address}
								<span class="event-detail">{event.ip_address}</span>
							{/if}
							{#if event.browser || event.os}
								<span class="event-detail">
									{[event.browser, event.os].filter(Boolean).join(' on ')}
								</span>
							{/if}
							{#if event.location}
								<span class="event-detail">{getLocationString(event.location)}</span>
							{/if}
						</div>
					</div>
				</div>
			{/each}
		</div>

		{#if hasMore}
			<div class="load-more">
				<button class="btn btn-secondary" onclick={handleLoadMore} disabled={loadingMore}>
					{#if loadingMore}
						<LoadingSpinner size={16} />
					{/if}
					Load More
				</button>
			</div>
		{/if}
	{/if}
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

	/* Filters */
	.filters {
		display: flex;
		gap: 1rem;
		flex-wrap: wrap;
		padding: 1rem;
		background-color: var(--color-surface);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-xl);
	}

	.filter-field {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
		flex: 1;
		min-width: 8rem;
	}

	.filter-label {
		font-size: 0.75rem;
		font-weight: 500;
		color: var(--color-text-secondary);
	}

	.filter-input {
		padding: 0.5rem 0.75rem;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		background-color: var(--color-bg);
		color: var(--color-text);
		font-size: 0.8125rem;
		outline: none;
		transition: border-color 0.15s;
	}

	.filter-input:focus {
		border-color: var(--color-primary);
	}

	/* Timeline */
	.activity-timeline {
		display: flex;
		flex-direction: column;
	}

	.timeline-date {
		font-size: 0.75rem;
		font-weight: 600;
		color: var(--color-text-secondary);
		text-transform: uppercase;
		letter-spacing: 0.05em;
		padding: 1rem 0 0.5rem;
		border-bottom: 1px solid var(--color-border);
		margin-bottom: 0.5rem;
	}

	.timeline-event {
		display: flex;
		gap: 0.875rem;
		padding: 0.75rem 0;
		border-bottom: 1px solid color-mix(in srgb, var(--color-border) 50%, transparent);
	}

	.timeline-event:last-child {
		border-bottom: none;
	}

	.event-dot {
		width: 2rem;
		height: 2rem;
		padding: 0.375rem;
		border-radius: var(--radius-md);
		background-color: var(--color-bg-tertiary);
		color: var(--color-text-secondary);
		flex-shrink: 0;
	}

	.event-dot svg {
		width: 100%;
		height: 100%;
	}

	.event-content {
		flex: 1;
		min-width: 0;
	}

	.event-header {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		margin-bottom: 0.25rem;
		flex-wrap: wrap;
	}

	.event-description {
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--color-text);
	}

	.event-meta {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
	}

	.event-time {
		font-size: 0.75rem;
		color: var(--color-text-secondary);
	}

	.event-detail {
		font-size: 0.75rem;
		color: var(--color-text-tertiary);
	}

	.event-detail::before {
		content: '\00B7';
		margin-right: 0.5rem;
	}

	.load-more {
		display: flex;
		justify-content: center;
		padding: 1rem 0;
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

	.btn-secondary {
		background-color: var(--color-bg-secondary);
		border-color: var(--color-border);
		color: var(--color-text);
	}

	.btn-secondary:hover:not(:disabled) {
		background-color: var(--color-bg-tertiary);
	}
</style>
