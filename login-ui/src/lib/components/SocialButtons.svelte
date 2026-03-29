<script lang="ts">
	import { env } from '$env/dynamic/public';
	import type { SocialProvider } from '$lib/api/types';
	import type { OAuthParams } from '$lib/stores/oauth';
	import LoadingSpinner from './LoadingSpinner.svelte';

	interface Props {
		providers: SocialProvider[];
		oauthParams?: OAuthParams | null;
		mode?: 'login' | 'register';
	}

	let { providers, oauthParams = null, mode = 'login' }: Props = $props();
	let loadingProvider = $state<string | null>(null);

	const providerIcons: Record<string, { path: string; viewBox: string; color: string }> = {
		google: {
			viewBox: '0 0 24 24',
			path: 'M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 01-2.2 3.32v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.1z M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z',
			color: '#4285F4'
		},
		apple: {
			viewBox: '0 0 24 24',
			path: 'M17.05 20.28c-.98.95-2.05.88-3.08.4-1.09-.5-2.08-.48-3.24 0-1.44.62-2.2.44-3.06-.4C2.79 15.25 3.51 7.59 9.05 7.31c1.35.07 2.29.74 3.08.8 1.18-.24 2.31-.93 3.57-.84 1.51.12 2.65.72 3.4 1.8-3.12 1.87-2.38 5.98.48 7.13-.57 1.5-1.31 2.99-2.54 4.09zM12.03 7.25c-.15-2.23 1.66-4.07 3.74-4.25.29 2.58-2.34 4.5-3.74 4.25z',
			color: '#000000'
		},
		microsoft: {
			viewBox: '0 0 24 24',
			path: 'M3 3h8.5v8.5H3V3zm9.5 0H21v8.5h-8.5V3zM3 12.5h8.5V21H3v-8.5zm9.5 0H21V21h-8.5v-8.5z',
			color: '#00A4EF'
		},
		github: {
			viewBox: '0 0 24 24',
			path: 'M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z',
			color: '#333333'
		},
		gitlab: {
			viewBox: '0 0 24 24',
			path: 'M22.65 14.39L12 22.13 1.35 14.39a.84.84 0 01-.3-.94l1.22-3.78 2.44-7.51A.42.42 0 014.82 2a.43.43 0 01.58 0 .42.42 0 01.11.18l2.44 7.49h8.1l2.44-7.51A.42.42 0 0118.6 2a.43.43 0 01.58 0 .42.42 0 01.11.18l2.44 7.51L23 13.45a.84.84 0 01-.35.94z',
			color: '#FC6D26'
		},
		twitter: {
			viewBox: '0 0 24 24',
			path: 'M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z',
			color: '#000000'
		}
	};

	function handleSocialLogin(provider: SocialProvider) {
		loadingProvider = provider.id;
		const baseUrl = env.PUBLIC_API_URL || 'http://localhost:5050';
		const params = new URLSearchParams({
			provider: provider.id,
			...(oauthParams?.client_id && { client_id: oauthParams.client_id }),
			...(oauthParams?.redirect_uri && { redirect_uri: oauthParams.redirect_uri }),
			...(oauthParams?.scope && { scope: oauthParams.scope }),
			...(oauthParams?.state && { state: oauthParams.state }),
			...(oauthParams?.code_challenge && { code_challenge: oauthParams.code_challenge }),
			...(oauthParams?.code_challenge_method && {
				code_challenge_method: oauthParams.code_challenge_method
			})
		});
		window.location.href = `${baseUrl}/api/v1/auth/social/${provider.id}?${params.toString()}`;
	}
</script>

{#if providers.length > 0}
	<div class="flex flex-col gap-3">
		{#each providers as provider}
			{@const iconData = providerIcons[provider.id]}
			<button
				type="button"
				class="af-btn af-btn-secondary w-full"
				disabled={loadingProvider !== null}
				onclick={() => handleSocialLogin(provider)}
				aria-label="{mode === 'login' ? 'Sign in' : 'Sign up'} with {provider.name}"
			>
				{#if loadingProvider === provider.id}
					<LoadingSpinner size="sm" />
				{:else if iconData}
					<svg
						class="h-5 w-5"
						viewBox={iconData.viewBox}
						fill={provider.color || iconData.color}
						aria-hidden="true"
					>
						<path d={iconData.path} />
					</svg>
				{:else if provider.icon}
					<img src={provider.icon} alt="" class="h-5 w-5" aria-hidden="true" />
				{/if}
				<span>
					{mode === 'login' ? 'Sign in' : 'Sign up'} with {provider.name}
				</span>
			</button>
		{/each}
	</div>
{/if}
