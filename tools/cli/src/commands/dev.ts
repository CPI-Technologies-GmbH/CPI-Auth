import { Command } from 'commander'
import { createServer } from 'node:http'
import { readFileSync, existsSync, watch } from 'node:fs'
import { parse as parseYAML } from 'yaml'
import { resolve } from 'node:path'
import chalk from 'chalk'
import { renderPreview, buildCSS, type LanguageString, type CustomFieldConfig, type DesignTokens, type ProjectConfig } from '@cpi-auth/sdk'
import { findConfig, info, success } from '../helpers.js'

const PAGE_TYPES = ['login', 'signup', 'verification', 'password_reset', 'mfa_challenge', 'error', 'consent', 'profile']

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

function loadLocalStringsAsLS(locale: string): LanguageString[] {
  const path = `strings/${locale}.yaml`
  if (!existsSync(path)) return []
  const raw = parseYAML(readFileSync(path, 'utf8')) ?? {}
  const flat = flattenStrings(raw)
  return Object.entries(flat).map(([key, value]) => ({
    id: '', tenant_id: '', string_key: key, locale, value, created_at: '', updated_at: '',
  }))
}

function buildShellHTML(config: ProjectConfig, activePage: string, activeLocale: string): string {
  const locales = config.locales ?? ['en']
  const templateEntries = config.templates ? Object.keys(config.templates) : []
  const pages = [...new Set([...PAGE_TYPES, ...templateEntries])]

  const tokensCSS = config.tokens ? buildCSS(config.tokens) : ''
  const strings = loadLocalStringsAsLS(activeLocale)
  const sampleData: Record<string, string> = {}
  if (config.preview?.sample_data) {
    for (const [k, v] of Object.entries(config.preview.sample_data)) {
      sampleData[`{{${k}}}`] = v
    }
  }
  const customFields = config.preview?.custom_fields as CustomFieldConfig[] | undefined

  // Load active template
  let html = '', css = ''
  const tmplConfig = config.templates?.[activePage]
  if (tmplConfig) {
    html = existsSync(tmplConfig.html) ? readFileSync(tmplConfig.html, 'utf8') : '<p>Template not found</p>'
    css = existsSync(tmplConfig.css) ? readFileSync(tmplConfig.css, 'utf8') : ''
  }

  const preview = renderPreview(html, css, { strings, sampleData, customFields, tokensCSS })

  const pageButtons = pages.map((p) => {
    const active = p === activePage ? 'active' : ''
    const label = p.charAt(0).toUpperCase() + p.slice(1).replace(/_/g, ' ')
    const hasLocal = config.templates?.[p] ? 'local' : ''
    return `<a href="/?page=${p}&locale=${activeLocale}" class="page-btn ${active} ${hasLocal}">${label}</a>`
  }).join('\n            ')

  const localeButtons = locales.map((l) => {
    const active = l === activeLocale ? 'active' : ''
    return `<a href="/?page=${activePage}&locale=${l}" class="locale-btn ${active}">${l.toUpperCase()}</a>`
  }).join(' ')

  return `<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>CPI Auth Dev — ${activePage}</title>
  <style>
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #0a0a0a; color: #e5e5e5; height: 100vh; display: flex; }
    .sidebar { width: 220px; background: #141414; border-right: 1px solid #262626; padding: 1rem; display: flex; flex-direction: column; gap: 1rem; overflow-y: auto; }
    .sidebar h2 { font-size: .75rem; text-transform: uppercase; letter-spacing: .1em; color: #737373; margin-bottom: .25rem; }
    .page-btn { display: block; padding: .5rem .75rem; border-radius: 6px; color: #a3a3a3; text-decoration: none; font-size: .8125rem; transition: all .15s; }
    .page-btn:hover { background: #1a1a1a; color: #e5e5e5; }
    .page-btn.active { background: #6366f1; color: white; }
    .page-btn.local::after { content: ''; display: inline-block; width: 6px; height: 6px; border-radius: 50%; background: #34d399; margin-left: .5rem; }
    .locale-btn { display: inline-block; padding: .25rem .5rem; border-radius: 4px; color: #737373; text-decoration: none; font-size: .75rem; font-weight: 600; }
    .locale-btn.active { background: #262626; color: #e5e5e5; }
    .viewport-btns { display: flex; gap: .25rem; }
    .viewport-btns button { background: #1a1a1a; border: 1px solid #262626; border-radius: 4px; color: #a3a3a3; padding: .25rem .5rem; cursor: pointer; font-size: .75rem; }
    .viewport-btns button.active { border-color: #6366f1; color: #e5e5e5; }
    .main { flex: 1; display: flex; flex-direction: column; }
    .toolbar { padding: .5rem 1rem; background: #141414; border-bottom: 1px solid #262626; display: flex; align-items: center; justify-content: space-between; font-size: .75rem; color: #737373; }
    .preview-frame { flex: 1; display: flex; align-items: center; justify-content: center; padding: 2rem; }
    iframe { border: none; border-radius: 12px; box-shadow: 0 0 0 1px #262626, 0 25px 50px -12px rgba(0,0,0,.5); background: white; transition: width .3s; }
    .status { padding: .5rem 1rem; background: #141414; border-top: 1px solid #262626; font-size: .75rem; color: #525252; }
    .status .dot { display: inline-block; width: 6px; height: 6px; border-radius: 50%; background: #34d399; margin-right: .5rem; }
  </style>
  <script>
    function setViewport(w) {
      document.getElementById('preview').style.width = w;
      document.querySelectorAll('.viewport-btns button').forEach(b => b.classList.remove('active'));
      event.target.classList.add('active');
    }
    // Auto-reload every 1s in dev
    setTimeout(() => location.reload(), 2000);
  </script>
</head>
<body>
  <div class="sidebar">
    <div>
      <h2>Pages</h2>
      ${pageButtons}
    </div>
    <div>
      <h2>Locale</h2>
      <div>${localeButtons}</div>
    </div>
    <div>
      <h2>Viewport</h2>
      <div class="viewport-btns">
        <button class="active" onclick="setViewport('100%')">Desktop</button>
        <button onclick="setViewport('768px')">Tablet</button>
        <button onclick="setViewport('375px')">Mobile</button>
      </div>
    </div>
  </div>
  <div class="main">
    <div class="toolbar">
      <span>${activePage} &middot; ${activeLocale}</span>
      <span>CPI Auth Dev Server</span>
    </div>
    <div class="preview-frame">
      <iframe id="preview" srcdoc="${escapeHTML(preview)}" style="width: 100%; height: 100%;"></iframe>
    </div>
    <div class="status"><span class="dot"></span>Watching for changes... Auto-reload active</div>
  </div>
</body>
</html>`
}

function escapeHTML(s: string): string {
  return s.replace(/&/g, '&amp;').replace(/"/g, '&quot;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
}

export function devCommand() {
  return new Command('dev')
    .description('Start local preview dev server with hot reload')
    .option('-p, --port <port>', 'Port', '4400')
    .action((opts) => {
      const port = parseInt(opts.port, 10)
      const config = findConfig()

      const server = createServer((req, res) => {
        // Re-read config on every request for hot-reload
        const freshConfig = findConfig()
        const url = new URL(req.url ?? '/', `http://localhost:${port}`)
        const activePage = url.searchParams.get('page') ?? 'login'
        const activeLocale = url.searchParams.get('locale') ?? (freshConfig.locales?.[0] ?? 'en')

        res.writeHead(200, { 'Content-Type': 'text/html; charset=utf-8' })
        res.end(buildShellHTML(freshConfig, activePage, activeLocale))
      })

      server.listen(port, () => {
        console.log()
        console.log(chalk.bold('  CPI Auth Dev Server'))
        console.log()
        console.log(`  ${chalk.green('➜')}  ${chalk.bold('Local:')}   http://localhost:${port}/`)
        console.log(`  ${chalk.dim('➜')}  ${chalk.dim('Pages:')}   ${Object.keys(config.templates ?? {}).join(', ') || 'none'}`)
        console.log(`  ${chalk.dim('➜')}  ${chalk.dim('Locales:')} ${(config.locales ?? ['en']).join(', ')}`)
        console.log()
        info('Watching for file changes (auto-reload every 2s)')
        console.log()
      })
    })
}
