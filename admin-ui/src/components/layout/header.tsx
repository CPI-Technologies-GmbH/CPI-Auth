import { useAuthStore } from '@/stores/auth'
import { Avatar } from '@/components/ui/avatar'
import { DropdownMenu, DropdownMenuItem, DropdownMenuSeparator, DropdownMenuLabel } from '@/components/ui/dropdown-menu'
import { Button } from '@/components/ui/button'
import { LogOut, User, Bell, Search, Building2, Check, ArrowLeftRight, Languages } from 'lucide-react'
import { useLocation, useNavigate } from 'react-router-dom'
import { useState } from 'react'
import { Input } from '@/components/ui/input'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import type { Tenant } from '@/types'
import { useI18n, localeNames, type Locale } from '@/lib/i18n'
import { parseTenantRoute, switchTenantPath } from '@/lib/tenant-router'

function TenantSwitcher() {
  const { user, activeTenantId, setActiveTenant } = useAuthStore()
  const queryClient = useQueryClient()
  const navigate = useNavigate()
  const location = useLocation()

  const { data: tenants } = useQuery({
    queryKey: ['tenants'],
    queryFn: () => api.getTenants(),
    enabled: user?.role === 'super_admin',
  })

  if (!tenants || tenants.length === 0) return null

  const currentTenant = tenants.find((t: Tenant) => t.id === activeTenantId)
  const isOverriding = activeTenantId !== user?.tenant_id

  const handleSelect = (tenantId: string) => {
    const target = tenants.find((t: Tenant) => t.id === tenantId)
    if (!target) return
    setActiveTenant(tenantId)
    queryClient.invalidateQueries()
    // Switch the URL to the equivalent path under the new tenant so the
    // browser bar reflects the active tenant. The auth store stays in
    // sync via <TenantSync> on the new route.
    navigate(switchTenantPath(location.pathname, target.slug))
  }

  return (
    <DropdownMenu
      align="left"
      trigger={
        <button className={`flex items-center gap-2 rounded-lg border px-3 py-1.5 text-sm transition-colors cursor-pointer ${isOverriding ? 'border-amber-500/50 bg-amber-500/10 text-amber-700 dark:text-amber-400' : 'border-border hover:bg-accent'}`}>
          <Building2 className="h-4 w-4" />
          <span className="max-w-[150px] truncate">{currentTenant?.name || 'Select tenant'}</span>
          {isOverriding && <ArrowLeftRight className="h-3 w-3" />}
        </button>
      }
    >
      <DropdownMenuLabel>Switch Tenant</DropdownMenuLabel>
      <DropdownMenuSeparator />
      {tenants.map((tenant: Tenant) => (
        <DropdownMenuItem key={tenant.id} onClick={() => handleSelect(tenant.id)}>
          <Check className={`mr-2 h-4 w-4 ${tenant.id === activeTenantId ? 'opacity-100' : 'opacity-0'}`} />
          <span className="truncate">{tenant.name}</span>
          {tenant.id === user?.tenant_id && (
            <span className="ml-auto text-xs text-muted-foreground">(home)</span>
          )}
        </DropdownMenuItem>
      ))}
    </DropdownMenu>
  )
}

function LanguageSwitcher() {
  const { locale, setLocale } = useI18n()

  return (
    <DropdownMenu
      align="right"
      trigger={
        <button className="flex items-center gap-1.5 rounded-lg border border-border px-2 py-1.5 text-sm hover:bg-accent transition-colors cursor-pointer">
          <Languages className="h-4 w-4" />
          <span className="hidden sm:inline">{localeNames[locale].slice(0, 2).toUpperCase()}</span>
        </button>
      }
    >
      <DropdownMenuLabel>Language</DropdownMenuLabel>
      <DropdownMenuSeparator />
      {(Object.entries(localeNames) as [Locale, string][]).map(([code, name]) => (
        <DropdownMenuItem key={code} onClick={() => setLocale(code)}>
          <Check className={`mr-2 h-4 w-4 ${code === locale ? 'opacity-100' : 'opacity-0'}`} />
          {name}
        </DropdownMenuItem>
      ))}
    </DropdownMenu>
  )
}

function Header() {
  const { user, logout } = useAuthStore()
  const navigate = useNavigate()
  const location = useLocation()
  const [searchOpen, setSearchOpen] = useState(false)

  const handleLogout = async () => {
    await logout()
    navigate('/login')
  }

  // Profile link goes to the active tenant's settings page when on a
  // tenant route, and to /system/settings on platform routes.
  const tenantCtx = parseTenantRoute(location.pathname)
  const profilePath = tenantCtx.slug
    ? `/t/${tenantCtx.slug}/settings`
    : '/system/settings'

  return (
    <header className="sticky top-0 z-30 flex h-16 items-center justify-between border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 px-6">
      <div className="flex items-center gap-4 flex-1">
        {searchOpen ? (
          <div className="flex items-center gap-2 max-w-md flex-1">
            <Input
              placeholder="Search users, apps, organizations..."
              autoFocus
              onBlur={() => setSearchOpen(false)}
              onKeyDown={(e) => e.key === 'Escape' && setSearchOpen(false)}
              className="h-9"
            />
          </div>
        ) : (
          <Button variant="ghost" size="sm" onClick={() => setSearchOpen(true)} className="text-muted-foreground">
            <Search className="h-4 w-4 mr-2" />
            Search...
            <kbd className="ml-4 pointer-events-none inline-flex h-5 select-none items-center gap-1 rounded border bg-muted px-1.5 font-mono text-[10px] font-medium text-muted-foreground">
              /
            </kbd>
          </Button>
        )}
      </div>

      <div className="flex items-center gap-3">
        {user?.role === 'super_admin' && <TenantSwitcher />}

        <LanguageSwitcher />

        <Button variant="ghost" size="icon-sm" className="relative">
          <Bell className="h-4 w-4" />
          <span className="absolute -top-0.5 -right-0.5 h-2 w-2 rounded-full bg-destructive" />
        </Button>

        <DropdownMenu
          trigger={
            <button className="flex items-center gap-2 rounded-lg px-2 py-1.5 hover:bg-accent transition-colors cursor-pointer">
              <Avatar name={user?.name} src={user?.avatar_url} size="sm" />
              <div className="hidden md:block text-left">
                <p className="text-sm font-medium leading-tight">{user?.name || 'Admin'}</p>
                <p className="text-xs text-muted-foreground leading-tight">{user?.email || 'admin@cpi-auth.io'}</p>
              </div>
            </button>
          }
        >
          <DropdownMenuLabel>My Account</DropdownMenuLabel>
          <DropdownMenuSeparator />
          <DropdownMenuItem onClick={() => navigate(profilePath)}>
            <User className="mr-2 h-4 w-4" />
            Profile
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem onClick={handleLogout} destructive>
            <LogOut className="mr-2 h-4 w-4" />
            Log out
          </DropdownMenuItem>
        </DropdownMenu>
      </div>
    </header>
  )
}

export { Header }
