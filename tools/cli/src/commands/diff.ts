import { Command } from 'commander'
import { computeDiff } from '@cpi-auth/sdk'
import { findProjectConfig, getClient, success, info, warn } from '../helpers.js'
import chalk from 'chalk'

export function diffCommand() {
  return new Command('diff')
    .description('Show differences between local and server')
    .action(async () => {
      const config = { ...findProjectConfig(), server: "", tenant_id: "" } as any
      const client = getClient()

      info('Comparing local vs server...\n')
      const diff = await computeDiff(client, config)

      // Templates
      console.log(chalk.bold('Templates:'))
      for (const t of diff.templates.added) console.log(chalk.green(`  + ${t} (new)`))
      for (const t of diff.templates.modified) console.log(chalk.yellow(`  ~ ${t} (modified)`))
      for (const t of diff.templates.unchanged) console.log(chalk.dim(`    ${t} (unchanged)`))
      if (!diff.templates.added.length && !diff.templates.modified.length && !diff.templates.unchanged.length) {
        console.log(chalk.dim('  No local templates'))
      }

      // Strings
      console.log(chalk.bold('\nLanguage Strings:'))
      if (diff.strings.added) console.log(chalk.green(`  + ${diff.strings.added} new`))
      if (diff.strings.modified) console.log(chalk.yellow(`  ~ ${diff.strings.modified} modified`))
      if (diff.strings.deleted) console.log(chalk.red(`  - ${diff.strings.deleted} on server, not local`))
      if (diff.strings.unchanged) console.log(chalk.dim(`    ${diff.strings.unchanged} unchanged`))

      const hasChanges = diff.templates.added.length + diff.templates.modified.length + diff.strings.added + diff.strings.modified > 0
      console.log()
      if (hasChanges) {
        warn('Run `cpi-auth push` to apply changes')
      } else {
        success('Everything is in sync')
      }
    })
}
