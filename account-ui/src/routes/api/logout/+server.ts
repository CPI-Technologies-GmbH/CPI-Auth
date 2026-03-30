import { redirect } from '@sveltejs/kit';
import { env } from '$env/dynamic/public';
import type { RequestHandler } from './$types';

const LOGIN_URL = env.PUBLIC_LOGIN_URL || 'http://localhost:5053/login';

export const GET: RequestHandler = async ({ cookies }) => {
	cookies.delete('session_token', { path: '/' });
	redirect(302, LOGIN_URL);
};
