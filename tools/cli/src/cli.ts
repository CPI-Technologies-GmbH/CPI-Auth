#!/usr/bin/env node
import { Command } from 'commander'

// Auth & status
import { loginCommand } from './commands/login.js'
import { logoutCommand } from './commands/logout.js'
import { statusCommand } from './commands/status.js'
import { updateCommand } from './commands/update.js'

// Config / context management
import { configCommand } from './commands/configCmd.js'

// Quick setup & resource management
import { setupCommand } from './commands/setup.js'
import { appsCommand } from './commands/apps.js'
import { usersCommand } from './commands/users.js'
import { rolesCommand } from './commands/roles.js'

// Template development
import { templatesCommand } from './commands/templates.js'
import { stringsCommand } from './commands/strings.js'
import { tokensCommand } from './commands/tokens.js'

const program = new Command()
  .name('cpi-auth')
  .description('CPI Auth CLI — manage your identity platform from the command line')
  .version('0.1.1')

// ─── Authentication ────────────────────────────────────────
program.addCommand(loginCommand())
program.addCommand(logoutCommand())
program.addCommand(statusCommand())

// ─── Configuration ─────────────────────────────────────────
program.addCommand(configCommand())
program.addCommand(updateCommand())

// ─── Resource Management ───────────────────────────────────
program.addCommand(setupCommand())
program.addCommand(appsCommand())
program.addCommand(usersCommand())
program.addCommand(rolesCommand())

// ─── Template Development ──────────────────────────────────
program.addCommand(templatesCommand())
program.addCommand(stringsCommand())
program.addCommand(tokensCommand())

program.parse()
