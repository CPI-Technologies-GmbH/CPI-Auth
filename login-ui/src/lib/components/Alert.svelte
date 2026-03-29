<script lang="ts">
	import type { Snippet } from 'svelte';

	type AlertType = 'error' | 'success' | 'warning' | 'info';

	interface Props {
		type?: AlertType;
		message?: string;
		dismissible?: boolean;
		children?: Snippet;
	}

	let { type = 'info', message = '', dismissible = false, children }: Props = $props();
	let visible = $state(true);

	const colorMap: Record<AlertType, { bg: string; border: string; text: string; icon: string }> = {
		error: {
			bg: 'var(--af-color-error-light)',
			border: 'var(--af-color-error)',
			text: 'var(--af-color-error)',
			icon: 'M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z'
		},
		success: {
			bg: 'var(--af-color-success-light)',
			border: 'var(--af-color-success)',
			text: 'var(--af-color-success)',
			icon: 'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z'
		},
		warning: {
			bg: 'var(--af-color-warning-light)',
			border: 'var(--af-color-warning)',
			text: 'var(--af-color-warning)',
			icon: 'M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z'
		},
		info: {
			bg: 'var(--af-color-info-light)',
			border: 'var(--af-color-info)',
			text: 'var(--af-color-info)',
			icon: 'M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z'
		}
	};

	function dismiss() {
		visible = false;
	}

	const colors = $derived(colorMap[type]);
</script>

{#if visible && (message || children)}
	<div
		class="flex items-start gap-3 rounded-lg border p-4"
		style="background-color: {colors.bg}; border-color: {colors.border}; color: {colors.text}"
		role="alert"
		aria-live="polite"
	>
		<svg
			class="mt-0.5 h-5 w-5 shrink-0"
			fill="none"
			viewBox="0 0 24 24"
			stroke="currentColor"
			stroke-width="2"
			aria-hidden="true"
		>
			<path stroke-linecap="round" stroke-linejoin="round" d={colors.icon} />
		</svg>

		<div class="flex-1 text-sm font-medium">
			{#if children}
				{@render children()}
			{:else}
				{message}
			{/if}
		</div>

		{#if dismissible}
			<button
				type="button"
				onclick={dismiss}
				class="shrink-0 cursor-pointer opacity-70 hover:opacity-100"
				aria-label="Dismiss"
			>
				<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
					<path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
				</svg>
			</button>
		{/if}
	</div>
{/if}
