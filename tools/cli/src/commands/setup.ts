import { Command } from 'commander'
import chalk from 'chalk'
import { getClient, success, error, info, warn } from '../helpers.js'

export function setupCommand() {
  return new Command('setup')
    .description('One-command setup: create application, roles, permissions, and users')
    .option('-s, --server <url>', 'CPI Auth server URL')
    .option('-t, --token <token>', 'API access token')
    .option('--app-name <name>', 'Application name', 'My Application')
    .option('--app-type <type>', 'Application type (spa, web, native, m2m)', 'spa')
    .option('--redirect-uri <uri>', 'OAuth redirect URI (comma-separated for multiple)')
    .option('--allowed-origin <origin>', 'Allowed CORS origin (comma-separated)')
    .option('--logout-url <url>', 'Post-logout redirect URL (comma-separated)')
    .option('--grant-types <types>', 'Grant types (comma-separated)', 'authorization_code,refresh_token')
    .option('--create-role <name>', 'Create a custom role (repeatable)', collect, [])
    .option('--create-permission <name>', 'Create a permission (repeatable)', collect, [])
    .option('--create-user <email>', 'Create a user (repeatable)', collect, [])
    .option('--user-password <password>', 'Password for created users')
    .option('--user-role <role>', 'Assign role to created users')
    .option('--output <format>', 'Output format: env, json, yaml', 'env')
    .action(async (opts) => {
      const client = getClient(opts)

      console.log()
      console.log(chalk.bold('  CPI Auth Setup'))
      console.log()

      const results: Record<string, any> = {}

      // 1. Create application
      info(`Creating ${opts.appType} application "${opts.appName}"...`)
      try {
        const redirectUris = opts.redirectUri ? opts.redirectUri.split(',').map((s: string) => s.trim()) : []
        const allowedOrigins = opts.allowedOrigin ? opts.allowedOrigin.split(',').map((s: string) => s.trim()) : []
        const logoutUrls = opts.logoutUrl ? opts.logoutUrl.split(',').map((s: string) => s.trim()) : []
        const grantTypes = opts.grantTypes.split(',').map((s: string) => s.trim())

        const app = await client.request<any>('POST', '/admin/applications', {
          name: opts.appName,
          type: opts.appType,
          redirect_uris: redirectUris,
          allowed_origins: allowedOrigins,
          allowed_logout_urls: logoutUrls,
          grant_types: grantTypes,
          is_active: true,
        })

        results.app = app
        success(`Application created`)
        console.log(chalk.dim(`  Client ID:     ${app.client_id}`))
        console.log(chalk.dim(`  Client Secret: ${app.client_secret || '(none — SPA type)'}`))
      } catch (e: any) {
        error(`Failed to create application: ${e.message}`)
        process.exit(1)
      }

      // 2. Create permissions
      const createdPermissions: string[] = []
      for (const perm of opts.createPermission) {
        try {
          const parts = perm.split(':')
          const name = perm
          const displayName = parts.map((p: string) => p.charAt(0).toUpperCase() + p.slice(1)).join(' ')
          await client.request('POST', '/admin/permissions', {
            name,
            display_name: displayName,
            description: `Auto-created by CLI setup`,
          })
          createdPermissions.push(name)
          success(`Permission "${name}" created`)
        } catch (e: any) {
          warn(`Permission "${perm}": ${e.message}`)
        }
      }

      // 3. Create roles
      const createdRoles: Record<string, any> = {}
      for (const role of opts.createRole) {
        try {
          const r = await client.request<any>('POST', '/admin/roles', {
            name: role,
            description: `Auto-created by CLI setup`,
            permissions: createdPermissions,
          })
          createdRoles[role] = r
          success(`Role "${role}" created with ${createdPermissions.length} permissions`)
        } catch (e: any) {
          warn(`Role "${role}": ${e.message}`)
        }
      }

      // 4. Create users
      for (const email of opts.createUser) {
        try {
          const password = opts.userPassword || generatePassword()
          const user = await client.request<any>('POST', '/admin/users', {
            email,
            password,
            name: email.split('@')[0],
          })

          // Assign role if specified
          if (opts.userRole && createdRoles[opts.userRole]) {
            await client.request('POST', `/admin/users/${user.id}/roles`, {
              role_id: createdRoles[opts.userRole].id,
            }).catch(() => {})
          }

          success(`User "${email}" created`)
          console.log(chalk.dim(`  Password: ${password}`))
        } catch (e: any) {
          warn(`User "${email}": ${e.message}`)
        }
      }

      // 5. Output configuration
      console.log()
      console.log(chalk.bold('  Configuration'))
      console.log()

      const app = results.app
      const server = opts.server || process.env.CPI_AUTH_SERVER || ''

      if (opts.output === 'json') {
        console.log(JSON.stringify({
          client_id: app.client_id,
          client_secret: app.client_secret,
          issuer: server,
          redirect_uri: app.redirect_uris?.[0] || '',
          authorization_endpoint: `${server}/oauth/authorize`,
          token_endpoint: `${server}/oauth/token`,
          userinfo_endpoint: `${server}/oauth/userinfo`,
          jwks_uri: `${server}/.well-known/jwks.json`,
        }, null, 2))
      } else if (opts.output === 'yaml') {
        console.log(`client_id: "${app.client_id}"`)
        console.log(`client_secret: "${app.client_secret || ''}"`)
        console.log(`issuer: "${server}"`)
        console.log(`redirect_uri: "${app.redirect_uris?.[0] || ''}"`)
      } else {
        // env format (default)
        console.log(chalk.cyan('  # Add to your .env file:'))
        console.log(`  CPI_AUTH_CLIENT_ID=${app.client_id}`)
        if (app.client_secret) console.log(`  CPI_AUTH_CLIENT_SECRET=${app.client_secret}`)
        console.log(`  CPI_AUTH_ISSUER=${server}`)
        if (app.redirect_uris?.[0]) console.log(`  CPI_AUTH_REDIRECT_URI=${app.redirect_uris[0]}`)
        console.log(`  CPI_AUTH_AUTHORIZATION_ENDPOINT=${server}/oauth/authorize`)
        console.log(`  CPI_AUTH_TOKEN_ENDPOINT=${server}/oauth/token`)
        console.log(`  CPI_AUTH_USERINFO_ENDPOINT=${server}/oauth/userinfo`)
        console.log(`  CPI_AUTH_JWKS_URI=${server}/.well-known/jwks.json`)
      }

      console.log()
      success('Setup complete!')
    })
}

function collect(value: string, previous: string[]): string[] {
  return previous.concat([value])
}

function generatePassword(): string {
  const chars = 'ABCDEFGHJKLMNPQRSTUVWXYZabcdefghjkmnpqrstuvwxyz23456789'
  const special = '!@#$%'
  let pw = ''
  for (let i = 0; i < 16; i++) pw += chars[Math.floor(Math.random() * chars.length)]
  pw += special[Math.floor(Math.random() * special.length)]
  return pw
}
