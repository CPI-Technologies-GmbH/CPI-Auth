<script lang="ts">
	let { size = 24, color = 'var(--color-primary)' }: { size?: number; color?: string } = $props();
</script>

<svg
	width={size}
	height={size}
	viewBox="0 0 24 24"
	fill="none"
	class="animate-spin"
	role="status"
	aria-label="Loading"
>
	<circle cx="12" cy="12" r="10" stroke={color} stroke-width="3" opacity="0.25" />
	<path
		d="M12 2a10 10 0 0 1 10 10"
		stroke={color}
		stroke-width="3"
		stroke-linecap="round"
	/>
</svg>

<style>
	.animate-spin {
		animation: spin 0.8s linear infinite;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}
</style>
