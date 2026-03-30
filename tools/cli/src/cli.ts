#!/usr/bin/env node
import { Command } from 'commander'
import { initCommand } from './commands/init.js'
import { loginCommand } from './commands/login.js'
import { statusCommand } from './commands/status.js'
import { setupCommand } from './commands/setup.js'
import { devCommand } from './commands/dev.js'
import { pullCommand } from './commands/pull.js'
import { pushCommand } from './commands/push.js'
import { diffCommand } from './commands/diff.js'
import { stringsCommand } from './commands/strings.js'
import { tokensCommand } from './commands/tokens.js'
import { validateCommand } from './commands/validate.js'
import { appsCommand } from './commands/apps.js'
import { usersCommand } from './commands/users.js'
import { rolesCommand } from './commands/roles.js'

const program = new Command()
  .name('cpi-auth')
  .description('CPI Auth CLI — manage your identity platform from the command line')
  .version('0.1.0')

// Authentication
program.addCommand(loginCommand())
program.addCommand(statusCommand())

// Quick setup
program.addCommand(setupCommand())

// Resource management
program.addCommand(appsCommand())
program.addCommand(usersCommand())
program.addCommand(rolesCommand())

// Template development
program.addCommand(initCommand())
program.addCommand(devCommand())
program.addCommand(pullCommand())
program.addCommand(pushCommand())
program.addCommand(diffCommand())
program.addCommand(stringsCommand())
program.addCommand(tokensCommand())
program.addCommand(validateCommand())

program.parse()
