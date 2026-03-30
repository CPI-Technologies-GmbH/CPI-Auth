import { Command } from 'commander'
import { push } from '@cpi-auth/sdk'
import { findProjectConfig, getClient, success, info, warn } from '../helpers.js'

export function pushCommand() {
  return new Command('push')
    .description('Push local templates and strings to the server')
    .option('--dry-run', 'Show what would change without applying')
    .option('-t, --template <type>', 'Only push a specific template')
    .action(async (opts) => {
      const config = { ...findProjectConfig(), server: "", tenant_id: "" } as any
      const client = getClient()

      if (opts.dryRun) info('Dry run mode — no changes will be applied')

      const diff = await push(client, config, {
        dryRun: opts.dryRun,
        templateFilter: opts.template,
      })

      // Report
      if (diff.templates.added.length) success(`Templates added: ${diff.templates.added.join(', ')}`)
      if (diff.templates.modified.length) success(`Templates modified: ${diff.templates.modified.join(', ')}`)
      if (diff.templates.unchanged.length) info(`Templates unchanged: ${diff.templates.unchanged.join(', ')}`)

      if (diff.strings.added) success(`Strings added: ${diff.strings.added}`)
      if (diff.strings.modified) success(`Strings modified: ${diff.strings.modified}`)
      if (diff.strings.deleted) warn(`Strings on server not in local: ${diff.strings.deleted}`)
      if (diff.strings.unchanged) info(`Strings unchanged: ${diff.strings.unchanged}`)

      if (!opts.dryRun) success('Push complete!')
    })
}
