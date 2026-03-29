<script lang="ts">
	import { onMount } from 'svelte';
	import type { Snippet } from 'svelte';
	import { applyBranding, branding, darkMode } from '$lib/stores';
	import type { BrandingConfig } from '$lib/api/types';

	interface Props {
		initialBranding?: BrandingConfig | null;
		children: Snippet;
	}

	let { initialBranding = null, children }: Props = $props();

	onMount(() => {
		if (initialBranding) {
			applyBranding(initialBranding);
		}

		// Check system dark mode preference
		if (!initialBranding?.dark_mode && typeof window !== 'undefined') {
			const prefersDark = window.matchMedia('(prefers-color-scheme: dark)');
			if (prefersDark.matches) {
				darkMode.set(true);
				document.documentElement.classList.add('dark');
			}
			prefersDark.addEventListener('change', (e) => {
				darkMode.set(e.matches);
				if (e.matches) {
					document.documentElement.classList.add('dark');
				} else {
					document.documentElement.classList.remove('dark');
				}
			});
		}
	});
</script>

{@render children()}
