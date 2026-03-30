import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { useUIStore } from '@/stores/ui'
import { PageHeader } from '@/components/shared/page-header'
import { ConfirmDialog } from '@/components/shared/confirm-dialog'
import { JsonEditor } from '@/components/shared/json-editor'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select } from '@/components/ui/select'
import { Avatar } from '@/components/ui/avatar'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Dialog,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  DialogClose,
} from '@/components/ui/dialog'
import {
  Ban,
  Trash2,
  KeyRound,
  LogOut,
  Shield,
  Smartphone,
  Link2,
  History,
  Monitor,
  Save,
  CheckCircle,
  Plus,
  X,
  UserCheck,
} from 'lucide-react'
import { formatDate, formatDateTime, formatRelativeTime, copyToClipboard } from '@/lib/utils'
import type { User, Role, Application } from '@/types'

export default function UserDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { addToast } = useUIStore()

  const [editedUser, setEditedUser] = useState<Partial<User>>({})
  const [confirmAction, setConfirmAction] = useState<'delete' | 'block' | 'reset-password' | 'force-logout' | null>(null)
  const [assignRoleOpen, setAssignRoleOpen] = useState(false)
  const [selectedRoleId, setSelectedRoleId] = useState('')
  const [impersonateOpen, setImpersonateOpen] = useState(false)
  const [impersonateAppId, setImpersonateAppId] = useState('')
  const [impersonateRedirect, setImpersonateRedirect] = useState('')
  const [impersonateResult, setImpersonateResult] = useState<{ access_token: string; redirect_url: string; expires_in: number } | null>(null)

  const { data: user, isLoading } = useQuery({
    queryKey: ['user', id],
    queryFn: () => api.getUser(id!),
    enabled: !!id,
  })

  const { data: sessions } = useQuery({
    queryKey: ['user-sessions', id],
    queryFn: () => api.getUserSessions(id!),
    enabled: !!id,
  })

  const { data: mfaEnrollments } = useQuery({
    queryKey: ['user-mfa', id],
    queryFn: () => api.getUserMfaEnrollments(id!),
    enabled: !!id,
  })

  const { data: identities } = useQuery({
    queryKey: ['user-identities', id],
    queryFn: () => api.getUserIdentities(id!),
    enabled: !!id,
  })

  const { data: auditLog } = useQuery({
    queryKey: ['user-audit', id],
    queryFn: () => api.getUserAuditLog(id!),
    enabled: !!id,
  })

  const { data: roles } = useQuery({
    queryKey: ['user-roles', id],
    queryFn: () => api.getUserRoles(id!),
    enabled: !!id,
  })

  const { data: allRoles } = useQuery({
    queryKey: ['roles'],
    queryFn: () => api.getRoles(),
  })

  const updateMutation = useMutation({
    mutationFn: (data: Partial<User>) => api.updateUser(id!, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['user', id] })
      addToast({ title: 'User updated', variant: 'success' })
    },
    onError: () => addToast({ title: 'Update failed', variant: 'error' }),
  })

  const deleteMutation = useMutation({
    mutationFn: () => api.deleteUser(id!),
    onSuccess: () => {
      navigate('/users')
      addToast({ title: 'User deleted', variant: 'success' })
    },
  })

  const blockMutation = useMutation({
    mutationFn: () =>
      user?.status === 'blocked' ? api.unblockUser(id!) : api.blockUser(id!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['user', id] })
      setConfirmAction(null)
      addToast({ title: user?.status === 'blocked' ? 'User unblocked' : 'User blocked', variant: 'success' })
    },
  })

  const resetPasswordMutation = useMutation({
    mutationFn: () => api.resetUserPassword(id!),
    onSuccess: () => {
      setConfirmAction(null)
      addToast({ title: 'Password reset email sent', variant: 'success' })
    },
  })

  const forceLogoutMutation = useMutation({
    mutationFn: () => api.forceLogout(id!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['user-sessions', id] })
      setConfirmAction(null)
      addToast({ title: 'User logged out from all sessions', variant: 'success' })
    },
  })

  const revokeSessionMutation = useMutation({
    mutationFn: (sessionId: string) => api.revokeUserSession(id!, sessionId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['user-sessions', id] })
      addToast({ title: 'Session revoked', variant: 'success' })
    },
  })

  const assignRoleMutation = useMutation({
    mutationFn: (roleId: string) => api.assignUserRole(id!, roleId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['user-roles', id] })
      setAssignRoleOpen(false)
      setSelectedRoleId('')
      addToast({ title: 'Role assigned', variant: 'success' })
    },
    onError: () => addToast({ title: 'Failed to assign role', variant: 'error' }),
  })

  const removeRoleMutation = useMutation({
    mutationFn: (roleId: string) => api.removeUserRole(id!, roleId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['user-roles', id] })
      addToast({ title: 'Role removed', variant: 'success' })
    },
    onError: () => addToast({ title: 'Failed to remove role', variant: 'error' }),
  })

  const { data: applications } = useQuery({
    queryKey: ['applications'],
    queryFn: () => api.getApplications(),
  })

  const impersonateMutation = useMutation({
    mutationFn: (appId?: string) => api.impersonateUser(id!, appId),
    onSuccess: (data) => {
      setImpersonateResult({
        access_token: data.access_token,
        redirect_url: data.redirect_url || impersonateRedirect,
        expires_in: data.expires_in,
      })
    },
    onError: () => addToast({ title: 'Failed to impersonate user', variant: 'error' }),
  })

  if (isLoading) {
    return (
      <div>
        <Skeleton className="h-8 w-48 mb-2" />
        <Skeleton className="h-4 w-64 mb-6" />
        <div className="grid gap-6 lg:grid-cols-3">
          <Skeleton className="h-64" />
          <Skeleton className="h-64 lg:col-span-2" />
        </div>
      </div>
    )
  }

  if (!user) return null

  const statusVariant: Record<string, 'success' | 'destructive' | 'warning' | 'muted'> = {
    active: 'success',
    blocked: 'destructive',
    inactive: 'warning',
    pending: 'muted',
  }

  return (
    <div>
      <PageHeader
        title={user.name}
        description={user.email}
        breadcrumbs={[
          { label: 'Dashboard', href: '/' },
          { label: 'Users', href: '/users' },
          { label: user.name },
        ]}
        actions={
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => {
                setImpersonateAppId('')
                setImpersonateRedirect('')
                setImpersonateResult(null)
                setImpersonateOpen(true)
              }}
            >
              <UserCheck className="mr-1 h-3 w-3" />
              Impersonate
            </Button>
            <Button variant="outline" size="sm" onClick={() => setConfirmAction('reset-password')}>
              <KeyRound className="mr-1 h-3 w-3" />
              Reset Password
            </Button>
            <Button variant="outline" size="sm" onClick={() => setConfirmAction('force-logout')}>
              <LogOut className="mr-1 h-3 w-3" />
              Force Logout
            </Button>
            <Button
              variant={user.status === 'blocked' ? 'success' : 'outline'}
              size="sm"
              onClick={() => setConfirmAction('block')}
            >
              <Ban className="mr-1 h-3 w-3" />
              {user.status === 'blocked' ? 'Unblock' : 'Block'}
            </Button>
            <Button variant="destructive" size="sm" onClick={() => setConfirmAction('delete')}>
              <Trash2 className="mr-1 h-3 w-3" />
              Delete
            </Button>
          </div>
        }
      />

      <div className="grid gap-6 lg:grid-cols-3">
        {/* Profile sidebar */}
        <Card>
          <CardContent className="p-6">
            <div className="flex flex-col items-center text-center">
              <Avatar name={user.name} src={user.avatar_url} size="lg" className="h-20 w-20 text-xl" />
              <h2 className="mt-3 text-lg font-semibold">{user.name}</h2>
              <p className="text-sm text-muted-foreground">{user.email}</p>
              <Badge variant={statusVariant[user.status]} className="mt-2">{user.status}</Badge>
            </div>

            <div className="mt-6 space-y-3 text-sm">
              <div className="flex justify-between">
                <span className="text-muted-foreground">User ID</span>
                <span className="font-mono text-xs">{user.id.slice(0, 12)}...</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Email Verified</span>
                <span>{user.email_verified ? <CheckCircle className="h-4 w-4 text-success" /> : 'No'}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Login Count</span>
                <span>{user.login_count}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Last Login</span>
                <span>{user.last_login ? formatRelativeTime(user.last_login) : 'Never'}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Created</span>
                <span>{formatDate(user.created_at)}</span>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Detail tabs */}
        <div className="lg:col-span-2">
          <Tabs defaultValue="profile">
            <TabsList className="w-full justify-start flex-wrap">
              <TabsTrigger value="profile">Profile</TabsTrigger>
              <TabsTrigger value="roles">Roles</TabsTrigger>
              <TabsTrigger value="sessions">Sessions</TabsTrigger>
              <TabsTrigger value="mfa">MFA</TabsTrigger>
              <TabsTrigger value="identities">Identities</TabsTrigger>
              <TabsTrigger value="audit">Audit Log</TabsTrigger>
              <TabsTrigger value="metadata">Metadata</TabsTrigger>
            </TabsList>

            <TabsContent value="profile">
              <Card>
                <CardHeader>
                  <CardTitle>Edit Profile</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="grid gap-4 sm:grid-cols-2">
                    <div className="space-y-2">
                      <Label>Name</Label>
                      <Input
                        defaultValue={user.name}
                        onChange={(e) => setEditedUser({ ...editedUser, name: e.target.value })}
                      />
                    </div>
                    <div className="space-y-2">
                      <Label>Email</Label>
                      <Input
                        defaultValue={user.email}
                        onChange={(e) => setEditedUser({ ...editedUser, email: e.target.value })}
                      />
                    </div>
                    <div className="space-y-2">
                      <Label>Phone</Label>
                      <Input
                        defaultValue={user.phone || ''}
                        onChange={(e) => setEditedUser({ ...editedUser, phone: e.target.value })}
                      />
                    </div>
                    <div className="space-y-2">
                      <Label>Status</Label>
                      <Select
                        defaultValue={user.status}
                        onChange={(e) => setEditedUser({ ...editedUser, status: e.target.value as User['status'] })}
                        options={[
                          { value: 'active', label: 'Active' },
                          { value: 'blocked', label: 'Blocked' },
                          { value: 'inactive', label: 'Inactive' },
                        ]}
                      />
                    </div>
                  </div>
                  <Button
                    onClick={() => updateMutation.mutate(editedUser)}
                    loading={updateMutation.isPending}
                    disabled={Object.keys(editedUser).length === 0}
                  >
                    <Save className="mr-2 h-4 w-4" />
                    Save Changes
                  </Button>
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="roles">
              <Card>
                <CardHeader>
                  <div className="flex items-center justify-between">
                    <CardTitle className="flex items-center gap-2">
                      <Shield className="h-4 w-4" />
                      Role Assignments
                    </CardTitle>
                    <Button size="sm" onClick={() => setAssignRoleOpen(true)}>
                      <Plus className="mr-1 h-3 w-3" />
                      Assign Role
                    </Button>
                  </div>
                </CardHeader>
                <CardContent>
                  {roles && roles.length > 0 ? (
                    <div className="space-y-2">
                      {roles.map((role) => (
                        <div
                          key={role.id}
                          className="flex items-center justify-between rounded-lg border p-3"
                        >
                          <div>
                            <p className="font-medium text-sm">{role.name}</p>
                            <p className="text-xs text-muted-foreground">{role.description}</p>
                          </div>
                          <div className="flex items-center gap-2">
                            {role.is_system && <Badge variant="muted">System</Badge>}
                            <Badge variant="secondary">{role.permissions.length} permissions</Badge>
                            <Button
                              variant="ghost"
                              size="icon-sm"
                              onClick={() => removeRoleMutation.mutate(role.id)}
                              loading={removeRoleMutation.isPending}
                            >
                              <X className="h-3 w-3 text-destructive" />
                            </Button>
                          </div>
                        </div>
                      ))}
                    </div>
                  ) : (
                    <p className="text-sm text-muted-foreground">No roles assigned</p>
                  )}
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="sessions">
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <Monitor className="h-4 w-4" />
                    Active Sessions
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  {sessions && sessions.length > 0 ? (
                    <div className="space-y-3">
                      {sessions.map((session) => (
                        <div
                          key={session.id}
                          className="flex items-center justify-between rounded-lg border p-3"
                        >
                          <div>
                            <p className="text-sm font-medium">{session.ip_address}</p>
                            <p className="text-xs text-muted-foreground truncate max-w-md">
                              {session.user_agent}
                            </p>
                            <p className="text-xs text-muted-foreground mt-1">
                              Last active: {formatRelativeTime(session.last_active)}
                            </p>
                          </div>
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => revokeSessionMutation.mutate(session.id)}
                            loading={revokeSessionMutation.isPending}
                          >
                            Revoke
                          </Button>
                        </div>
                      ))}
                    </div>
                  ) : (
                    <p className="text-sm text-muted-foreground">No active sessions</p>
                  )}
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="mfa">
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <Smartphone className="h-4 w-4" />
                    MFA Enrollments
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  {mfaEnrollments && mfaEnrollments.length > 0 ? (
                    <div className="space-y-2">
                      {mfaEnrollments.map((enrollment) => (
                        <div
                          key={enrollment.id}
                          className="flex items-center justify-between rounded-lg border p-3"
                        >
                          <div className="flex items-center gap-3">
                            <Badge variant="secondary">{enrollment.method.toUpperCase()}</Badge>
                            <span className="text-sm">
                              Enrolled {formatDate(enrollment.created_at)}
                            </span>
                          </div>
                          <Badge variant={enrollment.status === 'confirmed' ? 'success' : 'warning'}>
                            {enrollment.status}
                          </Badge>
                        </div>
                      ))}
                    </div>
                  ) : (
                    <p className="text-sm text-muted-foreground">No MFA enrollments</p>
                  )}
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="identities">
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <Link2 className="h-4 w-4" />
                    Linked Identities
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  {identities && identities.length > 0 ? (
                    <div className="space-y-2">
                      {identities.map((identity, idx) => (
                        <div
                          key={idx}
                          className="flex items-center justify-between rounded-lg border p-3"
                        >
                          <div>
                            <p className="font-medium text-sm capitalize">{identity.provider}</p>
                            <p className="text-xs text-muted-foreground">
                              {identity.email || identity.provider_id}
                            </p>
                          </div>
                          <span className="text-xs text-muted-foreground">
                            Linked {formatDate(identity.linked_at)}
                          </span>
                        </div>
                      ))}
                    </div>
                  ) : (
                    <p className="text-sm text-muted-foreground">No linked identities</p>
                  )}
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="audit">
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <History className="h-4 w-4" />
                    Audit Trail
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  {auditLog && auditLog.length > 0 ? (
                    <div className="space-y-3">
                      {auditLog.map((entry) => (
                        <div
                          key={entry.id}
                          className="flex items-start gap-3 border-b pb-3 last:border-0"
                        >
                          <div className="h-2 w-2 mt-2 rounded-full bg-primary shrink-0" />
                          <div className="flex-1">
                            <p className="text-sm font-medium">{entry.action}</p>
                            <p className="text-xs text-muted-foreground">
                              {entry.ip_address} - {formatDateTime(entry.created_at)}
                            </p>
                          </div>
                        </div>
                      ))}
                    </div>
                  ) : (
                    <p className="text-sm text-muted-foreground">No audit log entries</p>
                  )}
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="metadata">
              <Card>
                <CardHeader>
                  <CardTitle>User Metadata</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div>
                    <Label className="mb-2 block">User Metadata</Label>
                    <JsonEditor
                      value={user.metadata || {}}
                      onChange={(val) => setEditedUser({ ...editedUser, metadata: val })}
                    />
                  </div>
                  <div>
                    <Label className="mb-2 block">App Metadata</Label>
                    <JsonEditor
                      value={user.app_metadata || {}}
                      onChange={(val) => setEditedUser({ ...editedUser, app_metadata: val })}
                    />
                  </div>
                  <Button
                    onClick={() => updateMutation.mutate(editedUser)}
                    loading={updateMutation.isPending}
                  >
                    <Save className="mr-2 h-4 w-4" />
                    Save Metadata
                  </Button>
                </CardContent>
              </Card>
            </TabsContent>
          </Tabs>
        </div>
      </div>

      {/* Assign Role Dialog */}
      <Dialog
        open={assignRoleOpen}
        onClose={() => { setAssignRoleOpen(false); setSelectedRoleId('') }}
      >
        <DialogClose onClose={() => { setAssignRoleOpen(false); setSelectedRoleId('') }} />
        <DialogHeader>
          <DialogTitle>Assign Role</DialogTitle>
        </DialogHeader>
        <div className="mt-4 space-y-4">
          <div className="space-y-2">
            <Label>Select Role</Label>
            <Select
              value={selectedRoleId}
              onChange={(e) => setSelectedRoleId(e.target.value)}
              placeholder="Choose a role..."
              options={
                (allRoles ?? [])
                  .filter((r) => !(roles ?? []).some((assigned) => assigned.id === r.id))
                  .map((r) => ({ value: r.id, label: r.name }))
              }
            />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => { setAssignRoleOpen(false); setSelectedRoleId('') }}>
              Cancel
            </Button>
            <Button
              onClick={() => assignRoleMutation.mutate(selectedRoleId)}
              loading={assignRoleMutation.isPending}
              disabled={!selectedRoleId}
            >
              Assign Role
            </Button>
          </DialogFooter>
        </div>
      </Dialog>

      <ConfirmDialog
        open={confirmAction === 'delete'}
        onClose={() => setConfirmAction(null)}
        onConfirm={() => deleteMutation.mutate()}
        title="Delete User"
        description={`Are you sure you want to permanently delete ${user.name}? This action cannot be undone.`}
        confirmLabel="Delete"
        variant="destructive"
        loading={deleteMutation.isPending}
      />

      <ConfirmDialog
        open={confirmAction === 'block'}
        onClose={() => setConfirmAction(null)}
        onConfirm={() => blockMutation.mutate()}
        title={user.status === 'blocked' ? 'Unblock User' : 'Block User'}
        description={
          user.status === 'blocked'
            ? `Unblock ${user.name}? They will be able to log in again.`
            : `Block ${user.name}? They will be logged out and unable to log in.`
        }
        confirmLabel={user.status === 'blocked' ? 'Unblock' : 'Block'}
        variant={user.status === 'blocked' ? 'default' : 'destructive'}
        loading={blockMutation.isPending}
      />

      <ConfirmDialog
        open={confirmAction === 'reset-password'}
        onClose={() => setConfirmAction(null)}
        onConfirm={() => resetPasswordMutation.mutate()}
        title="Reset Password"
        description={`Send a password reset email to ${user.email}?`}
        confirmLabel="Send Reset Email"
        loading={resetPasswordMutation.isPending}
      />

      <ConfirmDialog
        open={confirmAction === 'force-logout'}
        onClose={() => setConfirmAction(null)}
        onConfirm={() => forceLogoutMutation.mutate()}
        title="Force Logout"
        description={`Force ${user.name} to log out from all active sessions?`}
        confirmLabel="Force Logout"
        variant="destructive"
        loading={forceLogoutMutation.isPending}
      />

      {/* ─── Impersonate Dialog ──────────────────────────── */}
      <Dialog open={impersonateOpen} onClose={() => setImpersonateOpen(false)} className="max-w-lg">
        <DialogClose onClose={() => setImpersonateOpen(false)} />
        <DialogHeader>
          <DialogTitle>Impersonate {user.name}</DialogTitle>
        </DialogHeader>
        <div className="mt-4 space-y-4">
          <p className="text-sm text-muted-foreground">
            Sign in as <strong>{user.email}</strong> to debug their experience. A short-lived token (15 min) will be issued with the selected application&apos;s permissions.
          </p>

          {!impersonateResult ? (
            <>
              {/* Application selector */}
              <div className="space-y-2">
                <Label>Application</Label>
                <Select
                  value={impersonateAppId}
                  onChange={(e) => {
                    setImpersonateAppId(e.target.value)
                    setImpersonateRedirect('')
                    // Auto-select first redirect URI
                    const app = (applications ?? []).find((a: Application) => a.id === e.target.value)
                    if (app?.redirect_uris?.length === 1) {
                      setImpersonateRedirect(app.redirect_uris[0])
                    }
                  }}
                  options={[
                    { value: '', label: 'No application (all permissions)' },
                    ...((applications ?? []) as Application[]).map((a) => ({
                      value: a.id,
                      label: `${a.name} (${a.type})`,
                    })),
                  ]}
                />
              </div>

              {/* Redirect URI selector (if app has multiple) */}
              {impersonateAppId && (() => {
                const app = (applications ?? []).find((a: Application) => a.id === impersonateAppId)
                if (!app?.redirect_uris?.length) return null
                if (app.redirect_uris.length === 1) {
                  return (
                    <div className="space-y-2">
                      <Label>Redirect URL</Label>
                      <Input value={app.redirect_uris[0]} readOnly className="text-xs font-mono" />
                    </div>
                  )
                }
                return (
                  <div className="space-y-2">
                    <Label>Redirect URL</Label>
                    <Select
                      value={impersonateRedirect}
                      onChange={(e) => setImpersonateRedirect(e.target.value)}
                      options={app.redirect_uris.map((uri: string) => ({ value: uri, label: uri }))}
                    />
                  </div>
                )
              })()}

              <DialogFooter>
                <Button variant="outline" onClick={() => setImpersonateOpen(false)}>Cancel</Button>
                <Button
                  onClick={() => impersonateMutation.mutate(impersonateAppId || undefined)}
                  loading={impersonateMutation.isPending}
                >
                  <UserCheck className="mr-2 h-4 w-4" />
                  Generate Token
                </Button>
              </DialogFooter>
            </>
          ) : (
            <>
              {/* Token result */}
              <div className="rounded-md border border-green-500/20 bg-green-500/5 p-3">
                <p className="text-sm text-green-400 font-medium mb-1">Impersonation token generated</p>
                <p className="text-xs text-muted-foreground">Expires in {Math.round(impersonateResult.expires_in / 60)} minutes</p>
              </div>

              <div className="space-y-2">
                <Label>Access Token</Label>
                <div className="flex gap-2">
                  <Input
                    value={impersonateResult.access_token}
                    readOnly
                    className="text-xs font-mono"
                  />
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      copyToClipboard(impersonateResult!.access_token)
                      addToast({ title: 'Token copied', variant: 'success' })
                    }}
                  >
                    Copy
                  </Button>
                </div>
              </div>

              <DialogFooter>
                <Button variant="outline" onClick={() => setImpersonateOpen(false)}>Close</Button>
                {impersonateResult.redirect_url && (
                  <Button
                    onClick={() => {
                      const sep = impersonateResult!.redirect_url.includes('?') ? '&' : '?'
                      const url = `${impersonateResult!.redirect_url}${sep}access_token=${impersonateResult!.access_token}&token_type=bearer&impersonated=true`
                      window.open(url, '_blank')
                    }}
                  >
                    <Monitor className="mr-2 h-4 w-4" />
                    Open as User
                  </Button>
                )}
              </DialogFooter>
            </>
          )}
        </div>
      </Dialog>
    </div>
  )
}
