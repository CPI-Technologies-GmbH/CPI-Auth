import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { useUIStore } from '@/stores/ui'
import { PageHeader } from '@/components/shared/page-header'
import { DataTable, type Column } from '@/components/shared/data-table'
import { EmptyState } from '@/components/shared/empty-state'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Dialog,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  DialogClose,
} from '@/components/ui/dialog'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'
import { Plus, Building, Users, Trash2, UserPlus, X } from 'lucide-react'
import { formatDate } from '@/lib/utils'
import type { Organization, OrganizationMember } from '@/types'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Select } from '@/components/ui/select'
import { Avatar } from '@/components/ui/avatar'
import { ConfirmDialog } from '@/components/shared/confirm-dialog'
import { useMemo } from 'react'

const createSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  display_name: z.string().min(1, 'Display name is required'),
})

type CreateForm = z.infer<typeof createSchema>

export default function OrganizationsPage() {
  const queryClient = useQueryClient()
  const { addToast } = useUIStore()
  const [createOpen, setCreateOpen] = useState(false)
  const [selectedOrg, setSelectedOrg] = useState<Organization | null>(null)
  const [detailOpen, setDetailOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [inviteOpen, setInviteOpen] = useState(false)
  const [inviteData, setInviteData] = useState({ user_id: '', role: 'member' })

  const { data: orgs, isLoading } = useQuery({
    queryKey: ['organizations'],
    queryFn: () => api.getOrganizations(),
  })

  const { data: members } = useQuery({
    queryKey: ['org-members', selectedOrg?.id],
    queryFn: () => api.getOrganizationMembers(selectedOrg!.id),
    enabled: !!selectedOrg,
  })

  const createMutation = useMutation({
    mutationFn: (data: CreateForm) => api.createOrganization(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['organizations'] })
      setCreateOpen(false)
      addToast({ title: 'Organization created', variant: 'success' })
    },
    onError: () => addToast({ title: 'Failed to create organization', variant: 'error' }),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.deleteOrganization(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['organizations'] })
      setDeleteOpen(false)
      setDetailOpen(false)
      setSelectedOrg(null)
      addToast({ title: 'Organization deleted', variant: 'success' })
    },
  })

  const addMemberMutation = useMutation({
    mutationFn: (data: { user_id: string; role: string }) =>
      api.addOrganizationMember(selectedOrg!.id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['org-members', selectedOrg?.id] })
      setInviteOpen(false)
      setInviteData({ user_id: '', role: 'member' })
      addToast({ title: 'Member added', variant: 'success' })
    },
  })

  const removeMemberMutation = useMutation({
    mutationFn: (memberId: string) =>
      api.removeOrganizationMember(selectedOrg!.id, memberId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['org-members', selectedOrg?.id] })
      addToast({ title: 'Member removed', variant: 'success' })
    },
  })

  const columns: Column<Organization>[] = useMemo(() => [
    {
      key: 'name',
      header: 'Organization',
      sortable: true,
      render: (org) => (
        <div className="flex items-center gap-3">
          <div className="rounded-lg bg-primary/10 p-2">
            <Building className="h-4 w-4 text-primary" />
          </div>
          <div>
            <p className="font-medium">{org.display_name}</p>
            <p className="text-xs text-muted-foreground">{org.name}</p>
          </div>
        </div>
      ),
    },
    {
      key: 'members',
      header: 'Members',
      render: (org) => (
        <div className="flex items-center gap-1">
          <Users className="h-3 w-3 text-muted-foreground" />
          <span>{org.member_count}</span>
        </div>
      ),
    },
    {
      key: 'created',
      header: 'Created',
      sortable: true,
      render: (org) => <span className="text-muted-foreground">{formatDate(org.created_at)}</span>,
    },
  ], [])

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<CreateForm>({ resolver: zodResolver(createSchema) })

  return (
    <div>
      <PageHeader
        title="Organizations"
        description="Manage organizations and their members"
        breadcrumbs={[{ label: 'Dashboard', href: '/' }, { label: 'Organizations' }]}
        actions={
          <Button onClick={() => { setCreateOpen(true); reset() }}>
            <Plus className="mr-2 h-4 w-4" />
            Create Organization
          </Button>
        }
      />

      {!isLoading && (!orgs || orgs.length === 0) ? (
        <EmptyState
          icon={<Building className="h-12 w-12" />}
          title="No organizations yet"
          description="Create your first organization to group users together."
          action={{ label: 'Create Organization', onClick: () => setCreateOpen(true) }}
        />
      ) : (
        <DataTable
          columns={columns}
          data={orgs ?? []}
          isLoading={isLoading}
          getRowId={(o) => o.id}
          onRowClick={(org) => {
            setSelectedOrg(org)
            setDetailOpen(true)
          }}
          emptyMessage="No organizations found"
        />
      )}

      {/* Create Dialog */}
      <Dialog open={createOpen} onClose={() => setCreateOpen(false)}>
        <DialogClose onClose={() => setCreateOpen(false)} />
        <DialogHeader>
          <DialogTitle>Create Organization</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit((d) => createMutation.mutate(d))} className="mt-4 space-y-4">
          <div className="space-y-2">
            <Label>Name (identifier)</Label>
            <Input {...register('name')} placeholder="my-org" error={errors.name?.message} />
          </div>
          <div className="space-y-2">
            <Label>Display Name</Label>
            <Input {...register('display_name')} placeholder="My Organization" error={errors.display_name?.message} />
          </div>
          <DialogFooter>
            <Button variant="outline" type="button" onClick={() => setCreateOpen(false)}>Cancel</Button>
            <Button type="submit" loading={createMutation.isPending}>Create</Button>
          </DialogFooter>
        </form>
      </Dialog>

      {/* Detail Dialog */}
      <Dialog open={detailOpen} onClose={() => setDetailOpen(false)} className="max-w-2xl">
        <DialogClose onClose={() => setDetailOpen(false)} />
        {selectedOrg && (
          <>
            <DialogHeader>
              <DialogTitle>{selectedOrg.display_name}</DialogTitle>
            </DialogHeader>

            <Tabs defaultValue="members" className="mt-4">
              <TabsList>
                <TabsTrigger value="members">Members</TabsTrigger>
                <TabsTrigger value="settings">Settings</TabsTrigger>
              </TabsList>

              <TabsContent value="members">
                <div className="flex items-center justify-between mb-4">
                  <p className="text-sm text-muted-foreground">{members?.length || 0} members</p>
                  <Button size="sm" onClick={() => setInviteOpen(true)}>
                    <UserPlus className="mr-1 h-3 w-3" />
                    Add Member
                  </Button>
                </div>
                <div className="space-y-2 max-h-80 overflow-y-auto">
                  {members?.map((member) => (
                    <div key={member.id} className="flex items-center justify-between rounded-lg border p-3">
                      <div className="flex items-center gap-3">
                        <Avatar name={member.user_name} size="sm" />
                        <div>
                          <p className="text-sm font-medium">{member.user_name}</p>
                          <p className="text-xs text-muted-foreground">{member.user_email}</p>
                        </div>
                      </div>
                      <div className="flex items-center gap-2">
                        <Badge variant="secondary">{member.role}</Badge>
                        <Button
                          variant="ghost"
                          size="icon-sm"
                          onClick={() => removeMemberMutation.mutate(member.id)}
                        >
                          <X className="h-3 w-3 text-destructive" />
                        </Button>
                      </div>
                    </div>
                  ))}
                  {(!members || members.length === 0) && (
                    <p className="text-sm text-muted-foreground text-center py-4">No members</p>
                  )}
                </div>
              </TabsContent>

              <TabsContent value="settings">
                <Card>
                  <CardHeader>
                    <CardTitle className="text-sm">Danger Zone</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <Button
                      variant="destructive"
                      onClick={() => setDeleteOpen(true)}
                    >
                      <Trash2 className="mr-2 h-4 w-4" />
                      Delete Organization
                    </Button>
                  </CardContent>
                </Card>
              </TabsContent>
            </Tabs>
          </>
        )}
      </Dialog>

      {/* Invite Dialog */}
      <Dialog open={inviteOpen} onClose={() => setInviteOpen(false)}>
        <DialogClose onClose={() => setInviteOpen(false)} />
        <DialogHeader>
          <DialogTitle>Add Member</DialogTitle>
        </DialogHeader>
        <div className="mt-4 space-y-4">
          <div className="space-y-2">
            <Label>User ID</Label>
            <Input
              value={inviteData.user_id}
              onChange={(e) => setInviteData({ ...inviteData, user_id: e.target.value })}
              placeholder="Enter user ID"
            />
          </div>
          <div className="space-y-2">
            <Label>Role</Label>
            <Select
              value={inviteData.role}
              onChange={(e) => setInviteData({ ...inviteData, role: e.target.value })}
              options={[
                { value: 'member', label: 'Member' },
                { value: 'admin', label: 'Admin' },
                { value: 'owner', label: 'Owner' },
              ]}
            />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setInviteOpen(false)}>Cancel</Button>
            <Button
              onClick={() => addMemberMutation.mutate(inviteData)}
              loading={addMemberMutation.isPending}
            >
              Add Member
            </Button>
          </DialogFooter>
        </div>
      </Dialog>

      <ConfirmDialog
        open={deleteOpen}
        onClose={() => setDeleteOpen(false)}
        onConfirm={() => selectedOrg && deleteMutation.mutate(selectedOrg.id)}
        title="Delete Organization"
        description={`Delete "${selectedOrg?.display_name}"? All members will be removed.`}
        confirmLabel="Delete"
        variant="destructive"
        loading={deleteMutation.isPending}
      />
    </div>
  )
}
