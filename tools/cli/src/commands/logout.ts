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
          deleteToken((ctx as any).server)
        }
        success('Logged out from all servers')
        return
      }

      const server = opts.server || getCurrentContext()?.server
      if (!server) {
        error('No server specified. Use --server or set a context first.')
        return
      }

      deleteToken(server)
      success(`Logged out from ${server}`)
    })
}
