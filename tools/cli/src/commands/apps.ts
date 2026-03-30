import { Command } from 'commander'
import chalk from 'chalk'
import { getClient, success, error, info } from '../helpers.js'

export function appsCommand() {
  const cmd = new Command('apps')
    .description('Manage applications (OAuth clients)')

  cmd
    .command('list')
    .description('List all applications')
    .option('-s, --server <url>', 'Server URL')
    .option('-t, --token <token>', 'Access token')
    .option('--json', 'Output as JSON')
    .action(async (opts) => {
      const client = getClient(opts)
      const apps = await client.request<any>('GET', '/admin/applications')
      const list = apps.data || apps

      if (opts.json) {
        console.log(JSON.stringify(list, null, 2))
        return
      }

      console.log(chalk.bold(`\n  Applications (${list.length})\n`))
      for (const app of list) {
        const status = app.is_active ? chalk.green('active') : chalk.dim('disabled')
        console.log(`  ${chalk.cyan(app.name)} ${chalk.dim(`(${app.type})`)} ${status}`)
        console.log(`    Client ID: ${chalk.dim(app.client_id)}`)
        if (app.redirect_uris?.length) {
          console.log(`    Redirect:  ${chalk.dim(app.redirect_uris.join(', '))}`)
        }
        console.log()
      }
    })

  cmd
    .command('create')
    .description('Create a new application')
    .requiredOption('--name <name>', 'Application name')
    .option('--type <type>', 'Type: spa, web, native, m2m', 'spa')
    .option('--redirect-uri <uris>', 'Redirect URIs (comma-separated)')
    .option('--allowed-origin <origins>', 'Allowed origins (comma-separated)')
    .option('--grant-types <types>', 'Grant types (comma-separated)', 'authorization_code,refresh_token')
    .option('-s, --server <url>', 'Server URL')
    .option('-t, --token <token>', 'Access token')
    .option('--json', 'Output as JSON')
    .action(async (opts) => {
      const client = getClient(opts)

      const app = await client.request<any>('POST', '/admin/applications', {
        name: opts.name,
        type: opts.type,
        redirect_uris: opts.redirectUri ? opts.redirectUri.split(',').map((s: string) => s.trim()) : [],
        allowed_origins: opts.allowedOrigin ? opts.allowedOrigin.split(',').map((s: string) => s.trim()) : [],
        grant_types: opts.grantTypes.split(',').map((s: string) => s.trim()),
        is_active: true,
      })

      if (opts.json) {
        console.log(JSON.stringify(app, null, 2))
        return
      }

      success(`Application "${app.name}" created`)
      console.log(`  Client ID:     ${chalk.cyan(app.client_id)}`)
      if (app.client_secret) console.log(`  Client Secret: ${chalk.yellow(app.client_secret)}`)
      console.log(`  Type:          ${app.type}`)
    })

  cmd
    .command('delete <id>')
    .description('Delete an application')
    .option('-s, --server <url>', 'Server URL')
    .option('-t, --token <token>', 'Access token')
    .action(async (id, opts) => {
      const client = getClient(opts)
      await client.request('DELETE', `/admin/applications/${id}`)
      success(`Application ${id} deleted`)
    })

  cmd
    .command('rotate-secret <id>')
    .description('Rotate client secret')
    .option('-s, --server <url>', 'Server URL')
    .option('-t, --token <token>', 'Access token')
    .action(async (id, opts) => {
      const client = getClient(opts)
      const app = await client.request<any>('POST', `/admin/applications/${id}/rotate-secret`)
      success(`Secret rotated`)
      console.log(`  New secret: ${chalk.yellow(app.client_secret)}`)
    })

  return cmd
}
