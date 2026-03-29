import { useState, useMemo } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { api } from '@/lib/api'
import { useI18n } from '@/lib/i18n'
import { useUIStore } from '@/stores/ui'
import { DataTable, type Column } from '@/components/shared/data-table'
import { PageHeader } from '@/components/shared/page-header'
import { Pagination } from '@/components/shared/pagination'
import { ConfirmDialog } from '@/components/shared/confirm-dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { Avatar } from '@/components/ui/avatar'
import { DropdownMenu, DropdownMenuItem, DropdownMenuSeparator } from '@/components/ui/dropdown-menu'
import {
  Dialog,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  DialogClose,
} from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import {
  Plus,
  Search,
  MoreHorizontal,
  Download,
  Ban,
  Trash2,
  Users as UsersIcon,
} from 'lucide-react'
import { formatDate, formatRelativeTime } from '@/lib/utils'
import type { User } from '@/types'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'

const statusBadgeVariant: Record<string, 'success' | 'destructive' | 'warning' | 'muted'> = {
  active: 'success',
  blocked: 'destructive',
  inactive: 'warning',
  pending: 'muted',
}

const createUserSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  email: z.string().email('Valid email required'),
  password: z.string().min(8, 'Min 8 characters'),
})

type CreateUserForm = z.infer<typeof createUserSchema>

export default function UsersPage() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { addToast } = useUIStore()
  const { t } = useI18n()

  const [search, setSearch] = useState('')
  const [statusFilter, setStatusFilter] = useState('')
  const [cursor, setCursor] = useState<string | undefined>()
  const [prevCursors, setPrevCursors] = useState<string[]>([])
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set())
  const [bulkAction, setBulkAction] = useState<'block' | 'delete' | null>(null)
  const [createOpen, setCreateOpen] = useState(false)

  const { data, isLoading } = useQuery({
    queryKey: ['users', { search, status: statusFilter, cursor }],
    queryFn: () =>
      api.getUsers({
        search: search || undefined,
        status: statusFilter || undefined,
        cursor,
        limit: 20,
      }),
  })

  const createMutation = useMutation({
    mutationFn: (data: CreateUserForm) => api.createUser(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] })
      setCreateOpen(false)
      addToast({ title: 'User created', variant: 'success' })
    },
    onError: () => {
      addToast({ title: 'Failed to create user', variant: 'error' })
    },
  })

  const bulkBlockMutation = useMutation({
    mutationFn: (ids: string[]) => api.bulkBlockUsers(ids),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] })
      setSelectedIds(new Set())
      setBulkAction(null)
      addToast({ title: 'Users blocked', variant: 'success' })
    },
  })

  const bulkDeleteMutation = useMutation({
    mutationFn: (ids: string[]) => api.bulkDeleteUsers(ids),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] })
      setSelectedIds(new Set())
      setBulkAction(null)
      addToast({ title: 'Users deleted', variant: 'success' })
    },
  })

  const handleExport = async () => {
    try {
      const blob = await api.exportUsers({ search, status: statusFilter })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = 'users-export.csv'
      a.click()
      URL.revokeObjectURL(url)
      addToast({ title: 'Export started', variant: 'success' })
    } catch {
      addToast({ title: 'Export failed', variant: 'error' })
    }
  }

  const columns: Column<User>[] = useMemo(
    () => [
      {
        key: 'user',
        header: 'User',
        render: (user) => (
          <div className="flex items-center gap-3">
            <Avatar name={user.name} src={user.avatar_url} size="sm" />
            <div>
              <p className="font-medium">{user.name}</p>
              <p className="text-xs text-muted-foreground">{user.email}</p>
            </div>
          </div>
        ),
      },
      {
        key: 'status',
        header: 'Status',
        sortable: true,
        render: (user) => (
          <Badge variant={statusBadgeVariant[user.status]}>{user.status}</Badge>
        ),
      },
      {
        key: 'last_login',
        header: 'Last Login',
        sortable: true,
        render: (user) => (
          <span className="text-muted-foreground">
            {user.last_login ? formatRelativeTime(user.last_login) : 'Never'}
          </span>
        ),
      },
      {
        key: 'created_at',
        header: 'Created',
        sortable: true,
        render: (user) => (
          <span className="text-muted-foreground">{formatDate(user.created_at)}</span>
        ),
      },
      {
        key: 'actions',
        header: '',
        className: 'w-12',
        render: (user) => (
          <div onClick={(e) => e.stopPropagation()}>
            <DropdownMenu
              trigger={
                <Button variant="ghost" size="icon-sm">
                  <MoreHorizontal className="h-4 w-4" />
                </Button>
              }
            >
              <DropdownMenuItem onClick={() => navigate(`/users/${user.id}`)}>
                View details
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem destructive onClick={() => {
                setSelectedIds(new Set([user.id]))
                setBulkAction('block')
              }}>
                Block user
              </DropdownMenuItem>
              <DropdownMenuItem destructive onClick={() => {
                setSelectedIds(new Set([user.id]))
                setBulkAction('delete')
              }}>
                Delete user
              </DropdownMenuItem>
            </DropdownMenu>
          </div>
        ),
      },
    ],
    [navigate]
  )

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<CreateUserForm>({
    resolver: zodResolver(createUserSchema),
  })

  const onCreateSubmit = (data: CreateUserForm) => {
    createMutation.mutate(data)
  }

  return (
    <div>
      <PageHeader
        title={t('users.title')}
        description={t('users.description')}
        breadcrumbs={[{ label: t('dashboard.title'), href: '/' }, { label: t('users.title') }]}
        actions={
          <Button onClick={() => { setCreateOpen(true); reset() }}>
            <Plus className="mr-2 h-4 w-4" />
            Create User
          </Button>
        }
      />

      <div className="flex flex-col sm:flex-row gap-3 mb-4">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search by name or email..."
            value={search}
            onChange={(e) => {
              setSearch(e.target.value)
              setCursor(undefined)
              setPrevCursors([])
            }}
            className="pl-9"
          />
        </div>
        <Select
          value={statusFilter}
          onChange={(e) => {
            setStatusFilter(e.target.value)
            setCursor(undefined)
            setPrevCursors([])
          }}
          options={[
            { value: '', label: 'All statuses' },
            { value: 'active', label: 'Active' },
            { value: 'blocked', label: 'Blocked' },
            { value: 'inactive', label: 'Inactive' },
            { value: 'pending', label: 'Pending' },
          ]}
          className="w-40"
        />

        {selectedIds.size > 0 && (
          <div className="flex items-center gap-2 ml-auto">
            <span className="text-sm text-muted-foreground">{selectedIds.size} selected</span>
            <Button variant="outline" size="sm" onClick={() => setBulkAction('block')}>
              <Ban className="mr-1 h-3 w-3" />
              Block
            </Button>
            <Button variant="destructive" size="sm" onClick={() => setBulkAction('delete')}>
              <Trash2 className="mr-1 h-3 w-3" />
              Delete
            </Button>
          </div>
        )}

        <Button variant="outline" onClick={handleExport} className="ml-auto">
          <Download className="mr-2 h-4 w-4" />
          Export
        </Button>
      </div>

      <DataTable
        columns={columns}
        data={data?.data ?? []}
        isLoading={isLoading}
        selectable
        selectedIds={selectedIds}
        onSelectionChange={setSelectedIds}
        getRowId={(u) => u.id}
        onRowClick={(u) => navigate(`/users/${u.id}`)}
        emptyMessage="No users found"
        emptyIcon={<UsersIcon className="h-10 w-10" />}
      />

      <Pagination
        hasMore={data?.has_more ?? false}
        hasPrevious={prevCursors.length > 0}
        onNext={() => {
          if (data?.cursor) {
            setPrevCursors([...prevCursors, cursor ?? ''])
            setCursor(data.cursor)
          }
        }}
        onPrevious={() => {
          const prev = [...prevCursors]
          const prevCursor = prev.pop()
          setPrevCursors(prev)
          setCursor(prevCursor || undefined)
        }}
        loading={isLoading}
      />

      {/* Bulk action confirm */}
      <ConfirmDialog
        open={bulkAction !== null}
        onClose={() => setBulkAction(null)}
        onConfirm={() => {
          const ids = Array.from(selectedIds)
          if (bulkAction === 'block') bulkBlockMutation.mutate(ids)
          if (bulkAction === 'delete') bulkDeleteMutation.mutate(ids)
        }}
        title={bulkAction === 'block' ? 'Block Users' : 'Delete Users'}
        description={`Are you sure you want to ${bulkAction} ${selectedIds.size} user(s)? This action ${bulkAction === 'delete' ? 'cannot be undone' : 'can be reversed later'}.`}
        confirmLabel={bulkAction === 'block' ? 'Block' : 'Delete'}
        variant="destructive"
        loading={bulkBlockMutation.isPending || bulkDeleteMutation.isPending}
      />

      {/* Create user dialog */}
      <Dialog open={createOpen} onClose={() => setCreateOpen(false)}>
        <DialogClose onClose={() => setCreateOpen(false)} />
        <DialogHeader>
          <DialogTitle>Create User</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit(onCreateSubmit)} className="mt-4 space-y-4">
          <div className="space-y-2">
            <Label>Name</Label>
            <Input {...register('name')} error={errors.name?.message} />
          </div>
          <div className="space-y-2">
            <Label>Email</Label>
            <Input type="email" {...register('email')} error={errors.email?.message} />
          </div>
          <div className="space-y-2">
            <Label>Password</Label>
            <Input type="password" {...register('password')} error={errors.password?.message} />
          </div>
          <DialogFooter>
            <Button variant="outline" type="button" onClick={() => setCreateOpen(false)}>
              Cancel
            </Button>
            <Button type="submit" loading={createMutation.isPending}>
              Create
            </Button>
          </DialogFooter>
        </form>
      </Dialog>
    </div>
  )
}
