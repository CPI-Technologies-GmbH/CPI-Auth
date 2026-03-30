import { Command } from 'commander'
import { execSync } from 'node:child_process'
import chalk from 'chalk'
import { success, error, info, warn } from '../helpers.js'

const CURRENT_VERSION = '0.1.1'
const NPM_PACKAGE = '@cpi-auth/cli'

export function updateCommand() {
  return new Command('update')
    .description('Check for updates and install the latest version')
    .option('--check', 'Only check, do not install')
    .action(async (opts) => {
      console.log()
      info(`Current version: ${CURRENT_VERSION}`)

      // Check npm for latest version
      let latest = ''
      try {
        latest = execSync(`npm view ${NPM_PACKAGE} version 2>/dev/null`, { encoding: 'utf8' }).trim()
      } catch {
        try {
          const res = await fetch(`https://registry.npmjs.org/${encodeURIComponent(NPM_PACKAGE)}/latest`)
          if (res.ok) {
            const data = await res.json() as any
            latest = data.version || ''
          }
        } catch {
          warn('Could not check for updates (network error)')
          return
        }
      }

      if (!latest) {
        info('Package not yet published to npm. Install from source:')
        console.log(chalk.dim('  cd tools/cli && npm run build && npm link'))
        return
      }

      if (latest === CURRENT_VERSION) {
        success(`You are on the latest version (${CURRENT_VERSION})`)
        return
      }

      info(`New version available: ${chalk.green(latest)} (current: ${CURRENT_VERSION})`)

      if (opts.check) {
        console.log()
        console.log(`  Run ${chalk.cyan('cpi-auth update')} to install.`)
        return
      }

      // Install update
      info('Updating...')
      try {
        execSync(`npm install -g ${NPM_PACKAGE}@latest`, { stdio: 'inherit' })
        success(`Updated to ${latest}`)
      } catch {
        error('Update failed. Try manually:')
        console.log(chalk.dim(`  npm install -g ${NPM_PACKAGE}@latest`))
      }
    })
}

// Check for updates in background (non-blocking, shown as hint)
export async function checkForUpdatesHint(): Promise<void> {
  try {
    const res = await fetch(`https://registry.npmjs.org/${encodeURIComponent(NPM_PACKAGE)}/latest`, {
      signal: AbortSignal.timeout(2000), // 2s timeout
    })
    if (!res.ok) return
    const data = await res.json() as any
    const latest = data.version
    if (latest && latest !== CURRENT_VERSION) {
      console.log(chalk.dim(`  Update available: ${CURRENT_VERSION} → ${latest}. Run \`cpi-auth update\` to install.`))
    }
  } catch {
    // Silently ignore
  }
}
