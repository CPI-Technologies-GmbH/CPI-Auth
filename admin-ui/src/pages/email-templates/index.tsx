import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { useUIStore } from '@/stores/ui'
import { PageHeader } from '@/components/shared/page-header'
import { CodeEditor } from '@/components/shared/code-editor'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
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
import { Mail, Eye, Save, Send, Code } from 'lucide-react'
import type { EmailTemplate, EmailTemplateType } from '@/types'

const templateTypeLabels: Record<EmailTemplateType, string> = {
  welcome: 'Welcome',
  verification: 'Email Verification',
  password_reset: 'Password Reset',
  mfa_code: 'MFA Code',
  invitation: 'Invitation',
  blocked: 'Account Blocked',
  password_changed: 'Password Changed',
}

const templateVariables = [
  '{{user.name}}',
  '{{user.email}}',
  '{{application.name}}',
  '{{tenant.name}}',
  '{{code}}',
  '{{link}}',
  '{{expiry}}',
]

export default function EmailTemplatesPage() {
  const queryClient = useQueryClient()
  const { addToast } = useUIStore()

  const [selectedTemplate, setSelectedTemplate] = useState<EmailTemplate | null>(null)
  const [editData, setEditData] = useState({ subject: '', body_mjml: '' })
  const [testEmailOpen, setTestEmailOpen] = useState(false)
  const [testEmail, setTestEmail] = useState('')
  const [previewMode, setPreviewMode] = useState<'code' | 'preview'>('code')

  const { data: templates, isLoading } = useQuery({
    queryKey: ['email-templates'],
    queryFn: () => api.getEmailTemplates(),
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<EmailTemplate> }) =>
      api.updateEmailTemplate(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['email-templates'] })
      addToast({ title: 'Template saved', variant: 'success' })
    },
    onError: () => addToast({ title: 'Failed to save template', variant: 'error' }),
  })

  const sendTestMutation = useMutation({
    mutationFn: ({ id, email }: { id: string; email: string }) =>
      api.sendTestEmail(id, email),
    onSuccess: () => {
      setTestEmailOpen(false)
      addToast({ title: 'Test email sent', variant: 'success' })
    },
    onError: () => addToast({ title: 'Failed to send test email', variant: 'error' }),
  })

  const handleSelectTemplate = (template: EmailTemplate) => {
    setSelectedTemplate(template)
    setEditData({
      subject: template.subject,
      body_mjml: template.body_mjml,
    })
    setPreviewMode('code')
  }

  const insertVariable = (variable: string) => {
    setEditData((prev) => ({
      ...prev,
      body_mjml: prev.body_mjml + variable,
    }))
  }

  return (
    <div>
      <PageHeader
        title="Email Templates"
        description="Customize email templates for authentication flows"
        breadcrumbs={[{ label: 'Dashboard', href: '/' }, { label: 'Email Templates' }]}
      />

      <div className="grid gap-6 lg:grid-cols-3">
        {/* Template List */}
        <div>
          <h2 className="text-sm font-medium text-muted-foreground mb-3">Templates</h2>
          {isLoading ? (
            <div className="space-y-2">
              {Array.from({ length: 5 }).map((_, i) => (
                <Skeleton key={i} className="h-16" />
              ))}
            </div>
          ) : (
            <div className="space-y-2">
              {(templates ?? []).map((template) => (
                <Card
                  key={template.id}
                  className={`cursor-pointer transition-colors ${
                    selectedTemplate?.id === template.id
                      ? 'border-primary bg-primary/5'
                      : 'hover:border-primary/50'
                  }`}
                  onClick={() => handleSelectTemplate(template)}
                >
                  <CardContent className="p-3">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-3">
                        <Mail className="h-4 w-4 text-primary" />
                        <div>
                          <p className="font-medium text-sm">
                            {templateTypeLabels[template.type] || template.type}
                          </p>
                          <p className="text-xs text-muted-foreground">{template.subject}</p>
                        </div>
                      </div>
                      <div className="flex items-center gap-1">
                        <Badge variant="muted" className="text-[10px]">{template.locale}</Badge>
                        <Badge variant={template.is_active ? 'success' : 'muted'} className="text-[10px]">
                          {template.is_active ? 'Active' : 'Draft'}
                        </Badge>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          )}
        </div>

        {/* Editor */}
        <div className="lg:col-span-2">
          {selectedTemplate ? (
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <h2 className="text-lg font-semibold">
                  {templateTypeLabels[selectedTemplate.type]}
                </h2>
                <div className="flex gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      setTestEmailOpen(true)
                      setTestEmail('')
                    }}
                  >
                    <Send className="mr-1 h-3 w-3" />
                    Send Test
                  </Button>
                  <Button
                    size="sm"
                    onClick={() =>
                      updateMutation.mutate({
                        id: selectedTemplate.id,
                        data: editData,
                      })
                    }
                    loading={updateMutation.isPending}
                  >
                    <Save className="mr-1 h-3 w-3" />
                    Save
                  </Button>
                </div>
              </div>

              <div className="space-y-2">
                <Label>Subject</Label>
                <Input
                  value={editData.subject}
                  onChange={(e) => setEditData({ ...editData, subject: e.target.value })}
                />
              </div>

              {/* Variable toolbar */}
              <div className="flex items-center gap-2 flex-wrap">
                <span className="text-xs text-muted-foreground">Insert variable:</span>
                {templateVariables.map((v) => (
                  <button
                    key={v}
                    onClick={() => insertVariable(v)}
                    className="rounded-md bg-muted px-2 py-1 text-xs font-mono hover:bg-accent cursor-pointer transition-colors"
                  >
                    {v}
                  </button>
                ))}
              </div>

              {/* Toggle code/preview */}
              <div className="flex gap-1 bg-muted rounded-md p-1 w-fit">
                <button
                  onClick={() => setPreviewMode('code')}
                  className={`px-3 py-1 rounded-sm text-xs font-medium transition-colors cursor-pointer ${
                    previewMode === 'code' ? 'bg-background text-foreground shadow-sm' : 'text-muted-foreground'
                  }`}
                >
                  <Code className="h-3 w-3 inline mr-1" />
                  MJML Code
                </button>
                <button
                  onClick={() => setPreviewMode('preview')}
                  className={`px-3 py-1 rounded-sm text-xs font-medium transition-colors cursor-pointer ${
                    previewMode === 'preview' ? 'bg-background text-foreground shadow-sm' : 'text-muted-foreground'
                  }`}
                >
                  <Eye className="h-3 w-3 inline mr-1" />
                  HTML Preview
                </button>
              </div>

              {previewMode === 'code' ? (
                <CodeEditor
                  value={editData.body_mjml}
                  onChange={(v) => setEditData({ ...editData, body_mjml: v })}
                  language="xml"
                  height="400px"
                />
              ) : (
                <Card>
                  <CardContent className="p-0">
                    <div
                      className="p-6 min-h-[400px] bg-white rounded-lg"
                      dangerouslySetInnerHTML={{ __html: selectedTemplate.body_html }}
                    />
                  </CardContent>
                </Card>
              )}
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center h-96 text-muted-foreground">
              <Mail className="h-12 w-12 mb-4" />
              <p className="text-sm">Select a template to edit</p>
            </div>
          )}
        </div>
      </div>

      {/* Send Test Email */}
      <Dialog open={testEmailOpen} onClose={() => setTestEmailOpen(false)}>
        <DialogClose onClose={() => setTestEmailOpen(false)} />
        <DialogHeader>
          <DialogTitle>Send Test Email</DialogTitle>
        </DialogHeader>
        <div className="mt-4 space-y-4">
          <div className="space-y-2">
            <Label>Recipient Email</Label>
            <Input
              type="email"
              value={testEmail}
              onChange={(e) => setTestEmail(e.target.value)}
              placeholder="test@example.com"
            />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setTestEmailOpen(false)}>Cancel</Button>
            <Button
              onClick={() =>
                selectedTemplate && sendTestMutation.mutate({ id: selectedTemplate.id, email: testEmail })
              }
              loading={sendTestMutation.isPending}
              disabled={!testEmail}
            >
              Send Test
            </Button>
          </DialogFooter>
        </div>
      </Dialog>
    </div>
  )
}
