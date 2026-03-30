import { Command } from 'commander'
import { writeFileSync, mkdirSync, existsSync } from 'node:fs'
import { stringify as stringifyYAML } from 'yaml'
import { success, warn, info } from '../helpers.js'

const defaultConfig = {
  server: 'http://localhost:5050',
  tenant_id: 'a0000000-0000-0000-0000-000000000001',
  tokens: {
    colors: {
      primary: '#6366f1',
      'primary-hover': '#4f46e5',
      background: '#0f172a',
      surface: '#1e293b',
      border: '#334155',
      text: '#e2e8f0',
      'text-muted': '#94a3b8',
      error: '#f87171',
      success: '#34d399',
    },
    spacing: { xs: '0.25rem', sm: '0.5rem', md: '1rem', lg: '1.5rem', xl: '2.5rem' },
    radius: { sm: '6px', md: '8px', lg: '16px' },
    typography: {
      'font-family': '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
      'font-size-sm': '0.8125rem',
      'font-size-base': '0.875rem',
      'font-size-lg': '1.5rem',
    },
  },
  templates: {
    login: { html: 'templates/login.html', css: 'templates/login.css' },
    signup: { html: 'templates/signup.html', css: 'templates/signup.css' },
    profile: { html: 'templates/profile.html', css: 'templates/profile.css' },
  },
  locales: ['en', 'de'],
  preview: {
    custom_fields: [
      { label: 'Company', type: 'text', required: true, placeholder: 'Acme Inc.' },
      { label: 'Phone', type: 'tel', required: false, placeholder: '+1 555-1234' },
    ],
    sample_data: {
      'user.name': 'Jane Smith',
      'user.email': 'jane@acme.com',
      'application.name': 'Acme Portal',
      'tenant.name': 'Acme Corp',
    },
  },
}

const starterHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{{application.name}}</title>
</head>
<body>
  <div class="card">
    <h1>{{t.login.title}}</h1>
    <p class="subtitle">{{t.login.subtitle}}</p>
    <form>
      <div class="field">
        <label>{{t.login.email}}</label>
        <input type="email" placeholder="you@example.com" />
      </div>
      <div class="field">
        <label>{{t.login.password}}</label>
        <input type="password" placeholder="Enter your password" />
      </div>
      <button type="submit" class="btn-primary">{{t.login.submit}}</button>
    </form>
  </div>
</body>
</html>`

const starterCSS = `/* Uses design tokens: var(--af-color-primary), var(--af-spacing-md), etc. */
* { margin: 0; padding: 0; box-sizing: border-box; }
body {
  font-family: var(--af-font-family);
  background: var(--af-color-background);
  color: var(--af-color-text);
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--af-spacing-md);
}
.card {
  background: var(--af-color-surface);
  border-radius: var(--af-radius-lg);
  padding: var(--af-spacing-xl);
  width: 100%;
  max-width: 420px;
  box-shadow: 0 25px 50px -12px rgba(0,0,0,.5);
}
h1 { text-align: center; font-size: var(--af-font-size-lg); font-weight: 700; margin-bottom: 0.25rem; }
.subtitle { text-align: center; color: var(--af-color-text-muted); font-size: var(--af-font-size-base); margin-bottom: 2rem; }
.field { margin-bottom: 1.25rem; }
.field label { display: block; font-size: var(--af-font-size-sm); font-weight: 500; margin-bottom: 0.5rem; color: var(--af-color-text); }
.field input {
  width: 100%; padding: 0.75rem 1rem;
  background: var(--af-color-background);
  border: 1px solid var(--af-color-border);
  border-radius: var(--af-radius-md);
  color: var(--af-color-text);
  font-size: var(--af-font-size-base);
  outline: none;
}
.field input:focus { border-color: var(--af-color-primary); }
.btn-primary {
  width: 100%; padding: 0.75rem;
  background: var(--af-color-primary);
  color: white; border: none;
  border-radius: var(--af-radius-md);
  font-size: var(--af-font-size-base);
  font-weight: 600; cursor: pointer;
}
.btn-primary:hover { background: var(--af-color-primary-hover); }`

const defaultStrings = {
  login: {
    title: 'Welcome back',
    subtitle: 'Sign in to your account',
    email: 'Email address',
    password: 'Password',
    submit: 'Sign in',
    forgot_password: 'Forgot password?',
  },
  signup: {
    title: 'Create your account',
    subtitle: 'Get started in seconds',
    name: 'Full name',
    email: 'Email address',
    password: 'Password',
    submit: 'Create account',
  },
  profile: {
    title: 'Your Profile',
    subtitle: 'Manage your account settings',
    name: 'Full name',
    email: 'Email address',
    save: 'Save changes',
  },
}

export function initCommand() {
  return new Command('init')
    .description('Initialize a new CPIAuth design project')
    .option('--server <url>', 'Server URL', 'http://localhost:5050')
    .action((opts) => {
      if (existsSync('cpi-auth.config.yaml')) {
        warn('cpi-auth.config.yaml already exists. Skipping.')
        return
      }

      const config = { ...defaultConfig, server: opts.server }

      // Create directories
      mkdirSync('templates', { recursive: true })
      mkdirSync('strings', { recursive: true })

      // Write config
      writeFileSync('cpi-auth.config.yaml', stringifyYAML(config, { lineWidth: 120 }))
      success('Created cpi-auth.config.yaml')

      // Write starter templates
      writeFileSync('templates/login.html', starterHTML)
      writeFileSync('templates/login.css', starterCSS)
      writeFileSync('templates/signup.html', starterHTML.replace(/login/g, 'signup'))
      writeFileSync('templates/signup.css', starterCSS)
      writeFileSync('templates/profile.html', starterHTML.replace(/login/g, 'profile'))
      writeFileSync('templates/profile.css', starterCSS)
      success('Created templates/ with starter files')

      // Write language strings
      writeFileSync('strings/en.yaml', stringifyYAML(defaultStrings))
      writeFileSync('strings/de.yaml', stringifyYAML({
        login: { title: 'Willkommen zurueck', subtitle: 'Melden Sie sich an', email: 'E-Mail-Adresse', password: 'Passwort', submit: 'Anmelden' },
        signup: { title: 'Konto erstellen', subtitle: 'Starten Sie in Sekunden', name: 'Name', email: 'E-Mail', password: 'Passwort', submit: 'Registrieren' },
        profile: { title: 'Ihr Profil', subtitle: 'Kontoeinstellungen verwalten', name: 'Name', email: 'E-Mail', save: 'Speichern' },
      }))
      success('Created strings/ with en.yaml and de.yaml')

      // .gitignore
      writeFileSync('.cpi-auth-token', '')
      if (!existsSync('.gitignore')) {
        writeFileSync('.gitignore', '.cpi-auth-token\nnode_modules/\ndist/\n')
      }

      info('')
      info('Next steps:')
      info('  1. cpi-auth login          — authenticate with your server')
      info('  2. cpi-auth pull           — pull existing templates')
      info('  3. cpi-auth dev            — start the dev server')
      info('  4. cpi-auth push           — deploy to your tenant')
    })
}
