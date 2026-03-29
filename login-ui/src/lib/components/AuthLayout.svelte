<script lang="ts">
	import type { Snippet } from 'svelte';
	import { branding } from '$lib/stores';
	import Logo from './Logo.svelte';

	interface Props {
		children: Snippet;
		maxWidth?: 'sm' | 'md' | 'lg';
		showLogo?: boolean;
	}

	let { children, maxWidth = 'sm', showLogo = true }: Props = $props();

	const maxWidthClasses = {
		sm: 'max-w-sm',
		md: 'max-w-md',
		lg: 'max-w-lg'
	};

	let layoutMode = $derived($branding?.layout_mode || 'centered');
</script>

{#if layoutMode === 'centered'}
	<div
		class="flex min-h-screen items-center justify-center px-4 py-8 sm:px-6 lg:px-8"
		style="background-color: var(--af-color-surface)"
	>
		<div class="w-full {maxWidthClasses[maxWidth]} animate-fade-in">
			{#if showLogo}
				<div class="mb-8 flex justify-center">
					<Logo size="lg" />
				</div>
			{/if}
			<div class="af-card">
				{@render children()}
			</div>
		</div>
	</div>
{:else if layoutMode === 'split-screen'}
	<div class="flex min-h-screen">
		<div
			class="hidden w-1/2 items-center justify-center lg:flex"
			style="background-color: var(--af-color-primary)"
		>
			<div class="px-12 text-center">
				<Logo size="lg" />
				<p
					class="mt-4 text-lg opacity-90"
					style="color: var(--af-color-text-inverted)"
				>
					{$branding?.app_name || 'CPI Auth'}
				</p>
			</div>
		</div>
		<div
			class="flex w-full items-center justify-center px-4 py-8 sm:px-6 lg:w-1/2 lg:px-8"
			style="background-color: var(--af-color-surface)"
		>
			<div class="w-full {maxWidthClasses[maxWidth]} animate-fade-in">
				{#if showLogo}
					<div class="mb-8 flex justify-center lg:hidden">
						<Logo size="lg" />
					</div>
				{/if}
				<div class="af-card">
					{@render children()}
				</div>
			</div>
		</div>
	</div>
{:else if layoutMode === 'sidebar'}
	<div class="flex min-h-screen">
		<div
			class="hidden w-80 flex-col items-center justify-between py-8 lg:flex"
			style="background-color: var(--af-color-primary)"
		>
			<div class="px-6">
				<Logo size="lg" />
			</div>
			<div class="px-6 text-center">
				<p class="text-sm opacity-75" style="color: var(--af-color-text-inverted)">
					Powered by CPI Auth
				</p>
			</div>
		</div>
		<div
			class="flex flex-1 items-center justify-center px-4 py-8 sm:px-6 lg:px-8"
			style="background-color: var(--af-color-surface)"
		>
			<div class="w-full {maxWidthClasses[maxWidth]} animate-fade-in">
				{#if showLogo}
					<div class="mb-8 flex justify-center lg:hidden">
						<Logo size="lg" />
					</div>
				{/if}
				<div class="af-card">
					{@render children()}
				</div>
			</div>
		</div>
	</div>
{/if}
