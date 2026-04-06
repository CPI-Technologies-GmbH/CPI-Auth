import type { LayoutServerLoad } from './$types';
import type { BrandingConfig } from '$lib/api/types';
import { env as publicEnv } from '$env/dynamic/public';
import { env as privateEnv } from '$env/dynamic/private';

export const load: LayoutServerLoad = async ({ url, fetch, request, locals }) => {
	// Server-side: use internal URL (INTERNAL_API_URL) or fallback to public
	const apiUrl = privateEnv.INTERNAL_API_URL || publicEnv.PUBLIC_API_URL || 'http://localhost:5050';
	const clientId = url.searchParams.get('client_id') || '';

	// Forward the original Host header so the backend can resolve the tenant
	const originalHost = request.headers.get('host') || url.host;

	let branding: BrandingConfig | null = null;

	try {
		// Tenant resolution priority: explicit /t/{slug}/ prefix > client_id.
		// The branding endpoint accepts client_id; for slug-based access we
		// pass the slug-resolved tenant via the X-Tenant-Slug header.
		const params = clientId ? `?client_id=${encodeURIComponent(clientId)}` : '';
		const headers: Record<string, string> = {
			Accept: 'application/json',
			Host: originalHost,
		};
		if (locals.tenantSlug) {
			headers['X-Tenant-Slug'] = locals.tenantSlug;
		}
		const res = await fetch(`${apiUrl}/api/v1/branding${params}`, { headers });
		if (res.ok) {
			branding = await res.json();
		}
	} catch {
		// Branding fetch failed - use defaults
	}

	return {
		branding,
		tenantSlug: locals.tenantSlug,
		tenantBase: locals.tenantBase,
	};
};
