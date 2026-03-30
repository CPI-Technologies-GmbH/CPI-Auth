# CLI & SDK Overview

The CPI Auth CLI and TypeScript SDK provide developer tools for building, previewing, and deploying design systems for CPI Auth authentication pages. They enable a code-first workflow for managing page templates, language strings, and design tokens.

## What It Provides

- **Local development server** with hot reload and preview rendering
- **Template management** -- pull, edit, validate, and push page templates
- **Language string management** -- list, add, sync, and export localized strings
- **Design token system** -- define colors, spacing, and typography as tokens, compiled to CSS custom properties
- **Diff and sync** -- compare local changes against the server before pushing
- **Validation** -- check templates for errors and design tokens for WCAG contrast compliance

## Use Case

The CLI is designed for teams that want to manage their CPI Auth design system as code, using version control and standard development workflows:

- Design system engineers building branded authentication experiences
- Frontend teams iterating on login/signup page designs
- DevOps teams deploying template changes through CI/CD pipelines
- Translation teams managing language strings across locales

## Two Packages

### @cpi-auth/cli

The command-line tool for interacting with CPI Auth from the terminal.

```bash
npm install -g @cpi-auth/cli
```

Provides commands like `cpi-auth init`, `cpi-auth dev`, `cpi-auth push`, etc.

### @cpi-auth/sdk

The TypeScript library for programmatic access to CPI Auth resources.

```bash
npm install @cpi-auth/sdk
```

Use it to build custom tooling, automation scripts, or CI/CD integrations.

## Workflow

The typical workflow follows four steps:

### 1. Initialize

```bash
cpi-auth init --server http://localhost:5054
```

Creates an `cpi-auth.config.yaml` and pulls current templates, strings, and tokens into a local project directory.

### 2. Develop

```bash
cpi-auth dev --port 4400
```

Starts a local development server with hot reload. Edit HTML, CSS, and design tokens -- changes preview instantly.

### 3. Validate

```bash
cpi-auth validate
```

Checks templates for syntax errors, validates design tokens for WCAG contrast compliance, and verifies language string completeness.

### 4. Push

```bash
cpi-auth push
```

Deploys local changes to the CPI Auth server. Use `--dry-run` to preview what would change before applying.

## Project Structure

After running `cpi-auth init`, your project directory looks like this:

```
my-cpi-auth-theme/
├── cpi-auth.config.yaml     # Configuration file
├── templates/
│   ├── login.html            # Login page template
│   ├── signup.html           # Signup page template
│   ├── verification.html     # Email verification template
│   ├── password_reset.html   # Password reset template
│   ├── mfa_challenge.html    # MFA challenge template
│   ├── error.html            # Error page template
│   ├── consent.html          # OAuth consent template
│   └── profile.html          # User profile template
├── styles/
│   └── main.css              # Global stylesheet
├── tokens/
│   └── design-tokens.yaml    # Design token definitions
└── strings/
    ├── en.json               # English language strings
    ├── de.json               # German language strings
    ├── fr.json               # French language strings
    └── es.json               # Spanish language strings
```

## Configuration

The `cpi-auth.config.yaml` file stores project settings:

```yaml
server: http://localhost:5054
tenant_id: your-tenant-uuid

templates:
  directory: ./templates

styles:
  directory: ./styles

tokens:
  directory: ./tokens
  output: ./styles/tokens.css

strings:
  directory: ./strings
  locales:
    - en
    - de
    - fr
    - es
```

## Next Steps

- [Installation](/cli/installation) -- install the CLI and SDK
- [Commands](/cli/commands) -- full command reference
- [SDK Reference](/cli/sdk) -- TypeScript API documentation
- [Design Tokens](/cli/design-tokens) -- token system documentation
- [Dev Server](/cli/dev-server) -- local development server guide
