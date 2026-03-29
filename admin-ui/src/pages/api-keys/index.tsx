import { useState, useMemo } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { useUIStore } from '@/stores/ui'
import { PageHeader } from '@/components/shared/page-header'
import { EmptyState } from '@/components/shared/empty-state'
import { DataTable, type Column } from '@/components/shared/data-table'
import { ConfirmDialog } from '@/components/shared/confirm-dialog'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Dialog,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
  DialogClose,
} from '@/components/ui/dialog'
import { Plus, Key, Copy, Trash2, AlertTriangle } from 'lucide-react'
import { formatDate, formatRelativeTime, copyToClipboard } from '@/lib/utils'
import type { ApiKey } from '@/types'

const availableScopes = [
  'read:users',
  'write:users',
  'read:applications',
  'write:applications',
  'read:organizations',
  'write:organizations',
  'read:roles',
  'write:roles',
  'read:logs',
  'read:settings',
  'write:settings',
  'read:tenants',
  'write:tenants',
]

export default function ApiKeysPage() {
  const queryClient = useQueryClient()
  const { addToast } = useUIStore()

  const [createOpen, setCreateOpen] = useState(false)
  const [revokeKey, setRevokeKey] = useState<ApiKey | null>(null)
  const [newKeyResult, setNewKeyResult] = useState<string | null>(null)
  const [formData, setFormData] = useState({
    name: '',
    scopes: [] as string[],
    rate_limit: 1000,
    expires_at: '',
  })

  const { data: keys, isLoading } = useQuery({
    queryKey: ['api-keys'],
    queryFn: () => api.getApiKeys(),
  })

  const createMutation = useMutation({
    mutationFn: (data: typeof formData) =>
      api.createApiKey({
        ...data,
        expires_at: data.expires_at || undefined,
      }),
    onSuccess: (result) => {
      queryClient.invalidateQueries({ queryKey: ['api-keys'] })
      setCreateOpen(false)
      setNewKeyResult(result.key)
      addToast({ title: 'API key created', variant: 'success' })
    },
    onError: () => addToast({ title: 'Failed to create API key', variant: 'error' }),
  })

  const revokeMutation = useMutation({
    mutationFn: (id: string) => api.revokeApiKey(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['api-keys'] })
      setRevokeKey(null)
      addToast({ title: 'API key revoked', variant: 'success' })
    },
  })

  const toggleScope = (scope: string) => {
    setFormData((prev) => ({
      ...prev,
      scopes: prev.scopes.includes(scope)
        ? prev.scopes.filter((s) => s !== scope)
        : [...prev.scopes, scope],
    }))
  }

  const columns: Column<ApiKey>[] = useMemo(() => [
    {
      key: 'name',
      header: 'Name',
      render: (key) => (
        <div className="flex items-center gap-3">
          <Key className="h-4 w-4 text-primary" />
          <div>
            <p className="font-medium">{key.name}</p>
            <p className="text-xs text-muted-foreground font-mono">{key.key_prefix}...</p>
          </div>
        </div>
      ),
    },
    {
      key: 'scopes',
      header: 'Scopes',
      render: (key) => (
        <div className="flex flex-wrap gap-1 max-w-xs">
          {key.scopes.slice(0, 3).map((s) => (
            <Badge key={s} variant="muted" className="text-[10px] font-mono">{s}</Badge>
          ))}
          {key.scopes.length > 3 && (
            <Badge variant="muted" className="text-[10px]">+{key.scopes.length - 3}</Badge>
          )}
        </div>
      ),
    },
    {
      key: 'rate_limit',
      header: 'Rate Limit',
      render: (key) => <span className="text-sm">{key.rate_limit}/min</span>,
    },
    {
      key: 'last_used',
      header: 'Last Used',
      render: (key) => (
        <span className="text-muted-foreground text-xs">
          {key.last_used_at ? formatRelativeTime(key.last_used_at) : 'Never'}
        </span>
      ),
    },
    {
      key: 'expires',
      header: 'Expires',
      render: (key) => (
        <span className="text-muted-foreground text-xs">
          {key.expires_at ? formatDate(key.expires_at) : 'Never'}
        </span>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      render: (key) => (
        <Badge variant={key.is_active ? 'success' : 'destructive'}>
          {key.is_active ? 'Active' : 'Revoked'}
        </Badge>
      ),
    },
    {
      key: 'actions',
      header: '',
      className: 'w-12',
      render: (key) =>
        key.is_active ? (
          <Button
            variant="ghost"
            size="icon-sm"
            onClick={(e) => {
              e.stopPropagation()
              setRevokeKey(key)
            }}
          >
            <Trash2 className="h-3 w-3 text-destructive" />
          </Button>
        ) : null,
    },
  ], [])

  return (
    <div>
      <PageHeader
        title="API Keys"
        description="Manage API keys for programmatic access"
        breadcrumbs={[{ label: 'Dashboard', href: '/' }, { label: 'API Keys' }]}
        actions={
          <Button onClick={() => {
            setCreateOpen(true)
            setFormData({ name: '', scopes: [], rate_limit: 1000, expires_at: '' })
          }}>
            <Plus className="mr-2 h-4 w-4" />
            Create API Key
          </Button>
        }
      />

      {!isLoading && (!keys || keys.length === 0) ? (
        <EmptyState
          icon={<Key className="h-12 w-12" />}
          title="No API keys"
          description="Create an API key for programmatic access to the admin API."
          action={{ label: 'Create API Key', onClick: () => setCreateOpen(true) }}
        />
      ) : (
        <DataTable
          columns={columns}
          data={keys ?? []}
          isLoading={isLoading}
          getRowId={(k) => k.id}
        />
      )}

      {/* Create */}
      <Dialog open={createOpen} onClose={() => setCreateOpen(false)} className="max-w-xl">
        <DialogClose onClose={() => setCreateOpen(false)} />
        <DialogHeader>
          <DialogTitle>Create API Key</DialogTitle>
          <DialogDescription>The full key will only be shown once after creation.</DialogDescription>
        </DialogHeader>
        <div className="mt-4 space-y-4">
          <div className="space-y-2">
            <Label>Name</Label>
            <Input
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              placeholder="Production API Key"
            />
          </div>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label>Rate Limit (req/min)</Label>
              <Input
                type="number"
                value={formData.rate_limit}
                onChange={(e) => setFormData({ ...formData, rate_limit: Number(e.target.value) })}
              />
            </div>
            <div className="space-y-2">
              <Label>Expiry Date (optional)</Label>
              <Input
                type="date"
                value={formData.expires_at}
                onChange={(e) => setFormData({ ...formData, expires_at: e.target.value })}
              />
            </div>
          </div>
          <div className="space-y-2">
            <Label>Scopes</Label>
            <div className="max-h-48 overflow-y-auto rounded-md border p-3 space-y-1.5">
              {availableScopes.map((scope) => (
                <div key={scope} className="flex items-center gap-2">
                  <Checkbox
                    checked={formData.scopes.includes(scope)}
                    onCheckedChange={() => toggleScope(scope)}
                  />
                  <span className="text-xs font-mono">{scope}</span>
                </div>
              ))}
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateOpen(false)}>Cancel</Button>
            <Button
              onClick={() => createMutation.mutate(formData)}
              loading={createMutation.isPending}
              disabled={!formData.name || formData.scopes.length === 0}
            >
              Create Key
            </Button>
          </DialogFooter>
        </div>
      </Dialog>

      {/* New Key Result */}
      <Dialog open={!!newKeyResult} onClose={() => setNewKeyResult(null)}>
        <DialogClose onClose={() => setNewKeyResult(null)} />
        <DialogHeader>
          <DialogTitle>API Key Created</DialogTitle>
        </DialogHeader>
        <div className="mt-4 space-y-4">
          <div className="flex items-start gap-2 rounded-lg border border-warning/50 bg-warning/10 p-3">
            <AlertTriangle className="h-4 w-4 text-warning mt-0.5" />
            <p className="text-sm text-warning">
              Copy this key now. You will not be able to see it again.
            </p>
          </div>
          <div className="flex gap-2">
            <Input value={newKeyResult || ''} readOnly className="font-mono text-xs" />
            <Button
              variant="outline"
              size="icon"
              onClick={() => {
                copyToClipboard(newKeyResult || '')
                addToast({ title: 'Copied to clipboard', variant: 'success' })
              }}
            >
              <Copy className="h-4 w-4" />
            </Button>
          </div>
          <DialogFooter>
            <Button onClick={() => setNewKeyResult(null)}>Done</Button>
          </DialogFooter>
        </div>
      </Dialog>

      <ConfirmDialog
        open={!!revokeKey}
        onClose={() => setRevokeKey(null)}
        onConfirm={() => revokeKey && revokeMutation.mutate(revokeKey.id)}
        title="Revoke API Key"
        description={`Revoke "${revokeKey?.name}"? Any applications using this key will lose access.`}
        confirmLabel="Revoke"
        variant="destructive"
        loading={revokeMutation.isPending}
      />
    </div>
  )
}
