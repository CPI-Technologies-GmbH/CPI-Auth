import { Command } from 'commander'
import chalk from 'chalk'
import { getClient, success, error } from '../helpers.js'

export function tenantsCommand() {
  const cmd = new Command('tenants')
    .description('Manage tenants')

  cmd
    .command('list')
    .description('List all tenants')
    .option('-s, --server <url>', 'Server URL')
    .option('-t, --token <token>', 'Access token')
    .option('--json', 'Output as JSON')
    .action(async (opts) => {
      const client = getClient(opts)
      const res = await client.request<any>('GET', '/admin/tenants')
      const tenants = res.data || res

      if (opts.json) {
        console.log(JSON.stringify(tenants, null, 2))
        return
      }

      console.log(chalk.bold(`\n  Tenants (${tenants.length})\n`))
      for (const t of tenants) {
        console.log(`  ${chalk.cyan(t.name)} ${chalk.dim(`(${t.slug})`)}`)
        console.log(`    ID:     ${chalk.dim(t.id)}`)
        if (t.domain) console.log(`    Domain: ${chalk.dim(t.domain)}`)
        console.log()
      }
    })

  cmd
    .command('create')
    .description('Create a new tenant')
    .requiredOption('--name <name>', 'Tenant name')
    .requiredOption('--slug <slug>', 'Tenant slug (URL-safe identifier)')
    .option('--domain <domain>', 'Custom domain')
    .option('-s, --server <url>', 'Server URL')
    .option('-t, --token <token>', 'Access token')
    .option('--json', 'Output as JSON')
    .action(async (opts) => {
      const client = getClient(opts)

      const tenant = await client.request<any>('POST', '/admin/tenants', {
        name: opts.name,
        slug: opts.slug,
        ...(opts.domain ? { domain: opts.domain } : {}),
      })

      if (opts.json) {
        console.log(JSON.stringify(tenant, null, 2))
        return
      }

      success(`Tenant "${tenant.name}" created`)
      console.log(`  ID:     ${chalk.dim(tenant.id)}`)
      console.log(`  Slug:   ${chalk.dim(tenant.slug)}`)
      if (tenant.domain) console.log(`  Domain: ${chalk.dim(tenant.domain)}`)
    })

  cmd
    .command('update <id>')
    .description('Update a tenant')
    .option('--name <name>', 'Tenant name')
    .option('--domain <domain>', 'Custom domain')
    .option('-s, --server <url>', 'Server URL')
    .option('-t, --token <token>', 'Access token')
    .action(async (id, opts) => {
      const client = getClient(opts)

      const body: any = {}
      if (opts.name) body.name = opts.name
      if (opts.domain) body.domain = opts.domain

      const tenant = await client.request<any>('PATCH', `/admin/tenants/${id}`, body)
      success(`Tenant "${tenant.name}" updated`)
    })

  cmd
    .command('delete <id>')
    .description('Delete a tenant')
    .option('-s, --server <url>', 'Server URL')
    .option('-t, --token <token>', 'Access token')
    .action(async (id, opts) => {
      const client = getClient(opts)
      await client.request('DELETE', `/admin/tenants/${id}`)
      success(`Tenant ${id} deleted`)
    })

  cmd
    .command('set-branding <id>')
    .description('Set tenant branding (colors, logo, name)')
    .option('--app-name <name>', 'Application display name')
    .option('--logo-url <url>', 'Logo URL')
    .option('--primary-color <hex>', 'Primary color (hex)')
    .option('--surface-color <hex>', 'Surface/background color (hex)')
    .option('--text-color <hex>', 'Text color (hex)')
    .option('-s, --server <url>', 'Server URL')
    .option('-t, --token <token>', 'Access token')
    .action(async (id, opts) => {
      const client = getClient(opts)

      const branding: any = {}
      if (opts.appName) branding.app_name = opts.appName
      if (opts.logoUrl) branding.logo_url = opts.logoUrl
      if (opts.primaryColor) branding.primary_color = opts.primaryColor
      if (opts.surfaceColor) branding.surface_color = opts.surfaceColor
      if (opts.textColor) branding.text_color = opts.textColor

      await client.request('PATCH', `/admin/tenants/${id}`, { branding })
      success(`Branding updated for tenant ${id}`)
    })

  return cmd
}
