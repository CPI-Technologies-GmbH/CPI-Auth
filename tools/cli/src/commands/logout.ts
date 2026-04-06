import { Command } from 'commander'
import { getCurrentContext, deleteToken } from '../config.js'
import { success, error, info } from '../helpers.js'

export function logoutCommand() {
  return new Command('logout')
    .description('Log out and clear saved credentials')
    .option('-s, --server <url>', 'Server to log out from (defaults to current context)')
    .option('--all', 'Log out from all servers')
    .action((opts) => {
      if (opts.all) {
        const { loadGlobalConfig } = require('../config.js')
        const config = loadGlobalConfig()
        for (const ctx of Object.values(config.contexts)) {
          const c = ctx as any
          deleteToken(c.server, c['tenant-id'])
        }
        success('Logged out from all servers')
        return
      }

      const ctx = getCurrentContext()
      const server = opts.server || ctx?.server
      const tenantId = opts.server ? undefined : ctx?.tenantId
      if (!server) {
        error('No server specified. Use --server or set a context first.')
        return
      }

      deleteToken(server, tenantId)
      success(`Logged out from ${server}`)
    })
}
