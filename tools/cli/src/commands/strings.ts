import { Command } from 'commander'
import { readFileSync, writeFileSync, existsSync } from 'node:fs'
import { parse as parseYAML, stringify as stringifyYAML } from 'yaml'
import chalk from 'chalk'
import { findConfig, getClient, success, info, warn } from '../helpers.js'

export function stringsCommand() {
  const cmd = new Command('strings').description('Manage language strings')

  cmd
    .command('list')
    .description('List all strings for a locale')
    .option('-l, --locale <locale>', 'Locale', 'en')
    .action(async (opts) => {
      const config = findConfig()
      const client = getClient(config)
      const strings = await client.listStrings(opts.locale)

      // Group by prefix
      const grouped: Record<string, { key: string; value: string }[]> = {}
      for (const s of strings) {
        const [prefix] = s.string_key.split('.')
        if (!grouped[prefix]) grouped[prefix] = []
        grouped[prefix].push({ key: s.string_key, value: s.value })
      }

      for (const [group, items] of Object.entries(grouped)) {
        console.log(chalk.bold(`\n${group}:`))
        for (const item of items) {
          console.log(`  ${chalk.cyan(item.key)} = ${chalk.dim('"')}${item.value}${chalk.dim('"')}`)
        }
      }
      info(`\nTotal: ${strings.length} strings (${opts.locale})`)
    })

  cmd
    .command('add <key> <value>')
    .description('Add or update a language string')
    .option('-l, --locale <locale>', 'Locale', 'en')
    .action(async (key, value, opts) => {
      const config = findConfig()
      const client = getClient(config)
      await client.upsertString({ string_key: key, locale: opts.locale, value })
      success(`Set ${key} (${opts.locale}) = "${value}"`)
    })

  cmd
    .command('sync')
    .description('Find missing translations across locales')
    .action(async () => {
      const config = findConfig()
      const client = getClient(config)
      const locales = config.locales ?? ['en']

      const allKeys = new Set<string>()
      const byLocale: Record<string, Set<string>> = {}

      for (const locale of locales) {
        const strings = await client.listStrings(locale)
        byLocale[locale] = new Set(strings.map((s) => s.string_key))
        strings.forEach((s) => allKeys.add(s.string_key))
      }

      let missing = 0
      for (const key of allKeys) {
        const missingIn: string[] = []
        for (const locale of locales) {
          if (!byLocale[locale]?.has(key)) missingIn.push(locale)
        }
        if (missingIn.length) {
          warn(`${key} missing in: ${missingIn.join(', ')}`)
          missing++
        }
      }

      if (missing === 0) {
        success(`All ${allKeys.size} keys are translated in ${locales.join(', ')}`)
      } else {
        warn(`${missing} keys have missing translations`)
      }
    })

  cmd
    .command('export')
    .description('Export strings as CSV')
    .option('-o, --output <file>', 'Output file', 'strings.csv')
    .action(async (opts) => {
      const config = findConfig()
      const client = getClient(config)
      const locales = config.locales ?? ['en']

      const data: Record<string, Record<string, string>> = {}
      for (const locale of locales) {
        const strings = await client.listStrings(locale)
        for (const s of strings) {
          if (!data[s.string_key]) data[s.string_key] = {}
          data[s.string_key][locale] = s.value
        }
      }

      const header = ['key', ...locales].join(',')
      const rows = Object.entries(data).map(([key, vals]) => {
        return [key, ...locales.map((l) => `"${(vals[l] ?? '').replace(/"/g, '""')}"`)]
          .join(',')
      })

      writeFileSync(opts.output, [header, ...rows].join('\n'))
      success(`Exported ${Object.keys(data).length} keys to ${opts.output}`)
    })

  return cmd
}
