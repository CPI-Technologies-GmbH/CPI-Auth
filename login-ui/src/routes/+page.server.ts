import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ url }) => {
	// Preserve query parameters when redirecting to login
	const params = url.searchParams.toString();
	const target = params ? `/login?${params}` : '/login';
	redirect(302, target);
};
