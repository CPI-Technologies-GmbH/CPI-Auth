<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		variant = 'info',
		dismissible = false,
		ondismiss,
		children
	}: {
		variant?: 'info' | 'success' | 'warning' | 'danger';
		dismissible?: boolean;
		ondismiss?: () => void;
		children: Snippet;
	} = $props();

	let visible = $state(true);

	function dismiss() {
		visible = false;
		ondismiss?.();
	}

	const icons: Record<string, string> = {
		info: 'M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z',
		success: 'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z',
		warning: 'M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.34 16.5c-.77.833.192 2.5 1.732 2.5z',
		danger: 'M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z'
	};
</script>

{#if visible}
	<div class="alert alert-{variant}" role="alert">
		<svg
			class="alert-icon"
			viewBox="0 0 24 24"
			fill="none"
			stroke="currentColor"
			stroke-width="2"
			stroke-linecap="round"
			stroke-linejoin="round"
		>
			<path d={icons[variant]} />
		</svg>
		<div class="alert-content">
			{@render children()}
		</div>
		{#if dismissible}
			<button class="alert-dismiss" onclick={dismiss} aria-label="Dismiss">
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M6 18L18 6M6 6l12 12" />
				</svg>
			</button>
		{/if}
	</div>
{/if}

<style>
	.alert {
		display: flex;
		align-items: flex-start;
		gap: 0.75rem;
		padding: 1rem;
		border-radius: var(--radius-lg);
		border: 1px solid;
		font-size: 0.875rem;
		line-height: 1.5;
	}

	.alert-info {
		background-color: color-mix(in srgb, var(--color-secondary) 10%, var(--color-bg));
		border-color: color-mix(in srgb, var(--color-secondary) 30%, transparent);
		color: var(--color-secondary);
	}

	.alert-success {
		background-color: color-mix(in srgb, var(--color-success) 10%, var(--color-bg));
		border-color: color-mix(in srgb, var(--color-success) 30%, transparent);
		color: var(--color-success);
	}

	.alert-warning {
		background-color: color-mix(in srgb, var(--color-warning) 10%, var(--color-bg));
		border-color: color-mix(in srgb, var(--color-warning) 30%, transparent);
		color: var(--color-warning);
	}

	.alert-danger {
		background-color: color-mix(in srgb, var(--color-danger) 10%, var(--color-bg));
		border-color: color-mix(in srgb, var(--color-danger) 30%, transparent);
		color: var(--color-danger);
	}

	.alert-icon {
		width: 1.25rem;
		height: 1.25rem;
		flex-shrink: 0;
		margin-top: 0.125rem;
	}

	.alert-content {
		flex: 1;
		color: var(--color-text);
	}

	.alert-dismiss {
		flex-shrink: 0;
		width: 1.25rem;
		height: 1.25rem;
		cursor: pointer;
		background: none;
		border: none;
		color: var(--color-text-tertiary);
		padding: 0;
	}

	.alert-dismiss:hover {
		color: var(--color-text);
	}

	.alert-dismiss svg {
		width: 100%;
		height: 100%;
	}
</style>
