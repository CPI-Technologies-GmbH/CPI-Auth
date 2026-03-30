import { Command } from 'commander'
import { createInterface } from 'node:readline/promises'
import { APIClient } from '@cpi-auth/sdk'
import { findConfig, saveToken, success, error } from '../helpers.js'

export function loginCommand() {
  return new Command('login')
    .description('Authenticate with the CPI Auth server')
    .option('-e, --email <email>', 'Admin email')
    .option('-p, --password <password>', 'Admin password')
    .action(async (opts) => {
      const config = findConfig()

      let email = opts.email || process.env.CPI_AUTH_EMAIL
      let password = opts.password || process.env.CPI_AUTH_PASSWORD

      if (!email || !password) {
        const rl = createInterface({ input: process.stdin, output: process.stdout })
        if (!email) email = await rl.question('Email: ')
        if (!password) password = await rl.question('Password: ')
        rl.close()
      }

      try {
        const client = new APIClient({ server: config.server, tenantId: config.tenant_id })
        const tokens = await client.login(email, password)
        saveToken(tokens.access_token)
        success(`Logged in as ${email}`)
        success(`Token saved to .cpi-auth-token (expires in ${tokens.expires_in}s)`)
      } catch (e: any) {
        error(`Login failed: ${e.message}`)
        process.exit(1)
      }
    })
}
