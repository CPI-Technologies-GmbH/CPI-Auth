import type { LayoutServerLoad } from './$types';
import type { BrandingConfig } from '$lib/api/types';
import { env } from '$env/dynamic/public';

export const load: LayoutServerLoad = async ({ url, fetch }) => {
	const apiUrl = env.PUBLIC_API_URL || 'http://localhost:5050';
	const clientId = url.searchParams.get('client_id') || '';

	let branding: BrandingConfig | null = null;

	try {
		const params = clientId ? `?client_id=${encodeURIComponent(clientId)}` : '';
		const res = await fetch(`${apiUrl}/api/v1/branding${params}`, {
			headers: { Accept: 'application/json' }
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
