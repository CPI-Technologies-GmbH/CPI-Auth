import { writable, derived, get } from 'svelte/store';

export type Locale = 'en' | 'de' | 'fr' | 'es';

const translations: Record<Locale, Record<string, string>> = {
	en: {
		// Profile
		'profile.title': 'Account Settings',
		'profile.description': 'Manage your account settings and preferences',
		'profile.personal': 'Personal Information',
		'profile.name': 'Full Name',
		'profile.email': 'Email Address',
		'profile.phone': 'Phone Number',
		'profile.avatar': 'Avatar URL',
		'profile.save': 'Save Changes',
		'profile.saved': 'Profile updated successfully',

		// Security
		'security.title': 'Security',
		'security.description': 'Manage your password and security settings',
		'security.change_password': 'Change Password',
		'security.current_password': 'Current Password',
		'security.new_password': 'New Password',
		'security.confirm_password': 'Confirm New Password',
		'security.update_password': 'Update Password',
		'security.mfa': 'Two-Factor Authentication',
		'security.mfa_description': 'Add an extra layer of security to your account',
		'security.mfa_enable': 'Enable MFA',
		'security.mfa_disable': 'Disable MFA',
		'security.mfa_enabled': 'MFA is enabled',
		'security.mfa_disabled': 'MFA is not enabled',

		// Sessions
		'sessions.title': 'Active Sessions',
		'sessions.description': 'Manage your active sessions across devices',
		'sessions.current': 'Current Session',
		'sessions.revoke': 'Revoke',
		'sessions.revoke_all': 'Revoke All Sessions',

		// Connected Accounts
		'accounts.title': 'Connected Accounts',
		'accounts.description': 'Manage linked social and external accounts',
		'accounts.connect': 'Connect',
		'accounts.disconnect': 'Disconnect',

		// Navigation
		'nav.profile': 'Profile',
		'nav.security': 'Security',
		'nav.sessions': 'Sessions',
		'nav.accounts': 'Connected Accounts',
		'nav.logout': 'Log Out',

		// Common
		'common.save': 'Save',
		'common.cancel': 'Cancel',
		'common.loading': 'Loading...',
		'common.error': 'An error occurred',
		'common.success': 'Success',
	},
	de: {
		'profile.title': 'Kontoeinstellungen',
		'profile.description': 'Verwalten Sie Ihre Kontoeinstellungen und Praeferenzen',
		'profile.personal': 'Persoenliche Informationen',
		'profile.name': 'Vollstaendiger Name',
		'profile.email': 'E-Mail-Adresse',
		'profile.phone': 'Telefonnummer',
		'profile.save': 'Aenderungen speichern',
		'profile.saved': 'Profil erfolgreich aktualisiert',
		'security.title': 'Sicherheit',
		'security.description': 'Passwort und Sicherheitseinstellungen verwalten',
		'security.change_password': 'Passwort aendern',
		'security.current_password': 'Aktuelles Passwort',
		'security.new_password': 'Neues Passwort',
		'security.confirm_password': 'Neues Passwort bestaetigen',
		'security.update_password': 'Passwort aktualisieren',
		'security.mfa': 'Zwei-Faktor-Authentifizierung',
		'security.mfa_enable': 'MFA aktivieren',
		'security.mfa_disable': 'MFA deaktivieren',
		'sessions.title': 'Aktive Sitzungen',
		'sessions.description': 'Aktive Sitzungen auf Ihren Geraeten verwalten',
		'sessions.revoke': 'Widerrufen',
		'sessions.revoke_all': 'Alle Sitzungen widerrufen',
		'accounts.title': 'Verbundene Konten',
		'accounts.connect': 'Verbinden',
		'accounts.disconnect': 'Trennen',
		'nav.profile': 'Profil',
		'nav.security': 'Sicherheit',
		'nav.sessions': 'Sitzungen',
		'nav.accounts': 'Verbundene Konten',
		'nav.logout': 'Abmelden',
		'common.save': 'Speichern',
		'common.cancel': 'Abbrechen',
		'common.error': 'Ein Fehler ist aufgetreten',
	},
	fr: {
		'profile.title': 'Parametres du compte',
		'profile.personal': 'Informations personnelles',
		'profile.name': 'Nom complet',
		'profile.email': 'Adresse e-mail',
		'profile.save': 'Enregistrer les modifications',
		'security.title': 'Securite',
		'security.change_password': 'Changer le mot de passe',
		'security.mfa': 'Authentification a deux facteurs',
		'sessions.title': 'Sessions actives',
		'sessions.revoke': 'Revoquer',
		'accounts.title': 'Comptes connectes',
		'nav.profile': 'Profil',
		'nav.security': 'Securite',
		'nav.sessions': 'Sessions',
		'nav.logout': 'Deconnexion',
		'common.save': 'Enregistrer',
		'common.cancel': 'Annuler',
	},
	es: {
		'profile.title': 'Configuracion de cuenta',
		'profile.personal': 'Informacion personal',
		'profile.name': 'Nombre completo',
		'profile.email': 'Correo electronico',
		'profile.save': 'Guardar cambios',
		'security.title': 'Seguridad',
		'security.change_password': 'Cambiar contrasena',
		'security.mfa': 'Autenticacion de dos factores',
		'sessions.title': 'Sesiones activas',
		'sessions.revoke': 'Revocar',
		'accounts.title': 'Cuentas conectadas',
		'nav.profile': 'Perfil',
		'nav.security': 'Seguridad',
		'nav.sessions': 'Sesiones',
		'nav.logout': 'Cerrar sesion',
		'common.save': 'Guardar',
		'common.cancel': 'Cancelar',
	},
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
	if (typeof window !== 'undefined') {
		localStorage.setItem('locale', newLocale);
	}
	locale.set(newLocale);
}

export function initLocale() {
	if (typeof window !== 'undefined') {
		const stored = localStorage.getItem('locale') as Locale | null;
		if (stored && translations[stored]) {
			locale.set(stored);
		}
	}
}

export function getTranslation(key: string, params?: Record<string, string>): string {
	const $t = get(t);
	return $t(key, params);
}

export const localeNames: Record<Locale, string> = {
	en: 'English',
	de: 'Deutsch',
	fr: 'Francais',
	es: 'Espanol',
};
