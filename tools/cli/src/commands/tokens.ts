import { Command } from 'commander'
import { writeFileSync } from 'node:fs'
import chalk from 'chalk'
import { buildCSS, validateContrasts } from '@cpi-auth/sdk'
import { findConfig, success, info, warn } from '../helpers.js'

export function tokensCommand() {
  const cmd = new Command('tokens').description('Manage design tokens')

  cmd
    .command('build')
    .description('Generate _tokens.css from config')
    .option('-o, --output <file>', 'Output file', 'templates/_tokens.css')
    .action((opts) => {
      const config = findConfig()
      if (!config.tokens) {
        warn('No tokens defined in cpi-auth.config.yaml')
        return
      }

      const css = buildCSS(config.tokens)
      writeFileSync(opts.output, css)
      success(`Generated ${opts.output} (${css.split('\n').length} properties)`)
    })

  cmd
    .command('validate')
    .description('Check color contrast ratios (WCAG)')
    .action(() => {
      const config = findConfig()
      if (!config.tokens?.colors) {
        warn('No color tokens defined')
        return
      }

      const results = validateContrasts(config.tokens)
      if (results.length === 0) {
        info('No contrast pairs to check (define text, background, surface colors)')
        return
      }

      console.log(chalk.bold('\nContrast Check (WCAG):'))
      for (const r of results) {
        const status = r.aa
          ? chalk.green('AA PASS')
          : r.aaLarge
            ? chalk.yellow('AA Large')
            : chalk.red('FAIL')
        const ratio = r.ratio.toFixed(1) + ':1'
        console.log(`  ${r.pair[0]} on ${r.pair[1]}: ${ratio} ${status}`)
      }

      const failures = results.filter((r) => !r.aa)
      console.log()
      if (failures.length === 0) {
        success('All color pairs pass WCAG AA')
      } else {
        warn(`${failures.length} pair(s) fail WCAG AA`)
      }
    })

  return cmd
}
