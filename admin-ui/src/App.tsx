import { useEffect, lazy, Suspense } from 'react'
import { BrowserRouter, Routes, Route, Navigate, useParams } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useAuthStore } from '@/stores/auth'
import { AppLayout } from '@/components/layout/app-layout'
import { Toaster } from '@/components/shared/toast'
import { ErrorBoundary } from '@/components/shared/error-boundary'
import { Loader2 } from 'lucide-react'

// Lazy load pages
const LoginPage = lazy(() => import('@/pages/login'))
const DashboardPage = lazy(() => import('@/pages/dashboard'))
const UsersPage = lazy(() => import('@/pages/users'))
const UserDetailPage = lazy(() => import('@/pages/users/detail'))
const ApplicationsPage = lazy(() => import('@/pages/applications'))
const ApplicationDetailPage = lazy(() => import('@/pages/applications/detail'))
const TenantsPage = lazy(() => import('@/pages/tenants'))
const OrganizationsPage = lazy(() => import('@/pages/organizations'))
const RolesPage = lazy(() => import('@/pages/roles'))
const BrandingPage = lazy(() => import('@/pages/branding'))
const WebhooksPage = lazy(() => import('@/pages/webhooks'))
const ActionsPage = lazy(() => import('@/pages/actions'))
const EmailTemplatesPage = lazy(() => import('@/pages/email-templates'))
const ApiKeysPage = lazy(() => import('@/pages/api-keys'))
const CustomFieldsPage = lazy(() => import('@/pages/custom-fields'))
const PageTemplatesPage = lazy(() => import('@/pages/page-templates'))
const LogsPage = lazy(() => import('@/pages/logs'))
const SettingsPage = lazy(() => import('@/pages/settings'))

/**
 * TenantSync syncs the active tenant id in the auth store from the URL slug.
 * Wrapped around all /t/:slug routes so the URL is the source of truth.
 */
function TenantSync({ children }: { children: React.ReactNode }) {
  const { slug } = useParams<{ slug: string }>()
  const { user, setActiveTenantBySlug } = useAuthStore()

  useEffect(() => {
    if (slug && user) {
      setActiveTenantBySlug(slug)
    }
  }, [slug, user, setActiveTenantBySlug])

  return <>{children}</>
}

/**
 * RootRedirect chooses where to send a freshly authenticated user. Super-admins
 * land on /system; tenant-scoped admins land on their home tenant.
 */
function RootRedirect() {
  const { user } = useAuthStore()
  if (!user) return <Navigate to="/login" replace />
  if (user.role === 'super_admin') return <Navigate to="/system" replace />
  if (user.tenant_slug) return <Navigate to={`/t/${user.tenant_slug}`} replace />
  return <Navigate to={`/t/${user.tenant_id}`} replace />
}

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      staleTime: 30_000,
      refetchOnWindowFocus: false,
    },
  },
})

function PageLoader() {
  return (
    <div className="flex items-center justify-center min-h-[400px]">
      <Loader2 className="h-8 w-8 animate-spin text-primary" />
    </div>
  )
}

function AppInitializer({ children }: { children: React.ReactNode }) {
  const { loadUser } = useAuthStore()

  useEffect(() => {
    loadUser()
  }, [loadUser])

  return <>{children}</>
}

function App() {
  return (
    <ErrorBoundary>
      <QueryClientProvider client={queryClient}>
        {/*
         * basename="/admin" — the admin UI is mounted under /admin/ behind
         * the ingress. All Route paths are relative to that. The browser
         * sees /admin/system, /admin/t/lastsoftware/users, etc.
         */}
        <BrowserRouter basename="/admin">
          <AppInitializer>
            <Suspense fallback={<PageLoader />}>
              <Routes>
                <Route path="/login" element={<LoginPage />} />

                {/*
                 * System (platform) routes — for super-admins managing the
                 * authforge installation itself, not individual tenants.
                 */}
                <Route element={<AppLayout />}>
                  <Route path="/system" element={<DashboardPage />} />
                  <Route path="/system/tenants" element={<TenantsPage />} />
                  <Route path="/system/settings" element={<SettingsPage />} />
                  <Route path="/system/logs" element={<LogsPage />} />
                </Route>

                {/*
                 * Tenant-scoped routes — every URL carries the active tenant
                 * slug. The TenantSync wrapper keeps the auth store in sync
                 * so existing API requests (which read activeTenantId from
                 * the store) target the right tenant.
                 */}
                <Route element={<AppLayout />}>
                  <Route
                    path="/t/:slug"
                    element={<TenantSync><DashboardPage /></TenantSync>}
                  />
                  <Route
                    path="/t/:slug/users"
                    element={<TenantSync><UsersPage /></TenantSync>}
                  />
                  <Route
                    path="/t/:slug/users/:id"
                    element={<TenantSync><UserDetailPage /></TenantSync>}
                  />
                  <Route
                    path="/t/:slug/applications"
                    element={<TenantSync><ApplicationsPage /></TenantSync>}
                  />
                  <Route
                    path="/t/:slug/applications/:id"
                    element={<TenantSync><ApplicationDetailPage /></TenantSync>}
                  />
                  <Route
                    path="/t/:slug/organizations"
                    element={<TenantSync><OrganizationsPage /></TenantSync>}
                  />
                  <Route
                    path="/t/:slug/roles"
                    element={<TenantSync><RolesPage /></TenantSync>}
                  />
                  <Route
                    path="/t/:slug/branding"
                    element={<TenantSync><BrandingPage /></TenantSync>}
                  />
                  <Route
                    path="/t/:slug/webhooks"
                    element={<TenantSync><WebhooksPage /></TenantSync>}
                  />
                  <Route
                    path="/t/:slug/actions"
                    element={<TenantSync><ActionsPage /></TenantSync>}
                  />
                  <Route
                    path="/t/:slug/email-templates"
                    element={<TenantSync><EmailTemplatesPage /></TenantSync>}
                  />
                  <Route
                    path="/t/:slug/api-keys"
                    element={<TenantSync><ApiKeysPage /></TenantSync>}
                  />
                  <Route
                    path="/t/:slug/custom-fields"
                    element={<TenantSync><CustomFieldsPage /></TenantSync>}
                  />
                  <Route
                    path="/t/:slug/page-templates"
                    element={<TenantSync><PageTemplatesPage /></TenantSync>}
                  />
                  <Route
                    path="/t/:slug/logs"
                    element={<TenantSync><LogsPage /></TenantSync>}
                  />
                  <Route
                    path="/t/:slug/settings"
                    element={<TenantSync><SettingsPage /></TenantSync>}
                  />
                </Route>

                {/* Root → redirect to /system or /t/{home}, depending on role */}
                <Route path="/" element={<RootRedirect />} />
              </Routes>
            </Suspense>
            <Toaster />
          </AppInitializer>
        </BrowserRouter>
      </QueryClientProvider>
    </ErrorBoundary>
  )
}

export default App
