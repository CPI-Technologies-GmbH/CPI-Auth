import { Command } from 'commander'
import { createInterface } from 'node:readline/promises'
import { APIClient } from '@cpi-auth/sdk'
import { findConfig, saveToken, success, error } from '../helpers.js'

export function loginCommand() {
  return new Command('login')
    .description('Authenticate with the CPI Auth server')
    .option('-s, --server <url>', 'Server URL (or set CPI_AUTH_SERVER)')
    .option('-e, --email <email>', 'Admin email (or set CPI_AUTH_EMAIL)')
    .option('-p, --password <password>', 'Admin password (or set CPI_AUTH_PASSWORD)')
    .action(async (opts) => {
      let server = opts.server || process.env.CPI_AUTH_SERVER
      let email = opts.email || process.env.CPI_AUTH_EMAIL
      let password = opts.password || process.env.CPI_AUTH_PASSWORD

      // Try config file for server URL
      if (!server) {
        try {
          const config = findConfig()
          server = config.server
        } catch {
          // No config file — server is required
        }
      }

      if (!server || !email || !password) {
        const rl = createInterface({ input: process.stdin, output: process.stdout })
        if (!server) server = await rl.question('Server URL: ')
        if (!email) email = await rl.question('Email: ')
        if (!password) password = await rl.question('Password: ')
        rl.close()
      }

      try {
        const client = new APIClient({ server })
        const tokens = await client.login(email, password)
        saveToken(tokens.access_token)
        success(`Logged in as ${email} on ${server}`)
        success(`Token saved to .cpi-auth-token (expires in ${tokens.expires_in}s)`)
      } catch (e: any) {
        error(`Login failed: ${e.message}`)
        process.exit(1)
      }
    })
}
