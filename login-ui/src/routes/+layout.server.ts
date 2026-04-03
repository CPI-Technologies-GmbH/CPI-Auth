import type { LayoutServerLoad } from './$types';
import type { BrandingConfig } from '$lib/api/types';
import { env as publicEnv } from '$env/dynamic/public';
import { env as privateEnv } from '$env/dynamic/private';

export const load: LayoutServerLoad = async ({ url, fetch, request }) => {
	// Server-side: use internal URL (INTERNAL_API_URL) or fallback to public
	const apiUrl = privateEnv.INTERNAL_API_URL || publicEnv.PUBLIC_API_URL || 'http://localhost:5050';
	const clientId = url.searchParams.get('client_id') || '';

	// Forward the original Host header so the backend can resolve the tenant
	const originalHost = request.headers.get('host') || url.host;

	let branding: BrandingConfig | null = null;

	try {
		const params = clientId ? `?client_id=${encodeURIComponent(clientId)}` : '';
		const res = await fetch(`${apiUrl}/api/v1/branding${params}`, {
			headers: {
				Accept: 'application/json',
				Host: originalHost,
			}
		});
		if (res.ok) {
			branding = await res.json();
		}
	} catch {
		// Branding fetch failed - use defaults
	}

	return {
		branding
	};
};
