export interface User {
	id: string;
	email: string;
	email_verified: boolean;
	phone?: string;
	phone_verified?: boolean;
	name?: string;
	given_name?: string;
	family_name?: string;
	nickname?: string;
	avatar_url?: string;
	locale?: string;
	timezone?: string;
	metadata?: Record<string, unknown>;
	user_metadata?: Record<string, unknown>;
	created_at: string;
	updated_at: string;
}

export interface UpdateUserRequest {
	name?: string;
	given_name?: string;
	family_name?: string;
	nickname?: string;
	avatar_url?: string;
	locale?: string;
	timezone?: string;
	user_metadata?: Record<string, unknown>;
}

export interface ChangePasswordRequest {
	current_password: string;
	new_password: string;
}

export interface Session {
	id: string;
	user_id: string;
	ip_address: string;
	user_agent: string;
	device_type?: string;
	browser?: string;
	os?: string;
	location?: {
		city?: string;
		region?: string;
		country?: string;
	};
	last_active_at: string;
	created_at: string;
	is_current: boolean;
}

export interface MFAMethod {
	id: string;
	type: 'totp' | 'email' | 'sms';
	name?: string;
	verified: boolean;
	created_at: string;
	phone_number?: string;
	email?: string;
}

export interface TOTPEnrollment {
	id: string;
	secret: string;
	uri: string;
	qr_code?: string;
}

export interface RecoveryCodes {
	codes: string[];
	generated_at: string;
}

export interface Passkey {
	id: string;
	name: string;
	credential_id: string;
	aaguid?: string;
	created_at: string;
	last_used_at?: string;
}

export interface PasskeyRegistrationBegin {
	options: PublicKeyCredentialCreationOptions;
}

export interface LinkedIdentity {
	id: string;
	provider: string;
	provider_user_id: string;
	email?: string;
	name?: string;
	username?: string;
	avatar_url?: string;
	linked_at: string;
}

export interface Organization {
	id: string;
	name: string;
	slug: string;
	logo_url?: string;
	role: string;
	member_since: string;
}

export interface ActivityEvent {
	id: string;
	type: string;
	description: string;
	ip_address?: string;
	user_agent?: string;
	device_type?: string;
	browser?: string;
	os?: string;
	location?: {
		city?: string;
		region?: string;
		country?: string;
	};
	metadata?: Record<string, unknown>;
	created_at: string;
}

export interface Consent {
	id: string;
	client_id: string;
	client_name: string;
	client_logo_url?: string;
	scopes: string[];
	granted_at: string;
	updated_at?: string;
}

export interface DataExportRequest {
	id: string;
	status: 'pending' | 'processing' | 'ready' | 'expired';
	download_url?: string;
	requested_at: string;
	expires_at?: string;
}

export interface Branding {
	logo_url?: string;
	logo_dark_url?: string;
	favicon_url?: string;
	primary_color?: string;
	background_color?: string;
	font_family?: string;
	company_name?: string;
	support_url?: string;
	privacy_url?: string;
	terms_url?: string;
}

export interface ApiError {
	error: string;
	message: string;
	status: number;
}

export type ActivityType =
	| 'login'
	| 'logout'
	| 'password_change'
	| 'mfa_enable'
	| 'mfa_disable'
	| 'profile_update'
	| 'session_revoke'
	| 'email_change'
	| 'phone_change'
	| 'passkey_register'
	| 'passkey_delete'
	| 'identity_link'
	| 'identity_unlink'
	| 'consent_grant'
	| 'consent_revoke'
	| 'data_export'
	| 'account_delete';
