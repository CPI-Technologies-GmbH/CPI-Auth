import { create } from 'zustand'

export type Locale = 'en' | 'de' | 'fr' | 'es'

const translations: Record<Locale, Record<string, string>> = {
  en: {
    // Navigation
    'nav.dashboard': 'Dashboard',
    'nav.users': 'Users',
    'nav.applications': 'Applications',
    'nav.tenants': 'Tenants',
    'nav.organizations': 'Organizations',
    'nav.roles': 'Roles & Permissions',
    'nav.branding': 'Branding',
    'nav.webhooks': 'Webhooks',
    'nav.actions': 'Actions',
    'nav.email_templates': 'Email Templates',
    'nav.page_templates': 'Page Templates',
    'nav.api_keys': 'API Keys',
    'nav.custom_fields': 'Custom Fields',
    'nav.logs': 'Logs',
    'nav.settings': 'Settings',

    // Common actions
    'action.create': 'Create',
    'action.edit': 'Edit',
    'action.delete': 'Delete',
    'action.save': 'Save',
    'action.cancel': 'Cancel',
    'action.confirm': 'Confirm',
    'action.search': 'Search...',
    'action.export': 'Export',
    'action.import': 'Import',
    'action.refresh': 'Refresh',
    'action.close': 'Close',
    'action.back': 'Back',

    // Dashboard
    'dashboard.title': 'Dashboard',
    'dashboard.description': 'Overview of your authentication platform',
    'dashboard.active_users': 'Active Users',
    'dashboard.login_success': 'Login Success Rate',
    'dashboard.mfa_adoption': 'MFA Adoption',
    'dashboard.total_sessions': 'Total Sessions',
    'dashboard.error_rate': 'Error Rate',
    'dashboard.recent_events': 'Recent Events',
    'dashboard.login_activity': 'Login Activity',

    // Users
    'users.title': 'Users',
    'users.description': 'Manage user accounts and identities',
    'users.create': 'Create User',
    'users.search': 'Search users...',
    'users.block': 'Block',
    'users.unblock': 'Unblock',
    'users.impersonate': 'Impersonate',
    'users.force_logout': 'Force Logout',
    'users.reset_password': 'Reset Password',

    // Applications
    'apps.title': 'Applications',
    'apps.description': 'Manage OAuth applications and API clients',
    'apps.create': 'Create Application',

    // Roles
    'roles.title': 'Roles & Permissions',
    'roles.description': 'Manage access control policies',
    'roles.create': 'Create Role',
    'roles.permissions': 'Permissions',

    // Settings
    'settings.title': 'Settings',
    'settings.description': 'Configure global platform settings',
    'settings.security': 'Security',
    'settings.mfa': 'MFA',
    'settings.email': 'Email',
    'settings.social': 'Social Providers',
    'settings.domain': 'Custom Domain',
    'settings.language': 'Language',

    // Status
    'status.active': 'Active',
    'status.blocked': 'Blocked',
    'status.inactive': 'Inactive',
    'status.pending': 'Pending',
    'status.verified': 'Verified',

    // Confirmations
    'confirm.delete_title': 'Are you sure?',
    'confirm.delete_description': 'This action cannot be undone.',

    // Toast messages
    'toast.saved': 'Changes saved',
    'toast.created': 'Created successfully',
    'toast.deleted': 'Deleted successfully',
    'toast.error': 'An error occurred',
    'toast.copied': 'Copied to clipboard',
  },
  de: {
    'nav.dashboard': 'Dashboard',
    'nav.users': 'Benutzer',
    'nav.applications': 'Anwendungen',
    'nav.tenants': 'Mandanten',
    'nav.organizations': 'Organisationen',
    'nav.roles': 'Rollen & Berechtigungen',
    'nav.branding': 'Branding',
    'nav.webhooks': 'Webhooks',
    'nav.actions': 'Aktionen',
    'nav.email_templates': 'E-Mail-Vorlagen',
    'nav.page_templates': 'Seitenvorlagen',
    'nav.api_keys': 'API-Schluessel',
    'nav.custom_fields': 'Benutzerdefinierte Felder',
    'nav.logs': 'Protokolle',
    'nav.settings': 'Einstellungen',
    'action.create': 'Erstellen',
    'action.edit': 'Bearbeiten',
    'action.delete': 'Loeschen',
    'action.save': 'Speichern',
    'action.cancel': 'Abbrechen',
    'action.confirm': 'Bestaetigen',
    'action.search': 'Suchen...',
    'action.export': 'Exportieren',
    'action.import': 'Importieren',
    'action.refresh': 'Aktualisieren',
    'action.close': 'Schliessen',
    'action.back': 'Zurueck',
    'dashboard.title': 'Dashboard',
    'dashboard.description': 'Ueberblick ueber Ihre Authentifizierungsplattform',
    'dashboard.active_users': 'Aktive Benutzer',
    'dashboard.login_success': 'Login-Erfolgsrate',
    'dashboard.mfa_adoption': 'MFA-Nutzung',
    'dashboard.total_sessions': 'Sitzungen gesamt',
    'dashboard.error_rate': 'Fehlerrate',
    'users.title': 'Benutzer',
    'users.description': 'Benutzerkonten und Identitaeten verwalten',
    'users.create': 'Benutzer erstellen',
    'users.search': 'Benutzer suchen...',
    'users.block': 'Sperren',
    'users.unblock': 'Entsperren',
    'users.impersonate': 'Identitaet annehmen',
    'apps.title': 'Anwendungen',
    'apps.description': 'OAuth-Anwendungen und API-Clients verwalten',
    'apps.create': 'Anwendung erstellen',
    'roles.title': 'Rollen & Berechtigungen',
    'roles.description': 'Zugriffsrichtlinien verwalten',
    'roles.create': 'Rolle erstellen',
    'settings.title': 'Einstellungen',
    'settings.description': 'Globale Plattformeinstellungen konfigurieren',
    'settings.security': 'Sicherheit',
    'settings.domain': 'Eigene Domain',
    'settings.language': 'Sprache',
    'status.active': 'Aktiv',
    'status.blocked': 'Gesperrt',
    'status.inactive': 'Inaktiv',
    'status.pending': 'Ausstehend',
    'status.verified': 'Verifiziert',
    'confirm.delete_title': 'Sind Sie sicher?',
    'confirm.delete_description': 'Diese Aktion kann nicht rueckgaengig gemacht werden.',
    'toast.saved': 'Aenderungen gespeichert',
    'toast.created': 'Erfolgreich erstellt',
    'toast.deleted': 'Erfolgreich geloescht',
    'toast.error': 'Ein Fehler ist aufgetreten',
    'toast.copied': 'In die Zwischenablage kopiert',
  },
  fr: {
    'nav.dashboard': 'Tableau de bord',
    'nav.users': 'Utilisateurs',
    'nav.applications': 'Applications',
    'nav.tenants': 'Locataires',
    'nav.organizations': 'Organisations',
    'nav.roles': 'Roles et Permissions',
    'nav.settings': 'Parametres',
    'action.create': 'Creer',
    'action.edit': 'Modifier',
    'action.delete': 'Supprimer',
    'action.save': 'Enregistrer',
    'action.cancel': 'Annuler',
    'action.search': 'Rechercher...',
    'dashboard.title': 'Tableau de bord',
    'users.title': 'Utilisateurs',
    'users.create': 'Creer un utilisateur',
    'apps.title': 'Applications',
    'apps.create': 'Creer une application',
    'roles.title': 'Roles et Permissions',
    'settings.title': 'Parametres',
    'settings.language': 'Langue',
    'status.active': 'Actif',
    'status.blocked': 'Bloque',
    'toast.saved': 'Modifications enregistrees',
    'toast.error': 'Une erreur est survenue',
  },
  es: {
    'nav.dashboard': 'Panel',
    'nav.users': 'Usuarios',
    'nav.applications': 'Aplicaciones',
    'nav.tenants': 'Inquilinos',
    'nav.organizations': 'Organizaciones',
    'nav.roles': 'Roles y Permisos',
    'nav.settings': 'Configuracion',
    'action.create': 'Crear',
    'action.edit': 'Editar',
    'action.delete': 'Eliminar',
    'action.save': 'Guardar',
    'action.cancel': 'Cancelar',
    'action.search': 'Buscar...',
    'dashboard.title': 'Panel',
    'users.title': 'Usuarios',
    'users.create': 'Crear usuario',
    'apps.title': 'Aplicaciones',
    'apps.create': 'Crear aplicacion',
    'roles.title': 'Roles y Permisos',
    'settings.title': 'Configuracion',
    'settings.language': 'Idioma',
    'status.active': 'Activo',
    'status.blocked': 'Bloqueado',
    'toast.saved': 'Cambios guardados',
    'toast.error': 'Ocurrio un error',
  },
}

interface I18nState {
  locale: Locale
  setLocale: (locale: Locale) => void
  t: (key: string, params?: Record<string, string>) => string
}

export const useI18n = create<I18nState>((set, get) => ({
  locale: (localStorage.getItem('locale') as Locale) || 'en',
  setLocale: (locale: Locale) => {
    localStorage.setItem('locale', locale)
    set({ locale })
  },
  t: (key: string, params?: Record<string, string>) => {
    const { locale } = get()
    let text = translations[locale]?.[key] || translations.en[key] || key
    if (params) {
      for (const [k, v] of Object.entries(params)) {
        text = text.replace(`{${k}}`, v)
      }
    }
    return text
  },
}))

export const localeNames: Record<Locale, string> = {
  en: 'English',
  de: 'Deutsch',
  fr: 'Francais',
  es: 'Espanol',
}
