import { Command } from 'commander'
import chalk from 'chalk'
import { loadGlobalConfig, addContext, useContext, removeContext, getCurrentContext, loadToken } from '../config.js'
import { success, error, info } from '../helpers.js'

export function configCommand() {
  const cmd = new Command('config')
    .description('Manage server connections and contexts')

  cmd
    .command('add-context <name>')
    .description('Add a new server context')
    .requiredOption('-s, --server <url>', 'Server URL')
    .option('--tenant-id <id>', 'Tenant ID')
    .action((name, opts) => {
      addContext(name, opts.server, opts.tenantId)
      success(`Context "${name}" added (${opts.server})`)
    })

  cmd
    .command('use-context <name>')
    .description('Switch to a different context')
    .action((name) => {
      if (useContext(name)) {
        success(`Switched to context "${name}"`)
      } else {
        error(`Context "${name}" not found. Use \`cpi-auth config list-contexts\` to see available contexts.`)
      }
    })

  cmd
    .command('list-contexts')
    .description('List all configured contexts')
    .action(() => {
      const config = loadGlobalConfig()
      const contexts = Object.entries(config.contexts)

      if (contexts.length === 0) {
        info('No contexts configured. Run `cpi-auth login --server <url>` to add one.')
        return
      }

      console.log()
      console.log(`  ${chalk.dim('CURRENT')}   ${chalk.dim('NAME'.padEnd(16))} ${chalk.dim('SERVER')}`)
      for (const [name, ctx] of contexts) {
        const current = name === config['current-context'] ? chalk.green('*') : ' '
        const token = loadToken(ctx.server, ctx['tenant-id'])
        const auth = token ? chalk.green('●') : chalk.dim('○')
        console.log(`  ${current}  ${auth}      ${name.padEnd(16)} ${chalk.dim(ctx.server)}`)
      }
      console.log()
    })

  cmd
    .command('remove-context <name>')
    .description('Remove a context')
    .action((name) => {
      if (removeContext(name)) {
        success(`Context "${name}" removed`)
      } else {
        error(`Context "${name}" not found.`)
      }
    })

  cmd
    .command('current-context')
    .description('Show current context')
    .action(() => {
      const ctx = getCurrentContext()
      if (ctx) {
        console.log(`  ${chalk.dim('Context:')} ${chalk.cyan(ctx.name)}`)
        console.log(`  ${chalk.dim('Server:')}  ${ctx.server}`)
        if (ctx.tenantId) console.log(`  ${chalk.dim('Tenant:')}  ${ctx.tenantId}`)
      } else {
        info('No current context. Run `cpi-auth login --server <url>` to set one.')
      }
    })

  return cmd
}
