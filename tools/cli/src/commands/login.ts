import { Command } from 'commander'
import { createInterface } from 'node:readline/promises'
import { execSync } from 'node:child_process'
import { platform } from 'node:os'
import { APIClient } from '@cpi-auth/sdk'
import { getCurrentContext, addContext, saveTokenToFile, type StoredToken } from '../config.js'
import { success, error, info } from '../helpers.js'

export function loginCommand() {
  return new Command('login')
    .description('Authenticate with a CPI Auth server (opens browser for secure login)')
    .option('-s, --server <url>', 'Server URL (or set CPI_AUTH_SERVER)')
    .option('-e, --email <email>', 'Email (CI/CD fallback, skips device flow)')
    .option('-p, --password <password>', 'Password (CI/CD fallback, skips device flow)')
    .option('--force', 'Force re-authentication even if already logged in')
    .action(async (opts) => {
      // Resolve server and tenantId from current context (unless overridden)
      let server = opts.server || process.env.CPI_AUTH_SERVER || ''
      let tenantId: string | undefined = process.env.CPI_AUTH_TENANT_ID || undefined
      if (!opts.server) {
        const ctx = getCurrentContext()
        if (ctx) {
          if (!server) server = ctx.server
          if (!tenantId) tenantId = ctx.tenantId
        }
      }
      if (!server) {
        const rl = createInterface({ input: process.stdin, output: process.stdout })
        server = await rl.question('Server URL: ')
        rl.close()
      }
      server = server.replace(/\/$/, '')

      // CI/CD fallback: email + password (no browser)
      if (opts.email && opts.password) {
        await loginWithPassword(server, opts.email, opts.password, tenantId)
        return
      }

      // Try device authorization flow
      await loginWithDeviceAuth(server, tenantId)
    })
}

async function loginWithDeviceAuth(server: string, tenantId?: string) {
  const client = new APIClient({ server, tenantId })

  // 1. Request device code
  let deviceResponse: any
  try {
    const res = await fetch(`${server}/oauth/device/code`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ client_id: 'cpi-auth-cli', scope: 'openid profile email admin:access' }),
    })
    if (!res.ok) {
      const text = await res.text()
      throw new Error(`${res.status}: ${text}`)
    }
    deviceResponse = await res.json()
  } catch (e: any) {
    // Fallback to password login if device auth not available
    info(`Device authorization not available on ${server}. Falling back to email/password.`)
    console.log()
    const rl = createInterface({ input: process.stdin, output: process.stdout })
    const email = await rl.question('  Email: ')
    const password = await rl.question('  Password: ')
    rl.close()
    await loginWithPassword(server, email, password, tenantId)
    return
  }

  // 2. Display code and open browser
  const { device_code, user_code, verification_uri, verification_uri_complete, interval } = deviceResponse

  console.log()
  console.log(`  ! First, copy your one-time code: \x1b[1m\x1b[36m${user_code}\x1b[0m`)
  console.log()

  // Try to open browser
  const url = verification_uri_complete || verification_uri
  try {
    const cmd = platform() === 'darwin' ? 'open' : platform() === 'win32' ? 'start' : 'xdg-open'
    execSync(`${cmd} "${url}"`, { stdio: 'ignore' })
    info(`Browser opened to ${url}`)
  } catch {
    info(`Open this URL in your browser: ${url}`)
  }

  console.log()

  // 3. Poll for authorization
  const spinner = ['⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏']
  let i = 0
  const pollInterval = (interval || 5) * 1000

  while (true) {
    await new Promise(r => setTimeout(r, pollInterval))

    process.stdout.write(`\r  ${spinner[i % spinner.length]} Waiting for authorization...`)
    i++

    try {
      const res = await fetch(`${server}/oauth/device/token`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          device_code,
          client_id: 'cpi-auth-cli',
          grant_type: 'urn:ietf:params:oauth:grant-type:device_code',
        }),
      })

      const data = await res.json() as any

      if (res.ok && data.access_token) {
        process.stdout.write('\r')

        // Decode token to get user info
        let email = ''
        let name = ''
        let expiresAt = 0
        try {
          const payload = JSON.parse(Buffer.from(data.access_token.split('.')[1], 'base64url').toString())
          email = payload.email || ''
          name = payload.name || ''
          expiresAt = payload.exp || (Date.now() / 1000 + (data.expires_in || 3600))
        } catch {}

        // Save token
        const stored: StoredToken = {
          access_token: data.access_token,
          refresh_token: data.refresh_token,
          expires_at: expiresAt,
          server,
          tenant_id: tenantId,
          email,
          name,
        }
        saveTokenToFile(server, stored, tenantId)

        // Auto-create context (only if no context for this server+tenant exists)
        const { loadGlobalConfig } = await import('../config.js')
        const config = loadGlobalConfig()
        const hasContext = Object.values(config.contexts).some(
          c => c.server === server && (c['tenant-id'] || undefined) === (tenantId || undefined)
        )
        if (!hasContext) {
          const contextName = server.includes('localhost') ? 'local' : new URL(server).hostname.split('.')[0]
          addContext(contextName, server, tenantId)
          info(`Context "${contextName}" created`)
        }

        success(`Logged in as ${email || 'user'} on ${server}`)
        success(`Token saved to ~/.cpi-auth/tokens/`)
        return
      }

      if (data.error === 'authorization_pending') continue
      if (data.error === 'slow_down') {
        await new Promise(r => setTimeout(r, 5000))
        continue
      }
      if (data.error === 'expired_token') {
        process.stdout.write('\r')
        error('Device code expired. Please try again.')
        process.exit(1)
      }
      if (data.error === 'access_denied') {
        process.stdout.write('\r')
        error('Authorization denied.')
        process.exit(1)
      }
    } catch {
      // Network error — retry
      continue
    }
  }
}

async function loginWithPassword(server: string, email: string, password: string, tenantId?: string) {
  try {
    const client = new APIClient({ server, tenantId })
    const tokens = await client.login(email, password)

    let expiresAt = Date.now() / 1000 + (tokens.expires_in || 3600)
    try {
      const payload = JSON.parse(Buffer.from(tokens.access_token.split('.')[1], 'base64url').toString())
      expiresAt = payload.exp || expiresAt
    } catch {}

    const stored: StoredToken = {
      access_token: tokens.access_token,
      refresh_token: tokens.refresh_token,
      expires_at: expiresAt,
      server,
      tenant_id: tenantId,
      email,
    }
    saveTokenToFile(server, stored, tenantId)

    // Auto-create context (only if no context for this server+tenant exists)
    const { loadGlobalConfig } = await import('../config.js')
    const config = loadGlobalConfig()
    const hasContext = Object.values(config.contexts).some(
      c => c.server === server && (c['tenant-id'] || undefined) === (tenantId || undefined)
    )
    if (!hasContext) {
      const contextName = server.includes('localhost') ? 'local' : new URL(server).hostname.split('.')[0]
      addContext(contextName, server, tenantId)
    }

    success(`Logged in as ${email} on ${server}`)
    success(`Token saved to ~/.cpi-auth/tokens/`)
  } catch (e: any) {
    error(`Login failed: ${e.message}`)
    process.exit(1)
  }
}
