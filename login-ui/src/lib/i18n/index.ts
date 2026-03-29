import { writable, derived, get } from 'svelte/store';

export type Locale = 'en' | 'de' | 'fr' | 'es';

const translations: Record<Locale, Record<string, string>> = {
	en: {
		// Login
		'login.title': 'Sign in to your account',
		'login.email': 'Email address',
		'login.email.placeholder': 'you@example.com',
		'login.password': 'Password',
		'login.password.placeholder': 'Enter your password',
		'login.remember': 'Remember me',
		'login.forgot': 'Forgot password?',
		'login.submit': 'Sign in',
		'login.no_account': "Don't have an account?",
		'login.signup_link': 'Sign up',
		'login.or': 'or continue with',
		'login.passkey': 'Sign in with Passkey',
		'login.magic_link': 'Sign in with Magic Link',
		'login.error.invalid': 'Invalid email or password',
		'login.error.locked': 'Your account has been locked. Please try again later.',
		'login.error.unverified': 'Please verify your email address first.',

		// Register
		'register.title': 'Create your account',
		'register.name': 'Full name',
		'register.name.placeholder': 'John Doe',
		'register.email': 'Email address',
		'register.email.placeholder': 'you@example.com',
		'register.password': 'Password',
		'register.password.placeholder': 'Create a strong password',
		'register.confirm_password': 'Confirm password',
		'register.confirm_password.placeholder': 'Confirm your password',
		'register.terms': 'I agree to the',
		'register.terms_link': 'Terms of Service',
		'register.and': 'and',
		'register.privacy_link': 'Privacy Policy',
		'register.submit': 'Create account',
		'register.has_account': 'Already have an account?',
		'register.signin_link': 'Sign in',
		'register.or': 'or sign up with',

		// Forgot Password
		'forgot.title': 'Reset your password',
		'forgot.description': "Enter your email address and we'll send you instructions to reset your password.",
		'forgot.email': 'Email address',
		'forgot.email.placeholder': 'you@example.com',
		'forgot.submit': 'Send reset instructions',
		'forgot.back': 'Back to sign in',
		'forgot.success': "We've sent password reset instructions to your email address. Please check your inbox.",

		// Reset Password
		'reset.title': 'Set new password',
		'reset.description': 'Enter your new password below.',
		'reset.password': 'New password',
		'reset.password.placeholder': 'Enter your new password',
		'reset.confirm': 'Confirm new password',
		'reset.confirm.placeholder': 'Confirm your new password',
		'reset.submit': 'Reset password',
		'reset.success': 'Your password has been reset successfully.',
		'reset.login_link': 'Sign in with your new password',
		'reset.error.token': 'Invalid or expired reset token. Please request a new password reset.',

		// Email Verification
		'verify.title': 'Email Verification',
		'verify.verifying': 'Verifying your email...',
		'verify.success': 'Your email has been verified successfully!',
		'verify.error': 'Email verification failed. The link may have expired.',
		'verify.resend': 'Resend verification email',
		'verify.login': 'Continue to sign in',

		// MFA
		'mfa.title': 'Two-factor authentication',
		'mfa.description': 'Enter the 6-digit code from your authenticator app.',
		'mfa.code': 'Verification code',
		'mfa.code.placeholder': '000000',
		'mfa.submit': 'Verify',
		'mfa.recovery': 'Use a recovery code',
		'mfa.different_method': 'Use a different method',
		'mfa.sms': 'Send code via SMS',
		'mfa.email': 'Send code via email',
		'mfa.recovery.title': 'Enter recovery code',
		'mfa.recovery.description': 'Enter one of your recovery codes.',
		'mfa.recovery.placeholder': 'xxxx-xxxx-xxxx',

		// MFA Enrollment
		'mfa_enroll.title': 'Set up two-factor authentication',
		'mfa_enroll.scan': 'Scan this QR code with your authenticator app',
		'mfa_enroll.manual': "Can't scan? Enter this code manually:",
		'mfa_enroll.verify': 'Enter the 6-digit code from your app to verify setup',
		'mfa_enroll.submit': 'Verify and enable',
		'mfa_enroll.recovery_title': 'Save your recovery codes',
		'mfa_enroll.recovery_description': 'Store these recovery codes in a safe place. Each code can only be used once.',
		'mfa_enroll.recovery_copied': 'Recovery codes copied to clipboard',
		'mfa_enroll.copy': 'Copy codes',
		'mfa_enroll.done': 'Done',

		// Consent
		'consent.title': 'Authorize',
		'consent.description': 'wants to access your account',
		'consent.scopes_title': 'This application is requesting access to:',
		'consent.remember': 'Remember this decision',
		'consent.allow': 'Allow',
		'consent.deny': 'Deny',

		// Passwordless
		'passwordless.title': 'Check your email',
		'passwordless.email_sent': "We've sent a verification code to",
		'passwordless.link_sent': "We've sent a sign-in link to",
		'passwordless.code': 'Verification code',
		'passwordless.code.placeholder': 'Enter code',
		'passwordless.submit': 'Verify',
		'passwordless.resend': 'Resend code',

		// Error
		'error.title': 'Something went wrong',
		'error.back': 'Back to sign in',
		'error.generic': 'An unexpected error occurred. Please try again.',

		// Validation
		'validation.required': 'This field is required',
		'validation.email': 'Please enter a valid email address',
		'validation.password.min': 'Password must be at least 8 characters',
		'validation.password.match': 'Passwords do not match',
		'validation.terms': 'You must agree to the terms of service',
		'validation.code': 'Please enter a valid 6-digit code',

		// Password Strength
		'password.strength.weak': 'Weak',
		'password.strength.fair': 'Fair',
		'password.strength.good': 'Good',
		'password.strength.strong': 'Strong',

		// General
		'general.loading': 'Loading...',
		'general.error': 'Error',
		'general.success': 'Success',
		'general.or': 'or'
	},
	de: {
		'login.title': 'Anmelden',
		'login.email': 'E-Mail-Adresse',
		'login.email.placeholder': 'du@beispiel.de',
		'login.password': 'Passwort',
		'login.password.placeholder': 'Passwort eingeben',
		'login.remember': 'Angemeldet bleiben',
		'login.forgot': 'Passwort vergessen?',
		'login.submit': 'Anmelden',
		'login.no_account': 'Kein Konto?',
		'login.signup_link': 'Registrieren',
		'login.or': 'oder weiter mit',
		'login.passkey': 'Mit Passkey anmelden',
		'login.magic_link': 'Mit Magic Link anmelden',
		'login.error.invalid': 'Ungueltige E-Mail oder Passwort',
		'login.error.locked': 'Ihr Konto wurde gesperrt. Bitte versuchen Sie es spaeter erneut.',
		'login.error.unverified': 'Bitte verifizieren Sie zuerst Ihre E-Mail-Adresse.',
		'register.title': 'Konto erstellen',
		'register.submit': 'Konto erstellen',
		'forgot.title': 'Passwort zuruecksetzen',
		'forgot.submit': 'Anweisungen senden',
		'error.title': 'Etwas ist schiefgelaufen',
		'error.back': 'Zurueck zur Anmeldung'
	},
	fr: {
		'login.title': 'Connectez-vous',
		'login.email': 'Adresse e-mail',
		'login.password': 'Mot de passe',
		'login.submit': 'Se connecter',
		'login.forgot': 'Mot de passe oublie?',
		'register.title': 'Creer un compte',
		'register.submit': 'Creer un compte',
		'error.title': "Quelque chose s'est mal passe",
		'error.back': 'Retour a la connexion'
	},
	es: {
		'login.title': 'Iniciar sesion',
		'login.email': 'Correo electronico',
		'login.password': 'Contrasena',
		'login.submit': 'Iniciar sesion',
		'login.forgot': 'Olvidaste tu contrasena?',
		'register.title': 'Crear cuenta',
		'register.submit': 'Crear cuenta',
		'error.title': 'Algo salio mal',
		'error.back': 'Volver al inicio de sesion'
	}
};

export const locale = writable<Locale>('en');

export const t = derived(locale, ($locale) => {
	return (key: string, params?: Record<string, string>): string => {
		let text = translations[$locale]?.[key] || translations.en[key] || key;
		if (params) {
			for (const [k, v] of Object.entries(params)) {
				text = text.replace(`{${k}}`, v);
			}
		}
		return text;
	};
});

export function setLocale(newLocale: Locale) {
	locale.set(newLocale);
}

export function getTranslation(key: string, params?: Record<string, string>): string {
	const $t = get(t);
	return $t(key, params);
}
