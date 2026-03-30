#!/usr/bin/env node
import { Command } from 'commander'
import { initCommand } from './commands/init.js'
import { loginCommand } from './commands/login.js'
import { devCommand } from './commands/dev.js'
import { pullCommand } from './commands/pull.js'
import { pushCommand } from './commands/push.js'
import { diffCommand } from './commands/diff.js'
import { stringsCommand } from './commands/strings.js'
import { tokensCommand } from './commands/tokens.js'
import { validateCommand } from './commands/validate.js'

const program = new Command()
  .name('cpi-auth')
  .description('CPI Auth CLI — build, preview, and deploy page templates')
  .version('0.1.0')

program.addCommand(initCommand())
program.addCommand(loginCommand())
program.addCommand(devCommand())
program.addCommand(pullCommand())
program.addCommand(pushCommand())
program.addCommand(diffCommand())
program.addCommand(stringsCommand())
program.addCommand(tokensCommand())
program.addCommand(validateCommand())

program.parse()
