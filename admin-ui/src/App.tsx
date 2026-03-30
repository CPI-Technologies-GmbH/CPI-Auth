import { useEffect, lazy, Suspense } from 'react'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
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
        <BrowserRouter>
          <AppInitializer>
            <Suspense fallback={<PageLoader />}>
              <Routes>
                <Route path="/login" element={<LoginPage />} />

                <Route element={<AppLayout />}>
                  <Route path="/" element={<DashboardPage />} />
                  <Route path="/users" element={<UsersPage />} />
                  <Route path="/users/:id" element={<UserDetailPage />} />
                  <Route path="/applications" element={<ApplicationsPage />} />
                  <Route path="/applications/:id" element={<ApplicationDetailPage />} />
                  <Route path="/tenants" element={<TenantsPage />} />
                  <Route path="/organizations" element={<OrganizationsPage />} />
                  <Route path="/roles" element={<RolesPage />} />
                  <Route path="/branding" element={<BrandingPage />} />
                  <Route path="/webhooks" element={<WebhooksPage />} />
                  <Route path="/actions" element={<ActionsPage />} />
                  <Route path="/email-templates" element={<EmailTemplatesPage />} />
                  <Route path="/api-keys" element={<ApiKeysPage />} />
                  <Route path="/custom-fields" element={<CustomFieldsPage />} />
                  <Route path="/page-templates" element={<PageTemplatesPage />} />
                  <Route path="/logs" element={<LogsPage />} />
                  <Route path="/settings" element={<SettingsPage />} />
                </Route>
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
