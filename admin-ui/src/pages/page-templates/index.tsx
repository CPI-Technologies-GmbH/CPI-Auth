import { useState, useMemo } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { useUIStore } from '@/stores/ui'
import { PageHeader } from '@/components/shared/page-header'
import { CodeEditor } from '@/components/shared/code-editor'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select } from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Dialog,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  DialogClose,
} from '@/components/ui/dialog'
import { ConfirmDialog } from '@/components/shared/confirm-dialog'
import { FileCode, Eye, Save, Code, Plus, Trash2, Copy, Lock, Languages, Search, X } from 'lucide-react'
import type { PageTemplate, PageTemplateType, LanguageString } from '@/types'

// ─── Constants ────────────────────────────────────────────────

const pageTypeLabels: Record<string, string> = {
  login: 'Login',
  signup: 'Sign Up',
  verification: 'Email Verification',
  password_reset: 'Password Reset',
  mfa_challenge: 'MFA Challenge',
  error: 'Error Page',
  consent: 'OAuth Consent',
  profile: 'User Profile',
  custom: 'Custom Page',
}

const pageTypeOptions = Object.entries(pageTypeLabels).map(([value, label]) => ({
  value,
  label,
}))

const templateVariables = [
  { key: '{{user.name}}', label: 'User Name' },
  { key: '{{user.email}}', label: 'User Email' },
  { key: '{{user.initials}}', label: 'Initials' },
  { key: '{{application.name}}', label: 'App Name' },
  { key: '{{tenant.name}}', label: 'Tenant' },
  { key: '{{code}}', label: 'Code' },
  { key: '{{link}}', label: 'Link' },
  { key: '{{error}}', label: 'Error' },
  { key: '{{custom_fields}}', label: 'Custom Fields (Registration)' },
  { key: '{{profile_fields}}', label: 'Custom Fields (Profile)' },
]

const sampleCustomFields = `
      <div class="custom-field">
        <label>Company <span class="required">*</span></label>
        <input type="text" placeholder="Acme Inc." />
      </div>
      <div class="custom-field">
        <label>Phone number</label>
        <input type="tel" placeholder="+1 (555) 123-4567" />
      </div>
      <div class="custom-field">
        <label>Role</label>
        <select><option>Developer</option><option>Designer</option><option>Manager</option></select>
      </div>`

const sampleData: Record<string, string> = {
  '{{user.name}}': 'John Doe',
  '{{user.email}}': 'john@example.com',
  '{{user.initials}}': 'JD',
  '{{application.name}}': 'My Application',
  '{{tenant.name}}': 'Acme Corp',
  '{{code}}': '847293',
  '{{link}}': '#',
  '{{error}}': 'The session has expired. Please try again.',
  '{{custom_fields}}': sampleCustomFields,
  '{{profile_fields}}': sampleCustomFields,
}

const defaultHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{{application.name}}</title>
</head>
<body>
  <div class="card">
    <h1>{{application.name}}</h1>
    <p>Your content here</p>
  </div>
</body>
</html>`

const defaultCSS = `* { margin: 0; padding: 0; box-sizing: border-box; }
body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  background: #0f172a;
  color: #e2e8f0;
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
}
.card {
  background: #1e293b;
  border-radius: 16px;
  padding: 2.5rem;
  max-width: 420px;
  width: 100%;
  box-shadow: 0 25px 50px -12px rgba(0,0,0,.5);
  text-align: center;
}
h1 { font-size: 1.5rem; font-weight: 700; margin-bottom: 1rem; }`

// ─── Main Component ───────────────────────────────────────────

export default function PageTemplatesPage() {
  const queryClient = useQueryClient()
  const { addToast } = useUIStore()

  const [selectedTemplate, setSelectedTemplate] = useState<PageTemplate | null>(null)
  const [editData, setEditData] = useState({ html_content: '', css_content: '', name: '', is_active: false })
  const [previewMode, setPreviewMode] = useState<'html' | 'css' | 'preview'>('html')
  const [createOpen, setCreateOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [duplicateOpen, setDuplicateOpen] = useState(false)
  const [duplicateName, setDuplicateName] = useState('')
  const [langOpen, setLangOpen] = useState(false)
  const [searchQuery, setSearchQuery] = useState('')
  const [newTemplate, setNewTemplate] = useState({ page_type: 'custom' as PageTemplateType, name: '' })

  // Language strings state
  const [langLocale, setLangLocale] = useState('en')
  const [newStringKey, setNewStringKey] = useState('')
  const [newStringValue, setNewStringValue] = useState('')

  // ─── Queries ──────────────────────────────────────────────

  const { data: templates, isLoading } = useQuery({
    queryKey: ['page-templates'],
    queryFn: () => api.getPageTemplates(),
  })

  const { data: langStrings } = useQuery({
    queryKey: ['language-strings', langLocale],
    queryFn: () => api.getLanguageStrings(langLocale),
  })

  // ─── Mutations ────────────────────────────────────────────

  const createMutation = useMutation({
    mutationFn: (data: Partial<PageTemplate>) => api.createPageTemplate(data),
    onSuccess: (tmpl) => {
      queryClient.invalidateQueries({ queryKey: ['page-templates'] })
      addToast({ title: 'Template created', variant: 'success' })
      setCreateOpen(false)
      handleSelectTemplate(tmpl)
    },
    onError: () => addToast({ title: 'Failed to create template', variant: 'error' }),
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<PageTemplate> }) =>
      api.updatePageTemplate(id, data),
    onSuccess: (tmpl) => {
      queryClient.invalidateQueries({ queryKey: ['page-templates'] })
      addToast({ title: 'Template saved', variant: 'success' })
      setSelectedTemplate(tmpl)
    },
    onError: () => addToast({ title: 'Failed to save template', variant: 'error' }),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.deletePageTemplate(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['page-templates'] })
      addToast({ title: 'Template deleted', variant: 'success' })
      setSelectedTemplate(null)
      setDeleteOpen(false)
    },
    onError: () => addToast({ title: 'Failed to delete template', variant: 'error' }),
  })

  const duplicateMutation = useMutation({
    mutationFn: ({ id, name }: { id: string; name: string }) =>
      api.duplicatePageTemplate(id, name),
    onSuccess: (tmpl) => {
      queryClient.invalidateQueries({ queryKey: ['page-templates'] })
      addToast({ title: 'Template duplicated', variant: 'success' })
      setDuplicateOpen(false)
      handleSelectTemplate(tmpl)
    },
    onError: () => addToast({ title: 'Failed to duplicate template', variant: 'error' }),
  })

  const upsertLangMutation = useMutation({
    mutationFn: (data: { string_key: string; locale: string; value: string }) =>
      api.upsertLanguageString(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['language-strings'] })
      setNewStringKey('')
      setNewStringValue('')
    },
    onError: () => addToast({ title: 'Failed to save string', variant: 'error' }),
  })

  const deleteLangMutation = useMutation({
    mutationFn: ({ key, locale }: { key: string; locale: string }) =>
      api.deleteLanguageString(key, locale),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['language-strings'] })
    },
  })

  // ─── Handlers ─────────────────────────────────────────────

  const handleSelectTemplate = (template: PageTemplate) => {
    setSelectedTemplate(template)
    setEditData({
      html_content: template.html_content,
      css_content: template.css_content,
      name: template.name,
      is_active: template.is_active,
    })
    setPreviewMode('html')
  }

  const handleCreate = () => {
    createMutation.mutate({
      page_type: newTemplate.page_type,
      name: newTemplate.name || pageTypeLabels[newTemplate.page_type],
      html_content: defaultHTML,
      css_content: defaultCSS,
    })
  }

  const handleDuplicate = () => {
    if (!selectedTemplate) return
    duplicateMutation.mutate({
      id: selectedTemplate.id,
      name: duplicateName || `${selectedTemplate.name} (Copy)`,
    })
  }

  const insertVariable = (variable: string) => {
    if (previewMode === 'html') {
      setEditData((prev) => ({ ...prev, html_content: prev.html_content + variable }))
    }
  }

  const buildPreview = () => {
    let html = editData.html_content || ''
    const css = editData.css_content || ''

    // Replace template variables with sample data
    for (const [key, value] of Object.entries(sampleData)) {
      html = html.replaceAll(key, value)
    }

    // Replace {{t.xxx}} with language string values
    if (langStrings) {
      for (const ls of langStrings) {
        html = html.replaceAll(`{{t.${ls.string_key}}}`, ls.value)
      }
    }
    // Replace any remaining {{t.xxx}} with the key itself (as fallback)
    html = html.replace(/\{\{t\.([^}]+)\}\}/g, '[$1]')

    return `${html.replace('</head>', `<style>${css}</style></head>`)}`
  }

  // ─── Filtered & grouped templates ─────────────────────────

  const filtered = useMemo(() => {
    if (!templates) return []
    if (!searchQuery) return templates
    const q = searchQuery.toLowerCase()
    return templates.filter(
      (t) =>
        t.name.toLowerCase().includes(q) ||
        t.page_type.toLowerCase().includes(q) ||
        (pageTypeLabels[t.page_type] || '').toLowerCase().includes(q)
    )
  }, [templates, searchQuery])

  const grouped = useMemo(() => {
    const defaults = filtered.filter((t) => t.is_default)
    const custom = filtered.filter((t) => !t.is_default)
    return { defaults, custom }
  }, [filtered])

  const isReadonly = selectedTemplate?.is_default ?? false

  // ─── Render ───────────────────────────────────────────────

  return (
    <div>
      <PageHeader
        title="Page Templates"
        description="Customize authentication flow pages with HTML, CSS, and language strings"
        breadcrumbs={[{ label: 'Dashboard', href: '/' }, { label: 'Page Templates' }]}
        actions={
          <div className="flex gap-2">
            <Button variant="outline" onClick={() => setLangOpen(true)}>
              <Languages className="mr-2 h-4 w-4" />
              Language Strings
            </Button>
            <Button onClick={() => { setCreateOpen(true); setNewTemplate({ page_type: 'custom', name: '' }) }}>
              <Plus className="mr-2 h-4 w-4" />
              Create Template
            </Button>
          </div>
        }
      />

      <div className="grid gap-6 lg:grid-cols-3">
        {/* ─── Template List ────────────────────────────────── */}
        <div className="space-y-4">
          {/* Search */}
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search templates..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-9"
            />
          </div>

          {isLoading ? (
            <div className="space-y-2">
              {Array.from({ length: 4 }).map((_, i) => (
                <Skeleton key={i} className="h-16" />
              ))}
            </div>
          ) : (
            <>
              {/* Default templates */}
              {grouped.defaults.length > 0 && (
                <div>
                  <h3 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2 flex items-center gap-1.5">
                    <Lock className="h-3 w-3" />
                    Default Templates
                  </h3>
                  <div className="space-y-1.5">
                    {grouped.defaults.map((template) => (
                      <TemplateCard
                        key={template.id}
                        template={template}
                        isSelected={selectedTemplate?.id === template.id}
                        onClick={() => handleSelectTemplate(template)}
                      />
                    ))}
                  </div>
                </div>
              )}

              {/* Custom templates */}
              {grouped.custom.length > 0 && (
                <div>
                  <h3 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2">
                    Custom Templates
                  </h3>
                  <div className="space-y-1.5">
                    {grouped.custom.map((template) => (
                      <TemplateCard
                        key={template.id}
                        template={template}
                        isSelected={selectedTemplate?.id === template.id}
                        onClick={() => handleSelectTemplate(template)}
                      />
                    ))}
                  </div>
                </div>
              )}

              {filtered.length === 0 && (
                <div className="flex flex-col items-center justify-center py-12 text-muted-foreground">
                  <FileCode className="h-10 w-10 mb-3" />
                  <p className="text-sm">No templates found</p>
                </div>
              )}
            </>
          )}
        </div>

        {/* ─── Editor Panel ─────────────────────────────────── */}
        <div className="lg:col-span-2">
          {selectedTemplate ? (
            <div className="space-y-4">
              {/* Header */}
              <div className="flex items-center justify-between">
                <div>
                  <div className="flex items-center gap-2">
                    <h2 className="text-lg font-semibold">{isReadonly ? selectedTemplate.name : editData.name}</h2>
                    {isReadonly && <Badge variant="muted"><Lock className="h-3 w-3 mr-1" />Default</Badge>}
                  </div>
                  <p className="text-xs text-muted-foreground">
                    {pageTypeLabels[selectedTemplate.page_type as PageTemplateType] || selectedTemplate.page_type}
                  </p>
                </div>
                <div className="flex gap-2">
                  {/* Duplicate button (available for all templates) */}
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      setDuplicateName(`${selectedTemplate.name} (Copy)`)
                      setDuplicateOpen(true)
                    }}
                  >
                    <Copy className="mr-1 h-3 w-3" />
                    Duplicate
                  </Button>
                  {!isReadonly && (
                    <>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => setDeleteOpen(true)}
                      >
                        <Trash2 className="mr-1 h-3 w-3" />
                        Delete
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() =>
                          updateMutation.mutate({
                            id: selectedTemplate.id,
                            data: { ...editData, is_active: !editData.is_active },
                          })
                        }
                      >
                        {editData.is_active ? 'Deactivate' : 'Activate'}
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
                    </>
                  )}
                </div>
              </div>

              {/* Name field (only for custom) */}
              {!isReadonly && (
                <div className="space-y-2">
                  <Label>Template Name</Label>
                  <Input
                    value={editData.name}
                    onChange={(e) => setEditData({ ...editData, name: e.target.value })}
                  />
                </div>
              )}

              {/* Variable toolbar */}
              <div className="flex items-center gap-2 flex-wrap">
                <span className="text-xs text-muted-foreground">Variables:</span>
                {templateVariables.map((v) => (
                  <button
                    key={v.key}
                    onClick={() => !isReadonly && insertVariable(v.key)}
                    disabled={isReadonly}
                    title={v.label}
                    className="rounded-md bg-muted px-2 py-1 text-xs font-mono hover:bg-accent cursor-pointer transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {v.key}
                  </button>
                ))}
                {/* Language string variables */}
                {langStrings && langStrings.length > 0 && (
                  <>
                    <span className="text-xs text-muted-foreground ml-2">Strings:</span>
                    {langStrings.slice(0, 5).map((ls) => (
                      <button
                        key={ls.string_key}
                        onClick={() => !isReadonly && insertVariable(`{{t.${ls.string_key}}}`)}
                        disabled={isReadonly}
                        title={`${ls.string_key} = "${ls.value}"`}
                        className="rounded-md bg-indigo-500/10 text-indigo-400 border border-indigo-500/20 px-2 py-1 text-xs font-mono hover:bg-indigo-500/20 cursor-pointer transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                      >
                        t.{ls.string_key}
                      </button>
                    ))}
                    {langStrings.length > 5 && (
                      <button
                        onClick={() => setLangOpen(true)}
                        className="rounded-md bg-muted px-2 py-1 text-xs text-muted-foreground hover:bg-accent cursor-pointer"
                      >
                        +{langStrings.length - 5} more
                      </button>
                    )}
                  </>
                )}
              </div>

              {/* Editor tabs */}
              <div className="flex gap-1 bg-muted rounded-md p-1 w-fit">
                {(['html', 'css', 'preview'] as const).map((mode) => (
                  <button
                    key={mode}
                    onClick={() => setPreviewMode(mode)}
                    className={`px-3 py-1 rounded-sm text-xs font-medium transition-colors cursor-pointer flex items-center gap-1 ${
                      previewMode === mode ? 'bg-background text-foreground shadow-sm' : 'text-muted-foreground'
                    }`}
                  >
                    {mode === 'preview' ? <Eye className="h-3 w-3" /> : <Code className="h-3 w-3" />}
                    {mode.toUpperCase()}
                  </button>
                ))}
              </div>

              {/* Editor/Preview content */}
              {previewMode === 'html' && (
                <CodeEditor
                  value={isReadonly ? selectedTemplate.html_content : editData.html_content}
                  onChange={(v) => !isReadonly && setEditData({ ...editData, html_content: v })}
                  language="html"
                  height="400px"
                />
              )}
              {previewMode === 'css' && (
                <CodeEditor
                  value={isReadonly ? selectedTemplate.css_content : editData.css_content}
                  onChange={(v) => !isReadonly && setEditData({ ...editData, css_content: v })}
                  language="css"
                  height="400px"
                />
              )}
              {previewMode === 'preview' && (
                <Card>
                  <CardContent className="p-0">
                    <iframe
                      title="Template Preview"
                      srcDoc={buildPreview()}
                      className="w-full rounded-lg border-0"
                      style={{ height: '500px' }}
                      sandbox="allow-same-origin"
                    />
                  </CardContent>
                </Card>
              )}

              {isReadonly && (
                <div className="flex items-start gap-2 rounded-md border border-amber-500/20 bg-amber-500/5 p-3">
                  <Lock className="h-4 w-4 text-amber-500 mt-0.5 shrink-0" />
                  <p className="text-sm text-amber-200/80">
                    This is a default template and cannot be edited. Click <strong>Duplicate</strong> to create an editable copy.
                  </p>
                </div>
              )}
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center h-96 text-muted-foreground">
              <FileCode className="h-12 w-12 mb-4" />
              <p className="text-sm">Select a template to preview</p>
              <p className="text-xs mt-1">or duplicate a default template to customize it</p>
            </div>
          )}
        </div>
      </div>

      {/* ─── Create Template Dialog ────────────────────────── */}
      <Dialog open={createOpen} onClose={() => setCreateOpen(false)}>
        <DialogClose onClose={() => setCreateOpen(false)} />
        <DialogHeader>
          <DialogTitle>Create Template</DialogTitle>
        </DialogHeader>
        <div className="mt-4 space-y-4">
          <div className="space-y-2">
            <Label>Page Type</Label>
            <Select
              value={newTemplate.page_type}
              onChange={(e) => setNewTemplate({ ...newTemplate, page_type: e.target.value as PageTemplateType })}
              options={pageTypeOptions}
            />
          </div>
          <div className="space-y-2">
            <Label>Template Name</Label>
            <Input
              value={newTemplate.name}
              onChange={(e) => setNewTemplate({ ...newTemplate, name: e.target.value })}
              placeholder={pageTypeLabels[newTemplate.page_type] || 'My Template'}
            />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateOpen(false)}>Cancel</Button>
            <Button onClick={handleCreate} loading={createMutation.isPending}>Create</Button>
          </DialogFooter>
        </div>
      </Dialog>

      {/* ─── Duplicate Dialog ──────────────────────────────── */}
      <Dialog open={duplicateOpen} onClose={() => setDuplicateOpen(false)}>
        <DialogClose onClose={() => setDuplicateOpen(false)} />
        <DialogHeader>
          <DialogTitle>Duplicate Template</DialogTitle>
        </DialogHeader>
        <div className="mt-4 space-y-4">
          <p className="text-sm text-muted-foreground">
            Create an editable copy of <strong>{selectedTemplate?.name}</strong>
          </p>
          <div className="space-y-2">
            <Label>New Template Name</Label>
            <Input
              value={duplicateName}
              onChange={(e) => setDuplicateName(e.target.value)}
              placeholder="My Custom Login"
            />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDuplicateOpen(false)}>Cancel</Button>
            <Button onClick={handleDuplicate} loading={duplicateMutation.isPending}>
              <Copy className="mr-2 h-4 w-4" />
              Duplicate
            </Button>
          </DialogFooter>
        </div>
      </Dialog>

      {/* ─── Language Strings Dialog ───────────────────────── */}
      <Dialog open={langOpen} onClose={() => setLangOpen(false)} className="max-w-2xl">
        <DialogClose onClose={() => setLangOpen(false)} />
        <DialogHeader>
          <DialogTitle>Language Strings</DialogTitle>
        </DialogHeader>
        <div className="mt-4 space-y-4">
          <p className="text-sm text-muted-foreground">
            Define reusable text strings for templates. Use <code className="bg-muted px-1 rounded text-xs">{'{{t.key}}'}</code> in your HTML.
          </p>

          <div className="flex gap-2">
            <Select
              value={langLocale}
              onChange={(e) => setLangLocale(e.target.value)}
              options={[
                { value: 'en', label: 'English' },
                { value: 'de', label: 'Deutsch' },
                { value: 'fr', label: 'Francais' },
                { value: 'es', label: 'Espanol' },
              ]}
            />
          </div>

          {/* Existing strings */}
          <div className="max-h-80 overflow-y-auto space-y-1">
            {(langStrings ?? []).map((ls) => (
              <div key={ls.id || ls.string_key} className="flex items-center gap-2 group rounded-md p-2 hover:bg-muted/50">
                <code className="text-xs font-mono text-indigo-400 min-w-[140px] shrink-0 truncate" title={ls.string_key}>
                  {ls.string_key}
                </code>
                <Input
                  defaultValue={ls.value}
                  className="text-xs h-8 flex-1"
                  onBlur={(e) => {
                    if (e.target.value !== ls.value) {
                      upsertLangMutation.mutate({
                        string_key: ls.string_key,
                        locale: langLocale,
                        value: e.target.value,
                      })
                    }
                  }}
                />
                <button
                  onClick={() => deleteLangMutation.mutate({ key: ls.string_key, locale: langLocale })}
                  className="opacity-0 group-hover:opacity-100 text-muted-foreground hover:text-destructive transition-opacity cursor-pointer"
                >
                  <X className="h-3.5 w-3.5" />
                </button>
              </div>
            ))}
            {(langStrings ?? []).length === 0 && (
              <p className="text-sm text-muted-foreground text-center py-4">No strings defined for this locale</p>
            )}
          </div>

          {/* Add new string */}
          <div className="flex gap-2 pt-2 border-t border-border">
            <Input
              value={newStringKey}
              onChange={(e) => setNewStringKey(e.target.value)}
              placeholder="login.welcome"
              className="text-xs font-mono flex-1"
            />
            <Input
              value={newStringValue}
              onChange={(e) => setNewStringValue(e.target.value)}
              placeholder="Welcome back!"
              className="text-xs flex-1"
            />
            <Button
              size="sm"
              variant="outline"
              disabled={!newStringKey || !newStringValue}
              onClick={() => {
                upsertLangMutation.mutate({
                  string_key: newStringKey,
                  locale: langLocale,
                  value: newStringValue,
                })
              }}
            >
              <Plus className="h-3 w-3" />
            </Button>
          </div>

          <DialogFooter>
            <Button onClick={() => setLangOpen(false)}>Done</Button>
          </DialogFooter>
        </div>
      </Dialog>

      {/* ─── Delete Confirmation ───────────────────────────── */}
      <ConfirmDialog
        open={deleteOpen}
        onClose={() => setDeleteOpen(false)}
        title="Delete Template"
        description={`Are you sure you want to delete "${selectedTemplate?.name}"? This action cannot be undone.`}
        onConfirm={() => selectedTemplate && deleteMutation.mutate(selectedTemplate.id)}
        loading={deleteMutation.isPending}
        destructive
      />
    </div>
  )
}

// ─── Template Card Component ──────────────────────────────────

function TemplateCard({
  template,
  isSelected,
  onClick,
}: {
  template: PageTemplate
  isSelected: boolean
  onClick: () => void
}) {
  return (
    <Card
      className={`cursor-pointer transition-colors ${
        isSelected ? 'border-primary bg-primary/5' : 'hover:border-primary/50'
      }`}
      onClick={onClick}
    >
      <CardContent className="p-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3 min-w-0">
            <div className="shrink-0">
              {template.is_default ? (
                <Lock className="h-4 w-4 text-amber-500" />
              ) : (
                <FileCode className="h-4 w-4 text-primary" />
              )}
            </div>
            <div className="min-w-0">
              <p className="font-medium text-sm truncate">{template.name}</p>
              <p className="text-[10px] text-muted-foreground">
                {pageTypeLabels[template.page_type as PageTemplateType] || template.page_type}
              </p>
            </div>
          </div>
          <div className="flex items-center gap-1 shrink-0">
            {template.is_default ? (
              <Badge variant="muted" className="text-[10px]">Default</Badge>
            ) : (
              <Badge variant={template.is_active ? 'success' : 'muted'} className="text-[10px]">
                {template.is_active ? 'Active' : 'Draft'}
              </Badge>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
