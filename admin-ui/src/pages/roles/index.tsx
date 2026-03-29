import { useState, useMemo } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { useUIStore } from '@/stores/ui'
import { PageHeader } from '@/components/shared/page-header'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Checkbox } from '@/components/ui/checkbox'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Dialog,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  DialogClose,
} from '@/components/ui/dialog'
import { ConfirmDialog } from '@/components/shared/confirm-dialog'
import { Plus, Shield, ChevronRight, ChevronDown, Trash2, Edit } from 'lucide-react'
import type { Role, Permission } from '@/types'

export default function RolesPage() {
  const queryClient = useQueryClient()
  const { addToast } = useUIStore()

  const [createOpen, setCreateOpen] = useState(false)
  const [editRole, setEditRole] = useState<Role | null>(null)
  const [deleteRole, setDeleteRole] = useState<Role | null>(null)
  const [expandedRoles, setExpandedRoles] = useState<Set<string>>(new Set())
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    permissions: [] as string[],
  })

  // Permission CRUD state
  const [createPermOpen, setCreatePermOpen] = useState(false)
  const [editPerm, setEditPerm] = useState<Permission | null>(null)
  const [deletePerm, setDeletePerm] = useState<Permission | null>(null)
  const [permFormData, setPermFormData] = useState({
    name: '',
    display_name: '',
    description: '',
    group_name: '',
  })

  const { data: roles, isLoading: rolesLoading } = useQuery({
    queryKey: ['roles'],
    queryFn: () => api.getRoles(),
  })

  const { data: permissions } = useQuery({
    queryKey: ['permissions'],
    queryFn: () => api.getPermissions(),
  })

  const createMutation = useMutation({
    mutationFn: (data: { name: string; description: string; permissions: string[] }) =>
      api.createRole(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['roles'] })
      setCreateOpen(false)
      resetForm()
      addToast({ title: 'Role created', variant: 'success' })
    },
    onError: () => addToast({ title: 'Failed to create role', variant: 'error' }),
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<Role> }) =>
      api.updateRole(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['roles'] })
      setEditRole(null)
      addToast({ title: 'Role updated', variant: 'success' })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.deleteRole(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['roles'] })
      setDeleteRole(null)
      addToast({ title: 'Role deleted', variant: 'success' })
    },
  })

  const createPermMutation = useMutation({
    mutationFn: (data: { name: string; display_name: string; description?: string; group_name: string }) =>
      api.createPermission(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['permissions'] })
      setCreatePermOpen(false)
      resetPermForm()
      addToast({ title: 'Permission created', variant: 'success' })
    },
    onError: () => addToast({ title: 'Failed to create permission', variant: 'error' }),
  })

  const updatePermMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<Permission> }) =>
      api.updatePermission(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['permissions'] })
      setEditPerm(null)
      addToast({ title: 'Permission updated', variant: 'success' })
    },
    onError: () => addToast({ title: 'Failed to update permission', variant: 'error' }),
  })

  const deletePermMutation = useMutation({
    mutationFn: (id: string) => api.deletePermission(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['permissions'] })
      setDeletePerm(null)
      addToast({ title: 'Permission deleted', variant: 'success' })
    },
    onError: () => addToast({ title: 'Failed to delete permission', variant: 'error' }),
  })

  const resetForm = () => {
    setFormData({ name: '', description: '', permissions: [] })
  }

  const resetPermForm = () => {
    setPermFormData({ name: '', display_name: '', description: '', group_name: '' })
  }

  const permissionsByGroup = useMemo(() => {
    if (!permissions) return {}
    return permissions.reduce<Record<string, Permission[]>>((acc, perm) => {
      if (!acc[perm.group]) acc[perm.group] = []
      acc[perm.group].push(perm)
      return acc
    }, {})
  }, [permissions])

  const togglePermission = (permName: string) => {
    setFormData((prev) => ({
      ...prev,
      permissions: prev.permissions.includes(permName)
        ? prev.permissions.filter((p) => p !== permName)
        : [...prev.permissions, permName],
    }))
  }

  const toggleExpanded = (roleId: string) => {
    const next = new Set(expandedRoles)
    if (next.has(roleId)) next.delete(roleId)
    else next.add(roleId)
    setExpandedRoles(next)
  }

  return (
    <div>
      <PageHeader
        title="Roles & Permissions"
        description="Manage role-based access control"
        breadcrumbs={[{ label: 'Dashboard', href: '/' }, { label: 'Roles & Permissions' }]}
        actions={
          <Button onClick={() => { setCreateOpen(true); resetForm() }}>
            <Plus className="mr-2 h-4 w-4" />
            Create Role
          </Button>
        }
      />

      <div className="grid gap-6 lg:grid-cols-2">
        {/* Role Tree */}
        <div>
          <h2 className="text-sm font-medium text-muted-foreground mb-3">Roles</h2>
          {rolesLoading ? (
            <div className="space-y-2">
              {Array.from({ length: 4 }).map((_, i) => (
                <Skeleton key={i} className="h-16" />
              ))}
            </div>
          ) : (
            <div className="space-y-2">
              {(roles ?? []).map((role) => (
                <Card key={role.id}>
                  <CardContent className="p-3">
                    <div
                      className="flex items-center justify-between cursor-pointer"
                      onClick={() => toggleExpanded(role.id)}
                    >
                      <div className="flex items-center gap-3">
                        {expandedRoles.has(role.id) ? (
                          <ChevronDown className="h-4 w-4 text-muted-foreground" />
                        ) : (
                          <ChevronRight className="h-4 w-4 text-muted-foreground" />
                        )}
                        <Shield className="h-4 w-4 text-primary" />
                        <div>
                          <p className="font-medium text-sm">{role.name}</p>
                          {role.description && (
                            <p className="text-xs text-muted-foreground">{role.description}</p>
                          )}
                        </div>
                      </div>
                      <div className="flex items-center gap-2">
                        {role.is_system && <Badge variant="muted">System</Badge>}
                        <Badge variant="secondary">{role.permissions.length} perms</Badge>
                        {!role.is_system && (
                          <>
                            <Button
                              variant="ghost"
                              size="icon-sm"
                              onClick={(e) => {
                                e.stopPropagation()
                                setEditRole(role)
                                setFormData({
                                  name: role.name,
                                  description: role.description || '',
                                  permissions: [...role.permissions],
                                })
                              }}
                            >
                              <Edit className="h-3 w-3" />
                            </Button>
                            <Button
                              variant="ghost"
                              size="icon-sm"
                              onClick={(e) => {
                                e.stopPropagation()
                                setDeleteRole(role)
                              }}
                            >
                              <Trash2 className="h-3 w-3 text-destructive" />
                            </Button>
                          </>
                        )}
                      </div>
                    </div>

                    {expandedRoles.has(role.id) && (
                      <div className="mt-3 ml-8 flex flex-wrap gap-1">
                        {role.permissions.map((perm) => (
                          <Badge key={perm} variant="outline" className="text-[10px] font-mono">
                            {perm}
                          </Badge>
                        ))}
                        {role.permissions.length === 0 && (
                          <span className="text-xs text-muted-foreground">No permissions assigned</span>
                        )}
                      </div>
                    )}
                  </CardContent>
                </Card>
              ))}
            </div>
          )}
        </div>

        {/* Permissions List */}
        <div>
          <div className="flex items-center justify-between mb-3">
            <h2 className="text-sm font-medium text-muted-foreground">All Permissions</h2>
            <Button size="sm" variant="outline" onClick={() => { setCreatePermOpen(true); resetPermForm() }}>
              <Plus className="mr-1 h-3 w-3" />
              Create Permission
            </Button>
          </div>
          <Card>
            <CardContent className="p-4">
              {Object.entries(permissionsByGroup).map(([group, perms]) => (
                <div key={group} className="mb-4 last:mb-0">
                  <h3 className="text-sm font-semibold mb-2 capitalize">{group}</h3>
                  <div className="space-y-1.5">
                    {perms.map((perm) => (
                      <div key={perm.id} className="flex items-center justify-between rounded-md p-1.5 hover:bg-muted/50">
                        <div className="flex items-start gap-2 min-w-0">
                          <Badge variant="outline" className="text-[10px] font-mono shrink-0 mt-0.5">
                            {perm.name}
                          </Badge>
                          {perm.description && (
                            <span className="text-xs text-muted-foreground truncate">{perm.description}</span>
                          )}
                        </div>
                        <div className="flex items-center gap-1 shrink-0 ml-2">
                          {perm.is_system ? (
                            <Badge variant="muted" className="text-[10px]">System</Badge>
                          ) : (
                            <>
                              <Button
                                variant="ghost"
                                size="icon-sm"
                                onClick={() => {
                                  setEditPerm(perm)
                                  setPermFormData({
                                    name: perm.name,
                                    display_name: perm.display_name || '',
                                    description: perm.description || '',
                                    group_name: perm.group,
                                  })
                                }}
                              >
                                <Edit className="h-3 w-3" />
                              </Button>
                              <Button
                                variant="ghost"
                                size="icon-sm"
                                onClick={() => setDeletePerm(perm)}
                              >
                                <Trash2 className="h-3 w-3 text-destructive" />
                              </Button>
                            </>
                          )}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              ))}
              {Object.keys(permissionsByGroup).length === 0 && (
                <p className="text-sm text-muted-foreground text-center py-4">No permissions defined</p>
              )}
            </CardContent>
          </Card>
        </div>
      </div>

      {/* Create/Edit Role Dialog */}
      <Dialog
        open={createOpen || editRole !== null}
        onClose={() => {
          setCreateOpen(false)
          setEditRole(null)
        }}
        className="max-w-xl"
      >
        <DialogClose onClose={() => { setCreateOpen(false); setEditRole(null) }} />
        <DialogHeader>
          <DialogTitle>{editRole ? 'Edit Role' : 'Create Role'}</DialogTitle>
        </DialogHeader>
        <div className="mt-4 space-y-4">
          <div className="space-y-2">
            <Label>Role Name</Label>
            <Input
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              placeholder="e.g. Editor"
            />
          </div>
          <div className="space-y-2">
            <Label>Description</Label>
            <Textarea
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              placeholder="What this role can do..."
              className="h-20"
            />
          </div>
          <div className="space-y-2">
            <Label>Permissions</Label>
            <div className="max-h-60 overflow-y-auto rounded-md border p-3 space-y-3">
              {Object.entries(permissionsByGroup).map(([group, perms]) => (
                <div key={group}>
                  <p className="text-xs font-semibold text-muted-foreground mb-1.5 uppercase">{group}</p>
                  <div className="space-y-1">
                    {perms.map((perm) => (
                      <div key={perm.id} className="flex items-center gap-2">
                        <Checkbox
                          checked={formData.permissions.includes(perm.name)}
                          onCheckedChange={() => togglePermission(perm.name)}
                        />
                        <span className="text-xs font-mono">{perm.name}</span>
                      </div>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => { setCreateOpen(false); setEditRole(null) }}>
              Cancel
            </Button>
            <Button
              onClick={() => {
                if (editRole) {
                  updateMutation.mutate({ id: editRole.id, data: formData })
                } else {
                  createMutation.mutate(formData)
                }
              }}
              loading={createMutation.isPending || updateMutation.isPending}
            >
              {editRole ? 'Save Changes' : 'Create Role'}
            </Button>
          </DialogFooter>
        </div>
      </Dialog>

      <ConfirmDialog
        open={deleteRole !== null}
        onClose={() => setDeleteRole(null)}
        onConfirm={() => deleteRole && deleteMutation.mutate(deleteRole.id)}
        title="Delete Role"
        description={`Delete the "${deleteRole?.name}" role? Users assigned this role will lose its permissions.`}
        confirmLabel="Delete Role"
        variant="destructive"
        loading={deleteMutation.isPending}
      />

      {/* Create/Edit Permission Dialog */}
      <Dialog
        open={createPermOpen || editPerm !== null}
        onClose={() => { setCreatePermOpen(false); setEditPerm(null) }}
      >
        <DialogClose onClose={() => { setCreatePermOpen(false); setEditPerm(null) }} />
        <DialogHeader>
          <DialogTitle>{editPerm ? 'Edit Permission' : 'Create Permission'}</DialogTitle>
        </DialogHeader>
        <div className="mt-4 space-y-4">
          <div className="space-y-2">
            <Label>Permission Name</Label>
            <Input
              value={permFormData.name}
              onChange={(e) => setPermFormData({ ...permFormData, name: e.target.value })}
              placeholder="e.g. billing:manage"
              disabled={editPerm !== null}
            />
          </div>
          <div className="space-y-2">
            <Label>Display Name</Label>
            <Input
              value={permFormData.display_name}
              onChange={(e) => setPermFormData({ ...permFormData, display_name: e.target.value })}
              placeholder="e.g. Manage Billing"
            />
          </div>
          <div className="space-y-2">
            <Label>Group</Label>
            <Input
              value={permFormData.group_name}
              onChange={(e) => setPermFormData({ ...permFormData, group_name: e.target.value })}
              placeholder="e.g. Billing"
            />
          </div>
          <div className="space-y-2">
            <Label>Description</Label>
            <Textarea
              value={permFormData.description}
              onChange={(e) => setPermFormData({ ...permFormData, description: e.target.value })}
              placeholder="What this permission allows..."
              className="h-20"
            />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => { setCreatePermOpen(false); setEditPerm(null) }}>
              Cancel
            </Button>
            <Button
              onClick={() => {
                if (editPerm) {
                  updatePermMutation.mutate({
                    id: editPerm.id,
                    data: {
                      display_name: permFormData.display_name,
                      description: permFormData.description,
                      group: permFormData.group_name,
                    },
                  })
                } else {
                  createPermMutation.mutate(permFormData)
                }
              }}
              loading={createPermMutation.isPending || updatePermMutation.isPending}
              disabled={!permFormData.name || !permFormData.display_name || !permFormData.group_name}
            >
              {editPerm ? 'Save Changes' : 'Create Permission'}
            </Button>
          </DialogFooter>
        </div>
      </Dialog>

      <ConfirmDialog
        open={deletePerm !== null}
        onClose={() => setDeletePerm(null)}
        onConfirm={() => deletePerm && deletePermMutation.mutate(deletePerm.id)}
        title="Delete Permission"
        description={`Delete the "${deletePerm?.name}" permission? It will be removed from all roles that use it.`}
        confirmLabel="Delete Permission"
        variant="destructive"
        loading={deletePermMutation.isPending}
      />
    </div>
  )
}
