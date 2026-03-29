import { describe, it, expect } from 'vitest'
import { renderPreview, renderCustomFieldsHTML } from './preview.js'
import type { LanguageString, CustomFieldConfig } from './types.js'

describe('renderPreview', () => {
  const baseHTML = '<html><head></head><body><h1>{{user.name}}</h1><p>{{user.email}}</p></body></html>'
  const baseCSS = 'body { color: red; }'

  it('should replace built-in variables with sample data', () => {
    const result = renderPreview(baseHTML, baseCSS)
    expect(result).toContain('John Doe')
    expect(result).toContain('john@example.com')
    expect(result).not.toContain('{{user.name}}')
    expect(result).not.toContain('{{user.email}}')
  })

  it('should inject CSS before </head>', () => {
    const result = renderPreview(baseHTML, baseCSS)
    expect(result).toContain('<style>body { color: red; }</style></head>')
  })

  it('should replace {{application.name}} and {{tenant.name}}', () => {
    const html = '<div>{{application.name}} by {{tenant.name}}</div>'
    const result = renderPreview(html, '')
    expect(result).toContain('My Application')
    expect(result).toContain('Acme Corp')
  })

  it('should replace {{code}} and {{link}}', () => {
    const html = '<span>{{code}}</span><a href="{{link}}">Verify</a>'
    const result = renderPreview(html, '')
    expect(result).toContain('847293')
    expect(result).toContain('href="#"')
  })

  it('should replace {{error}}', () => {
    const html = '<div class="error">{{error}}</div>'
    const result = renderPreview(html, '')
    expect(result).toContain('The session has expired')
  })

  it('should replace {{user.initials}}', () => {
    const html = '<div>{{user.initials}}</div>'
    const result = renderPreview(html, '')
    expect(result).toContain('JD')
  })

  it('should use custom sample data when provided', () => {
    const result = renderPreview(baseHTML, '', {
      sampleData: { '{{user.name}}': 'Max Mustermann' },
    })
    expect(result).toContain('Max Mustermann')
  })

  it('should replace {{t.xxx}} with language strings', () => {
    const html = '<h1>{{t.login.title}}</h1><p>{{t.login.subtitle}}</p>'
    const strings: LanguageString[] = [
      { id: '1', tenant_id: '', string_key: 'login.title', locale: 'en', value: 'Welcome', created_at: '', updated_at: '' },
      { id: '2', tenant_id: '', string_key: 'login.subtitle', locale: 'en', value: 'Sign in', created_at: '', updated_at: '' },
    ]
    const result = renderPreview(html, '', { strings })
    expect(result).toContain('Welcome')
    expect(result).toContain('Sign in')
    expect(result).not.toContain('{{t.')
  })

  it('should fallback unreplaced {{t.xxx}} to [xxx]', () => {
    const html = '<p>{{t.missing.key}}</p>'
    const result = renderPreview(html, '')
    expect(result).toContain('[missing.key]')
    expect(result).not.toContain('{{t.')
  })

  it('should render custom fields for {{custom_fields}}', () => {
    const html = '<form>{{custom_fields}}</form>'
    const fields: CustomFieldConfig[] = [
      { label: 'Company', type: 'text', required: true, placeholder: 'Acme' },
      { label: 'Role', type: 'select', options: ['Dev', 'PM'] },
    ]
    const result = renderPreview(html, '', { customFields: fields })
    expect(result).toContain('Company')
    expect(result).toContain('required')
    expect(result).toContain('Acme')
    expect(result).toContain('Role')
    expect(result).toContain('<option>Dev</option>')
    expect(result).toContain('<option>PM</option>')
  })

  it('should render {{profile_fields}} same as {{custom_fields}}', () => {
    const html = '<form>{{profile_fields}}</form>'
    const fields: CustomFieldConfig[] = [
      { label: 'Phone', type: 'tel', placeholder: '+1 555' },
    ]
    const result = renderPreview(html, '', { customFields: fields })
    expect(result).toContain('Phone')
    expect(result).toContain('type="tel"')
  })

  it('should prepend tokens CSS when provided', () => {
    const html = '<html><head></head><body></body></html>'
    const result = renderPreview(html, 'body {}', {
      tokensCSS: ':root { --af-color-primary: #6366f1; }',
    })
    expect(result).toContain('--af-color-primary')
    expect(result).toContain('body {}')
    // Tokens should come before template CSS
    const tokensIdx = result.indexOf('--af-color-primary')
    const bodyIdx = result.indexOf('body {}')
    expect(tokensIdx).toBeLessThan(bodyIdx)
  })

  it('should handle empty HTML gracefully', () => {
    const result = renderPreview('', '')
    expect(result).toBe('')
  })

  it('should handle multiple occurrences of same variable', () => {
    const html = '{{user.name}} said hello to {{user.name}}'
    const result = renderPreview(html, '')
    expect(result).toBe('John Doe said hello to John Doe')
  })
})

describe('renderCustomFieldsHTML', () => {
  it('should render text input', () => {
    const fields: CustomFieldConfig[] = [
      { label: 'Name', type: 'text', placeholder: 'John' },
    ]
    const html = renderCustomFieldsHTML(fields)
    expect(html).toContain('type="text"')
    expect(html).toContain('placeholder="John"')
    expect(html).toContain('Name')
  })

  it('should render required indicator', () => {
    const fields: CustomFieldConfig[] = [
      { label: 'Email', type: 'email', required: true },
    ]
    const html = renderCustomFieldsHTML(fields)
    expect(html).toContain('required')
    expect(html).toContain('*')
  })

  it('should render select with options', () => {
    const fields: CustomFieldConfig[] = [
      { label: 'Country', type: 'select', options: ['US', 'DE', 'FR'] },
    ]
    const html = renderCustomFieldsHTML(fields)
    expect(html).toContain('<select>')
    expect(html).toContain('<option>US</option>')
    expect(html).toContain('<option>DE</option>')
    expect(html).toContain('<option>FR</option>')
  })

  it('should render textarea', () => {
    const fields: CustomFieldConfig[] = [
      { label: 'Bio', type: 'textarea', placeholder: 'Tell us about yourself' },
    ]
    const html = renderCustomFieldsHTML(fields)
    expect(html).toContain('<textarea')
    expect(html).toContain('Tell us about yourself')
  })

  it('should render checkbox', () => {
    const fields: CustomFieldConfig[] = [
      { label: 'Agree to terms', type: 'checkbox', required: true },
    ]
    const html = renderCustomFieldsHTML(fields)
    expect(html).toContain('type="checkbox"')
    expect(html).toContain('Agree to terms')
  })

  it('should render multiple field types', () => {
    const fields: CustomFieldConfig[] = [
      { label: 'Name', type: 'text' },
      { label: 'Phone', type: 'tel' },
      { label: 'Website', type: 'url' },
      { label: 'Age', type: 'number' },
      { label: 'DOB', type: 'date' },
    ]
    const html = renderCustomFieldsHTML(fields)
    expect(html).toContain('type="text"')
    expect(html).toContain('type="tel"')
    expect(html).toContain('type="url"')
    expect(html).toContain('type="number"')
    expect(html).toContain('type="date"')
  })

  it('should return empty string for empty array', () => {
    expect(renderCustomFieldsHTML([])).toBe('')
  })
})
