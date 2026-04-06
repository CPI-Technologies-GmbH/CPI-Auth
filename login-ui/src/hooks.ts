import type { Reroute } from '@sveltejs/kit';
import { stripTenantPrefix } from '$lib/tenant';

/**
 * Universal reroute hook — runs on both server and client. When the request
 * URL has a /t/{slug}/ prefix, route the request as if it came in without
 * the prefix so SvelteKit's filesystem router matches the same Svelte page
 * that handles the legacy URL. The URL the browser sees is unchanged;
 * tenant detection itself happens in hooks.server.ts via the request URL.
 *
 * Reroute MUST live in hooks.ts (not hooks.server.ts), per the SvelteKit
 * docs, otherwise it does not run for client-side navigation and the
 * server-only version is silently ignored by route resolution.
 */
export const reroute: Reroute = ({ url }) => {
	if (url.pathname.startsWith('/t/')) {
		return stripTenantPrefix(url.pathname);
	}
};
