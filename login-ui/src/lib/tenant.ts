/**
 * Tenant-aware path helpers for the login UI.
 *
 * The login UI runs at two URL shapes simultaneously:
 *
 *   /login                       — legacy host-derived tenant
 *   /t/{slug}/login              — explicit path-based tenant
 *
 * The reroute hook strips /t/{slug}/ before SvelteKit's route matcher
 * sees the URL, so the same Svelte page handles both. The handle hook
 * stores the slug in event.locals + page data so server-side load
 * functions and client-side scripts can build links that put the
 * prefix back when needed.
 */

const SLUG_RE = /^\/t\/([a-z0-9](?:[a-z0-9-]{0,30}[a-z0-9])?)(\/.*)?$/;

export interface TenantContext {
	/** The tenant slug, e.g. "lastsoftware". Empty when no /t/ prefix. */
	slug: string;
	/** The path prefix to prepend to internal URLs, e.g. "/t/lastsoftware". Empty when no prefix. */
	base: string;
}

/** Parse the original request URL and return the tenant context. */
export function parseTenantFromPath(pathname: string): TenantContext {
	const match = pathname.match(SLUG_RE);
	if (!match) return { slug: '', base: '' };
	return { slug: match[1], base: '/t/' + match[1] };
}

/** Strip the /t/{slug}/ prefix from a pathname, returning the remainder. */
export function stripTenantPrefix(pathname: string): string {
	const match = pathname.match(SLUG_RE);
	if (!match) return pathname;
	return match[2] || '/';
}

/**
 * Build a tenant-aware path. Prepends the tenant base to a same-origin
 * path. External URLs and absolute URLs are returned untouched.
 */
export function tenantPath(base: string | undefined, path: string): string {
	if (!base) return path;
	if (!path.startsWith('/') || path.startsWith('//')) return path;
	if (path.startsWith(base + '/') || path === base) return path;
	return base + path;
}
