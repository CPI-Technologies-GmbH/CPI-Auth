import type { Handle } from '@sveltejs/kit';
import { redirect } from '@sveltejs/kit';
import { env } from '$env/dynamic/public';

const API_URL = env.PUBLIC_API_URL || 'http://localhost:5050';
const LOGIN_URL = env.PUBLIC_LOGIN_URL || 'http://localhost:5053/login';

export const handle: Handle = async ({ event, resolve }) => {
	const sessionToken = event.cookies.get('session_token') || null;
	event.locals.sessionToken = sessionToken;
	event.locals.user = null;

	// Public routes that don't require authentication
	const publicPaths = ['/health'];
	if (publicPaths.some((path) => event.url.pathname.startsWith(path))) {
		return resolve(event);
	}

	if (!sessionToken) {
		const returnTo = encodeURIComponent(event.url.pathname + event.url.search);
		redirect(302, `${LOGIN_URL}?return_to=${returnTo}`);
	}

	try {
		const response = await fetch(`${API_URL}/v1/users/me`, {
			headers: {
				Authorization: `Bearer ${sessionToken}`,
				'Content-Type': 'application/json'
			}
		});

		if (!response.ok) {
			// Clear invalid cookie and redirect to login
			event.cookies.delete('session_token', { path: '/' });
			const returnTo = encodeURIComponent(event.url.pathname + event.url.search);
			redirect(302, `${LOGIN_URL}?return_to=${returnTo}`);
		}

		event.locals.user = await response.json();
	} catch (error) {
		// If the API is unreachable, redirect to login
		if ((error as { status?: number })?.status === 302) throw error;
		event.cookies.delete('session_token', { path: '/' });
		const returnTo = encodeURIComponent(event.url.pathname + event.url.search);
		redirect(302, `${LOGIN_URL}?return_to=${returnTo}`);
	}

	return resolve(event);
};
