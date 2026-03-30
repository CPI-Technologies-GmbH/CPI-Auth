import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { useUIStore } from '@/stores/ui'
import { PageHeader } from '@/components/shared/page-header'
import { CodeEditor } from '@/components/shared/code-editor'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select } from '@/components/ui/select'
import { Slider } from '@/components/ui/slider'
import { Save, Upload } from 'lucide-react'
import type { BrandingConfig } from '@/types'

const fontOptions = [
  { value: 'Inter', label: 'Inter' },
  { value: 'Roboto', label: 'Roboto' },
  { value: 'Open Sans', label: 'Open Sans' },
  { value: 'Lato', label: 'Lato' },
  { value: 'Poppins', label: 'Poppins' },
  { value: 'Montserrat', label: 'Montserrat' },
  { value: 'Source Sans Pro', label: 'Source Sans Pro' },
  { value: 'Nunito', label: 'Nunito' },
  { value: 'Raleway', label: 'Raleway' },
  { value: 'DM Sans', label: 'DM Sans' },
]

const layoutOptions = [
  { value: 'centered', label: 'Centered' },
  { value: 'split-screen', label: 'Split Screen' },
  { value: 'sidebar', label: 'Sidebar' },
]

const defaultBranding: BrandingConfig = {
  primary_color: '#6366f1',
  secondary_color: '#1e1e2e',
  background_color: '#0a0a0f',
  text_color: '#fafafa',
  font_family: 'Inter',
  border_radius: 8,
  layout_mode: 'centered',
  custom_css: '',
}

function ColorPicker({
  label,
  value,
  onChange,
}: {
  label: string
  value: string
  onChange: (val: string) => void
}) {
  return (
    <div className="space-y-2">
      <Label>{label}</Label>
      <div className="flex items-center gap-2">
        <div className="relative">
          <input
            type="color"
            value={value}
            onChange={(e) => onChange(e.target.value)}
            className="h-10 w-10 cursor-pointer rounded-md border border-input bg-transparent p-0.5"
          />
        </div>
        <Input
          value={value}
          onChange={(e) => onChange(e.target.value)}
          className="font-mono text-xs flex-1"
          maxLength={7}
        />
      </div>
    </div>
  )
}

function LoginPreview({ branding }: { branding: BrandingConfig }) {
  return (
    <div
      className="rounded-lg border overflow-hidden"
      style={{ backgroundColor: branding.background_color, fontFamily: branding.font_family }}
    >
      {branding.layout_mode === 'split-screen' ? (
        <div className="flex h-[400px]">
          <div
            className="w-1/2 flex items-center justify-center"
            style={{ backgroundColor: branding.primary_color }}
          >
            <div className="text-center text-white p-6">
              {branding.logo_url ? (
                <img src={branding.logo_url} alt="Logo" className="h-12 mx-auto mb-4" />
              ) : (
                <div className="h-12 w-12 rounded-xl bg-white/20 mx-auto mb-4" />
              )}
              <h2 className="text-xl font-bold">Welcome Back</h2>
              <p className="text-sm opacity-80 mt-1">Sign in to continue</p>
            </div>
          </div>
          <div className="w-1/2 flex items-center justify-center p-6">
            <div className="w-full max-w-xs space-y-4">
              <div className="space-y-1.5">
                <div className="text-xs font-medium" style={{ color: branding.text_color }}>Email</div>
                <div className="h-9 rounded-md border" style={{ borderRadius: `${branding.border_radius}px` }} />
              </div>
              <div className="space-y-1.5">
                <div className="text-xs font-medium" style={{ color: branding.text_color }}>Password</div>
                <div className="h-9 rounded-md border" style={{ borderRadius: `${branding.border_radius}px` }} />
              </div>
              <div
                className="h-9 rounded-md flex items-center justify-center text-xs text-white font-medium"
                style={{
                  backgroundColor: branding.primary_color,
                  borderRadius: `${branding.border_radius}px`,
                }}
              >
                Sign In
              </div>
            </div>
          </div>
        </div>
      ) : branding.layout_mode === 'sidebar' ? (
        <div className="flex h-[400px]">
          <div className="w-60 border-r p-6" style={{ backgroundColor: branding.secondary_color }}>
            {branding.logo_url ? (
              <img src={branding.logo_url} alt="Logo" className="h-8 mb-6" />
            ) : (
              <div className="h-8 w-8 rounded-lg bg-white/10 mb-6" />
            )}
            <div className="space-y-2">
              <div className="h-3 w-20 rounded bg-white/10" />
              <div className="h-3 w-16 rounded bg-white/10" />
            </div>
          </div>
          <div className="flex-1 flex items-center justify-center p-6">
            <div className="w-full max-w-xs space-y-4">
              <h2 className="text-lg font-bold" style={{ color: branding.text_color }}>Sign In</h2>
              <div className="space-y-1.5">
                <div className="text-xs font-medium" style={{ color: branding.text_color }}>Email</div>
                <div className="h-9 rounded-md border" style={{ borderRadius: `${branding.border_radius}px` }} />
              </div>
              <div className="space-y-1.5">
                <div className="text-xs font-medium" style={{ color: branding.text_color }}>Password</div>
                <div className="h-9 rounded-md border" style={{ borderRadius: `${branding.border_radius}px` }} />
              </div>
              <div
                className="h-9 rounded-md flex items-center justify-center text-xs text-white font-medium"
                style={{
                  backgroundColor: branding.primary_color,
                  borderRadius: `${branding.border_radius}px`,
                }}
              >
                Sign In
              </div>
            </div>
          </div>
        </div>
      ) : (
        <div className="flex items-center justify-center h-[400px] p-6">
          <div className="w-full max-w-sm space-y-6 text-center">
            {branding.logo_url ? (
              <img src={branding.logo_url} alt="Logo" className="h-10 mx-auto" />
            ) : (
              <div
                className="h-10 w-10 rounded-xl mx-auto"
                style={{ backgroundColor: branding.primary_color }}
              />
            )}
            <h2 className="text-xl font-bold" style={{ color: branding.text_color }}>
              Welcome Back
            </h2>
            <div className="space-y-3 text-left">
              <div className="space-y-1.5">
                <div className="text-xs font-medium" style={{ color: branding.text_color }}>Email</div>
                <div className="h-9 rounded-md border" style={{ borderRadius: `${branding.border_radius}px` }} />
              </div>
              <div className="space-y-1.5">
                <div className="text-xs font-medium" style={{ color: branding.text_color }}>Password</div>
                <div className="h-9 rounded-md border" style={{ borderRadius: `${branding.border_radius}px` }} />
              </div>
              <div
                className="h-9 rounded-md flex items-center justify-center text-xs text-white font-medium"
                style={{
                  backgroundColor: branding.primary_color,
                  borderRadius: `${branding.border_radius}px`,
                }}
              >
                Sign In
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default function BrandingPage() {
  const queryClient = useQueryClient()
  const { addToast } = useUIStore()

  const { data: settings } = useQuery({
    queryKey: ['settings'],
    queryFn: () => api.getSettings(),
  })

  const [branding, setBranding] = useState<BrandingConfig>(defaultBranding)

  useEffect(() => {
    if (settings?.branding) {
      setBranding({ ...defaultBranding, ...settings.branding })
    }
  }, [settings])

  const saveMutation = useMutation({
    mutationFn: (data: Partial<BrandingConfig>) => api.updateBranding(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['settings'] })
      addToast({ title: 'Branding saved', variant: 'success' })
    },
    onError: () => addToast({ title: 'Failed to save branding', variant: 'error' }),
  })

  const updateField = <K extends keyof BrandingConfig>(key: K, value: BrandingConfig[K]) => {
    setBranding((prev) => ({ ...prev, [key]: value }))
  }

  return (
    <div>
      <PageHeader
        title="Branding"
        description="Customize the look and feel of your login experience"
        breadcrumbs={[{ label: 'Dashboard', href: '/' }, { label: 'Branding' }]}
        actions={
          <Button onClick={() => saveMutation.mutate(branding)} loading={saveMutation.isPending}>
            <Save className="mr-2 h-4 w-4" />
            Save Changes
          </Button>
        }
      />

      <div className="grid gap-6 lg:grid-cols-2">
        {/* Editor Controls */}
        <div className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Colors</CardTitle>
            </CardHeader>
            <CardContent className="grid gap-4 sm:grid-cols-2">
              <ColorPicker
                label="Primary Color"
                value={branding.primary_color}
                onChange={(v) => updateField('primary_color', v)}
              />
              <ColorPicker
                label="Secondary Color"
                value={branding.secondary_color}
                onChange={(v) => updateField('secondary_color', v)}
              />
              <ColorPicker
                label="Background Color"
                value={branding.background_color}
                onChange={(v) => updateField('background_color', v)}
              />
              <ColorPicker
                label="Text Color"
                value={branding.text_color}
                onChange={(v) => updateField('text_color', v)}
              />
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Logo</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-4 sm:grid-cols-2">
                <div className="space-y-2">
                  <Label>Logo (Light)</Label>
                  <div className="flex items-center gap-2">
                    <Input
                      value={branding.logo_url || ''}
                      onChange={(e) => updateField('logo_url', e.target.value)}
                      placeholder="URL or upload"
                    />
                    <Button variant="outline" size="icon">
                      <Upload className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
                <div className="space-y-2">
                  <Label>Logo (Dark)</Label>
                  <div className="flex items-center gap-2">
                    <Input
                      value={branding.logo_dark_url || ''}
                      onChange={(e) => updateField('logo_dark_url', e.target.value)}
                      placeholder="URL or upload"
                    />
                    <Button variant="outline" size="icon">
                      <Upload className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Typography & Layout</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label>Font Family</Label>
                <Select
                  value={branding.font_family}
                  onChange={(e) => updateField('font_family', e.target.value)}
                  options={fontOptions}
                />
              </div>
              <div className="space-y-2">
                <Label>Border Radius: {branding.border_radius}px</Label>
                <Slider
                  value={branding.border_radius}
                  onValueChange={(v) => updateField('border_radius', v)}
                  min={0}
                  max={24}
                  step={1}
                />
              </div>
              <div className="space-y-2">
                <Label>Layout Mode</Label>
                <Select
                  value={branding.layout_mode}
                  onChange={(e) => updateField('layout_mode', e.target.value as BrandingConfig['layout_mode'])}
                  options={layoutOptions}
                />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Custom CSS</CardTitle>
            </CardHeader>
            <CardContent>
              <CodeEditor
                value={branding.custom_css || ''}
                onChange={(v) => updateField('custom_css', v)}
                language="css"
                height="150px"
                placeholder="/* Add your custom CSS here */"
              />
            </CardContent>
          </Card>
        </div>

        {/* Live Preview */}
        <div className="lg:sticky lg:top-24">
          <Card>
            <CardHeader>
              <CardTitle>Live Preview</CardTitle>
            </CardHeader>
            <CardContent>
              <LoginPreview branding={branding} />
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}
