import { Command } from 'commander'
import { writeFileSync, mkdirSync } from 'node:fs'
import { stringify as stringifyYAML } from 'yaml'
import { pull } from '@cpi-auth/sdk'
import { findConfig, getClient, success, info } from '../helpers.js'

export function pullCommand() {
  return new Command('pull')
    .description('Pull templates and strings from the server')
    .action(async () => {
      const config = findConfig()
      const client = getClient(config)

      info('Pulling from server...')
      const result = await pull(client, config)

      mkdirSync('templates', { recursive: true })
      mkdirSync('strings', { recursive: true })

      // Write templates
      let count = 0
      for (const tmpl of result.templates) {
        if (tmpl.is_default) continue // Skip defaults, only pull custom
        const base = tmpl.page_type
        writeFileSync(`templates/${base}.html`, tmpl.html_content)
        writeFileSync(`templates/${base}.css`, tmpl.css_content)
        count++
      }
      success(`Pulled ${count} custom templates`)

      // Write strings by locale
      for (const [locale, strings] of result.strings) {
        const nested: Record<string, Record<string, string>> = {}
        for (const s of strings) {
          const [prefix, ...rest] = s.string_key.split('.')
          const key = rest.join('.')
          if (!nested[prefix]) nested[prefix] = {}
          nested[prefix][key] = s.value
        }
        writeFileSync(`strings/${locale}.yaml`, stringifyYAML(nested))
        success(`Pulled ${strings.length} strings for locale "${locale}"`)
      }
    })
}
