import { env } from '$env/dynamic/public';

export const API_URL = env.PUBLIC_API_URL || 'http://localhost:5050';
export const LOGIN_URL = env.PUBLIC_LOGIN_URL || 'http://localhost:5053/login';
