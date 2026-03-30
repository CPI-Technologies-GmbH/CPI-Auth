import { Command } from 'commander'
import chalk from 'chalk'
import { getClient, success, error } from '../helpers.js'

export function usersCommand() {
  const cmd = new Command('users')
    .description('Manage users')

  cmd
    .command('list')
    .description('List users')
    .option('-s, --server <url>', 'Server URL')
    .option('-t, --token <token>', 'Access token')
    .option('--search <query>', 'Search by email or name')
    .option('--limit <n>', 'Limit results', '20')
    .option('--json', 'Output as JSON')
    .action(async (opts) => {
      const client = getClient(opts)
      const params = new URLSearchParams({ per_page: opts.limit })
      if (opts.search) params.set('search', opts.search)
      const res = await client.request<any>('GET', `/admin/users?${params}`)
      const users = res.data || res

      if (opts.json) {
        console.log(JSON.stringify(users, null, 2))
        return
      }

      console.log(chalk.bold(`\n  Users (${users.length})\n`))
      for (const u of users) {
        const status = u.status === 'active' ? chalk.green('active') : chalk.red(u.status)
        console.log(`  ${chalk.cyan(u.email)} ${chalk.dim(u.name || '')} ${status}`)
        console.log(`    ID: ${chalk.dim(u.id)}`)
      }
    })

  cmd
    .command('create')
    .description('Create a new user')
    .requiredOption('--email <email>', 'User email')
    .requiredOption('--password <password>', 'User password')
    .option('--name <name>', 'User name')
    .option('--role <role>', 'Assign role by name')
    .option('-s, --server <url>', 'Server URL')
    .option('-t, --token <token>', 'Access token')
    .option('--json', 'Output as JSON')
    .action(async (opts) => {
      const client = getClient(opts)

      const user = await client.request<any>('POST', '/admin/users', {
        email: opts.email,
        password: opts.password,
        name: opts.name || opts.email.split('@')[0],
      })

      if (opts.role) {
        try {
          const roles = await client.request<any>('GET', '/admin/roles')
          const roleList = roles.data || roles
          const role = roleList.find((r: any) => r.name === opts.role)
          if (role) {
            await client.request('POST', `/admin/users/${user.id}/roles`, { role_id: role.id })
            success(`Assigned role "${opts.role}"`)
          }
        } catch { /* role assignment failed — user still created */ }
      }

      if (opts.json) {
        console.log(JSON.stringify(user, null, 2))
        return
      }

      success(`User created: ${user.email}`)
      console.log(`  ID: ${chalk.dim(user.id)}`)
    })

  cmd
    .command('delete <id>')
    .description('Delete a user')
    .option('-s, --server <url>', 'Server URL')
    .option('-t, --token <token>', 'Access token')
    .action(async (id, opts) => {
      const client = getClient(opts)
      await client.request('DELETE', `/admin/users/${id}`)
      success(`User ${id} deleted`)
    })

  cmd
    .command('block <id>')
    .description('Block a user')
    .option('-s, --server <url>', 'Server URL')
    .option('-t, --token <token>', 'Access token')
    .action(async (id, opts) => {
      const client = getClient(opts)
      await client.request('POST', `/admin/users/${id}/block`)
      success(`User ${id} blocked`)
    })

  cmd
    .command('unblock <id>')
    .description('Unblock a user')
    .option('-s, --server <url>', 'Server URL')
    .option('-t, --token <token>', 'Access token')
    .action(async (id, opts) => {
      const client = getClient(opts)
      await client.request('POST', `/admin/users/${id}/unblock`)
      success(`User ${id} unblocked`)
    })

  return cmd
}
