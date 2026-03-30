import { useState, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { useUIStore } from '@/stores/ui'
import { PageHeader } from '@/components/shared/page-header'
import { DataTable, type Column } from '@/components/shared/data-table'
import { Pagination } from '@/components/shared/pagination'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import {
  Dialog,
  DialogHeader,
  DialogTitle,
  DialogClose,
} from '@/components/ui/dialog'
import { JsonEditor } from '@/components/shared/json-editor'
import { Download, Search, ScrollText, RefreshCw } from 'lucide-react'
import { formatDateTime } from '@/lib/utils'
import type { AuditLogEntry } from '@/types'

const actionOptions = [
  { value: '', label: 'All actions' },
  { value: 'user.login', label: 'User Login' },
  { value: 'user.logout', label: 'User Logout' },
  { value: 'user.created', label: 'User Created' },
  { value: 'user.updated', label: 'User Updated' },
  { value: 'user.deleted', label: 'User Deleted' },
  { value: 'user.blocked', label: 'User Blocked' },
  { value: 'user.password_changed', label: 'Password Changed' },
  { value: 'application.created', label: 'App Created' },
  { value: 'application.updated', label: 'App Updated' },
  { value: 'role.assigned', label: 'Role Assigned' },
  { value: 'settings.updated', label: 'Settings Updated' },
]

const actionColors: Record<string, string> = {
  'user.login': 'text-success',
  'user.logout': 'text-muted-foreground',
  'user.created': 'text-primary',
  'user.deleted': 'text-destructive',
  'user.blocked': 'text-destructive',
  'user.updated': 'text-blue-400',
  'user.password_changed': 'text-warning',
  'application.created': 'text-primary',
  'settings.updated': 'text-warning',
}

export default function LogsPage() {
  const { addToast } = useUIStore()

  const [actionFilter, setActionFilter] = useState('')
  const [search, setSearch] = useState('')
  const [dateFrom, setDateFrom] = useState('')
  const [dateTo, setDateTo] = useState('')
  const [cursor, setCursor] = useState<string | undefined>()
  const [prevCursors, setPrevCursors] = useState<string[]>([])
  const [selectedLog, setSelectedLog] = useState<AuditLogEntry | null>(null)

  const { data, isLoading, refetch } = useQuery({
    queryKey: ['audit-logs', { action: actionFilter, search, dateFrom, dateTo, cursor }],
    queryFn: () =>
      api.getAuditLogs({
        action: actionFilter || undefined,
        actor_id: search || undefined,
        date_from: dateFrom || undefined,
        date_to: dateTo || undefined,
        cursor,
        limit: 50,
      }),
    refetchInterval: 10000, // Real-time polling every 10s
  })

  const handleExport = async () => {
    try {
      const blob = await api.exportAuditLogs({
        action: actionFilter || undefined,
        date_from: dateFrom || undefined,
        date_to: dateTo || undefined,
      })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = 'audit-logs.csv'
      a.click()
      URL.revokeObjectURL(url)
      addToast({ title: 'Logs exported', variant: 'success' })
    } catch {
      addToast({ title: 'Export failed', variant: 'error' })
    }
  }

  const columns: Column<AuditLogEntry>[] = useMemo(() => [
    {
      key: 'action',
      header: 'Action',
      render: (log) => (
        <div className="flex items-center gap-2">
          <div className={`h-2 w-2 rounded-full bg-current ${actionColors[log.action] || 'text-muted-foreground'}`} />
          <Badge variant="outline" className="font-mono text-[10px]">{log.action}</Badge>
        </div>
      ),
    },
    {
      key: 'actor',
      header: 'Actor',
      render: (log) => (
        <div>
          <p className="text-sm">{log.actor_email || (log.actor_id?.slice(0, 12) ?? 'unknown') + '...'}</p>
        </div>
      ),
    },
    {
      key: 'target',
      header: 'Target',
      render: (log) => (
        <span className="text-muted-foreground text-xs">
          {log.target_type ? `${log.target_type}:${log.target_id?.slice(0, 8)}` : '-'}
        </span>
      ),
    },
    {
      key: 'ip',
      header: 'IP Address',
      render: (log) => (
        <span className="text-muted-foreground text-xs font-mono">{log.ip_address || '-'}</span>
      ),
    },
    {
      key: 'timestamp',
      header: 'Timestamp',
      sortable: true,
      render: (log) => (
        <span className="text-muted-foreground text-xs">{formatDateTime(log.created_at)}</span>
      ),
    },
  ], [])

  return (
    <div>
      <PageHeader
        title="Audit Logs"
        description="Real-time audit trail of all platform events"
        breadcrumbs={[{ label: 'Dashboard', href: '/' }, { label: 'Logs' }]}
        actions={
          <div className="flex gap-2">
            <Button variant="outline" onClick={() => refetch()}>
              <RefreshCw className="mr-2 h-4 w-4" />
              Refresh
            </Button>
            <Button variant="outline" onClick={handleExport}>
              <Download className="mr-2 h-4 w-4" />
              Export
            </Button>
          </div>
        }
      />

      <div className="flex flex-col sm:flex-row gap-3 mb-4">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search by actor..."
            value={search}
            onChange={(e) => {
              setSearch(e.target.value)
              setCursor(undefined)
              setPrevCursors([])
            }}
            className="pl-9"
          />
        </div>
        <Select
          value={actionFilter}
          onChange={(e) => {
            setActionFilter(e.target.value)
            setCursor(undefined)
            setPrevCursors([])
          }}
          options={actionOptions}
          className="w-44"
        />
        <Input
          type="date"
          value={dateFrom}
          onChange={(e) => {
            setDateFrom(e.target.value)
            setCursor(undefined)
            setPrevCursors([])
          }}
          className="w-40"
          placeholder="From"
        />
        <Input
          type="date"
          value={dateTo}
          onChange={(e) => {
            setDateTo(e.target.value)
            setCursor(undefined)
            setPrevCursors([])
          }}
          className="w-40"
          placeholder="To"
        />
      </div>

      <div className="flex items-center gap-2 mb-2">
        <div className="h-2 w-2 rounded-full bg-success animate-pulse" />
        <span className="text-xs text-muted-foreground">Live - auto-refreshing every 10 seconds</span>
        {data?.total !== undefined && (
          <Badge variant="muted" className="ml-2">{data.total} total entries</Badge>
        )}
      </div>

      <DataTable
        columns={columns}
        data={data?.data ?? []}
        isLoading={isLoading}
        getRowId={(l) => l.id}
        onRowClick={(log) => setSelectedLog(log)}
        emptyMessage="No log entries found"
        emptyIcon={<ScrollText className="h-10 w-10" />}
      />

      <Pagination
        hasMore={data?.has_more ?? false}
        hasPrevious={prevCursors.length > 0}
        onNext={() => {
          if (data?.cursor) {
            setPrevCursors([...prevCursors, cursor ?? ''])
            setCursor(data.cursor)
          }
        }}
        onPrevious={() => {
          const prev = [...prevCursors]
          const prevCursor = prev.pop()
          setPrevCursors(prev)
          setCursor(prevCursor || undefined)
        }}
        loading={isLoading}
      />

      {/* Log Detail Modal */}
      <Dialog open={!!selectedLog} onClose={() => setSelectedLog(null)} className="max-w-xl">
        <DialogClose onClose={() => setSelectedLog(null)} />
        {selectedLog && (
          <>
            <DialogHeader>
              <DialogTitle>Log Entry Detail</DialogTitle>
            </DialogHeader>
            <div className="mt-4 space-y-3 text-sm">
              <div className="grid grid-cols-2 gap-2">
                <div>
                  <p className="text-xs text-muted-foreground">Action</p>
                  <Badge variant="outline" className="font-mono">{selectedLog.action}</Badge>
                </div>
                <div>
                  <p className="text-xs text-muted-foreground">Timestamp</p>
                  <p>{formatDateTime(selectedLog.created_at)}</p>
                </div>
                <div>
                  <p className="text-xs text-muted-foreground">Actor</p>
                  <p className="font-mono text-xs">{selectedLog.actor_email || selectedLog.actor_id}</p>
                </div>
                <div>
                  <p className="text-xs text-muted-foreground">IP Address</p>
                  <p className="font-mono text-xs">{selectedLog.ip_address || '-'}</p>
                </div>
                <div>
                  <p className="text-xs text-muted-foreground">Target</p>
                  <p className="font-mono text-xs">
                    {selectedLog.target_type ? `${selectedLog.target_type}:${selectedLog.target_id}` : '-'}
                  </p>
                </div>
                <div>
                  <p className="text-xs text-muted-foreground">Tenant</p>
                  <p className="font-mono text-xs">{selectedLog.tenant_id}</p>
                </div>
              </div>
              {selectedLog.user_agent && (
                <div>
                  <p className="text-xs text-muted-foreground mb-1">User Agent</p>
                  <p className="text-xs font-mono bg-muted p-2 rounded">{selectedLog.user_agent}</p>
                </div>
              )}
              {selectedLog.metadata && Object.keys(selectedLog.metadata).length > 0 && (
                <div>
                  <p className="text-xs text-muted-foreground mb-1">Metadata</p>
                  <JsonEditor value={selectedLog.metadata} readOnly height="150px" />
                </div>
              )}
            </div>
          </>
        )}
      </Dialog>
    </div>
  )
}
