<script lang="ts">
	import type { Snippet } from 'svelte';
	import LoadingSpinner from './LoadingSpinner.svelte';

	let {
		open = false,
		title = 'Confirm',
		confirmText = 'Confirm',
		cancelText = 'Cancel',
		variant = 'danger',
		loading = false,
		onconfirm,
		oncancel,
		children
	}: {
		open: boolean;
		title?: string;
		confirmText?: string;
		cancelText?: string;
		variant?: 'danger' | 'warning' | 'primary';
		loading?: boolean;
		onconfirm: () => void;
		oncancel: () => void;
		children: Snippet;
	} = $props();

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape' && !loading) {
			oncancel();
		}
	}

	function handleBackdropClick() {
		if (!loading) {
			oncancel();
		}
	}
</script>

{#if open}
	<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
	<!-- svelte-ignore a11y_interactive_supports_focus -->
	<div class="overlay" role="dialog" aria-modal="true" aria-labelledby="dialog-title" onkeydown={handleKeydown}>
		<div class="backdrop" onclick={handleBackdropClick} role="presentation"></div>
		<div class="dialog">
			<div class="dialog-header">
				<h3 id="dialog-title">{title}</h3>
			</div>
			<div class="dialog-body">
				{@render children()}
			</div>
			<div class="dialog-footer">
				<button class="btn btn-secondary" onclick={oncancel} disabled={loading}>
					{cancelText}
				</button>
				<button class="btn btn-{variant}" onclick={onconfirm} disabled={loading}>
					{#if loading}
						<LoadingSpinner size={16} color="white" />
					{/if}
					{confirmText}
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.overlay {
		position: fixed;
		inset: 0;
		z-index: 50;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 1rem;
	}

	.backdrop {
		position: fixed;
		inset: 0;
		background-color: rgba(0, 0, 0, 0.5);
		backdrop-filter: blur(2px);
	}

	.dialog {
		position: relative;
		background-color: var(--color-surface);
		border-radius: var(--radius-xl);
		box-shadow: var(--shadow-lg);
		width: 100%;
		max-width: 28rem;
		border: 1px solid var(--color-border);
		animation: dialog-enter 0.15s ease-out;
	}

	@keyframes dialog-enter {
		from {
			opacity: 0;
			transform: scale(0.95);
		}
		to {
			opacity: 1;
			transform: scale(1);
		}
	}

	.dialog-header {
		padding: 1.5rem 1.5rem 0;
	}

	.dialog-header h3 {
		font-size: 1.125rem;
		font-weight: 600;
		color: var(--color-text);
		margin: 0;
	}

	.dialog-body {
		padding: 1rem 1.5rem;
		color: var(--color-text-secondary);
		font-size: 0.875rem;
		line-height: 1.5;
	}

	.dialog-footer {
		display: flex;
		justify-content: flex-end;
		gap: 0.75rem;
		padding: 0 1.5rem 1.5rem;
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

	.btn-danger {
		background-color: var(--color-danger);
		color: white;
	}

	.btn-danger:hover:not(:disabled) {
		background-color: var(--color-danger-hover);
	}

	.btn-warning {
		background-color: var(--color-warning);
		color: white;
	}

	.btn-primary {
		background-color: var(--color-primary);
		color: var(--color-text-on-primary);
	}

	.btn-primary:hover:not(:disabled) {
		background-color: var(--color-primary-hover);
	}
</style>
