<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { User } from '$lib/types';
	import Avatar from './Avatar.svelte';
	import { page } from '$app/stores';

	let {
		user,
		children
	}: {
		user: User;
		children: Snippet;
	} = $props();

	let mobileMenuOpen = $state(false);

	const navItems = [
		{ href: '/profile', label: 'Profile', icon: 'M20 21v-2a4 4 0 00-4-4H8a4 4 0 00-4 4v2M12 3a4 4 0 100 8 4 4 0 000-8z' },
		{ href: '/security', label: 'Security', icon: 'M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z' },
		{ href: '/sessions', label: 'Sessions', icon: 'M20 3H4a1 1 0 00-1 1v12a1 1 0 001 1h16a1 1 0 001-1V4a1 1 0 00-1-1zM8 21h8M12 17v4' },
		{ href: '/linked-accounts', label: 'Linked Accounts', icon: 'M10 13a5 5 0 007.54.54l3-3a5 5 0 00-7.07-7.07l-1.72 1.71M14 11a5 5 0 00-7.54-.54l-3 3a5 5 0 007.07 7.07l1.71-1.71' },
		{ href: '/organizations', label: 'Organizations', icon: 'M3 21h18M3 7v14M21 7v14M6 11h2M6 15h2M14 11h2M14 15h2M10 21v-4h4v4M9 7h6V3H9v4' },
		{ href: '/privacy', label: 'Privacy & Data', icon: 'M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z' },
		{ href: '/activity', label: 'Activity Log', icon: 'M22 12h-4l-3 9L9 3l-3 9H2' }
	];

	function isActive(href: string, pathname: string): boolean {
		return pathname === href || pathname.startsWith(href + '/');
	}
</script>

<div class="layout">
	<!-- Mobile header -->
	<header class="mobile-header">
		<button class="menu-toggle" onclick={() => (mobileMenuOpen = !mobileMenuOpen)} aria-label="Toggle menu">
			<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
				{#if mobileMenuOpen}
					<path d="M18 6L6 18M6 6l12 12" />
				{:else}
					<path d="M3 12h18M3 6h18M3 18h18" />
				{/if}
			</svg>
		</button>
		<h1 class="mobile-title">Account</h1>
		<Avatar src={user.avatar_url} name={user.name} email={user.email} size={32} />
	</header>

	<!-- Sidebar -->
	<aside class="sidebar" class:sidebar-open={mobileMenuOpen}>
		<div class="sidebar-header">
			<div class="brand">
				<svg viewBox="0 0 24 24" fill="none" stroke="var(--color-primary)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="brand-icon">
					<path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
				</svg>
				<span class="brand-text">CPI Auth</span>
			</div>
		</div>

		<div class="sidebar-user">
			<Avatar src={user.avatar_url} name={user.name} email={user.email} size={40} />
			<div class="user-info">
				<span class="user-name">{user.name || user.email}</span>
				{#if user.name}
					<span class="user-email">{user.email}</span>
				{/if}
			</div>
		</div>

		<nav class="sidebar-nav">
			{#each navItems as item}
				<a
					href={item.href}
					class="nav-item"
					class:active={isActive(item.href, $page.url.pathname)}
					onclick={() => (mobileMenuOpen = false)}
				>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="nav-icon">
						<path d={item.icon} />
					</svg>
					<span>{item.label}</span>
				</a>
			{/each}
		</nav>

		<div class="sidebar-footer">
			<a href="/api/logout" class="nav-item logout-item" data-sveltekit-preload-data="off">
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="nav-icon">
					<path d="M9 21H5a2 2 0 01-2-2V5a2 2 0 012-2h4M16 17l5-5-5-5M21 12H9" />
				</svg>
				<span>Sign Out</span>
			</a>
		</div>
	</aside>

	<!-- Overlay for mobile menu -->
	{#if mobileMenuOpen}
		<div class="sidebar-overlay" onclick={() => (mobileMenuOpen = false)} role="presentation"></div>
	{/if}

	<!-- Main content -->
	<main class="main-content">
		{@render children()}
	</main>
</div>

<style>
	.layout {
		display: flex;
		min-height: 100vh;
	}

	/* Mobile header */
	.mobile-header {
		display: none;
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		height: 3.5rem;
		padding: 0 1rem;
		background-color: var(--color-surface);
		border-bottom: 1px solid var(--color-border);
		align-items: center;
		justify-content: space-between;
		z-index: 40;
	}

	.menu-toggle {
		width: 2rem;
		height: 2rem;
		display: flex;
		align-items: center;
		justify-content: center;
		background: none;
		border: none;
		color: var(--color-text);
		cursor: pointer;
	}

	.menu-toggle svg {
		width: 1.25rem;
		height: 1.25rem;
	}

	.mobile-title {
		font-size: 1rem;
		font-weight: 600;
		color: var(--color-text);
		margin: 0;
	}

	/* Sidebar */
	.sidebar {
		position: fixed;
		left: 0;
		top: 0;
		bottom: 0;
		width: 16rem;
		background-color: var(--color-surface);
		border-right: 1px solid var(--color-border);
		display: flex;
		flex-direction: column;
		z-index: 50;
		overflow-y: auto;
	}

	.sidebar-header {
		padding: 1.25rem 1.25rem 0;
	}

	.brand {
		display: flex;
		align-items: center;
		gap: 0.625rem;
	}

	.brand-icon {
		width: 1.75rem;
		height: 1.75rem;
	}

	.brand-text {
		font-size: 1.125rem;
		font-weight: 700;
		color: var(--color-text);
	}

	.sidebar-user {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 1.25rem;
		margin: 0.5rem 0.75rem;
		border-radius: var(--radius-lg);
		background-color: var(--color-bg-secondary);
	}

	.user-info {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
	}

	.user-name {
		font-size: 0.875rem;
		font-weight: 600;
		color: var(--color-text);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.user-email {
		font-size: 0.75rem;
		color: var(--color-text-secondary);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.sidebar-nav {
		flex: 1;
		padding: 0.5rem 0.75rem;
		display: flex;
		flex-direction: column;
		gap: 0.125rem;
	}

	.nav-item {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 0.625rem 0.75rem;
		border-radius: var(--radius-md);
		color: var(--color-text-secondary);
		text-decoration: none;
		font-size: 0.875rem;
		font-weight: 500;
		transition: all 0.15s;
	}

	.nav-item:hover {
		background-color: var(--color-bg-secondary);
		color: var(--color-text);
	}

	.nav-item.active {
		background-color: var(--color-primary-light);
		color: var(--color-primary);
	}

	.nav-icon {
		width: 1.25rem;
		height: 1.25rem;
		flex-shrink: 0;
	}

	.sidebar-footer {
		padding: 0.75rem;
		border-top: 1px solid var(--color-border);
	}

	.logout-item {
		color: var(--color-text-secondary);
	}

	.logout-item:hover {
		color: var(--color-danger);
		background-color: color-mix(in srgb, var(--color-danger) 8%, transparent);
	}

	.sidebar-overlay {
		display: none;
	}

	/* Main content */
	.main-content {
		flex: 1;
		margin-left: 16rem;
		padding: 2rem 2.5rem;
		max-width: 56rem;
	}

	@media (max-width: 768px) {
		.mobile-header {
			display: flex;
		}

		.sidebar {
			transform: translateX(-100%);
			transition: transform 0.25s ease;
		}

		.sidebar.sidebar-open {
			transform: translateX(0);
		}

		.sidebar-overlay {
			display: block;
			position: fixed;
			inset: 0;
			background-color: rgba(0, 0, 0, 0.5);
			z-index: 45;
		}

		.main-content {
			margin-left: 0;
			margin-top: 3.5rem;
			padding: 1.5rem 1rem;
		}
	}
</style>
