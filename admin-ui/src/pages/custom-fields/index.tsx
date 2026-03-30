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
import { Checkbox } from '@/components/ui/checkbox'
import {
  Dialog,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
  DialogClose,
} from '@/components/ui/dialog'
import { Plus, FormInput, Pencil, Trash2, GripVertical } from 'lucide-react'
import { formatDate } from '@/lib/utils'
import type { CustomFieldDefinition, CustomFieldType } from '@/types'

const fieldTypes: { value: CustomFieldType; label: string }[] = [
  { value: 'text', label: 'Text' },
  { value: 'textarea', label: 'Textarea' },
  { value: 'number', label: 'Number' },
  { value: 'checkbox', label: 'Checkbox' },
  { value: 'date', label: 'Date' },
  { value: 'select', label: 'Select' },
  { value: 'url', label: 'URL' },
  { value: 'email', label: 'Email' },
  { value: 'tel', label: 'Phone' },
]

const visibilityOptions = [
  { value: 'both', label: 'Registration & Profile' },
  { value: 'registration', label: 'Registration only' },
  { value: 'profile', label: 'Profile only' },
]

interface FieldFormData {
  name: string
  label: string
  field_type: CustomFieldType
  placeholder: string
  description: string
  options: string
  required: boolean
  visible_on: string
  position: number
}

const emptyForm: FieldFormData = {
  name: '',
  label: '',
  field_type: 'text',
  placeholder: '',
  description: '',
  options: '',
  required: false,
  visible_on: 'both',
  position: 0,
}

export default function CustomFieldsPage() {
  const queryClient = useQueryClient()
  const { addToast } = useUIStore()

  const [createOpen, setCreateOpen] = useState(false)
  const [editField, setEditField] = useState<CustomFieldDefinition | null>(null)
  const [deleteField, setDeleteField] = useState<CustomFieldDefinition | null>(null)
  const [formData, setFormData] = useState<FieldFormData>(emptyForm)

  const { data: fields, isLoading } = useQuery({
    queryKey: ['custom-fields'],
    queryFn: () => api.getCustomFields(),
  })

  const createMutation = useMutation({
    mutationFn: (data: FieldFormData) =>
      api.createCustomField({
        ...data,
        options: data.options ? data.options.split('\n').filter(Boolean) : undefined,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['custom-fields'] })
      setCreateOpen(false)
      addToast({ title: 'Custom field created', variant: 'success' })
    },
    onError: () => addToast({ title: 'Failed to create custom field', variant: 'error' }),
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: FieldFormData }) =>
      api.updateCustomField(id, {
        ...data,
        options: data.options ? data.options.split('\n').filter(Boolean) : undefined,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['custom-fields'] })
      setEditField(null)
      addToast({ title: 'Custom field updated', variant: 'success' })
    },
    onError: () => addToast({ title: 'Failed to update custom field', variant: 'error' }),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.deleteCustomField(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['custom-fields'] })
      setDeleteField(null)
      addToast({ title: 'Custom field deleted', variant: 'success' })
    },
  })

  const openCreate = () => {
    setFormData({ ...emptyForm, position: (fields?.length ?? 0) + 1 })
    setCreateOpen(true)
  }

  const openEdit = (field: CustomFieldDefinition) => {
    setFormData({
      name: field.name,
      label: field.label,
      field_type: field.field_type,
      placeholder: field.placeholder || '',
      description: field.description || '',
      options: field.options?.join('\n') || '',
      required: field.required,
      visible_on: field.visible_on,
      position: field.position,
    })
    setEditField(field)
  }

  const columns: Column<CustomFieldDefinition>[] = useMemo(() => [
    {
      key: 'position',
      header: '',
      className: 'w-8',
      render: () => <GripVertical className="h-4 w-4 text-muted-foreground" />,
    },
    {
      key: 'label',
      header: 'Field',
      render: (field) => (
        <div>
          <p className="font-medium">{field.label}</p>
          <p className="text-xs text-muted-foreground font-mono">{field.name}</p>
        </div>
      ),
    },
    {
      key: 'field_type',
      header: 'Type',
      render: (field) => (
        <Badge variant="muted">{field.field_type}</Badge>
      ),
    },
    {
      key: 'visible_on',
      header: 'Visible On',
      render: (field) => (
        <span className="text-sm text-muted-foreground capitalize">{field.visible_on}</span>
      ),
    },
    {
      key: 'required',
      header: 'Required',
      render: (field) => (
        <Badge variant={field.required ? 'default' : 'muted'}>
          {field.required ? 'Yes' : 'No'}
        </Badge>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      render: (field) => (
        <Badge variant={field.is_active ? 'success' : 'destructive'}>
          {field.is_active ? 'Active' : 'Inactive'}
        </Badge>
      ),
    },
    {
      key: 'created_at',
      header: 'Created',
      render: (field) => (
        <span className="text-xs text-muted-foreground">{formatDate(field.created_at)}</span>
      ),
    },
    {
      key: 'actions',
      header: '',
      className: 'w-20',
      render: (field) => (
        <div className="flex gap-1">
          <Button variant="ghost" size="icon-sm" onClick={(e) => { e.stopPropagation(); openEdit(field) }}>
            <Pencil className="h-3 w-3" />
          </Button>
          <Button variant="ghost" size="icon-sm" onClick={(e) => { e.stopPropagation(); setDeleteField(field) }}>
            <Trash2 className="h-3 w-3 text-destructive" />
          </Button>
        </div>
      ),
    },
  ], [])

  const showOptions = formData.field_type === 'select'

  const fieldForm = (
    <div className="mt-4 space-y-4">
      <div className="grid gap-4 sm:grid-cols-2">
        <div className="space-y-2">
          <Label>Name (identifier)</Label>
          <Input
            value={formData.name}
            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
            placeholder="company_name"
            disabled={!!editField}
          />
        </div>
        <div className="space-y-2">
          <Label>Label</Label>
          <Input
            value={formData.label}
            onChange={(e) => setFormData({ ...formData, label: e.target.value })}
            placeholder="Company Name"
          />
        </div>
      </div>

      <div className="grid gap-4 sm:grid-cols-2">
        <div className="space-y-2">
          <Label>Field Type</Label>
          <select
            className="flex h-9 w-full rounded-md border border-input bg-background px-3 py-1 text-sm shadow-sm"
            value={formData.field_type}
            onChange={(e) => setFormData({ ...formData, field_type: e.target.value as CustomFieldType })}
          >
            {fieldTypes.map((t) => (
              <option key={t.value} value={t.value}>{t.label}</option>
            ))}
          </select>
        </div>
        <div className="space-y-2">
          <Label>Visibility</Label>
          <select
            className="flex h-9 w-full rounded-md border border-input bg-background px-3 py-1 text-sm shadow-sm"
            value={formData.visible_on}
            onChange={(e) => setFormData({ ...formData, visible_on: e.target.value })}
          >
            {visibilityOptions.map((o) => (
              <option key={o.value} value={o.value}>{o.label}</option>
            ))}
          </select>
        </div>
      </div>

      <div className="grid gap-4 sm:grid-cols-2">
        <div className="space-y-2">
          <Label>Placeholder</Label>
          <Input
            value={formData.placeholder}
            onChange={(e) => setFormData({ ...formData, placeholder: e.target.value })}
            placeholder="Enter value..."
          />
        </div>
        <div className="space-y-2">
          <Label>Position</Label>
          <Input
            type="number"
            value={formData.position}
            onChange={(e) => setFormData({ ...formData, position: Number(e.target.value) })}
          />
        </div>
      </div>

      <div className="space-y-2">
        <Label>Description</Label>
        <Input
          value={formData.description}
          onChange={(e) => setFormData({ ...formData, description: e.target.value })}
          placeholder="Help text shown below the field"
        />
      </div>

      {showOptions && (
        <div className="space-y-2">
          <Label>Options (one per line)</Label>
          <textarea
            className="flex min-h-[80px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm shadow-sm"
            value={formData.options}
            onChange={(e) => setFormData({ ...formData, options: e.target.value })}
            placeholder={"Option A\nOption B\nOption C"}
          />
        </div>
      )}

      <div className="flex items-center gap-2">
        <Checkbox
          checked={formData.required}
          onCheckedChange={(checked) => setFormData({ ...formData, required: !!checked })}
        />
        <Label>Required field</Label>
      </div>
    </div>
  )

  return (
    <div>
      <PageHeader
        title="Custom Fields"
        description="Define custom user profile fields for registration and profile forms"
        breadcrumbs={[{ label: 'Dashboard', href: '/' }, { label: 'Custom Fields' }]}
        actions={
          <Button onClick={openCreate}>
            <Plus className="mr-2 h-4 w-4" />
            Add Field
          </Button>
        }
      />

      {!isLoading && (!fields || fields.length === 0) ? (
        <EmptyState
          icon={<FormInput className="h-12 w-12" />}
          title="No custom fields"
          description="Add custom fields to collect additional information from users during registration or on their profile."
          action={{ label: 'Add Field', onClick: openCreate }}
        />
      ) : (
        <DataTable
          columns={columns}
          data={fields ?? []}
          isLoading={isLoading}
          getRowId={(f) => f.id}
        />
      )}

      {/* Create */}
      <Dialog open={createOpen} onClose={() => setCreateOpen(false)} className="max-w-xl">
        <DialogClose onClose={() => setCreateOpen(false)} />
        <DialogHeader>
          <DialogTitle>Add Custom Field</DialogTitle>
          <DialogDescription>Define a new field for user profiles and registration.</DialogDescription>
        </DialogHeader>
        {fieldForm}
        <DialogFooter>
          <Button variant="outline" onClick={() => setCreateOpen(false)}>Cancel</Button>
          <Button
            onClick={() => createMutation.mutate(formData)}
            loading={createMutation.isPending}
            disabled={!formData.name || !formData.label}
          >
            Create Field
          </Button>
        </DialogFooter>
      </Dialog>

      {/* Edit */}
      <Dialog open={!!editField} onClose={() => setEditField(null)} className="max-w-xl">
        <DialogClose onClose={() => setEditField(null)} />
        <DialogHeader>
          <DialogTitle>Edit Custom Field</DialogTitle>
          <DialogDescription>Update field settings. The field name cannot be changed.</DialogDescription>
        </DialogHeader>
        {fieldForm}
        <DialogFooter>
          <Button variant="outline" onClick={() => setEditField(null)}>Cancel</Button>
          <Button
            onClick={() => editField && updateMutation.mutate({ id: editField.id, data: formData })}
            loading={updateMutation.isPending}
            disabled={!formData.label}
          >
            Save Changes
          </Button>
        </DialogFooter>
      </Dialog>

      <ConfirmDialog
        open={!!deleteField}
        onClose={() => setDeleteField(null)}
        onConfirm={() => deleteField && deleteMutation.mutate(deleteField.id)}
        title="Delete Custom Field"
        description={`Delete "${deleteField?.label}"? This will remove the field definition. Existing user data for this field will be preserved.`}
        confirmLabel="Delete"
        variant="destructive"
        loading={deleteMutation.isPending}
      />
    </div>
  )
}
