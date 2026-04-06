import type { Handle, Reroute } from '@sveltejs/kit';
import { parseTenantFromPath, stripTenantPrefix, tenantPath } from '$lib/tenant';

/**
 * Reroute hook: when the request URL has a /t/{slug}/ prefix, route the
 * request as if it came in without the prefix so SvelteKit's filesystem
 * router matches the same Svelte page that handles the legacy URL. The
 * URL the browser sees is unchanged; tenant detection happens in handle.
 */
export const reroute: Reroute = ({ url }) => {
	if (url.pathname.startsWith('/t/')) {
		return stripTenantPrefix(url.pathname);
	}
};

export const handle: Handle = async ({ event, resolve }) => {
	// Detect tenant from the original (browser-visible) URL.
	const { slug, base } = parseTenantFromPath(event.url.pathname);
	if (slug) {
		event.locals.tenantSlug = slug;
		event.locals.tenantBase = base;
	}

	// Generate CSRF token for forms
	event.locals.csrfToken = crypto.randomUUID();

	const response = await resolve(event, {
		// Rewrite root-relative href / src / action attributes in HTML so
		// client-side navigation stays within the same tenant prefix. Static
		// asset paths (/_app, /favicon.svg) are emitted by SvelteKit at /
		// and are served by the same login-ui pod regardless of prefix, so
		// we don't rewrite them.
		transformPageChunk: ({ html }) => {
			if (!base) return html;
			return html
				.replace(/(href|action)="(\/(?!\/|t\/|_app|favicon|@|api))/g, `$1="${base}$2`)
				.replace(/(href|action)='(\/(?!\/|t\/|_app|favicon|@|api))/g, `$1='${base}$2`);
		}
	});

	// Rewrite Location headers from server-side redirects so SvelteKit's
	// `redirect(303, '/login?...')` calls inside +page.server.ts files
	// stay within the tenant prefix without each call having to know about it.
	if (base) {
		const location = response.headers.get('location');
		if (location) {
			const rewritten = tenantPath(base, location);
			if (rewritten !== location) {
				response.headers.set('location', rewritten);
			}
		}
	}

	response.headers.set('X-Frame-Options', 'DENY');
	response.headers.set('X-Content-Type-Options', 'nosniff');
	response.headers.set('Referrer-Policy', 'strict-origin-when-cross-origin');
	response.headers.set(
		'Permissions-Policy',
		'camera=(), microphone=(), geolocation=(), payment=(self)'
	);

	return response;
};
