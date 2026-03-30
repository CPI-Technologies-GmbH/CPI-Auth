<script lang="ts">
	import type { Session } from '$lib/types';
	import { formatRelativeTime, getLocationString } from '$lib/utils/format';
	import Badge from './Badge.svelte';

	let {
		session,
		onrevoke
	}: {
		session: Session;
		onrevoke: (id: string) => void;
	} = $props();

	const deviceIcon = $derived.by(() => {
		const type = session.device_type?.toLowerCase() || '';
		if (type === 'mobile') return 'M12 18h.01M8 21h8a2 2 0 002-2V5a2 2 0 00-2-2H8a2 2 0 00-2 2v14a2 2 0 002 2z';
		if (type === 'tablet') return 'M12 18h.01M7 21h10a2 2 0 002-2V5a2 2 0 00-2-2H7a2 2 0 00-2 2v14a2 2 0 002 2z';
		return 'M20 3H4a1 1 0 00-1 1v12a1 1 0 001 1h16a1 1 0 001-1V4a1 1 0 00-1-1zM8 21h8M12 17v4';
	});
</script>

<div class="session-card" class:current={session.is_current}>
	<div class="session-icon">
		<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
			<path d={deviceIcon} />
		</svg>
	</div>
	<div class="session-info">
		<div class="session-header">
			<span class="session-browser">
				{session.browser || 'Unknown Browser'} on {session.os || 'Unknown OS'}
			</span>
			{#if session.is_current}
				<Badge variant="success">Current</Badge>
			{/if}
		</div>
		<div class="session-details">
			<span class="session-detail">
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="detail-icon">
					<path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0118 0z" />
					<circle cx="12" cy="10" r="3" />
				</svg>
				{getLocationString(session.location)}
			</span>
			<span class="session-detail">
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="detail-icon">
					<rect x="2" y="2" width="20" height="20" rx="5" ry="5" />
					<path d="M16 11.37A4 4 0 1112.63 8 4 4 0 0116 11.37zM17.5 6.5h.01" />
				</svg>
				{session.ip_address}
			</span>
			<span class="session-detail">
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="detail-icon">
					<circle cx="12" cy="12" r="10" />
					<polyline points="12,6 12,12 16,14" />
				</svg>
				{formatRelativeTime(session.last_active_at)}
			</span>
		</div>
	</div>
	{#if !session.is_current}
		<button class="revoke-btn" onclick={() => onrevoke(session.id)} title="Revoke session">
			<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
				<path d="M18 6L6 18M6 6l12 12" />
			</svg>
		</button>
	{/if}
</div>

<style>
	.session-card {
		display: flex;
		align-items: flex-start;
		gap: 1rem;
		padding: 1rem;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-lg);
		background-color: var(--color-surface);
		transition: border-color 0.15s;
	}

	.session-card:hover {
		border-color: var(--color-border-hover);
	}

	.session-card.current {
		border-color: var(--color-primary);
		background-color: color-mix(in srgb, var(--color-primary) 5%, var(--color-surface));
	}

	.session-icon {
		width: 2.5rem;
		height: 2.5rem;
		padding: 0.5rem;
		border-radius: var(--radius-md);
		background-color: var(--color-bg-tertiary);
		color: var(--color-text-secondary);
		flex-shrink: 0;
	}

	.session-icon svg {
		width: 100%;
		height: 100%;
	}

	.session-info {
		flex: 1;
		min-width: 0;
	}

	.session-header {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		margin-bottom: 0.375rem;
	}

	.session-browser {
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--color-text);
	}

	.session-details {
		display: flex;
		flex-wrap: wrap;
		gap: 0.75rem;
	}

	.session-detail {
		display: flex;
		align-items: center;
		gap: 0.25rem;
		font-size: 0.75rem;
		color: var(--color-text-secondary);
	}

	.detail-icon {
		width: 0.875rem;
		height: 0.875rem;
		flex-shrink: 0;
	}

	.revoke-btn {
		flex-shrink: 0;
		width: 2rem;
		height: 2rem;
		display: flex;
		align-items: center;
		justify-content: center;
		border: none;
		background: none;
		color: var(--color-text-tertiary);
		cursor: pointer;
		border-radius: var(--radius-sm);
		transition: all 0.15s;
	}

	.revoke-btn:hover {
		background-color: color-mix(in srgb, var(--color-danger) 10%, transparent);
		color: var(--color-danger);
	}

	.revoke-btn svg {
		width: 1rem;
		height: 1rem;
	}
</style>
