import { env } from '$env/dynamic/public';
import type {
	AuthResponse,
	BrandingConfig,
	ConsentDecisionRequest,
	ConsentInfo,
	ForgotPasswordRequest,
	LoginRequest,
	MfaChallengeRequest,
	MfaChallengeResponse,
	MfaEnrollRequest,
	MfaEnrollResponse,
	MfaEnrollVerifyRequest,
	MfaVerifyRequest,
	PasswordlessStartRequest,
	PasswordlessStartResponse,
	PasswordlessVerifyRequest,
	RegisterRequest,
	ResetPasswordRequest,
	VerifyEmailRequest,
	WebAuthnLoginBeginRequest,
	WebAuthnLoginFinishRequest,
	WebAuthnRegisterBeginRequest,
	WebAuthnRegisterFinishRequest,
	WebAuthnBeginResponse,
	ApiError
} from './types';

class CPIAuthApiClient {
	private baseUrl: string;

	constructor() {
		this.baseUrl = env.PUBLIC_API_URL || '';
	}

	private async request<T>(
		method: string,
		path: string,
		body?: unknown,
		options?: { headers?: Record<string, string> }
	): Promise<T> {
		const url = `${this.baseUrl}${path}`;
		const headers: Record<string, string> = {
			'Content-Type': 'application/json',
			Accept: 'application/json',
			...options?.headers
		};

		const res = await fetch(url, {
			method,
			headers,
			body: body ? JSON.stringify(body) : undefined,
			credentials: 'include'
		});

		if (!res.ok) {
			let errorBody: ApiError;
			try {
				errorBody = await res.json();
			} catch {
				errorBody = {
					error: 'request_failed',
					error_description: `Request failed with status ${res.status}`,
					status_code: res.status
				};
			}
			throw new ApiClientError(
				errorBody.error_description || errorBody.error || 'Unknown error',
				errorBody.error || 'unknown_error',
				res.status
			);
		}

		const contentType = res.headers.get('content-type');
		if (contentType?.includes('application/json')) {
			return res.json();
		}

		return {} as T;
	}

	// ── Branding ──

	async getBranding(clientId?: string): Promise<BrandingConfig> {
		const params = clientId ? `?client_id=${encodeURIComponent(clientId)}` : '';
		return this.request<BrandingConfig>('GET', `/api/v1/branding${params}`);
	}

	// ── Authentication ──

	async login(data: LoginRequest): Promise<AuthResponse> {
		return this.request<AuthResponse>('POST', '/api/v1/auth/login', data);
	}

	async register(data: RegisterRequest): Promise<AuthResponse> {
		return this.request<AuthResponse>('POST', '/api/v1/auth/register', data);
	}

	async forgotPassword(data: ForgotPasswordRequest): Promise<{ message: string }> {
		return this.request<{ message: string }>('POST', '/api/v1/auth/forgot-password', data);
	}

	async resetPassword(data: ResetPasswordRequest): Promise<{ message: string }> {
		return this.request<{ message: string }>('POST', '/api/v1/auth/reset-password', data);
	}

	async verifyEmail(data: VerifyEmailRequest): Promise<{ message: string }> {
		return this.request<{ message: string }>('POST', '/api/v1/auth/verify-email', data);
	}

	// ── MFA ──

	async mfaChallenge(data: MfaChallengeRequest): Promise<MfaChallengeResponse> {
		return this.request<MfaChallengeResponse>('POST', '/api/v1/auth/mfa/challenge', data);
	}

	async mfaVerify(data: MfaVerifyRequest): Promise<AuthResponse> {
		return this.request<AuthResponse>('POST', '/api/v1/auth/mfa/verify', data);
	}

	async mfaEnroll(data: MfaEnrollRequest): Promise<MfaEnrollResponse> {
		return this.request<MfaEnrollResponse>('POST', '/api/v1/auth/mfa/enroll', data);
	}

	async mfaEnrollVerify(data: MfaEnrollVerifyRequest): Promise<AuthResponse> {
		return this.request<AuthResponse>('POST', '/api/v1/auth/mfa/verify', data);
	}

	// ── Passwordless ──

	async passwordlessStart(data: PasswordlessStartRequest): Promise<PasswordlessStartResponse> {
		return this.request<PasswordlessStartResponse>(
			'POST',
			'/api/v1/auth/passwordless/start',
			data
		);
	}

	async passwordlessVerify(data: PasswordlessVerifyRequest): Promise<AuthResponse> {
		return this.request<AuthResponse>('POST', '/api/v1/auth/passwordless/verify', data);
	}

	// ── WebAuthn ──

	async webAuthnRegisterBegin(data: WebAuthnRegisterBeginRequest): Promise<WebAuthnBeginResponse> {
		return this.request<WebAuthnBeginResponse>(
			'POST',
			'/api/v1/auth/webauthn/register/begin',
			data
		);
	}

	async webAuthnRegisterFinish(data: WebAuthnRegisterFinishRequest): Promise<AuthResponse> {
		return this.request<AuthResponse>('POST', '/api/v1/auth/webauthn/register/finish', data);
	}

	async webAuthnLoginBegin(data: WebAuthnLoginBeginRequest): Promise<WebAuthnBeginResponse> {
		return this.request<WebAuthnBeginResponse>(
			'POST',
			'/api/v1/auth/webauthn/login/begin',
			data
		);
	}

	async webAuthnLoginFinish(data: WebAuthnLoginFinishRequest): Promise<AuthResponse> {
		return this.request<AuthResponse>('POST', '/api/v1/auth/webauthn/login/finish', data);
	}

	// ── OAuth ──

	async oauthAuthorize(params: Record<string, string>): Promise<AuthResponse> {
		const searchParams = new URLSearchParams(params);
		return this.request<AuthResponse>('POST', `/oauth/authorize?${searchParams.toString()}`);
	}

	async oauthToken(data: Record<string, string>): Promise<AuthResponse> {
		return this.request<AuthResponse>('POST', '/oauth/token', data);
	}

	// ── Consent ──

	async getConsentInfo(challenge: string): Promise<ConsentInfo> {
		return this.request<ConsentInfo>(
			'GET',
			`/api/v1/auth/consent?consent_challenge=${encodeURIComponent(challenge)}`
		);
	}

	async submitConsent(data: ConsentDecisionRequest): Promise<AuthResponse> {
		return this.request<AuthResponse>('POST', '/api/v1/auth/consent', data);
	}
}

export class ApiClientError extends Error {
	code: string;
	statusCode: number;

	constructor(message: string, code: string, statusCode: number) {
		super(message);
		this.name = 'ApiClientError';
		this.code = code;
		this.statusCode = statusCode;
	}
}

export const api = new CPIAuthApiClient();
