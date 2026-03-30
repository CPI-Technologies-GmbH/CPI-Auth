import { readFileSync, existsSync } from 'node:fs'
import { parse as parseYAML } from 'yaml'
import { APIClient } from './client.js'
import type { ProjectConfig, SyncDiff, PageTemplate, LanguageString } from './types.js'

export function loadConfig(path: string): ProjectConfig {
  if (!existsSync(path)) throw new Error(`Config not found: ${path}`)
  return parseYAML(readFileSync(path, 'utf8'))
}

export function loadStringFile(path: string): Record<string, string> {
  if (!existsSync(path)) return {}
  return parseYAML(readFileSync(path, 'utf8')) ?? {}
}

function flattenStrings(obj: Record<string, unknown>, prefix = ''): Record<string, string> {
  const result: Record<string, string> = {}
  for (const [key, value] of Object.entries(obj)) {
    const fullKey = prefix ? `${prefix}.${key}` : key
    if (typeof value === 'object' && value !== null && !Array.isArray(value)) {
      Object.assign(result, flattenStrings(value as Record<string, unknown>, fullKey))
    } else {
      result[fullKey] = String(value)
    }
  }
  return result
}

export interface LocalTemplate {
  pageType: string
  name: string
  html: string
  css: string
}

export function loadLocalTemplates(config: ProjectConfig): LocalTemplate[] {
  const templates: LocalTemplate[] = []
  if (!config.templates) return templates

  for (const [pageType, paths] of Object.entries(config.templates)) {
    const html = existsSync(paths.html) ? readFileSync(paths.html, 'utf8') : ''
    const css = existsSync(paths.css) ? readFileSync(paths.css, 'utf8') : ''
    if (!html && !css) continue

    // Inject _shared.css if it exists
    const sharedCssPath = paths.css.replace(/[^/]+$/, '_shared.css')
    const sharedCss = existsSync(sharedCssPath) ? readFileSync(sharedCssPath, 'utf8') + '\n' : ''

    templates.push({
      pageType,
      name: pageType.charAt(0).toUpperCase() + pageType.slice(1).replace(/_/g, ' '),
      html,
      css: sharedCss + css,
    })
  }
  return templates
}

export function loadLocalStrings(config: ProjectConfig): Map<string, Record<string, string>> {
  const allStrings = new Map<string, Record<string, string>>()
  const locales = config.locales ?? ['en']
  for (const locale of locales) {
    const path = `strings/${locale}.yaml`
    if (!existsSync(path)) continue
    const raw = parseYAML(readFileSync(path, 'utf8')) ?? {}
    allStrings.set(locale, flattenStrings(raw))
  }
  return allStrings
}

export async function computeDiff(
  client: APIClient,
  config: ProjectConfig
): Promise<SyncDiff> {
  const remoteTemplates = await client.listTemplates()
  const localTemplates = loadLocalTemplates(config)

  const templateDiff: SyncDiff['templates'] = { added: [], modified: [], unchanged: [] }

  for (const local of localTemplates) {
    const remote = remoteTemplates.find(
      (t) => t.page_type === local.pageType && !t.is_default
    )
    if (!remote) {
      templateDiff.added.push(local.pageType)
    } else if (remote.html_content !== local.html || remote.css_content !== local.css) {
      templateDiff.modified.push(local.pageType)
    } else {
      templateDiff.unchanged.push(local.pageType)
    }
  }

  // String diff
  const locales = config.locales ?? ['en']
  let sAdded = 0, sModified = 0, sDeleted = 0, sUnchanged = 0
  const localStrings = loadLocalStrings(config)

  for (const locale of locales) {
    const remote = await client.listStrings(locale)
    const remoteMap = new Map(remote.map((s) => [s.string_key, s.value]))
    const local = localStrings.get(locale) ?? {}

    for (const [key, value] of Object.entries(local)) {
      if (!remoteMap.has(key)) sAdded++
      else if (remoteMap.get(key) !== value) sModified++
      else sUnchanged++
    }
    for (const key of remoteMap.keys()) {
      if (!(key in local)) sDeleted++
    }
  }

  return {
    templates: templateDiff,
    strings: { added: sAdded, modified: sModified, deleted: sDeleted, unchanged: sUnchanged },
  }
}

export async function push(
  client: APIClient,
  config: ProjectConfig,
  options: { dryRun?: boolean; templateFilter?: string } = {}
): Promise<SyncDiff> {
  const diff = await computeDiff(client, config)

  if (options.dryRun) return diff

  const remoteTemplates = await client.listTemplates()
  const localTemplates = loadLocalTemplates(config)

  for (const local of localTemplates) {
    if (options.templateFilter && local.pageType !== options.templateFilter) continue

    const remote = remoteTemplates.find(
      (t) => t.page_type === local.pageType && !t.is_default
    )

    if (remote) {
      await client.updateTemplate(remote.id, {
        html_content: local.html,
        css_content: local.css,
      })
    } else {
      await client.createTemplate({
        page_type: local.pageType as PageTemplate['page_type'],
        name: local.name,
        html_content: local.html,
        css_content: local.css,
        is_active: true,
      })
    }
  }

  // Push strings
  const localStrings = loadLocalStrings(config)
  for (const [locale, strings] of localStrings) {
    for (const [key, value] of Object.entries(strings)) {
      await client.upsertString({ string_key: key, locale, value })
    }
  }

  return diff
}

export async function pull(
  client: APIClient,
  config: ProjectConfig
): Promise<{ templates: PageTemplate[]; strings: Map<string, LanguageString[]> }> {
  const templates = await client.listTemplates()
  const strings = new Map<string, LanguageString[]>()

  for (const locale of config.locales ?? ['en']) {
    strings.set(locale, await client.listStrings(locale))
  }

  return { templates, strings }
}
