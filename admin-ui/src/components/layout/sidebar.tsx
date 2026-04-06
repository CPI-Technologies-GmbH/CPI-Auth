import { Link, useLocation } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { useUIStore } from '@/stores/ui'
import { useI18n } from '@/lib/i18n'
import { useAuthStore } from '@/stores/auth'
import { parseTenantRoute, buildTenantPath, buildSystemPath } from '@/lib/tenant-router'
import {
  LayoutDashboard,
  Users,
  AppWindow,
  Building2,
  Building,
  Shield,
  Palette,
  Webhook,
  Zap,
  Mail,
  FileCode,
  Key,
  FormInput,
  ScrollText,
  Settings,
  ChevronLeft,
  ChevronRight,
} from 'lucide-react'

// Tenant-scoped navigation: every link is built relative to the active
// tenant slug parsed from the URL.
const tenantNavItems = [
  { i18nKey: 'nav.dashboard', sub: '/', icon: LayoutDashboard },
  { i18nKey: 'nav.users', sub: '/users', icon: Users },
  { i18nKey: 'nav.applications', sub: '/applications', icon: AppWindow },
  { i18nKey: 'nav.organizations', sub: '/organizations', icon: Building },
  { i18nKey: 'nav.roles', sub: '/roles', icon: Shield },
  { i18nKey: 'nav.branding', sub: '/branding', icon: Palette },
  { i18nKey: 'nav.webhooks', sub: '/webhooks', icon: Webhook },
  { i18nKey: 'nav.actions', sub: '/actions', icon: Zap },
  { i18nKey: 'nav.email_templates', sub: '/email-templates', icon: Mail },
  { i18nKey: 'nav.page_templates', sub: '/page-templates', icon: FileCode },
  { i18nKey: 'nav.api_keys', sub: '/api-keys', icon: Key },
  { i18nKey: 'nav.custom_fields', sub: '/custom-fields', icon: FormInput },
  { i18nKey: 'nav.logs', sub: '/logs', icon: ScrollText },
  { i18nKey: 'nav.settings', sub: '/settings', icon: Settings },
]

// System (platform) navigation: super-admin only.
const systemNavItems = [
  { i18nKey: 'nav.system_dashboard', sub: '', icon: LayoutDashboard },
  { i18nKey: 'nav.tenants', sub: '/tenants', icon: Building2 },
  { i18nKey: 'nav.system_logs', sub: '/logs', icon: ScrollText },
  { i18nKey: 'nav.system_settings', sub: '/settings', icon: Settings },
]

function Sidebar() {
  const location = useLocation()
  const { sidebarCollapsed, setSidebarCollapsed } = useUIStore()
  const { t } = useI18n()
  const { user } = useAuthStore()

  // Resolve the current scope from the URL: tenant page (slug present),
  // system page, or login. Tenant slug is the source of truth for nav links.
  const tenantCtx = parseTenantRoute(location.pathname)
  const isSystemRoute = location.pathname.startsWith('/system')
  const activeSlug = tenantCtx.slug || user?.tenant_slug || ''

  const navItems = isSystemRoute
    ? systemNavItems.map((item) => ({ ...item, path: buildSystemPath(item.sub) }))
    : tenantNavItems.map((item) => ({ ...item, path: buildTenantPath(activeSlug, item.sub) }))

  const isActive = (path: string) => {
    if (path === buildTenantPath(activeSlug)) {
      return location.pathname === path
    }
    if (path === buildSystemPath()) {
      return location.pathname === path
    }
    return location.pathname.startsWith(path)
  }

  const homePath = isSystemRoute ? buildSystemPath() : buildTenantPath(activeSlug)

  return (
    <aside
      className={cn(
        'fixed left-0 top-0 z-40 h-screen border-r bg-sidebar flex flex-col transition-all duration-300',
        sidebarCollapsed ? 'w-16' : 'w-64'
      )}
    >
      <div className="flex h-16 items-center border-b px-4">
        {!sidebarCollapsed && (
          <Link to={homePath} className="flex items-center gap-2">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary">
              <Shield className="h-4 w-4 text-white" />
            </div>
            <span className="text-lg font-bold text-foreground">CPI Auth</span>
          </Link>
        )}
        {sidebarCollapsed && (
          <Link to={homePath} className="flex items-center justify-center w-full">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary">
              <Shield className="h-4 w-4 text-white" />
            </div>
          </Link>
        )}
      </div>

      <nav className="flex-1 overflow-y-auto py-4 px-2">
        <ul className="space-y-1">
          {navItems.map((item) => (
            <li key={item.path}>
              <Link
                to={item.path}
                className={cn(
                  'flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors',
                  isActive(item.path)
                    ? 'bg-sidebar-active/10 text-sidebar-active'
                    : 'text-sidebar-foreground hover:bg-accent hover:text-accent-foreground',
                  sidebarCollapsed && 'justify-center px-2'
                )}
                title={sidebarCollapsed ? t(item.i18nKey) : undefined}
              >
                <item.icon className="h-5 w-5 shrink-0" />
                {!sidebarCollapsed && <span>{t(item.i18nKey)}</span>}
              </Link>
            </li>
          ))}
        </ul>
      </nav>

      <div className="border-t p-2">
        <button
          onClick={() => setSidebarCollapsed(!sidebarCollapsed)}
          className="flex w-full items-center justify-center rounded-lg px-3 py-2 text-sm text-sidebar-foreground hover:bg-accent hover:text-accent-foreground transition-colors cursor-pointer"
        >
          {sidebarCollapsed ? (
            <ChevronRight className="h-4 w-4" />
          ) : (
            <>
              <ChevronLeft className="h-4 w-4 mr-2" />
              <span>Collapse</span>
            </>
          )}
        </button>
      </div>
    </aside>
  )
}

export { Sidebar }
