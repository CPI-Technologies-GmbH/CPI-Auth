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
import { Switch } from '@/components/ui/switch'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Dialog,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
  DialogClose,
} from '@/components/ui/dialog'
import { Plus, Webhook, Play, Trash2, Eye, CheckCircle, XCircle } from 'lucide-react'
import { formatRelativeTime, formatDateTime } from '@/lib/utils'
import type { Webhook as WebhookType, WebhookEvent, WebhookDelivery } from '@/types'

const allEvents: { group: string; events: WebhookEvent[] }[] = [
  {
    group: 'User',
    events: ['user.created', 'user.updated', 'user.deleted', 'user.login', 'user.logout', 'user.blocked', 'user.password_changed', 'user.mfa_enrolled'],
  },
  {
    group: 'Application',
    events: ['application.created', 'application.updated'],
  },
  {
    group: 'Organization',
    events: ['organization.created', 'organization.member_added'],
  },
  {
    group: 'Tenant',
    events: ['tenant.created'],
  },
]

export default function WebhooksPage() {
  const queryClient = useQueryClient()
  const { addToast } = useUIStore()

  const [createOpen, setCreateOpen] = useState(false)
  const [detailWebhook, setDetailWebhook] = useState<WebhookType | null>(null)
  const [deleteWebhook, setDeleteWebhook] = useState<WebhookType | null>(null)
  const [formData, setFormData] = useState({
    name: '',
    url: '',
    events: [] as WebhookEvent[],
    secret: '',
  })

  const { data: webhooks, isLoading } = useQuery({
    queryKey: ['webhooks'],
    queryFn: () => api.getWebhooks(),
  })

  const { data: deliveries } = useQuery({
    queryKey: ['webhook-deliveries', detailWebhook?.id],
    queryFn: () => api.getWebhookDeliveries(detailWebhook!.id),
    enabled: !!detailWebhook,
  })

  const createMutation = useMutation({
    mutationFn: (data: typeof formData) => api.createWebhook(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['webhooks'] })
      setCreateOpen(false)
      resetForm()
      addToast({ title: 'Webhook created', variant: 'success' })
    },
    onError: () => addToast({ title: 'Failed to create webhook', variant: 'error' }),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.deleteWebhook(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['webhooks'] })
      setDeleteWebhook(null)
      addToast({ title: 'Webhook deleted', variant: 'success' })
    },
  })

  const toggleMutation = useMutation({
    mutationFn: ({ id, is_active }: { id: string; is_active: boolean }) =>
      api.updateWebhook(id, { is_active }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['webhooks'] })
    },
  })

  const testMutation = useMutation({
    mutationFn: (id: string) => api.testWebhook(id),
    onSuccess: () => {
      addToast({ title: 'Test webhook sent', variant: 'success' })
      if (detailWebhook) {
        queryClient.invalidateQueries({ queryKey: ['webhook-deliveries', detailWebhook.id] })
      }
    },
    onError: () => addToast({ title: 'Test webhook failed', variant: 'error' }),
  })

  const resetForm = () => {
    setFormData({ name: '', url: '', events: [], secret: '' })
  }

  const toggleEvent = (event: WebhookEvent) => {
    setFormData((prev) => ({
      ...prev,
      events: prev.events.includes(event)
        ? prev.events.filter((e) => e !== event)
        : [...prev.events, event],
    }))
  }

  const columns: Column<WebhookType>[] = useMemo(() => [
    {
      key: 'name',
      header: 'Webhook',
      render: (wh) => (
        <div>
          <p className="font-medium">{wh.name}</p>
          <p className="text-xs text-muted-foreground font-mono truncate max-w-xs">{wh.url}</p>
        </div>
      ),
    },
    {
      key: 'events',
      header: 'Events',
      render: (wh) => (
        <div className="flex flex-wrap gap-1 max-w-xs">
          {wh.events.slice(0, 3).map((e) => (
            <Badge key={e} variant="muted" className="text-[10px]">{e}</Badge>
          ))}
          {wh.events.length > 3 && (
            <Badge variant="muted" className="text-[10px]">+{wh.events.length - 3}</Badge>
          )}
        </div>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      render: (wh) => (
        <Switch
          checked={wh.is_active}
          onCheckedChange={(checked) => toggleMutation.mutate({ id: wh.id, is_active: checked })}
        />
      ),
    },
    {
      key: 'last_triggered',
      header: 'Last Triggered',
      render: (wh) => (
        <span className="text-muted-foreground text-xs">
          {wh.last_triggered ? formatRelativeTime(wh.last_triggered) : 'Never'}
        </span>
      ),
    },
    {
      key: 'actions',
      header: '',
      className: 'w-28',
      render: (wh) => (
        <div className="flex gap-1" onClick={(e) => e.stopPropagation()}>
          <Button variant="ghost" size="icon-sm" onClick={() => setDetailWebhook(wh)}>
            <Eye className="h-3 w-3" />
          </Button>
          <Button variant="ghost" size="icon-sm" onClick={() => testMutation.mutate(wh.id)}>
            <Play className="h-3 w-3" />
          </Button>
          <Button variant="ghost" size="icon-sm" onClick={() => setDeleteWebhook(wh)}>
            <Trash2 className="h-3 w-3 text-destructive" />
          </Button>
        </div>
      ),
    },
  ], [toggleMutation, testMutation])

  return (
    <div>
      <PageHeader
        title="Webhooks"
        description="Configure webhook endpoints for event notifications"
        breadcrumbs={[{ label: 'Dashboard', href: '/' }, { label: 'Webhooks' }]}
        actions={
          <Button onClick={() => { setCreateOpen(true); resetForm() }}>
            <Plus className="mr-2 h-4 w-4" />
            Create Webhook
          </Button>
        }
      />

      {!isLoading && (!webhooks || webhooks.length === 0) ? (
        <EmptyState
          icon={<Webhook className="h-12 w-12" />}
          title="No webhooks configured"
          description="Create a webhook to receive real-time event notifications."
          action={{ label: 'Create Webhook', onClick: () => setCreateOpen(true) }}
        />
      ) : (
        <DataTable
          columns={columns}
          data={webhooks ?? []}
          isLoading={isLoading}
          getRowId={(w) => w.id}
        />
      )}

      {/* Create */}
      <Dialog open={createOpen} onClose={() => setCreateOpen(false)} className="max-w-xl">
        <DialogClose onClose={() => setCreateOpen(false)} />
        <DialogHeader>
          <DialogTitle>Create Webhook</DialogTitle>
          <DialogDescription>Configure a new webhook endpoint</DialogDescription>
        </DialogHeader>
        <div className="mt-4 space-y-4">
          <div className="space-y-2">
            <Label>Name</Label>
            <Input
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              placeholder="My Webhook"
            />
          </div>
          <div className="space-y-2">
            <Label>URL</Label>
            <Input
              value={formData.url}
              onChange={(e) => setFormData({ ...formData, url: e.target.value })}
              placeholder="https://example.com/webhook"
            />
          </div>
          <div className="space-y-2">
            <Label>Secret</Label>
            <Input
              value={formData.secret}
              onChange={(e) => setFormData({ ...formData, secret: e.target.value })}
              placeholder="whsec_..."
            />
          </div>
          <div className="space-y-2">
            <Label>Events</Label>
            <div className="max-h-48 overflow-y-auto rounded-md border p-3 space-y-3">
              {allEvents.map((group) => (
                <div key={group.group}>
                  <p className="text-xs font-semibold text-muted-foreground mb-1">{group.group}</p>
                  <div className="space-y-1">
                    {group.events.map((event) => (
                      <div key={event} className="flex items-center gap-2">
                        <Checkbox
                          checked={formData.events.includes(event)}
                          onCheckedChange={() => toggleEvent(event)}
                        />
                        <span className="text-xs font-mono">{event}</span>
                      </div>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateOpen(false)}>Cancel</Button>
            <Button
              onClick={() => createMutation.mutate(formData)}
              loading={createMutation.isPending}
              disabled={!formData.name || !formData.url || formData.events.length === 0}
            >
              Create
            </Button>
          </DialogFooter>
        </div>
      </Dialog>

      {/* Detail - Delivery History */}
      <Dialog open={!!detailWebhook} onClose={() => setDetailWebhook(null)} className="max-w-2xl">
        <DialogClose onClose={() => setDetailWebhook(null)} />
        {detailWebhook && (
          <>
            <DialogHeader>
              <DialogTitle>{detailWebhook.name}</DialogTitle>
              <DialogDescription className="font-mono text-xs">{detailWebhook.url}</DialogDescription>
            </DialogHeader>
            <div className="mt-4">
              <div className="flex items-center justify-between mb-3">
                <h3 className="text-sm font-medium">Delivery History</h3>
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => testMutation.mutate(detailWebhook.id)}
                  loading={testMutation.isPending}
                >
                  <Play className="mr-1 h-3 w-3" />
                  Send Test
                </Button>
              </div>
              <div className="space-y-2 max-h-80 overflow-y-auto">
                {deliveries?.map((d) => (
                  <div key={d.id} className="flex items-center justify-between rounded-lg border p-3">
                    <div className="flex items-center gap-3">
                      {d.success ? (
                        <CheckCircle className="h-4 w-4 text-success" />
                      ) : (
                        <XCircle className="h-4 w-4 text-destructive" />
                      )}
                      <div>
                        <p className="text-sm font-medium">{d.event}</p>
                        <p className="text-xs text-muted-foreground">
                          {formatDateTime(d.created_at)} - {d.duration_ms}ms
                        </p>
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      <Badge variant={d.success ? 'success' : 'destructive'}>{d.status_code}</Badge>
                      <Badge variant="muted">Attempt {d.attempts}</Badge>
                    </div>
                  </div>
                ))}
                {(!deliveries || deliveries.length === 0) && (
                  <p className="text-sm text-muted-foreground text-center py-6">No deliveries yet</p>
                )}
              </div>
            </div>
          </>
        )}
      </Dialog>

      <ConfirmDialog
        open={!!deleteWebhook}
        onClose={() => setDeleteWebhook(null)}
        onConfirm={() => deleteWebhook && deleteMutation.mutate(deleteWebhook.id)}
        title="Delete Webhook"
        description={`Delete "${deleteWebhook?.name}"? This cannot be undone.`}
        confirmLabel="Delete"
        variant="destructive"
        loading={deleteMutation.isPending}
      />
    </div>
  )
}
