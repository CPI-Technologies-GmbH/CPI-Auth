import { useState, useMemo } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { useUIStore } from '@/stores/ui'
import { PageHeader } from '@/components/shared/page-header'
import { EmptyState } from '@/components/shared/empty-state'
import { CodeEditor } from '@/components/shared/code-editor'
import { ConfirmDialog } from '@/components/shared/confirm-dialog'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import {
  Dialog,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  DialogClose,
} from '@/components/ui/dialog'
import { Plus, Zap, GripVertical, Trash2, Edit, ChevronUp, ChevronDown } from 'lucide-react'
import type { Action, ActionTrigger } from '@/types'

const triggerLabels: Record<ActionTrigger, string> = {
  'pre-login': 'Pre Login',
  'post-login': 'Post Login',
  'pre-register': 'Pre Register',
  'post-register': 'Post Register',
  'pre-token': 'Pre Token Issue',
  'post-change-password': 'Post Change Password',
  'pre-user-update': 'Pre User Update',
}

const defaultCode = `/**
 * Handler that will be called during the execution of an action.
 *
 * @param {Event} event - Details about the context of the action.
 * @param {API} api - Interface for manipulating the action flow.
 */
exports.onExecute = async (event, api) => {
  // Add your logic here
  console.log('Action triggered:', event.type);
};
`

const triggerOptions = Object.entries(triggerLabels).map(([value, label]) => ({ value, label }))

export default function ActionsPage() {
  const queryClient = useQueryClient()
  const { addToast } = useUIStore()

  const [createOpen, setCreateOpen] = useState(false)
  const [editAction, setEditAction] = useState<Action | null>(null)
  const [deleteAction, setDeleteAction] = useState<Action | null>(null)
  const [formData, setFormData] = useState({
    name: '',
    trigger: 'post-login' as ActionTrigger,
    code: defaultCode,
  })

  const { data: actions, isLoading } = useQuery({
    queryKey: ['actions'],
    queryFn: () => api.getActions(),
  })

  const createMutation = useMutation({
    mutationFn: (data: typeof formData) => api.createAction(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['actions'] })
      setCreateOpen(false)
      addToast({ title: 'Action created', variant: 'success' })
    },
    onError: () => addToast({ title: 'Failed to create action', variant: 'error' }),
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<Action> }) =>
      api.updateAction(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['actions'] })
      setEditAction(null)
      addToast({ title: 'Action updated', variant: 'success' })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.deleteAction(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['actions'] })
      setDeleteAction(null)
      addToast({ title: 'Action deleted', variant: 'success' })
    },
  })

  const toggleMutation = useMutation({
    mutationFn: ({ id, is_active }: { id: string; is_active: boolean }) =>
      api.updateAction(id, { is_active }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['actions'] }),
  })

  const reorderMutation = useMutation({
    mutationFn: ({ trigger, ids }: { trigger: string; ids: string[] }) =>
      api.reorderActions(trigger, ids),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['actions'] }),
  })

  const actionsByTrigger = useMemo(() => {
    if (!actions) return {}
    return actions.reduce<Record<string, Action[]>>((acc, action) => {
      if (!acc[action.trigger]) acc[action.trigger] = []
      acc[action.trigger].push(action)
      return acc
    }, {})
  }, [actions])

  const moveAction = (trigger: string, index: number, direction: 'up' | 'down') => {
    const group = [...(actionsByTrigger[trigger] || [])]
    const newIndex = direction === 'up' ? index - 1 : index + 1
    if (newIndex < 0 || newIndex >= group.length) return
    ;[group[index], group[newIndex]] = [group[newIndex], group[index]]
    reorderMutation.mutate({ trigger, ids: group.map((a) => a.id) })
  }

  return (
    <div>
      <PageHeader
        title="Actions"
        description="Custom code that runs during authentication flows"
        breadcrumbs={[{ label: 'Dashboard', href: '/' }, { label: 'Actions' }]}
        actions={
          <Button onClick={() => {
            setCreateOpen(true)
            setFormData({ name: '', trigger: 'post-login', code: defaultCode })
          }}>
            <Plus className="mr-2 h-4 w-4" />
            Create Action
          </Button>
        }
      />

      {!isLoading && (!actions || actions.length === 0) ? (
        <EmptyState
          icon={<Zap className="h-12 w-12" />}
          title="No actions configured"
          description="Create actions to add custom logic to your authentication flows."
          action={{ label: 'Create Action', onClick: () => setCreateOpen(true) }}
        />
      ) : (
        <div className="space-y-6">
          {Object.entries(triggerLabels).map(([trigger, label]) => {
            const group = actionsByTrigger[trigger] || []
            if (group.length === 0) return null

            return (
              <div key={trigger}>
                <h2 className="text-sm font-medium text-muted-foreground mb-2 flex items-center gap-2">
                  <Zap className="h-3 w-3" />
                  {label}
                </h2>
                <div className="space-y-2">
                  {group.map((action, idx) => (
                    <Card key={action.id}>
                      <CardContent className="p-3 flex items-center justify-between">
                        <div className="flex items-center gap-3">
                          <div className="flex flex-col gap-0.5">
                            <button
                              onClick={() => moveAction(trigger, idx, 'up')}
                              disabled={idx === 0}
                              className="text-muted-foreground hover:text-foreground disabled:opacity-30 cursor-pointer"
                            >
                              <ChevronUp className="h-3 w-3" />
                            </button>
                            <button
                              onClick={() => moveAction(trigger, idx, 'down')}
                              disabled={idx === group.length - 1}
                              className="text-muted-foreground hover:text-foreground disabled:opacity-30 cursor-pointer"
                            >
                              <ChevronDown className="h-3 w-3" />
                            </button>
                          </div>
                          <GripVertical className="h-4 w-4 text-muted-foreground" />
                          <div>
                            <p className="font-medium text-sm">{action.name}</p>
                            <p className="text-xs text-muted-foreground">
                              Order: {action.order} | Runtime: {action.runtime} | Timeout: {action.timeout_ms}ms
                            </p>
                          </div>
                        </div>
                        <div className="flex items-center gap-2">
                          <Switch
                            checked={action.is_active}
                            onCheckedChange={(checked) =>
                              toggleMutation.mutate({ id: action.id, is_active: checked })
                            }
                          />
                          <Button
                            variant="ghost"
                            size="icon-sm"
                            onClick={() => {
                              setEditAction(action)
                              setFormData({
                                name: action.name,
                                trigger: action.trigger,
                                code: action.code,
                              })
                            }}
                          >
                            <Edit className="h-3 w-3" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="icon-sm"
                            onClick={() => setDeleteAction(action)}
                          >
                            <Trash2 className="h-3 w-3 text-destructive" />
                          </Button>
                        </div>
                      </CardContent>
                    </Card>
                  ))}
                </div>
              </div>
            )
          })}
        </div>
      )}

      {/* Create/Edit */}
      <Dialog
        open={createOpen || !!editAction}
        onClose={() => { setCreateOpen(false); setEditAction(null) }}
        className="max-w-3xl"
      >
        <DialogClose onClose={() => { setCreateOpen(false); setEditAction(null) }} />
        <DialogHeader>
          <DialogTitle>{editAction ? 'Edit Action' : 'Create Action'}</DialogTitle>
        </DialogHeader>
        <div className="mt-4 space-y-4">
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label>Name</Label>
              <Input
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                placeholder="My Action"
              />
            </div>
            <div className="space-y-2">
              <Label>Trigger</Label>
              <Select
                value={formData.trigger}
                onChange={(e) => setFormData({ ...formData, trigger: e.target.value as ActionTrigger })}
                options={triggerOptions}
              />
            </div>
          </div>
          <div className="space-y-2">
            <Label>Code</Label>
            <CodeEditor
              value={formData.code}
              onChange={(v) => setFormData({ ...formData, code: v })}
              language="javascript"
              height="300px"
            />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => { setCreateOpen(false); setEditAction(null) }}>
              Cancel
            </Button>
            <Button
              onClick={() => {
                if (editAction) {
                  updateMutation.mutate({ id: editAction.id, data: formData })
                } else {
                  createMutation.mutate(formData)
                }
              }}
              loading={createMutation.isPending || updateMutation.isPending}
            >
              {editAction ? 'Save' : 'Create'}
            </Button>
          </DialogFooter>
        </div>
      </Dialog>

      <ConfirmDialog
        open={!!deleteAction}
        onClose={() => setDeleteAction(null)}
        onConfirm={() => deleteAction && deleteMutation.mutate(deleteAction.id)}
        title="Delete Action"
        description={`Delete "${deleteAction?.name}"? This cannot be undone.`}
        confirmLabel="Delete"
        variant="destructive"
        loading={deleteMutation.isPending}
      />
    </div>
  )
}
