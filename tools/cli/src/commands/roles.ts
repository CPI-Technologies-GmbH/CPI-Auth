import { Command } from 'commander'
import chalk from 'chalk'
import { getClient, success } from '../helpers.js'

export function rolesCommand() {
  const cmd = new Command('roles')
    .description('Manage roles and permissions')

  cmd
    .command('list')
    .description('List all roles')
    .option('-s, --server <url>', 'Server URL')
    .option('-t, --token <token>', 'Access token')
    .option('--json', 'Output as JSON')
    .action(async (opts) => {
      const client = getClient(opts)
      const res = await client.request<any>('GET', '/admin/roles')
      const roles = res.data || res

      if (opts.json) {
        console.log(JSON.stringify(roles, null, 2))
        return
      }

      console.log(chalk.bold(`\n  Roles (${roles.length})\n`))
      for (const r of roles) {
        const system = r.is_system ? chalk.dim(' (system)') : ''
        console.log(`  ${chalk.cyan(r.name)}${system}`)
        if (r.description) console.log(`    ${chalk.dim(r.description)}`)
        if (r.permissions?.length) console.log(`    Permissions: ${chalk.dim(r.permissions.join(', '))}`)
      }
    })

  cmd
    .command('create')
    .description('Create a role')
    .requiredOption('--name <name>', 'Role name')
    .option('--description <desc>', 'Role description')
    .option('--permissions <perms>', 'Permissions (comma-separated)')
    .option('-s, --server <url>', 'Server URL')
    .option('-t, --token <token>', 'Access token')
    .action(async (opts) => {
      const client = getClient(opts)
      const role = await client.request<any>('POST', '/admin/roles', {
        name: opts.name,
        description: opts.description || '',
        permissions: opts.permissions ? opts.permissions.split(',').map((s: string) => s.trim()) : [],
      })
      success(`Role "${role.name}" created (ID: ${role.id})`)
    })

  cmd
    .command('permissions')
    .description('List all permissions')
    .option('-s, --server <url>', 'Server URL')
    .option('-t, --token <token>', 'Access token')
    .option('--json', 'Output as JSON')
    .action(async (opts) => {
      const client = getClient(opts)
      const res = await client.request<any>('GET', '/admin/permissions')
      const perms = res.data || res

      if (opts.json) {
        console.log(JSON.stringify(perms, null, 2))
        return
      }

      console.log(chalk.bold(`\n  Permissions (${perms.length})\n`))
      for (const p of perms) {
        console.log(`  ${chalk.cyan(p.name)} ${chalk.dim(p.display_name || '')}`)
      }
    })

  cmd
    .command('create-permission')
    .description('Create a permission')
    .requiredOption('--name <name>', 'Permission name (e.g. users:read)')
    .option('--display-name <name>', 'Display name')
    .option('--description <desc>', 'Description')
    .option('-s, --server <url>', 'Server URL')
    .option('-t, --token <token>', 'Access token')
    .action(async (opts) => {
      const client = getClient(opts)
      const perm = await client.request<any>('POST', '/admin/permissions', {
        name: opts.name,
        display_name: opts.displayName || opts.name.replace(/[:.]/g, ' '),
        description: opts.description || '',
      })
      success(`Permission "${perm.name}" created`)
    })

  return cmd
}
