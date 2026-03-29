import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { useUIStore } from '@/stores/ui'
import { PageHeader } from '@/components/shared/page-header'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Checkbox } from '@/components/ui/checkbox'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'
import { Slider } from '@/components/ui/slider'
import { Badge } from '@/components/ui/badge'
import { Save, Shield, Key, Mail, Globe, TestTube2, Link2, CheckCircle, Clock, RefreshCw, Trash2, Copy } from 'lucide-react'
import { copyToClipboard } from '@/lib/utils'
import { useI18n } from '@/lib/i18n'
import type { SecuritySettings, MfaSettings, DomainVerification } from '@/types'

const defaultSecurity: SecuritySettings = {
  password_min_length: 8,
  password_require_uppercase: true,
  password_require_lowercase: true,
  password_require_numbers: true,
  password_require_special: false,
  brute_force_protection: true,
  max_login_attempts: 10,
  lockout_duration: 900,
  session_lifetime: 86400,
  session_idle_timeout: 3600,
}

const defaultMfa: MfaSettings = {
  enabled: true,
  required: false,
  allowed_methods: ['totp', 'sms'],
}

export default function SettingsPage() {
  const queryClient = useQueryClient()
  const { addToast } = useUIStore()
  const { t } = useI18n()

  const { data: settings } = useQuery({
    queryKey: ['settings'],
    queryFn: () => api.getSettings(),
  })

  const [security, setSecurity] = useState<SecuritySettings>(defaultSecurity)
  const [mfa, setMfa] = useState<MfaSettings>(defaultMfa)
  const [smtp, setSmtp] = useState({
    host: '',
    port: '587',
    username: '',
    password: '',
    from_email: '',
    from_name: '',
    encryption: 'tls',
  })
  const [socialProviders, setSocialProviders] = useState<
    Record<string, { client_id: string; client_secret: string; enabled: boolean }>
  >({
    google: { client_id: '', client_secret: '', enabled: false },
    github: { client_id: '', client_secret: '', enabled: false },
    microsoft: { client_id: '', client_secret: '', enabled: false },
    apple: { client_id: '', client_secret: '', enabled: false },
    facebook: { client_id: '', client_secret: '', enabled: false },
  })

  useEffect(() => {
    if (settings?.security) setSecurity({ ...defaultSecurity, ...settings.security })
    if (settings?.mfa) setMfa({ ...defaultMfa, ...settings.mfa })
  }, [settings])

  const saveMutation = useMutation({
    mutationFn: (data: { security?: SecuritySettings; mfa?: MfaSettings }) =>
      api.updateSettings(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['settings'] })
      addToast({ title: 'Settings saved', variant: 'success' })
    },
    onError: () => addToast({ title: 'Failed to save settings', variant: 'error' }),
  })

  const testSmtpMutation = useMutation({
    mutationFn: (config: Record<string, string>) => api.testSmtp(config),
    onSuccess: (result) => {
      addToast({
        title: result.success ? 'SMTP connection successful' : 'SMTP test failed',
        variant: result.success ? 'success' : 'error',
      })
    },
    onError: () => addToast({ title: 'SMTP test failed', variant: 'error' }),
  })

  const toggleMfaMethod = (method: 'totp' | 'sms' | 'email' | 'webauthn') => {
    setMfa((prev) => ({
      ...prev,
      allowed_methods: prev.allowed_methods.includes(method)
        ? prev.allowed_methods.filter((m) => m !== method)
        : [...prev.allowed_methods, method],
    }))
  }

  return (
    <div>
      <PageHeader
        title={t('settings.title')}
        description={t('settings.description')}
        breadcrumbs={[{ label: t('dashboard.title'), href: '/' }, { label: t('settings.title') }]}
      />

      <Tabs defaultValue="security">
        <TabsList className="mb-6">
          <TabsTrigger value="security">
            <Shield className="mr-1 h-3 w-3" />
            Security
          </TabsTrigger>
          <TabsTrigger value="mfa">
            <Key className="mr-1 h-3 w-3" />
            MFA
          </TabsTrigger>
          <TabsTrigger value="email">
            <Mail className="mr-1 h-3 w-3" />
            Email
          </TabsTrigger>
          <TabsTrigger value="social">
            <Globe className="mr-1 h-3 w-3" />
            Social Providers
          </TabsTrigger>
          <TabsTrigger value="domain">
            <Link2 className="mr-1 h-3 w-3" />
            Custom Domain
          </TabsTrigger>
        </TabsList>

        <TabsContent value="security">
          <div className="grid gap-4 lg:grid-cols-2">
            <Card>
              <CardHeader>
                <CardTitle>Password Policy</CardTitle>
                <CardDescription>Configure password requirements for users</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-2">
                  <Label>Minimum Length: {security.password_min_length}</Label>
                  <Slider
                    value={security.password_min_length}
                    onValueChange={(v) => setSecurity({ ...security, password_min_length: v })}
                    min={6}
                    max={32}
                    step={1}
                  />
                </div>
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <Label>Require Uppercase</Label>
                    <Switch
                      checked={security.password_require_uppercase}
                      onCheckedChange={(v) => setSecurity({ ...security, password_require_uppercase: v })}
                    />
                  </div>
                  <div className="flex items-center justify-between">
                    <Label>Require Lowercase</Label>
                    <Switch
                      checked={security.password_require_lowercase}
                      onCheckedChange={(v) => setSecurity({ ...security, password_require_lowercase: v })}
                    />
                  </div>
                  <div className="flex items-center justify-between">
                    <Label>Require Numbers</Label>
                    <Switch
                      checked={security.password_require_numbers}
                      onCheckedChange={(v) => setSecurity({ ...security, password_require_numbers: v })}
                    />
                  </div>
                  <div className="flex items-center justify-between">
                    <Label>Require Special Characters</Label>
                    <Switch
                      checked={security.password_require_special}
                      onCheckedChange={(v) => setSecurity({ ...security, password_require_special: v })}
                    />
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Brute Force Protection</CardTitle>
                <CardDescription>Protect against credential stuffing attacks</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="flex items-center justify-between">
                  <Label>Enable Protection</Label>
                  <Switch
                    checked={security.brute_force_protection}
                    onCheckedChange={(v) => setSecurity({ ...security, brute_force_protection: v })}
                  />
                </div>
                <div className="space-y-2">
                  <Label>Max Login Attempts</Label>
                  <Input
                    type="number"
                    value={security.max_login_attempts}
                    onChange={(e) => setSecurity({ ...security, max_login_attempts: Number(e.target.value) })}
                  />
                </div>
                <div className="space-y-2">
                  <Label>Lockout Duration (seconds)</Label>
                  <Input
                    type="number"
                    value={security.lockout_duration}
                    onChange={(e) => setSecurity({ ...security, lockout_duration: Number(e.target.value) })}
                  />
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Session Configuration</CardTitle>
                <CardDescription>Manage session lifetimes and behavior</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-2">
                  <Label>Session Lifetime (seconds)</Label>
                  <Input
                    type="number"
                    value={security.session_lifetime}
                    onChange={(e) => setSecurity({ ...security, session_lifetime: Number(e.target.value) })}
                  />
                  <p className="text-xs text-muted-foreground">
                    {(security.session_lifetime / 3600).toFixed(1)} hours
                  </p>
                </div>
                <div className="space-y-2">
                  <Label>Idle Timeout (seconds)</Label>
                  <Input
                    type="number"
                    value={security.session_idle_timeout}
                    onChange={(e) => setSecurity({ ...security, session_idle_timeout: Number(e.target.value) })}
                  />
                  <p className="text-xs text-muted-foreground">
                    {(security.session_idle_timeout / 60).toFixed(0)} minutes
                  </p>
                </div>
              </CardContent>
            </Card>

            <div className="flex items-end">
              <Button
                onClick={() => saveMutation.mutate({ security })}
                loading={saveMutation.isPending}
              >
                <Save className="mr-2 h-4 w-4" />
                Save Security Settings
              </Button>
            </div>
          </div>
        </TabsContent>

        <TabsContent value="mfa">
          <div className="max-w-xl">
            <Card>
              <CardHeader>
                <CardTitle>Multi-Factor Authentication</CardTitle>
                <CardDescription>Configure MFA requirements and allowed methods</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="flex items-center justify-between">
                  <div>
                    <Label>Enable MFA</Label>
                    <p className="text-xs text-muted-foreground mt-0.5">Allow users to enroll in MFA</p>
                  </div>
                  <Switch
                    checked={mfa.enabled}
                    onCheckedChange={(v) => setMfa({ ...mfa, enabled: v })}
                  />
                </div>

                <div className="flex items-center justify-between">
                  <div>
                    <Label>Require MFA</Label>
                    <p className="text-xs text-muted-foreground mt-0.5">Force all users to enroll in MFA</p>
                  </div>
                  <Switch
                    checked={mfa.required}
                    onCheckedChange={(v) => setMfa({ ...mfa, required: v })}
                    disabled={!mfa.enabled}
                  />
                </div>

                <div className="space-y-2">
                  <Label>Allowed Methods</Label>
                  <div className="space-y-2">
                    {[
                      { value: 'totp' as const, label: 'Authenticator App (TOTP)' },
                      { value: 'sms' as const, label: 'SMS' },
                      { value: 'email' as const, label: 'Email' },
                      { value: 'webauthn' as const, label: 'WebAuthn / Passkeys' },
                    ].map((method) => (
                      <div key={method.value} className="flex items-center gap-2">
                        <Checkbox
                          checked={mfa.allowed_methods.includes(method.value)}
                          onCheckedChange={() => toggleMfaMethod(method.value)}
                          disabled={!mfa.enabled}
                        />
                        <Label>{method.label}</Label>
                      </div>
                    ))}
                  </div>
                </div>

                <Button
                  onClick={() => saveMutation.mutate({ mfa })}
                  loading={saveMutation.isPending}
                >
                  <Save className="mr-2 h-4 w-4" />
                  Save MFA Settings
                </Button>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="email">
          <div className="max-w-xl">
            <Card>
              <CardHeader>
                <CardTitle>SMTP Configuration</CardTitle>
                <CardDescription>Configure the email server for sending transactional emails</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid gap-4 sm:grid-cols-2">
                  <div className="space-y-2">
                    <Label>SMTP Host</Label>
                    <Input
                      value={smtp.host}
                      onChange={(e) => setSmtp({ ...smtp, host: e.target.value })}
                      placeholder="smtp.example.com"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>Port</Label>
                    <Input
                      value={smtp.port}
                      onChange={(e) => setSmtp({ ...smtp, port: e.target.value })}
                      placeholder="587"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>Username</Label>
                    <Input
                      value={smtp.username}
                      onChange={(e) => setSmtp({ ...smtp, username: e.target.value })}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>Password</Label>
                    <Input
                      type="password"
                      value={smtp.password}
                      onChange={(e) => setSmtp({ ...smtp, password: e.target.value })}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>From Email</Label>
                    <Input
                      value={smtp.from_email}
                      onChange={(e) => setSmtp({ ...smtp, from_email: e.target.value })}
                      placeholder="noreply@example.com"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>From Name</Label>
                    <Input
                      value={smtp.from_name}
                      onChange={(e) => setSmtp({ ...smtp, from_name: e.target.value })}
                      placeholder="CPI Auth"
                    />
                  </div>
                </div>

                <div className="flex gap-2">
                  <Button
                    variant="outline"
                    onClick={() => testSmtpMutation.mutate(smtp)}
                    loading={testSmtpMutation.isPending}
                  >
                    <TestTube2 className="mr-2 h-4 w-4" />
                    Test Connection
                  </Button>
                  <Button onClick={() => addToast({ title: 'SMTP settings saved', variant: 'success' })}>
                    <Save className="mr-2 h-4 w-4" />
                    Save
                  </Button>
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="social">
          <div className="grid gap-4 lg:grid-cols-2">
            {Object.entries(socialProviders).map(([provider, config]) => (
              <Card key={provider}>
                <CardHeader>
                  <div className="flex items-center justify-between">
                    <CardTitle className="capitalize">{provider}</CardTitle>
                    <Switch
                      checked={config.enabled}
                      onCheckedChange={(checked) =>
                        setSocialProviders({
                          ...socialProviders,
                          [provider]: { ...config, enabled: checked },
                        })
                      }
                    />
                  </div>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="space-y-2">
                    <Label>Client ID</Label>
                    <Input
                      value={config.client_id}
                      onChange={(e) =>
                        setSocialProviders({
                          ...socialProviders,
                          [provider]: { ...config, client_id: e.target.value },
                        })
                      }
                      placeholder={`${provider}_client_id`}
                      disabled={!config.enabled}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>Client Secret</Label>
                    <Input
                      type="password"
                      value={config.client_secret}
                      onChange={(e) =>
                        setSocialProviders({
                          ...socialProviders,
                          [provider]: { ...config, client_secret: e.target.value },
                        })
                      }
                      placeholder={`${provider}_client_secret`}
                      disabled={!config.enabled}
                    />
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
          <div className="mt-4">
            <Button onClick={() => addToast({ title: 'Social provider settings saved', variant: 'success' })}>
              <Save className="mr-2 h-4 w-4" />
              Save Social Providers
            </Button>
          </div>
        </TabsContent>
        <TabsContent value="domain">
          <DomainVerificationPanel />
        </TabsContent>
      </Tabs>
    </div>
  )
}

function DomainVerificationPanel() {
  const queryClient = useQueryClient()
  const { addToast } = useUIStore()
  const [newDomain, setNewDomain] = useState('')

  const { data: dv, isLoading } = useQuery({
    queryKey: ['domain-verification'],
    queryFn: () => api.getDomainVerification(),
  })

  const initiateMutation = useMutation({
    mutationFn: (domain: string) => api.initiateDomainVerification(domain),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['domain-verification'] })
      setNewDomain('')
      addToast({ title: 'Domain verification initiated', variant: 'success' })
    },
    onError: () => addToast({ title: 'Failed to initiate verification', variant: 'error' }),
  })

  const checkMutation = useMutation({
    mutationFn: (id: string) => api.checkDomainVerification(id),
    onSuccess: (result) => {
      queryClient.invalidateQueries({ queryKey: ['domain-verification'] })
      if (result.is_verified) {
        addToast({ title: 'Domain verified successfully!', variant: 'success' })
      } else {
        addToast({ title: 'DNS record not found yet. Please wait for propagation.', variant: 'error' })
      }
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.deleteDomainVerification(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['domain-verification'] })
      addToast({ title: 'Domain verification removed', variant: 'success' })
    },
  })

  if (isLoading) return null

  const hasVerification = dv && dv.status !== 'none'

  return (
    <div className="max-w-xl">
      <Card>
        <CardHeader>
          <CardTitle>Custom Domain</CardTitle>
          <CardDescription>
            Use your own domain for authentication pages (e.g., auth.yourcompany.com)
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {!hasVerification ? (
            <div className="space-y-4">
              <div className="space-y-2">
                <Label>Domain</Label>
                <div className="flex gap-2">
                  <Input
                    value={newDomain}
                    onChange={(e) => setNewDomain(e.target.value)}
                    placeholder="auth.yourcompany.com"
                  />
                  <Button
                    onClick={() => initiateMutation.mutate(newDomain)}
                    loading={initiateMutation.isPending}
                    disabled={!newDomain}
                  >
                    Verify
                  </Button>
                </div>
              </div>
            </div>
          ) : (
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="font-medium">{dv.domain}</p>
                  <div className="flex items-center gap-2 mt-1">
                    {dv.is_verified ? (
                      <Badge variant="success" className="flex items-center gap-1">
                        <CheckCircle className="h-3 w-3" />
                        Verified
                      </Badge>
                    ) : (
                      <Badge variant="warning" className="flex items-center gap-1">
                        <Clock className="h-3 w-3" />
                        Pending Verification
                      </Badge>
                    )}
                  </div>
                </div>
                <div className="flex gap-1">
                  {!dv.is_verified && (
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => checkMutation.mutate(dv.id)}
                      loading={checkMutation.isPending}
                    >
                      <RefreshCw className="mr-1 h-3 w-3" />
                      Check DNS
                    </Button>
                  )}
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    onClick={() => deleteMutation.mutate(dv.id)}
                  >
                    <Trash2 className="h-3 w-3 text-destructive" />
                  </Button>
                </div>
              </div>

              {!dv.is_verified && dv.dns_record && (
                <Card className="bg-muted/50">
                  <CardContent className="pt-4 space-y-3">
                    <p className="text-sm font-medium">Add this DNS record to verify ownership:</p>
                    <div className="space-y-2 text-sm">
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">Type:</span>
                        <span className="font-mono">{dv.dns_record.record_type}</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">Host:</span>
                        <div className="flex items-center gap-1">
                          <span className="font-mono text-xs">{dv.dns_record.host}</span>
                          <button
                            onClick={() => {
                              copyToClipboard(dv.dns_record!.host)
                              addToast({ title: 'Copied', variant: 'success' })
                            }}
                            className="text-muted-foreground hover:text-foreground"
                          >
                            <Copy className="h-3 w-3" />
                          </button>
                        </div>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">Value:</span>
                        <div className="flex items-center gap-1">
                          <span className="font-mono text-xs">{dv.dns_record.value}</span>
                          <button
                            onClick={() => {
                              copyToClipboard(dv.dns_record!.value)
                              addToast({ title: 'Copied', variant: 'success' })
                            }}
                            className="text-muted-foreground hover:text-foreground"
                          >
                            <Copy className="h-3 w-3" />
                          </button>
                        </div>
                      </div>
                    </div>
                    <p className="text-xs text-muted-foreground">
                      DNS changes may take up to 48 hours to propagate. Click "Check DNS" to verify.
                    </p>
                  </CardContent>
                </Card>
              )}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
