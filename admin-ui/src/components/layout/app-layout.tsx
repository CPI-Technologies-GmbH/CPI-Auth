import { Outlet, Navigate } from 'react-router-dom'
import { Sidebar } from './sidebar'
import { Header } from './header'
import { useAuthStore } from '@/stores/auth'
import { useUIStore } from '@/stores/ui'
import { cn } from '@/lib/utils'
import { Loader2 } from 'lucide-react'

function AppLayout() {
  const { isAuthenticated, isLoading } = useAuthStore()
  const { sidebarCollapsed } = useUIStore()

  if (isLoading) {
    return (
      <div className="flex h-screen items-center justify-center bg-background">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    )
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }

  return (
    <div className="min-h-screen bg-background">
      <Sidebar />
      <div
        className={cn(
          'transition-all duration-300',
          sidebarCollapsed ? 'ml-16' : 'ml-64'
        )}
      >
        <Header />
        <main className="p-6">
          <Outlet />
        </main>
      </div>
    </div>
  )
}

export { AppLayout }
