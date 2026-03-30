import { Command } from 'commander'
import { readFileSync, existsSync } from 'node:fs'
import { parse as parseYAML } from 'yaml'
import chalk from 'chalk'
import { validateContrasts } from '@cpi-auth/sdk'
import { findProjectConfig, success, warn, error, info } from '../helpers.js'

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

export function validateCommand() {
  return new Command('validate')
    .description('Validate templates, strings, and tokens')
    .action(() => {
      const config = findProjectConfig()
      let errors = 0

      console.log(chalk.bold('\nValidation Report\n'))

      // 1. Check templates exist
      console.log(chalk.bold('Templates:'))
      if (config.templates) {
        for (const [type, paths] of Object.entries(config.templates)) {
          const htmlExists = existsSync(paths.html)
          const cssExists = existsSync(paths.css)
          if (htmlExists && cssExists) {
            success(`  ${type}: ${paths.html}, ${paths.css}`)
          } else {
            if (!htmlExists) { error(`  ${type}: ${paths.html} not found`); errors++ }
            if (!cssExists) { error(`  ${type}: ${paths.css} not found`); errors++ }
          }
        }
      } else {
        warn('  No templates configured')
      }

      // 2. Check {{t.xxx}} references have matching strings
      console.log(chalk.bold('\nString References:'))
      const usedKeys = new Set<string>()

      if (config.templates) {
        for (const [, paths] of Object.entries(config.templates)) {
          if (!existsSync(paths.html)) continue
          const html = readFileSync(paths.html, 'utf8')
          const matches = html.matchAll(/\{\{t\.([^}]+)\}\}/g)
          for (const match of matches) {
            usedKeys.add(match[1])
          }
        }
      }

      if (usedKeys.size === 0) {
        info('  No {{t.xxx}} references found in templates')
      } else {
        const locales = config.locales ?? ['en']
        for (const locale of locales) {
          const path = `strings/${locale}.yaml`
          if (!existsSync(path)) {
            warn(`  strings/${locale}.yaml not found`)
            errors++
            continue
          }
          const raw = parseYAML(readFileSync(path, 'utf8')) ?? {}
          const flat = flattenStrings(raw)
          let localeMissing = 0

          for (const key of usedKeys) {
            if (!(key in flat)) {
              error(`  Missing: ${key} in ${locale}`)
              localeMissing++
              errors++
            }
          }
          if (localeMissing === 0) {
            success(`  All ${usedKeys.size} string refs found in ${locale}`)
          }
        }

        // Check for unused strings
        for (const locale of locales) {
          const path = `strings/${locale}.yaml`
          if (!existsSync(path)) continue
          const raw = parseYAML(readFileSync(path, 'utf8')) ?? {}
          const flat = flattenStrings(raw)
          const unused = Object.keys(flat).filter((k) => !usedKeys.has(k))
          if (unused.length > 0) {
            warn(`  ${unused.length} unused string(s) in ${locale}: ${unused.slice(0, 5).join(', ')}${unused.length > 5 ? '...' : ''}`)
          }
        }
      }

      // 3. Contrast validation
      console.log(chalk.bold('\nColor Contrast:'))
      if (config.tokens?.colors) {
        const results = validateContrasts(config.tokens)
        const failures = results.filter((r) => !r.aa)
        if (results.length === 0) {
          info('  No contrast pairs to check')
        } else if (failures.length === 0) {
          success(`  All ${results.length} color pairs pass WCAG AA`)
        } else {
          for (const f of failures) {
            error(`  ${f.pair[0]} on ${f.pair[1]}: ${f.ratio.toFixed(1)}:1 (need 4.5:1)`)
            errors++
          }
        }
      } else {
        info('  No color tokens defined')
      }

      // Summary
      console.log()
      if (errors === 0) {
        success('All checks passed!')
      } else {
        error(`${errors} issue(s) found`)
        process.exit(1)
      }
    })
}
