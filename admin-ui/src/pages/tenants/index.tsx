import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { useUIStore } from '@/stores/ui'
import { PageHeader } from '@/components/shared/page-header'
import { EmptyState } from '@/components/shared/empty-state'
import { ConfirmDialog } from '@/components/shared/confirm-dialog'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Dialog,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  DialogClose,
} from '@/components/ui/dialog'
import { Plus, Building2, Globe, Trash2, Settings, ChevronRight } from 'lucide-react'
import { formatDate } from '@/lib/utils'
import type { Tenant } from '@/types'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'

const createSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  slug: z.string().min(1, 'Slug is required').regex(/^[a-z0-9-]+$/, 'Lowercase letters, numbers, hyphens only'),
  domain: z.string().optional(),
})

type CreateForm = z.infer<typeof createSchema>

function TenantTree({ tenants, parentId, level = 0 }: { tenants: Tenant[]; parentId?: string; level?: number }) {
  const children = tenants.filter((t) => t.parent_id === parentId)
  if (children.length === 0) return null

  return (
    <div className={level > 0 ? 'ml-6 border-l pl-4' : ''}>
      {children.map((tenant) => (
        <div key={tenant.id} className="mb-3">
          <Card className="hover:border-primary/50 transition-colors cursor-pointer">
            <CardContent className="p-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div className="rounded-lg bg-primary/10 p-2">
                    <Building2 className="h-4 w-4 text-primary" />
                  </div>
                  <div>
                    <h3 className="font-semibold">{tenant.name}</h3>
                    <p className="text-xs text-muted-foreground">{tenant.slug}</p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  {tenant.domain && (
                    <Badge variant="muted" className="text-[10px]">
                      <Globe className="mr-1 h-3 w-3" />
                      {tenant.domain}
                    </Badge>
                  )}
                  <Badge variant={tenant.is_active ? 'success' : 'muted'}>
                    {tenant.is_active ? 'Active' : 'Disabled'}
                  </Badge>
                  <span className="text-xs text-muted-foreground">{formatDate(tenant.created_at)}</span>
                  <ChevronRight className="h-4 w-4 text-muted-foreground" />
                </div>
              </div>
            </CardContent>
          </Card>
          <TenantTree tenants={tenants} parentId={tenant.id} level={level + 1} />
        </div>
      ))}
    </div>
  )
}

export default function TenantsPage() {
  const queryClient = useQueryClient()
  const { addToast } = useUIStore()
  const [createOpen, setCreateOpen] = useState(false)
  const [selectedTenant, setSelectedTenant] = useState<Tenant | null>(null)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [settingsOpen, setSettingsOpen] = useState(false)

  const { data: tenants, isLoading } = useQuery({
    queryKey: ['tenants'],
    queryFn: () => api.getTenants(),
  })

  const createMutation = useMutation({
    mutationFn: (data: CreateForm) => api.createTenant(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tenants'] })
      setCreateOpen(false)
      addToast({ title: 'Tenant created', variant: 'success' })
    },
    onError: () => addToast({ title: 'Failed to create tenant', variant: 'error' }),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.deleteTenant(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tenants'] })
      setDeleteOpen(false)
      setSelectedTenant(null)
      addToast({ title: 'Tenant deleted', variant: 'success' })
    },
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<Tenant> }) => api.updateTenant(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tenants'] })
      setSettingsOpen(false)
      addToast({ title: 'Tenant updated', variant: 'success' })
    },
  })

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<CreateForm>({ resolver: zodResolver(createSchema) })

  const [editData, setEditData] = useState<Partial<Tenant>>({})

  return (
    <div>
      <PageHeader
        title="Tenants"
        description="Manage tenants and their configurations"
        breadcrumbs={[{ label: 'Dashboard', href: '/' }, { label: 'Tenants' }]}
        actions={
          <Button onClick={() => { setCreateOpen(true); reset() }}>
            <Plus className="mr-2 h-4 w-4" />
            Create Tenant
          </Button>
        }
      />

      {isLoading ? (
        <div className="space-y-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} className="h-20" />
          ))}
        </div>
      ) : !tenants || tenants.length === 0 ? (
        <EmptyState
          icon={<Building2 className="h-12 w-12" />}
          title="No tenants yet"
          description="Create your first tenant to start organizing your platform."
          action={{ label: 'Create Tenant', onClick: () => setCreateOpen(true) }}
        />
      ) : (
        <div>
          {tenants.map((tenant) => (
            <div key={tenant.id} className="mb-3">
              <Card className="hover:border-primary/50 transition-colors">
                <CardContent className="p-4">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <div className="rounded-lg bg-primary/10 p-2">
                        <Building2 className="h-4 w-4 text-primary" />
                      </div>
                      <div>
                        <h3 className="font-semibold">{tenant.name}</h3>
                        <p className="text-xs text-muted-foreground">{tenant.slug}</p>
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      {tenant.domain && (
                        <Badge variant="muted" className="text-[10px]">
                          <Globe className="mr-1 h-3 w-3" />
                          {tenant.domain}
                        </Badge>
                      )}
                      {tenant.custom_domain && (
                        <Badge variant="secondary" className="text-[10px]">
                          Custom: {tenant.custom_domain}
                        </Badge>
                      )}
                      <Badge variant={tenant.is_active ? 'success' : 'muted'}>
                        {tenant.is_active ? 'Active' : 'Disabled'}
                      </Badge>
                      <Button
                        variant="ghost"
                        size="icon-sm"
                        onClick={() => {
                          setSelectedTenant(tenant)
                          setEditData({ name: tenant.name, domain: tenant.domain, custom_domain: tenant.custom_domain })
                          setSettingsOpen(true)
                        }}
                      >
                        <Settings className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon-sm"
                        onClick={() => {
                          setSelectedTenant(tenant)
                          setDeleteOpen(true)
                        }}
                      >
                        <Trash2 className="h-4 w-4 text-destructive" />
                      </Button>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>
          ))}
        </div>
      )}

      {/* Create */}
      <Dialog open={createOpen} onClose={() => setCreateOpen(false)}>
        <DialogClose onClose={() => setCreateOpen(false)} />
        <DialogHeader>
          <DialogTitle>Create Tenant</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit((d) => createMutation.mutate(d))} className="mt-4 space-y-4">
          <div className="space-y-2">
            <Label>Name</Label>
            <Input {...register('name')} error={errors.name?.message} />
          </div>
          <div className="space-y-2">
            <Label>Slug</Label>
            <Input {...register('slug')} placeholder="my-tenant" error={errors.slug?.message} />
          </div>
          <div className="space-y-2">
            <Label>Domain (optional)</Label>
            <Input {...register('domain')} placeholder="tenant.example.com" />
          </div>
          <DialogFooter>
            <Button variant="outline" type="button" onClick={() => setCreateOpen(false)}>Cancel</Button>
            <Button type="submit" loading={createMutation.isPending}>Create</Button>
          </DialogFooter>
        </form>
      </Dialog>

      {/* Settings */}
      <Dialog open={settingsOpen} onClose={() => setSettingsOpen(false)}>
        <DialogClose onClose={() => setSettingsOpen(false)} />
        <DialogHeader>
          <DialogTitle>Tenant Settings: {selectedTenant?.name}</DialogTitle>
        </DialogHeader>
        <div className="mt-4 space-y-4">
          <div className="space-y-2">
            <Label>Name</Label>
            <Input
              value={editData.name || ''}
              onChange={(e) => setEditData({ ...editData, name: e.target.value })}
            />
          </div>
          <div className="space-y-2">
            <Label>Domain</Label>
            <Input
              value={editData.domain || ''}
              onChange={(e) => setEditData({ ...editData, domain: e.target.value })}
            />
          </div>
          <div className="space-y-2">
            <Label>Custom Domain</Label>
            <Input
              value={editData.custom_domain || ''}
              onChange={(e) => setEditData({ ...editData, custom_domain: e.target.value })}
            />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setSettingsOpen(false)}>Cancel</Button>
            <Button
              onClick={() => selectedTenant && updateMutation.mutate({ id: selectedTenant.id, data: editData })}
              loading={updateMutation.isPending}
            >
              Save
            </Button>
          </DialogFooter>
        </div>
      </Dialog>

      {/* Delete */}
      <ConfirmDialog
        open={deleteOpen}
        onClose={() => setDeleteOpen(false)}
        onConfirm={() => selectedTenant && deleteMutation.mutate(selectedTenant.id)}
        title="Delete Tenant"
        description={`Delete "${selectedTenant?.name}"? All associated data will be permanently removed.`}
        confirmLabel="Delete Tenant"
        variant="destructive"
        loading={deleteMutation.isPending}
      />
    </div>
  )
}
