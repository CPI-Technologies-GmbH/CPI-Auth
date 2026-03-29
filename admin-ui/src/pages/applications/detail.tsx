import { useState, useMemo } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { useUIStore } from '@/stores/ui'
import { PageHeader } from '@/components/shared/page-header'
import { ConfirmDialog } from '@/components/shared/confirm-dialog'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Switch } from '@/components/ui/switch'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'
import { Skeleton } from '@/components/ui/skeleton'
import { Checkbox } from '@/components/ui/checkbox'
import { Copy, Eye, EyeOff, RefreshCw, Trash2, Save, Plus, X, Info } from 'lucide-react'
import { copyToClipboard } from '@/lib/utils'
import type { Application, GrantType, Permission } from '@/types'

function TagInput({
  value,
  onChange,
  placeholder,
}: {
  value: string[]
  onChange: (val: string[]) => void
  placeholder?: string
}) {
  const [input, setInput] = useState('')

  const addTag = () => {
    const trimmed = input.trim()
    if (trimmed && !value.includes(trimmed)) {
      onChange([...value, trimmed])
      setInput('')
    }
  }

  return (
    <div className="space-y-2">
      <div className="flex flex-wrap gap-1.5">
        {value.map((tag, i) => (
          <span
            key={i}
            className="inline-flex items-center gap-1 rounded-md bg-secondary px-2 py-1 text-xs font-mono"
          >
            {tag}
            <button onClick={() => onChange(value.filter((_, idx) => idx !== i))} className="hover:text-destructive cursor-pointer">
              <X className="h-3 w-3" />
            </button>
          </span>
        ))}
      </div>
      <div className="flex gap-2">
        <Input
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === 'Enter') {
              e.preventDefault()
              addTag()
            }
          }}
          placeholder={placeholder}
          className="flex-1"
        />
        <Button type="button" variant="outline" size="sm" onClick={addTag}>
          <Plus className="h-3 w-3" />
        </Button>
      </div>
    </div>
  )
}

export default function ApplicationDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { addToast } = useUIStore()

  const [showSecret, setShowSecret] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [rotateOpen, setRotateOpen] = useState(false)
  const [formData, setFormData] = useState<Partial<Application>>({})

  const { data: app, isLoading } = useQuery({
    queryKey: ['application', id],
    queryFn: () => api.getApplication(id!),
    enabled: !!id,
  })

  const updateMutation = useMutation({
    mutationFn: (data: Partial<Application>) => api.updateApplication(id!, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['application', id] })
      addToast({ title: 'Application updated', variant: 'success' })
    },
    onError: () => addToast({ title: 'Update failed', variant: 'error' }),
  })

  const deleteMutation = useMutation({
    mutationFn: () => api.deleteApplication(id!),
    onSuccess: () => {
      navigate('/applications')
      addToast({ title: 'Application deleted', variant: 'success' })
    },
  })

  const rotateMutation = useMutation({
    mutationFn: () => api.rotateClientSecret(id!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['application', id] })
      setRotateOpen(false)
      addToast({ title: 'Client secret rotated', variant: 'success' })
    },
  })

  // Permissions tab
  const { data: allPermissions } = useQuery({
    queryKey: ['permissions'],
    queryFn: () => api.getPermissions(),
  })

  const { data: appPermissions } = useQuery({
    queryKey: ['app-permissions', id],
    queryFn: () => api.getApplicationPermissions(id!),
    enabled: !!id,
  })

  const [selectedPerms, setSelectedPerms] = useState<string[] | null>(null)

  const currentPerms = selectedPerms ?? appPermissions?.permissions ?? []

  const permsByGroup = useMemo(() => {
    if (!allPermissions) return {}
    return allPermissions.reduce<Record<string, Permission[]>>((acc, perm) => {
      if (!acc[perm.group]) acc[perm.group] = []
      acc[perm.group].push(perm)
      return acc
    }, {})
  }, [allPermissions])

  const togglePerm = (permName: string) => {
    const base = selectedPerms ?? appPermissions?.permissions ?? []
    if (base.includes(permName)) {
      setSelectedPerms(base.filter((p) => p !== permName))
    } else {
      setSelectedPerms([...base, permName])
    }
  }

  const savePermsMutation = useMutation({
    mutationFn: (perms: string[]) => api.setApplicationPermissions(id!, perms),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['app-permissions', id] })
      setSelectedPerms(null)
      addToast({ title: 'Application permissions updated', variant: 'success' })
    },
    onError: () => addToast({ title: 'Failed to update permissions', variant: 'error' }),
  })

  if (isLoading) {
    return (
      <div>
        <Skeleton className="h-8 w-48 mb-6" />
        <Skeleton className="h-96 w-full" />
      </div>
    )
  }

  if (!app) return null

  const allGrantTypes: { value: GrantType; label: string }[] = [
    { value: 'authorization_code', label: 'Authorization Code' },
    { value: 'client_credentials', label: 'Client Credentials' },
    { value: 'refresh_token', label: 'Refresh Token' },
    { value: 'implicit', label: 'Implicit (Legacy)' },
    { value: 'password', label: 'Password (Legacy)' },
  ]

  return (
    <div>
      <PageHeader
        title={app.name}
        description={app.description || `${app.type.toUpperCase()} Application`}
        breadcrumbs={[
          { label: 'Dashboard', href: '/' },
          { label: 'Applications', href: '/applications' },
          { label: app.name },
        ]}
        actions={
          <div className="flex gap-2">
            <Button variant="destructive" size="sm" onClick={() => setDeleteOpen(true)}>
              <Trash2 className="mr-1 h-3 w-3" />
              Delete
            </Button>
          </div>
        }
      />

      <Tabs defaultValue="overview">
        <TabsList>
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="settings">Settings</TabsTrigger>
          <TabsTrigger value="permissions">Permissions</TabsTrigger>
          <TabsTrigger value="connections">Connections</TabsTrigger>
          <TabsTrigger value="api">API</TabsTrigger>
        </TabsList>

        <TabsContent value="overview">
          <div className="grid gap-4 lg:grid-cols-2">
            <Card>
              <CardHeader>
                <CardTitle>Client Information</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-2">
                  <Label>Client ID</Label>
                  <div className="flex gap-2">
                    <Input value={app.client_id} readOnly className="font-mono text-xs" />
                    <Button
                      variant="outline"
                      size="icon"
                      onClick={() => {
                        copyToClipboard(app.client_id)
                        addToast({ title: 'Copied', variant: 'success' })
                      }}
                    >
                      <Copy className="h-4 w-4" />
                    </Button>
                  </div>
                </div>

                {app.client_secret && (
                  <div className="space-y-2">
                    <Label>Client Secret</Label>
                    <div className="flex gap-2">
                      <Input
                        value={showSecret ? app.client_secret : '***************************'}
                        readOnly
                        className="font-mono text-xs"
                      />
                      <Button variant="outline" size="icon" onClick={() => setShowSecret(!showSecret)}>
                        {showSecret ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                      </Button>
                      <Button
                        variant="outline"
                        size="icon"
                        onClick={() => {
                          copyToClipboard(app.client_secret!)
                          addToast({ title: 'Copied', variant: 'success' })
                        }}
                      >
                        <Copy className="h-4 w-4" />
                      </Button>
                    </div>
                    <Button variant="outline" size="sm" onClick={() => setRotateOpen(true)}>
                      <RefreshCw className="mr-1 h-3 w-3" />
                      Rotate Secret
                    </Button>
                  </div>
                )}

                <div className="space-y-2">
                  <Label>Application Type</Label>
                  <Badge variant="secondary" className="text-sm">{app.type.toUpperCase()}</Badge>
                </div>

                <div className="space-y-2">
                  <Label>Status</Label>
                  <div className="flex items-center gap-2">
                    <Switch
                      checked={app.is_active}
                      onCheckedChange={(checked) => updateMutation.mutate({ is_active: checked })}
                    />
                    <span className="text-sm">{app.is_active ? 'Active' : 'Disabled'}</span>
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Grant Types</CardTitle>
                <CardDescription>Select the OAuth grant types this application can use</CardDescription>
              </CardHeader>
              <CardContent className="space-y-3">
                {allGrantTypes.map((gt) => (
                  <div key={gt.value} className="flex items-center gap-2">
                    <Checkbox
                      checked={(app.grant_types || []).includes(gt.value)}
                      onCheckedChange={(checked) => {
                        const grants = app.grant_types || []
                        const newGrants = checked
                          ? [...grants, gt.value]
                          : grants.filter((g) => g !== gt.value)
                        updateMutation.mutate({ grant_types: newGrants })
                      }}
                    />
                    <Label className="cursor-pointer">{gt.label}</Label>
                  </div>
                ))}
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="settings">
          <div className="grid gap-4 lg:grid-cols-2">
            <Card>
              <CardHeader>
                <CardTitle>Redirect URIs</CardTitle>
                <CardDescription>URLs where users can be redirected after authentication</CardDescription>
              </CardHeader>
              <CardContent>
                <TagInput
                  value={formData.redirect_uris ?? app.redirect_uris ?? []}
                  onChange={(val) => setFormData({ ...formData, redirect_uris: val })}
                  placeholder="https://example.com/callback"
                />
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Allowed Origins</CardTitle>
                <CardDescription>Origins allowed for cross-origin requests</CardDescription>
              </CardHeader>
              <CardContent>
                <TagInput
                  value={formData.allowed_origins ?? app.allowed_origins ?? []}
                  onChange={(val) => setFormData({ ...formData, allowed_origins: val })}
                  placeholder="https://example.com"
                />
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Allowed Logout URLs</CardTitle>
              </CardHeader>
              <CardContent>
                <TagInput
                  value={formData.allowed_logout_urls ?? app.allowed_logout_urls ?? []}
                  onChange={(val) => setFormData({ ...formData, allowed_logout_urls: val })}
                  placeholder="https://example.com/logout"
                />
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Token Lifetimes</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-2">
                  <Label>Access Token TTL (seconds)</Label>
                  <Input
                    type="number"
                    defaultValue={app.access_token_ttl}
                    onChange={(e) =>
                      setFormData({ ...formData, access_token_ttl: Number(e.target.value) })
                    }
                  />
                </div>
                <div className="space-y-2">
                  <Label>Refresh Token TTL (seconds)</Label>
                  <Input
                    type="number"
                    defaultValue={app.refresh_token_ttl}
                    onChange={(e) =>
                      setFormData({ ...formData, refresh_token_ttl: Number(e.target.value) })
                    }
                  />
                </div>
                <div className="space-y-2">
                  <Label>ID Token TTL (seconds)</Label>
                  <Input
                    type="number"
                    defaultValue={app.id_token_ttl}
                    onChange={(e) =>
                      setFormData({ ...formData, id_token_ttl: Number(e.target.value) })
                    }
                  />
                </div>
              </CardContent>
            </Card>
          </div>

          <div className="mt-4">
            <Button
              onClick={() => updateMutation.mutate(formData)}
              loading={updateMutation.isPending}
              disabled={Object.keys(formData).length === 0}
            >
              <Save className="mr-2 h-4 w-4" />
              Save Settings
            </Button>
          </div>
        </TabsContent>

        <TabsContent value="permissions">
          <Card>
            <CardHeader>
              <CardTitle>Application Permissions</CardTitle>
              <CardDescription>
                Select which permissions are included in tokens issued for this application.
                If no permissions are selected, all user permissions will be included.
              </CardDescription>
            </CardHeader>
            <CardContent>
              {currentPerms.length === 0 && selectedPerms === null && (
                <div className="flex items-start gap-2 rounded-md border border-blue-200 bg-blue-50 p-3 mb-4 dark:border-blue-900 dark:bg-blue-950">
                  <Info className="h-4 w-4 text-blue-600 mt-0.5 shrink-0 dark:text-blue-400" />
                  <p className="text-sm text-blue-800 dark:text-blue-300">
                    No permissions selected — all user permissions will be included in tokens for this application.
                  </p>
                </div>
              )}
              <div className="space-y-4">
                {Object.entries(permsByGroup).map(([group, perms]) => (
                  <div key={group}>
                    <h3 className="text-sm font-semibold mb-2 capitalize">{group}</h3>
                    <div className="space-y-1.5">
                      {perms.map((perm) => (
                        <div key={perm.id} className="flex items-center gap-2 rounded-md p-1.5 hover:bg-muted/50">
                          <Checkbox
                            checked={currentPerms.includes(perm.name)}
                            onCheckedChange={() => togglePerm(perm.name)}
                          />
                          <span className="text-xs font-mono">{perm.name}</span>
                          {perm.description && (
                            <span className="text-xs text-muted-foreground ml-1">— {perm.description}</span>
                          )}
                        </div>
                      ))}
                    </div>
                  </div>
                ))}
                {Object.keys(permsByGroup).length === 0 && (
                  <p className="text-sm text-muted-foreground text-center py-4">No permissions defined</p>
                )}
              </div>
              {selectedPerms !== null && (
                <div className="mt-4 flex gap-2">
                  <Button
                    onClick={() => savePermsMutation.mutate(currentPerms)}
                    loading={savePermsMutation.isPending}
                  >
                    <Save className="mr-2 h-4 w-4" />
                    Save Permissions
                  </Button>
                  <Button variant="outline" onClick={() => setSelectedPerms(null)}>
                    Cancel
                  </Button>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="connections">
          <Card>
            <CardHeader>
              <CardTitle>Social Connections</CardTitle>
              <CardDescription>Link social identity providers to this application</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
                {['Google', 'GitHub', 'Microsoft', 'Apple', 'Facebook', 'Twitter'].map((provider) => (
                  <div key={provider} className="flex items-center justify-between rounded-lg border p-3">
                    <span className="text-sm font-medium">{provider}</span>
                    <Switch checked={false} onCheckedChange={() => addToast({ title: `${provider} connection toggled`, variant: 'success' })} />
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="api">
          <Card>
            <CardHeader>
              <CardTitle>API Scopes & Permissions</CardTitle>
              <CardDescription>Configure the scopes and permissions this application can request</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                {['openid', 'profile', 'email', 'offline_access', 'read:users', 'write:users', 'read:applications'].map(
                  (scope) => (
                    <div key={scope} className="flex items-center gap-2">
                      <Checkbox checked={true} onCheckedChange={() => {}} />
                      <Label className="font-mono text-xs">{scope}</Label>
                    </div>
                  )
                )}
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      <ConfirmDialog
        open={deleteOpen}
        onClose={() => setDeleteOpen(false)}
        onConfirm={() => deleteMutation.mutate()}
        title="Delete Application"
        description={`Permanently delete "${app.name}"? This will invalidate all client credentials and cannot be undone.`}
        confirmLabel="Delete Application"
        variant="destructive"
        loading={deleteMutation.isPending}
      />

      <ConfirmDialog
        open={rotateOpen}
        onClose={() => setRotateOpen(false)}
        onConfirm={() => rotateMutation.mutate()}
        title="Rotate Client Secret"
        description="This will generate a new client secret and invalidate the current one. Make sure to update your application configuration."
        confirmLabel="Rotate Secret"
        variant="destructive"
        loading={rotateMutation.isPending}
      />
    </div>
  )
}
