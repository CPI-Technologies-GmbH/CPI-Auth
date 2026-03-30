<script lang="ts">
	import { generateQRCode, qrToSvg } from '$lib/utils/qr';

	let {
		data,
		size = 200
	}: {
		data: string;
		size?: number;
	} = $props();

	const svgHtml = $derived.by(() => {
		if (!data) return '';
		try {
			const matrix = generateQRCode(data);
			return qrToSvg(matrix);
		} catch {
			return '';
		}
	});
</script>

<div class="qr-container" style="width: {size}px; height: {size}px;">
	{#if svgHtml}
		{@html svgHtml}
	{:else}
		<div class="qr-error">
			<p>Unable to generate QR code</p>
		</div>
	{/if}
</div>

<style>
	.qr-container {
		display: flex;
		align-items: center;
		justify-content: center;
		background: white;
		border-radius: var(--radius-lg);
		padding: 0.5rem;
		border: 1px solid var(--color-border);
	}

	.qr-container :global(svg) {
		width: 100%;
		height: 100%;
	}

	.qr-error {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 100%;
		height: 100%;
		color: var(--color-text-tertiary);
		font-size: 0.75rem;
		text-align: center;
	}
</style>
