import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { api } from '@/lib/api'
import { useUIStore } from '@/stores/ui'
import { PageHeader } from '@/components/shared/page-header'
import { EmptyState } from '@/components/shared/empty-state'
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
  DialogDescription,
  DialogFooter,
  DialogClose,
} from '@/components/ui/dialog'
import { Plus, AppWindow, Globe, Smartphone, Server, ArrowRight } from 'lucide-react'
import { formatDate } from '@/lib/utils'
import type { Application, ApplicationType } from '@/types'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'

const appTypeConfig: Record<ApplicationType, { label: string; icon: React.ElementType; color: string }> = {
  spa: { label: 'Single Page App', icon: Globe, color: 'text-blue-400' },
  web: { label: 'Regular Web App', icon: AppWindow, color: 'text-green-400' },
  native: { label: 'Native App', icon: Smartphone, color: 'text-purple-400' },
  m2m: { label: 'Machine to Machine', icon: Server, color: 'text-orange-400' },
}

const createSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  type: z.enum(['spa', 'web', 'native', 'm2m']),
  description: z.string().optional(),
})

type CreateForm = z.infer<typeof createSchema>

export default function ApplicationsPage() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { addToast } = useUIStore()
  const [createOpen, setCreateOpen] = useState(false)
  const [wizardStep, setWizardStep] = useState(0)

  const { data: apps, isLoading } = useQuery({
    queryKey: ['applications'],
    queryFn: () => api.getApplications(),
  })

  const createMutation = useMutation({
    mutationFn: (data: CreateForm) => api.createApplication(data),
    onSuccess: (app) => {
      queryClient.invalidateQueries({ queryKey: ['applications'] })
      setCreateOpen(false)
      setWizardStep(0)
      addToast({ title: 'Application created', variant: 'success' })
      navigate(`/applications/${app.id}`)
    },
    onError: () => addToast({ title: 'Failed to create application', variant: 'error' }),
  })

  const {
    register,
    handleSubmit,
    watch,
    setValue,
    reset,
    formState: { errors },
  } = useForm<CreateForm>({
    resolver: zodResolver(createSchema),
    defaultValues: { type: 'spa' },
  })

  const selectedType = watch('type')

  return (
    <div>
      <PageHeader
        title="Applications"
        description="Manage your registered applications"
        breadcrumbs={[{ label: 'Dashboard', href: '/' }, { label: 'Applications' }]}
        actions={
          <Button onClick={() => { setCreateOpen(true); reset(); setWizardStep(0) }}>
            <Plus className="mr-2 h-4 w-4" />
            Create Application
          </Button>
        }
      />

      {isLoading ? (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 6 }).map((_, i) => (
            <Skeleton key={i} className="h-40" />
          ))}
        </div>
      ) : !apps || apps.length === 0 ? (
        <EmptyState
          icon={<AppWindow className="h-12 w-12" />}
          title="No applications yet"
          description="Create your first application to start integrating authentication."
          action={{ label: 'Create Application', onClick: () => setCreateOpen(true) }}
        />
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {apps.map((app) => {
            const config = appTypeConfig[app.type]
            const Icon = config.icon

            return (
              <Card
                key={app.id}
                className="cursor-pointer hover:border-primary/50 transition-colors"
                onClick={() => navigate(`/applications/${app.id}`)}
              >
                <CardContent className="p-6">
                  <div className="flex items-start justify-between">
                    <div className="flex items-center gap-3">
                      <div className="rounded-lg bg-muted p-2">
                        <Icon className={`h-5 w-5 ${config.color}`} />
                      </div>
                      <div>
                        <h3 className="font-semibold">{app.name}</h3>
                        <Badge variant="muted" className="mt-1">{config.label}</Badge>
                      </div>
                    </div>
                    <ArrowRight className="h-4 w-4 text-muted-foreground" />
                  </div>

                  {app.description && (
                    <p className="mt-3 text-sm text-muted-foreground line-clamp-2">{app.description}</p>
                  )}

                  <div className="mt-4 flex items-center justify-between text-xs text-muted-foreground">
                    <span>Client ID: {app.client_id.slice(0, 12)}...</span>
                    <span>{formatDate(app.created_at)}</span>
                  </div>

                  <div className="mt-2 flex items-center gap-2">
                    <Badge variant={app.is_active ? 'success' : 'muted'} className="text-[10px]">
                      {app.is_active ? 'Active' : 'Disabled'}
                    </Badge>
                  </div>
                </CardContent>
              </Card>
            )
          })}
        </div>
      )}

      {/* Create Application Wizard */}
      <Dialog open={createOpen} onClose={() => setCreateOpen(false)} className="max-w-xl">
        <DialogClose onClose={() => setCreateOpen(false)} />
        <DialogHeader>
          <DialogTitle>Create Application</DialogTitle>
          <DialogDescription>
            {wizardStep === 0
              ? 'Choose the type of application you want to create'
              : 'Configure your application details'}
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit((d) => createMutation.mutate(d))}>
          {wizardStep === 0 && (
            <div className="mt-4 grid grid-cols-2 gap-3">
              {(Object.entries(appTypeConfig) as [ApplicationType, typeof appTypeConfig[ApplicationType]][]).map(
                ([type, config]) => {
                  const Icon = config.icon
                  return (
                    <button
                      type="button"
                      key={type}
                      onClick={() => {
                        setValue('type', type)
                        setWizardStep(1)
                      }}
                      className={`flex flex-col items-center gap-2 rounded-lg border p-4 text-center transition-colors hover:border-primary cursor-pointer ${
                        selectedType === type ? 'border-primary bg-primary/5' : ''
                      }`}
                    >
                      <Icon className={`h-8 w-8 ${config.color}`} />
                      <span className="text-sm font-medium">{config.label}</span>
                    </button>
                  )
                }
              )}
            </div>
          )}

          {wizardStep === 1 && (
            <div className="mt-4 space-y-4">
              <div className="flex items-center gap-2 rounded-lg bg-muted p-3">
                {(() => {
                  const config = appTypeConfig[selectedType]
                  const Icon = config.icon
                  return (
                    <>
                      <Icon className={`h-5 w-5 ${config.color}`} />
                      <span className="text-sm font-medium">{config.label}</span>
                    </>
                  )
                })()}
                <button
                  type="button"
                  onClick={() => setWizardStep(0)}
                  className="ml-auto text-xs text-primary hover:underline cursor-pointer"
                >
                  Change
                </button>
              </div>

              <div className="space-y-2">
                <Label>Application Name</Label>
                <Input {...register('name')} placeholder="My Application" error={errors.name?.message} />
              </div>
              <div className="space-y-2">
                <Label>Description (optional)</Label>
                <Input {...register('description')} placeholder="A brief description" />
              </div>
            </div>
          )}

          {wizardStep === 1 && (
            <DialogFooter>
              <Button variant="outline" type="button" onClick={() => setWizardStep(0)}>
                Back
              </Button>
              <Button type="submit" loading={createMutation.isPending}>
                Create Application
              </Button>
            </DialogFooter>
          )}
        </form>
      </Dialog>
    </div>
  )
}
