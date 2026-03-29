<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		variant = 'default',
		children
	}: {
		variant?: 'default' | 'success' | 'warning' | 'danger' | 'info';
		children: Snippet;
	} = $props();

	const variantClasses: Record<string, string> = {
		default: 'badge-default',
		success: 'badge-success',
		warning: 'badge-warning',
		danger: 'badge-danger',
		info: 'badge-info'
	};
</script>

<span class="badge {variantClasses[variant]}">
	{@render children()}
</span>

<style>
	.badge {
		display: inline-flex;
		align-items: center;
		padding: 2px 10px;
		border-radius: 9999px;
		font-size: 0.75rem;
		font-weight: 500;
		line-height: 1.25rem;
		white-space: nowrap;
	}

	.badge-default {
		background-color: var(--color-bg-tertiary);
		color: var(--color-text-secondary);
	}

	.badge-success {
		background-color: color-mix(in srgb, var(--color-success) 15%, transparent);
		color: var(--color-success);
	}

	.badge-warning {
		background-color: color-mix(in srgb, var(--color-warning) 15%, transparent);
		color: var(--color-warning);
	}

	.badge-danger {
		background-color: color-mix(in srgb, var(--color-danger) 15%, transparent);
		color: var(--color-danger);
	}

	.badge-info {
		background-color: color-mix(in srgb, var(--color-secondary) 15%, transparent);
		color: var(--color-secondary);
	}
</style>
