import { Command } from 'commander'
import { initCommand } from './init.js'
import { devCommand } from './dev.js'
import { pullCommand } from './pull.js'
import { pushCommand } from './push.js'
import { diffCommand } from './diff.js'
import { validateCommand } from './validate.js'

export function templatesCommand() {
  const cmd = new Command('templates')
    .description('Template development: init, preview, push/pull, validate')

  cmd.addCommand(initCommand())
  cmd.addCommand(devCommand())
  cmd.addCommand(pullCommand())
  cmd.addCommand(pushCommand())
  cmd.addCommand(diffCommand())
  cmd.addCommand(validateCommand())

  return cmd
}
