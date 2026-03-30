<script lang="ts">
	import { branding, darkMode } from '$lib/stores';

	interface Props {
		size?: 'sm' | 'md' | 'lg';
		class?: string;
	}

	let { size = 'md', class: className = '' }: Props = $props();

	const sizeMap = {
		sm: 'h-8',
		md: 'h-10',
		lg: 'h-14'
	};

	let logoUrl = $derived(
		$darkMode && $branding?.logo_dark_url ? $branding.logo_dark_url : $branding?.logo_url || ''
	);
	let appName = $derived($branding?.app_name || 'CPI Auth');
</script>

<div class="flex items-center gap-3 {className}">
	{#if logoUrl}
		<img src={logoUrl} alt="{appName} logo" class="{sizeMap[size]} w-auto object-contain" />
	{:else}
		<div
			class="flex items-center justify-center rounded-xl {sizeMap[size]} aspect-square"
			style="background-color: var(--af-color-primary)"
		>
			<svg
				class="h-2/3 w-2/3"
				viewBox="0 0 24 24"
				fill="none"
				stroke="var(--af-color-text-inverted)"
				stroke-width="2"
				aria-hidden="true"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
				/>
			</svg>
		</div>
		<span class="text-lg font-bold" style="color: var(--af-color-text)">{appName}</span>
	{/if}
</div>
