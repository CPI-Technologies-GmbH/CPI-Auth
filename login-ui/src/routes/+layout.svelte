<script lang="ts">
	import '../app.css';
	import ThemeProvider from '$lib/components/ThemeProvider.svelte';
	import type { LayoutData } from './$types';

	let { children, data }: { children: any; data: LayoutData } = $props();
</script>

<svelte:head>
	<meta name="viewport" content="width=device-width, initial-scale=1" />
	<meta name="theme-color" content="#4f46e5" />
	<link rel="preconnect" href="https://fonts.googleapis.com" />
	<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin="anonymous" />
	<link
		href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap"
		rel="stylesheet"
	/>
	<title>{data.branding?.app_name || 'CPI Auth'}</title>
</svelte:head>

<ThemeProvider initialBranding={data.branding}>
	{@render children()}
</ThemeProvider>
