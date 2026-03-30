# Installation

Install the CPI Auth CLI to manage page templates, design tokens, and language strings from the command line.

## Prerequisites

- **Node.js 20+** -- required runtime
- **npm 9+** or **pnpm** or **yarn** -- package manager
- **Access to an CPI Auth server** -- local or remote instance

Verify your Node.js version:

```bash
node --version
# v20.0.0 or higher
```

## Installation Methods

### Global Install (Recommended)

Install the CLI globally for system-wide access:

```bash
npm install -g @cpi-auth/cli
```

After installation, the `cpi-auth` command is available in your terminal:

```bash
cpi-auth --version
# @cpi-auth/cli v1.0.0
```

### npx (No Install)

Run the CLI without installing it globally using npx:

```bash
npx @cpi-auth/cli init
npx @cpi-auth/cli dev
npx @cpi-auth/cli push
```

This downloads and runs the latest version each time. Useful for one-off commands or CI/CD pipelines.

### From Source

Clone the CPI Auth repository and link the CLI locally:

```bash
git clone https://github.com/your-org/cpi-auth.git
cd cpi-auth/sdks/typescript
npm install
npm run build
npm link
```

This links the locally built CLI to your global `cpi-auth` command. Changes to the source code take effect after rebuilding.

### Project-Local Install

Install as a dev dependency in your project:

```bash
npm install --save-dev @cpi-auth/cli
```

Then use it via npx or npm scripts:

```json
{
  "scripts": {
    "af:dev": "cpi-auth dev --port 4400",
    "af:push": "cpi-auth push",
    "af:validate": "cpi-auth validate",
    "af:pull": "cpi-auth pull"
  }
}
```

```bash
npm run af:dev
```

## Installing the SDK

If you need programmatic access to CPI Auth (for custom tooling, CI/CD scripts, or automation), install the SDK:

```bash
npm install @cpi-auth/sdk
```

The SDK can be used independently or alongside the CLI:

```typescript
import { CPI Auth } from '@cpi-auth/sdk';

const client = new CPI Auth({
  server: 'http://localhost:5054',
  token: 'your-access-token',
  tenantId: 'your-tenant-uuid'
});

const templates = await client.templates.list();
```

## Verify Installation

After installation, verify that the CLI is working:

```bash
# Check version
cpi-auth --version

# View help
cpi-auth --help

# View command-specific help
cpi-auth init --help
```

Expected output for `--help`:

```
CPI Auth CLI - Manage authentication page design systems

Usage:
  cpi-auth <command> [options]

Commands:
  init          Initialize a new CPI Auth project
  login         Authenticate with the CPI Auth server
  dev           Start the local development server
  pull          Pull templates and strings from server
  push          Push local changes to server
  diff          Show differences between local and server
  strings       Manage language strings
  tokens        Build and validate design tokens
  validate      Validate templates and tokens

Options:
  --version     Show version number
  --help        Show help
```

## Authentication Setup

After installing, authenticate with your CPI Auth server:

```bash
# Initialize a project (creates config file)
cpi-auth init --server http://localhost:5054

# Log in with admin credentials
cpi-auth login -e admin@example.com -p your-password
```

The login command stores a session token locally for subsequent commands. Tokens are stored in `~/.cpi-auth/credentials.json` and refresh automatically.

## Upgrading

### Global Install

```bash
npm update -g @cpi-auth/cli
```

### Project-Local Install

```bash
npm update @cpi-auth/cli
```

### Verify After Upgrade

```bash
cpi-auth --version
```

## Troubleshooting

### Command Not Found

If `cpi-auth` is not recognized after global installation, ensure the npm global bin directory is in your PATH:

```bash
# Find npm global bin directory
npm config get prefix

# Add to PATH (add to your ~/.zshrc or ~/.bashrc)
export PATH="$(npm config get prefix)/bin:$PATH"
```

### Permission Errors

On macOS/Linux, if you get EACCES errors during global install:

```bash
# Option 1: Use a node version manager (recommended)
# nvm, fnm, or volta handle permissions automatically

# Option 2: Fix npm permissions
mkdir -p ~/.npm-global
npm config set prefix '~/.npm-global'
export PATH="~/.npm-global/bin:$PATH"
```

### Connection Issues

If the CLI cannot connect to the CPI Auth server:

```bash
# Verify server is running
curl http://localhost:5054/.well-known/openid-configuration

# Check your config
cat cpi-auth.config.yaml

# Re-authenticate
cpi-auth login -e admin@example.com -p your-password
```
