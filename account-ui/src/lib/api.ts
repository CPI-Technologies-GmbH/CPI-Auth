import { PUBLIC_API_URL } from '$env/static/public';
import type {
	User,
	UpdateUserRequest,
	ChangePasswordRequest,
	Session,
	MFAMethod,
	TOTPEnrollment,
	RecoveryCodes,
	Passkey,
	LinkedIdentity,
	Organization,
	ActivityEvent,
	Consent,
	DataExportRequest,
	Branding,
	ApiError
} from '$lib/types';

const API_BASE = PUBLIC_API_URL || 'http://localhost:5050';

class ApiClient {
	private baseUrl: string;
	private token: string | null = null;

	constructor(baseUrl: string) {
		this.baseUrl = baseUrl;
	}

	setToken(token: string | null) {
		this.token = token;
	}

	private async request<T>(
		method: string,
		path: string,
		body?: unknown,
		customHeaders?: Record<string, string>
	): Promise<T> {
		const headers: Record<string, string> = {
			'Content-Type': 'application/json',
			...customHeaders
		};

		if (this.token) {
			headers['Authorization'] = `Bearer ${this.token}`;
		}

		const response = await fetch(`${this.baseUrl}${path}`, {
			method,
			headers,
			body: body ? JSON.stringify(body) : undefined,
			credentials: 'include'
		});

		if (!response.ok) {
			let errorBody: ApiError;
			try {
				errorBody = await response.json();
			} catch {
				errorBody = {
					error: 'unknown_error',
					message: `Request failed with status ${response.status}`,
					status: response.status
				};
			}
			throw errorBody;
		}

		if (response.status === 204) {
			return undefined as T;
		}

		return response.json();
	}

	// User Profile
	async getMe(): Promise<User> {
		return this.request<User>('GET', '/v1/users/me');
	}

	async updateMe(data: UpdateUserRequest): Promise<User> {
		return this.request<User>('PATCH', '/v1/users/me', data);
	}

	async changePassword(data: ChangePasswordRequest): Promise<void> {
		return this.request<void>('POST', '/v1/users/me/change-password', data);
	}

	async deleteAccount(password?: string): Promise<void> {
		return this.request<void>('DELETE', '/v1/users/me', password ? { password } : undefined);
	}

	// Sessions
	async getSessions(): Promise<Session[]> {
		return this.request<Session[]>('GET', '/v1/users/me/sessions');
	}

	async revokeSession(sessionId: string): Promise<void> {
		return this.request<void>('DELETE', `/v1/users/me/sessions/${sessionId}`);
	}

	async revokeAllOtherSessions(): Promise<void> {
		return this.request<void>('DELETE', '/v1/users/me/sessions');
	}

	// MFA
	async getMFAMethods(): Promise<MFAMethod[]> {
		return this.request<MFAMethod[]>('GET', '/v1/users/me/mfa');
	}

	async enrollTOTP(): Promise<TOTPEnrollment> {
		return this.request<TOTPEnrollment>('POST', '/v1/users/me/mfa/totp/enroll');
	}

	async verifyTOTP(id: string, code: string): Promise<void> {
		return this.request<void>('POST', '/v1/users/me/mfa/totp/verify', { id, code });
	}

	async enrollEmailOTP(): Promise<{ id: string }> {
		return this.request<{ id: string }>('POST', '/v1/users/me/mfa/email/enroll');
	}

	async enrollSMSOTP(phoneNumber: string): Promise<{ id: string }> {
		return this.request<{ id: string }>('POST', '/v1/users/me/mfa/sms/enroll', {
			phone_number: phoneNumber
		});
	}

	async removeMFA(id: string): Promise<void> {
		return this.request<void>('DELETE', `/v1/users/me/mfa/${id}`);
	}

	async getRecoveryCodes(): Promise<RecoveryCodes> {
		return this.request<RecoveryCodes>('GET', '/v1/users/me/mfa/recovery-codes');
	}

	async regenerateRecoveryCodes(): Promise<RecoveryCodes> {
		return this.request<RecoveryCodes>('POST', '/v1/users/me/mfa/recovery-codes/regenerate');
	}

	// Passkeys
	async getPasskeys(): Promise<Passkey[]> {
		return this.request<Passkey[]>('GET', '/v1/users/me/passkeys');
	}

	async beginPasskeyRegistration(): Promise<{ options: unknown }> {
		return this.request<{ options: unknown }>('POST', '/v1/users/me/passkeys/register/begin');
	}

	async finishPasskeyRegistration(credential: unknown): Promise<Passkey> {
		return this.request<Passkey>('POST', '/v1/users/me/passkeys/register/finish', credential);
	}

	async renamePasskey(id: string, name: string): Promise<Passkey> {
		return this.request<Passkey>('PATCH', `/v1/users/me/passkeys/${id}`, { name });
	}

	async deletePasskey(id: string): Promise<void> {
		return this.request<void>('DELETE', `/v1/users/me/passkeys/${id}`);
	}

	// Linked Identities
	async getIdentities(): Promise<LinkedIdentity[]> {
		return this.request<LinkedIdentity[]>('GET', '/v1/users/me/identities');
	}

	async unlinkIdentity(id: string): Promise<void> {
		return this.request<void>('DELETE', `/v1/users/me/identities/${id}`);
	}

	// Organizations
	async getOrganizations(): Promise<Organization[]> {
		return this.request<Organization[]>('GET', '/v1/users/me/organizations');
	}

	// Data Export
	async requestDataExport(): Promise<DataExportRequest> {
		return this.request<DataExportRequest>('POST', '/v1/users/me/export');
	}

	// Activity
	async getActivity(params?: {
		type?: string;
		from?: string;
		to?: string;
		limit?: number;
		offset?: number;
	}): Promise<ActivityEvent[]> {
		const searchParams = new URLSearchParams();
		if (params?.type) searchParams.set('type', params.type);
		if (params?.from) searchParams.set('from', params.from);
		if (params?.to) searchParams.set('to', params.to);
		if (params?.limit) searchParams.set('limit', params.limit.toString());
		if (params?.offset) searchParams.set('offset', params.offset.toString());

		const query = searchParams.toString();
		return this.request<ActivityEvent[]>(
			'GET',
			`/v1/users/me/activity${query ? `?${query}` : ''}`
		);
	}

	// Consents
	async getConsents(): Promise<Consent[]> {
		return this.request<Consent[]>('GET', '/v1/users/me/consents');
	}

	async revokeConsent(id: string): Promise<void> {
		return this.request<void>('DELETE', `/v1/users/me/consents/${id}`);
	}

	// Branding
	async getBranding(): Promise<Branding> {
		return this.request<Branding>('GET', '/v1/branding');
	}
}

export const api = new ApiClient(API_BASE);

// Server-side API client factory (uses token directly)
export function createServerApi(token: string): ApiClient {
	const client = new ApiClient(API_BASE);
	client.setToken(token);
	return client;
}

export { API_BASE };
