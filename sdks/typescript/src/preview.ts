import type { LanguageString, CustomFieldConfig } from './types.js'

const defaultSampleData: Record<string, string> = {
  '{{user.name}}': 'John Doe',
  '{{user.email}}': 'john@example.com',
  '{{user.initials}}': 'JD',
  '{{application.name}}': 'My Application',
  '{{tenant.name}}': 'Acme Corp',
  '{{code}}': '847293',
  '{{link}}': '#',
  '{{error}}': 'The session has expired. Please try again.',
}

export function renderCustomFieldsHTML(fields: CustomFieldConfig[]): string {
  return fields
    .map((f) => {
      const req = f.required ? ' <span class="required">*</span>' : ''
      let input: string
      switch (f.type) {
        case 'select':
          input = `<select>${(f.options ?? []).map((o) => `<option>${o}</option>`).join('')}</select>`
          break
        case 'textarea':
          input = `<textarea placeholder="${f.placeholder ?? ''}"></textarea>`
          break
        case 'checkbox':
          return `<div class="custom-field"><label><input type="checkbox" /> ${f.label}${req}</label></div>`
        default:
          input = `<input type="${f.type}" placeholder="${f.placeholder ?? ''}" />`
      }
      return `<div class="custom-field"><label>${f.label}${req}</label>${input}</div>`
    })
    .join('\n      ')
}

export interface RenderOptions {
  strings?: LanguageString[]
  sampleData?: Record<string, string>
  customFields?: CustomFieldConfig[]
  tokensCSS?: string
}

export function renderPreview(
  htmlContent: string,
  cssContent: string,
  options: RenderOptions = {}
): string {
  let html = htmlContent

  // 1. Replace built-in variables with sample data
  const data = { ...defaultSampleData, ...options.sampleData }

  if (options.customFields && options.customFields.length > 0) {
    const fieldsHTML = renderCustomFieldsHTML(options.customFields)
    data['{{custom_fields}}'] = fieldsHTML
    data['{{profile_fields}}'] = fieldsHTML
  }

  for (const [key, value] of Object.entries(data)) {
    html = html.replaceAll(key, value)
  }

  // 2. Replace {{t.xxx}} language strings
  if (options.strings) {
    for (const ls of options.strings) {
      html = html.replaceAll(`{{t.${ls.string_key}}}`, ls.value)
    }
  }

  // 3. Fallback: unreplaced {{t.xxx}} → [xxx]
  html = html.replace(/\{\{t\.([^}]+)\}\}/g, '[$1]')

  // 4. Inject CSS before </head>
  let css = cssContent
  if (options.tokensCSS) {
    css = options.tokensCSS + '\n' + css
  }
  html = html.replace('</head>', `<style>${css}</style></head>`)

  return html
}
