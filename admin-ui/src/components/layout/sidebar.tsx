import { Link, useLocation } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { useUIStore } from '@/stores/ui'
import { useI18n } from '@/lib/i18n'
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

const navItems = [
  { i18nKey: 'nav.dashboard', path: '/', icon: LayoutDashboard },
  { i18nKey: 'nav.users', path: '/users', icon: Users },
  { i18nKey: 'nav.applications', path: '/applications', icon: AppWindow },
  { i18nKey: 'nav.tenants', path: '/tenants', icon: Building2 },
  { i18nKey: 'nav.organizations', path: '/organizations', icon: Building },
  { i18nKey: 'nav.roles', path: '/roles', icon: Shield },
  { i18nKey: 'nav.branding', path: '/branding', icon: Palette },
  { i18nKey: 'nav.webhooks', path: '/webhooks', icon: Webhook },
  { i18nKey: 'nav.actions', path: '/actions', icon: Zap },
  { i18nKey: 'nav.email_templates', path: '/email-templates', icon: Mail },
  { i18nKey: 'nav.page_templates', path: '/page-templates', icon: FileCode },
  { i18nKey: 'nav.api_keys', path: '/api-keys', icon: Key },
  { i18nKey: 'nav.custom_fields', path: '/custom-fields', icon: FormInput },
  { i18nKey: 'nav.logs', path: '/logs', icon: ScrollText },
  { i18nKey: 'nav.settings', path: '/settings', icon: Settings },
]

function Sidebar() {
  const location = useLocation()
  const { sidebarCollapsed, setSidebarCollapsed } = useUIStore()
  const { t } = useI18n()

  const isActive = (path: string) => {
    if (path === '/') return location.pathname === '/'
    return location.pathname.startsWith(path)
  }

  return (
    <aside
      className={cn(
        'fixed left-0 top-0 z-40 h-screen border-r bg-sidebar flex flex-col transition-all duration-300',
        sidebarCollapsed ? 'w-16' : 'w-64'
      )}
    >
      <div className="flex h-16 items-center border-b px-4">
        {!sidebarCollapsed && (
          <Link to="/" className="flex items-center gap-2">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary">
              <Shield className="h-4 w-4 text-white" />
            </div>
            <span className="text-lg font-bold text-foreground">CPI Auth</span>
          </Link>
        )}
        {sidebarCollapsed && (
          <Link to="/" className="flex items-center justify-center w-full">
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
