// ── Branding / Theming ──

export interface BrandingConfig {
	primary_color: string;
	primary_hover_color: string;
	primary_light_color: string;
	secondary_color: string;
	background_color: string;
	surface_color: string;
	text_color: string;
	text_muted_color: string;
	text_inverted_color: string;
	border_color: string;
	error_color: string;
	success_color: string;
	warning_color: string;
	info_color: string;
	logo_url: string;
	logo_dark_url: string;
	font_family: string;
	border_radius: string;
	layout_mode: 'centered' | 'split-screen' | 'sidebar';
	custom_css: string;
	dark_mode: boolean;
	app_name: string;
	social_providers: SocialProvider[];
	custom_fields: CustomField[];
	passkeys_enabled: boolean;
	magic_link_enabled: boolean;
	mfa_enabled: boolean;
}

export interface SocialProvider {
	id: string;
	name: string;
	icon?: string;
	color?: string;
}

export interface CustomField {
	name: string;
	label: string;
	type: 'text' | 'email' | 'tel' | 'select' | 'checkbox';
	required: boolean;
	placeholder?: string;
	options?: string[];
}

// ── Auth Requests ──

export interface LoginRequest {
	email: string;
	password: string;
	remember_me?: boolean;
	client_id?: string;
	redirect_uri?: string;
	scope?: string;
	state?: string;
	code_challenge?: string;
	code_challenge_method?: string;
	response_type?: string;
}

export interface RegisterRequest {
	email: string;
	password: string;
	name: string;
	custom_fields?: Record<string, string>;
	client_id?: string;
	redirect_uri?: string;
}

export interface ForgotPasswordRequest {
	email: string;
	client_id?: string;
}

export interface ResetPasswordRequest {
	token: string;
	password: string;
}

export interface VerifyEmailRequest {
	token: string;
}

export interface MfaChallengeRequest {
	mfa_token: string;
	method?: 'totp' | 'sms' | 'email';
}

export interface MfaVerifyRequest {
	mfa_token: string;
	code: string;
	method: 'totp' | 'sms' | 'email' | 'recovery';
}

export interface MfaEnrollRequest {
	mfa_token: string;
	method: 'totp';
}

export interface MfaEnrollVerifyRequest {
	mfa_token: string;
	code: string;
}

export interface PasswordlessStartRequest {
	email: string;
	method: 'email_link' | 'email_otp' | 'sms_otp';
	client_id?: string;
	redirect_uri?: string;
	scope?: string;
	state?: string;
}

export interface PasswordlessVerifyRequest {
	token: string;
	code?: string;
}

export interface WebAuthnRegisterBeginRequest {
	mfa_token?: string;
}

export interface WebAuthnRegisterFinishRequest {
	credential: unknown;
}

export interface WebAuthnLoginBeginRequest {
	client_id?: string;
	redirect_uri?: string;
}

export interface WebAuthnLoginFinishRequest {
	credential: unknown;
	client_id?: string;
	redirect_uri?: string;
}

export interface ConsentRequest {
	consent_challenge: string;
}

export interface ConsentDecisionRequest {
	consent_challenge: string;
	grant: boolean;
	remember: boolean;
	scopes?: string[];
}

export interface OAuthAuthorizeRequest {
	client_id: string;
	redirect_uri: string;
	response_type: string;
	scope: string;
	state: string;
	code_challenge?: string;
	code_challenge_method?: string;
}

// ── Auth Responses ──

export interface AuthResponse {
	redirect_url?: string;
	access_token?: string;
	refresh_token?: string;
	token_type?: string;
	expires_in?: number;
	mfa_token?: string;
	mfa_required?: boolean;
	mfa_methods?: string[];
}

export interface MfaEnrollResponse {
	secret: string;
	qr_code: string;
	recovery_codes: string[];
	otpauth_url: string;
}

export interface MfaChallengeResponse {
	challenge_id: string;
	method: string;
	expires_at: string;
}

export interface ConsentInfo {
	client_name: string;
	client_logo: string;
	client_uri: string;
	requested_scopes: ScopeInfo[];
	subject: string;
}

export interface ScopeInfo {
	name: string;
	description: string;
}

export interface WebAuthnBeginResponse {
	publicKey: unknown;
}

export interface ApiError {
	error: string;
	error_description?: string;
	status_code?: number;
}

export interface PasswordlessStartResponse {
	message: string;
	method: string;
	expires_at: string;
}
