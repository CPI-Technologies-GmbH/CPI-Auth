import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { PageHeader } from '@/components/shared/page-header'
import {
  Users,
  CheckCircle,
  ShieldCheck,
  Monitor,
  AlertTriangle,
  ArrowUpRight,
  ArrowDownRight,
  TrendingUp,
} from 'lucide-react'
import {
  LineChart,
  Line,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts'
import { formatRelativeTime, cn } from '@/lib/utils'
import { useI18n } from '@/lib/i18n'
import type { DashboardMetrics, LoginChartData, AuthMethodData, RecentEvent } from '@/types'

function MetricCard({
  title,
  value,
  change,
  icon: Icon,
  loading,
  format = 'number',
}: {
  title: string
  value: number
  change: number
  icon: React.ElementType
  loading?: boolean
  format?: 'number' | 'percent'
}) {
  if (loading) {
    return (
      <Card>
        <CardContent className="p-6">
          <div className="flex items-center justify-between">
            <Skeleton className="h-4 w-24" />
            <Skeleton className="h-8 w-8 rounded" />
          </div>
          <Skeleton className="h-8 w-20 mt-3" />
          <Skeleton className="h-3 w-16 mt-2" />
        </CardContent>
      </Card>
    )
  }

  const isPositive = change >= 0

  return (
    <Card>
      <CardContent className="p-6">
        <div className="flex items-center justify-between">
          <p className="text-sm font-medium text-muted-foreground">{title}</p>
          <div className="rounded-lg bg-primary/10 p-2">
            <Icon className="h-4 w-4 text-primary" />
          </div>
        </div>
        <div className="mt-3">
          <p className="text-2xl font-bold">
            {format === 'percent' ? `${value.toFixed(1)}%` : value.toLocaleString()}
          </p>
        </div>
        <div className="mt-1 flex items-center gap-1 text-xs">
          {isPositive ? (
            <ArrowUpRight className="h-3 w-3 text-success" />
          ) : (
            <ArrowDownRight className="h-3 w-3 text-destructive" />
          )}
          <span className={cn(isPositive ? 'text-success' : 'text-destructive')}>
            {isPositive ? '+' : ''}{change.toFixed(1)}%
          </span>
          <span className="text-muted-foreground">vs last period</span>
        </div>
      </CardContent>
    </Card>
  )
}

function LoginChart({ data, loading }: { data?: LoginChartData[]; loading: boolean }) {
  const [period, setPeriod] = useState<'7d' | '30d'>('7d')

  const { data: chartData, isLoading: chartLoading } = useQuery({
    queryKey: ['login-chart', period],
    queryFn: () => api.getLoginChart(period),
    placeholderData: data,
  })

  const isLoading = loading || chartLoading

  return (
    <Card className="col-span-2">
      <CardHeader className="flex flex-row items-center justify-between">
        <CardTitle className="text-base">Logins Over Time</CardTitle>
        <div className="flex gap-1">
          <Button
            variant={period === '7d' ? 'secondary' : 'ghost'}
            size="sm"
            onClick={() => setPeriod('7d')}
          >
            7D
          </Button>
          <Button
            variant={period === '30d' ? 'secondary' : 'ghost'}
            size="sm"
            onClick={() => setPeriod('30d')}
          >
            30D
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <Skeleton className="h-[300px] w-full" />
        ) : (
          <ResponsiveContainer width="100%" height={300}>
            <LineChart data={chartData || []}>
              <CartesianGrid strokeDasharray="3 3" stroke="#27272a" />
              <XAxis dataKey="date" stroke="#71717a" fontSize={12} />
              <YAxis stroke="#71717a" fontSize={12} />
              <Tooltip
                contentStyle={{
                  backgroundColor: '#111118',
                  border: '1px solid #27272a',
                  borderRadius: '8px',
                  fontSize: '12px',
                }}
              />
              <Line
                type="monotone"
                dataKey="logins"
                stroke="#6366f1"
                strokeWidth={2}
                dot={false}
                name="Successful"
              />
              <Line
                type="monotone"
                dataKey="failures"
                stroke="#ef4444"
                strokeWidth={2}
                dot={false}
                name="Failed"
              />
            </LineChart>
          </ResponsiveContainer>
        )}
      </CardContent>
    </Card>
  )
}

function AuthMethodsChart({ data, loading }: { data?: AuthMethodData[]; loading: boolean }) {
  if (loading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Auth Methods</CardTitle>
        </CardHeader>
        <CardContent>
          <Skeleton className="h-[300px] w-full" />
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Auth Methods</CardTitle>
      </CardHeader>
      <CardContent>
        <ResponsiveContainer width="100%" height={300}>
          <BarChart data={data || []} layout="vertical">
            <CartesianGrid strokeDasharray="3 3" stroke="#27272a" horizontal={false} />
            <XAxis type="number" stroke="#71717a" fontSize={12} />
            <YAxis dataKey="method" type="category" stroke="#71717a" fontSize={12} width={100} />
            <Tooltip
              contentStyle={{
                backgroundColor: '#111118',
                border: '1px solid #27272a',
                borderRadius: '8px',
                fontSize: '12px',
              }}
            />
            <Bar dataKey="count" fill="#6366f1" radius={[0, 4, 4, 0]} />
          </BarChart>
        </ResponsiveContainer>
      </CardContent>
    </Card>
  )
}

function RecentEventsList({ data, loading }: { data?: RecentEvent[]; loading: boolean }) {
  if (loading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Recent Events</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {Array.from({ length: 5 }).map((_, i) => (
              <div key={i} className="flex items-center gap-3">
                <Skeleton className="h-8 w-8 rounded-full" />
                <div className="flex-1">
                  <Skeleton className="h-3 w-3/4" />
                  <Skeleton className="h-2 w-1/2 mt-1" />
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    )
  }

  const eventIcons: Record<string, string> = {
    'user.login': 'text-success',
    'user.created': 'text-primary',
    'user.blocked': 'text-destructive',
    'user.deleted': 'text-destructive',
    'application.created': 'text-primary',
  }

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <CardTitle className="text-base">Recent Events</CardTitle>
        <Button variant="ghost" size="sm">
          View all
        </Button>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {(data || []).slice(0, 8).map((event) => (
            <div key={event.id} className="flex items-start gap-3">
              <div className={cn('mt-0.5 h-2 w-2 rounded-full bg-current shrink-0', eventIcons[event.type] || 'text-muted-foreground')} />
              <div className="flex-1 min-w-0">
                <p className="text-sm leading-tight">{event.description}</p>
                <div className="flex items-center gap-2 mt-0.5">
                  <span className="text-xs text-muted-foreground">{event.actor}</span>
                  <span className="text-xs text-muted-foreground">{formatRelativeTime(event.created_at)}</span>
                </div>
              </div>
              <Badge variant="muted" className="shrink-0 text-[10px]">
                {event.type.split('.')[1]}
              </Badge>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}

function ErrorRateWidget({ metrics, loading }: { metrics?: DashboardMetrics; loading: boolean }) {
  if (loading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Error Rate</CardTitle>
        </CardHeader>
        <CardContent>
          <Skeleton className="h-24 w-full" />
        </CardContent>
      </Card>
    )
  }

  const rate = metrics?.error_rate ?? 0
  const change = metrics?.error_rate_change ?? 0
  const isHigh = rate > 5

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base flex items-center gap-2">
          <AlertTriangle className={cn('h-4 w-4', isHigh ? 'text-destructive' : 'text-success')} />
          Error Rate
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="flex items-baseline gap-2">
          <span className={cn('text-3xl font-bold', isHigh ? 'text-destructive' : 'text-success')}>
            {rate.toFixed(2)}%
          </span>
          <div className="flex items-center gap-0.5 text-xs">
            {change >= 0 ? (
              <TrendingUp className="h-3 w-3 text-destructive" />
            ) : (
              <TrendingUp className="h-3 w-3 text-success rotate-180" />
            )}
            <span className={cn(change >= 0 ? 'text-destructive' : 'text-success')}>
              {change >= 0 ? '+' : ''}{change.toFixed(1)}%
            </span>
          </div>
        </div>
        <p className="mt-2 text-xs text-muted-foreground">
          {isHigh ? 'Error rate is above threshold. Check logs for details.' : 'Error rate is within normal range.'}
        </p>
      </CardContent>
    </Card>
  )
}

export default function DashboardPage() {
  const { t } = useI18n()
  const { data: metrics, isLoading: metricsLoading } = useQuery({
    queryKey: ['dashboard-metrics'],
    queryFn: () => api.getDashboardMetrics(),
  })

  const { data: authMethods, isLoading: authMethodsLoading } = useQuery({
    queryKey: ['auth-methods'],
    queryFn: () => api.getAuthMethodsChart(),
  })

  const { data: events, isLoading: eventsLoading } = useQuery({
    queryKey: ['recent-events'],
    queryFn: () => api.getRecentEvents(),
  })

  return (
    <div>
      <PageHeader title={t('dashboard.title')} description={t('dashboard.description')} />

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4 xl:grid-cols-5">
        <MetricCard
          title={t('dashboard.active_users')}
          value={metrics?.active_users ?? 0}
          change={metrics?.active_users_change ?? 0}
          icon={Users}
          loading={metricsLoading}
        />
        <MetricCard
          title={t('dashboard.login_success')}
          value={metrics?.login_success_rate ?? 0}
          change={metrics?.login_success_rate_change ?? 0}
          icon={CheckCircle}
          loading={metricsLoading}
          format="percent"
        />
        <MetricCard
          title={t('dashboard.mfa_adoption')}
          value={metrics?.mfa_adoption ?? 0}
          change={metrics?.mfa_adoption_change ?? 0}
          icon={ShieldCheck}
          loading={metricsLoading}
          format="percent"
        />
        <MetricCard
          title={t('dashboard.total_sessions')}
          value={metrics?.total_sessions ?? 0}
          change={metrics?.total_sessions_change ?? 0}
          icon={Monitor}
          loading={metricsLoading}
        />
        <MetricCard
          title={t('dashboard.error_rate')}
          value={metrics?.error_rate ?? 0}
          change={metrics?.error_rate_change ?? 0}
          icon={AlertTriangle}
          loading={metricsLoading}
          format="percent"
        />
      </div>

      <div className="mt-6 grid gap-4 lg:grid-cols-3">
        <LoginChart loading={metricsLoading} />
        <AuthMethodsChart data={authMethods} loading={authMethodsLoading} />
      </div>

      <div className="mt-6 grid gap-4 md:grid-cols-2">
        <RecentEventsList data={events} loading={eventsLoading} />
        <ErrorRateWidget metrics={metrics} loading={metricsLoading} />
      </div>
    </div>
  )
}
