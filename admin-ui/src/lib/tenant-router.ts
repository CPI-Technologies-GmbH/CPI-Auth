/**
 * Tenant-aware routing for the admin UI.
 *
 * The admin UI runs at the URL shapes:
 *
 *   /admin/login                — sign-in page (no tenant)
 *   /admin/system/...           — platform/super-admin pages (no tenant)
 *   /admin/t/{slug}/...         — tenant-scoped pages (per-tenant data)
 *
 * The active tenant is encoded in the URL via /t/{slug}/ so that:
 *  - bookmarks land on the right tenant
 *  - browser back/forward navigates between tenants
 *  - the URL is the source of truth instead of localStorage
 *
 * The TenantSwitcher rewrites the URL on selection and the auth store
 * derives the active tenant from the URL slug (with the user's home
 * tenant as fallback).
 */

const TENANT_SLUG_RE = /^\/t\/([a-z0-9](?:[a-z0-9-]{0,30}[a-z0-9])?)(\/.*)?$/;

export interface TenantRouteContext {
	/** The tenant slug from the URL, e.g. "lastsoftware". Empty for system routes. */
	slug: string;
	/** The sub-path within the tenant scope, e.g. "/users" or "/applications/123". */
	subPath: string;
}

/** Parse a React Router path (basename-relative) and return tenant context. */
export function parseTenantRoute(pathname: string): TenantRouteContext {
	const match = pathname.match(TENANT_SLUG_RE);
	if (!match) return { slug: '', subPath: pathname };
	return { slug: match[1], subPath: match[2] || '/' };
}

/** Build a tenant-scoped path for a given slug. */
export function buildTenantPath(slug: string, subPath = '/'): string {
	const normalized = subPath.startsWith('/') ? subPath : '/' + subPath;
	return `/t/${slug}${normalized === '/' ? '' : normalized}`;
}

/** Build a system-scoped path. */
export function buildSystemPath(subPath = ''): string {
	if (!subPath || subPath === '/') return '/system';
	return '/system' + (subPath.startsWith('/') ? subPath : '/' + subPath);
}

/**
 * When the user picks a different tenant, compute the equivalent path
 * within the new tenant. E.g. on /t/lastsoftware/users → switch to "default"
 * → /t/default/users.
 */
export function switchTenantPath(currentPath: string, newSlug: string): string {
	const ctx = parseTenantRoute(currentPath);
	if (!ctx.slug) {
		// We're on a system or login path — landing on tenant root is sensible.
		return buildTenantPath(newSlug);
	}
	return buildTenantPath(newSlug, ctx.subPath);
}
