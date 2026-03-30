# Command Reference

Complete reference for all CPI Auth CLI commands.

## cpi-auth init

Initialize a new CPI Auth project in the current directory. Creates the configuration file and directory structure.

```bash
cpi-auth init [--server <url>]
```

### Options

| Option | Default | Description |
|--------|---------|-------------|
| `--server <url>` | `http://localhost:5054` | CPI Auth server URL |

### Example

```bash
mkdir my-theme && cd my-theme
cpi-auth init --server http://localhost:5054
```

### Output

```
Initializing CPI Auth project...
Created cpi-auth.config.yaml
Created templates/
Created styles/
Created tokens/
Created strings/

Project initialized. Run 'cpi-auth login' to authenticate.
```

This creates the following structure:

```
./
├── cpi-auth.config.yaml
├── templates/
├── styles/
├── tokens/
│   └── design-tokens.yaml
└── strings/
```

---

## cpi-auth login

Authenticate with the CPI Auth server. Stores credentials locally for subsequent commands.

```bash
cpi-auth login [-e <email>] [-p <password>]
```

### Options

| Option | Description |
|--------|-------------|
| `-e, --email <email>` | Admin email address |
| `-p, --password <password>` | Admin password |

### Interactive Mode

If email or password are omitted, the CLI prompts for them interactively:

```bash
cpi-auth login
# Email: admin@example.com
# Password: ********
# Logged in successfully. Token stored in ~/.cpi-auth/credentials.json
```

### Non-Interactive Mode

```bash
cpi-auth login -e admin@example.com -p my-password
```

### Output

```
Authenticating with http://localhost:5054...
Login successful. Token expires in 3600s.
Credentials stored in ~/.cpi-auth/credentials.json
```

---

## cpi-auth dev

Start the local development server with hot reload and live preview.

```bash
cpi-auth dev [--port <port>]
```

### Options

| Option | Default | Description |
|--------|---------|-------------|
| `--port <port>` | `4400` | Port for the dev server |

### Example

```bash
cpi-auth dev --port 4400
```

### Output

```
CPI Auth dev server running at http://localhost:4400
Watching for changes in templates/, styles/, tokens/, strings/
Press Ctrl+C to stop.
```

See the [Dev Server](/cli/dev-server) page for full details on the development server features.

---

## cpi-auth pull

Pull the latest templates, language strings, and branding settings from the CPI Auth server to the local project.

```bash
cpi-auth pull
```

### Example

```bash
cpi-auth pull
```

### Output

```
Pulling from http://localhost:5054...
  ↓ templates/login.html (updated)
  ↓ templates/signup.html (updated)
  ↓ templates/verification.html (no changes)
  ↓ templates/password_reset.html (no changes)
  ↓ templates/mfa_challenge.html (no changes)
  ↓ templates/error.html (no changes)
  ↓ templates/consent.html (no changes)
  ↓ templates/profile.html (no changes)
  ↓ strings/en.json (updated)
  ↓ strings/de.json (updated)
  ↓ strings/fr.json (no changes)
  ↓ strings/es.json (no changes)

Pull complete. 4 files updated.
```

Local files are overwritten with server versions. Uncommitted local changes will be lost.

---

## cpi-auth push

Push local templates, strings, and tokens to the CPI Auth server.

```bash
cpi-auth push [--dry-run] [--template <type>]
```

### Options

| Option | Description |
|--------|-------------|
| `--dry-run` | Preview changes without applying them |
| `--template <type>` | Push only a specific template type (e.g., `login`) |

### Dry Run

```bash
cpi-auth push --dry-run
```

```
Dry run -- no changes will be applied.

Changes to push:
  → templates/login.html (modified)
  → templates/signup.html (modified)
  → strings/en.json (3 strings added)
  → strings/de.json (3 strings added)
  → tokens/design-tokens.yaml (2 tokens changed)

5 files would be updated. Run without --dry-run to apply.
```

### Push All

```bash
cpi-auth push
```

```
Pushing to http://localhost:5054...
  → templates/login.html (updated)
  → templates/signup.html (updated)
  → strings/en.json (3 strings added)
  → strings/de.json (3 strings added)
  → tokens/design-tokens.yaml (2 tokens changed)

Push complete. 5 files updated.
```

### Push Single Template

```bash
cpi-auth push --template login
```

```
Pushing login template to http://localhost:5054...
  → templates/login.html (updated)

Push complete. 1 file updated.
```

---

## cpi-auth diff

Show the differences between local files and the server state.

```bash
cpi-auth diff
```

### Example

```bash
cpi-auth diff
```

### Output

``` v-pre
Comparing local files with server...

templates/login.html
  - <h1>Welcome</h1>
  + <h1>&#123;&#123;t.login_title&#125;&#125;</h1>

strings/en.json
  + "login.subtitle": "Sign in to continue"
  + "login.remember_me": "Remember me"

tokens/design-tokens.yaml
  ~ colors.primary: #4F46E5 → #2563EB
  ~ spacing.md: 16px → 20px

3 files differ from server.
```

---

## cpi-auth strings list

List language strings for a locale.

```bash
cpi-auth strings list [-l <locale>]
```

### Options

| Option | Default | Description |
|--------|---------|-------------|
| `-l, --locale <locale>` | `en` | Locale to list strings for |

### Example

```bash
cpi-auth strings list -l de
```

### Output

```
Language strings (de):

  login.title              = Willkommen zurueck
  login.email_placeholder  = E-Mail-Adresse
  login.password_placeholder = Passwort
  login.submit_button      = Anmelden
  login.forgot_password    = Passwort vergessen?
  login.signup_link        = Noch kein Konto? Registrieren
  signup.title             = Konto erstellen
  signup.submit_button     = Registrieren
  ...

24 strings found.
```

---

## cpi-auth strings add

Add or update a language string.

```bash
cpi-auth strings add <key> <value> [-l <locale>]
```

### Options

| Option | Default | Description |
|--------|---------|-------------|
| `-l, --locale <locale>` | `en` | Target locale |

### Example

```bash
cpi-auth strings add "login.subtitle" "Sign in to your account" -l en
cpi-auth strings add "login.subtitle" "Melden Sie sich bei Ihrem Konto an" -l de
```

### Output

```
Added string: login.subtitle = "Sign in to your account" (en)
```

---

## cpi-auth strings sync

Synchronize local language string files with the server. Pushes local additions and pulls server updates.

```bash
cpi-auth strings sync
```

### Output

```
Syncing language strings...
  ↑ en: 3 strings pushed
  ↑ de: 3 strings pushed
  ↓ fr: 1 string pulled
  = es: no changes

Sync complete.
```

---

## cpi-auth strings export

Export all language strings to a CSV file.

```bash
cpi-auth strings export [-o <file>]
```

### Options

| Option | Default | Description |
|--------|---------|-------------|
| `-o, --output <file>` | `strings.csv` | Output file path |

### Example

```bash
cpi-auth strings export -o translations.csv
```

### Output

```
Exported 96 strings (4 locales) to translations.csv
```

### CSV Format

```csv
key,en,de,fr,es
login.title,Welcome back,Willkommen zurueck,Bon retour,Bienvenido de nuevo
login.submit_button,Sign in,Anmelden,Se connecter,Iniciar sesion
```

---

## cpi-auth tokens build

Compile design tokens into a CSS file with custom properties.

```bash
cpi-auth tokens build [-o <file>]
```

### Options

| Option | Default | Description |
|--------|---------|-------------|
| `-o, --output <file>` | `./styles/tokens.css` | Output CSS file path |

### Example

```bash
cpi-auth tokens build -o ./styles/tokens.css
```

### Output

```
Building design tokens...
  Read 18 tokens from tokens/design-tokens.yaml
  Generated CSS custom properties
  Written to ./styles/tokens.css

Build complete.
```

### Generated CSS

```css
:root {
  --af-color-primary: #4F46E5;
  --af-color-secondary: #7C3AED;
  --af-color-background: #FFFFFF;
  --af-color-text: #1F2937;
  --af-color-error: #DC2626;
  --af-color-success: #16A34A;
  --af-spacing-xs: 4px;
  --af-spacing-sm: 8px;
  --af-spacing-md: 16px;
  --af-spacing-lg: 24px;
  --af-spacing-xl: 32px;
  --af-radius-sm: 4px;
  --af-radius-md: 8px;
  --af-radius-lg: 16px;
  --af-font-family: 'Inter', sans-serif;
  --af-font-size-sm: 0.875rem;
  --af-font-size-md: 1rem;
  --af-font-size-lg: 1.25rem;
}
```

---

## cpi-auth tokens validate

Validate design tokens for correctness and WCAG contrast compliance.

```bash
cpi-auth tokens validate
```

### Output (Pass)

```
Validating design tokens...
  ✓ All color values are valid
  ✓ Primary/background contrast ratio: 7.2:1 (AAA pass)
  ✓ Text/background contrast ratio: 14.1:1 (AAA pass)
  ✓ All spacing values are valid
  ✓ All radius values are valid

All 18 tokens are valid.
```

### Output (Fail)

```
Validating design tokens...
  ✓ All color values are valid
  ✗ Primary/background contrast ratio: 2.1:1 (WCAG AA requires 4.5:1)
  ✓ Text/background contrast ratio: 14.1:1 (AAA pass)
  ✓ All spacing values are valid
  ✓ All radius values are valid

1 issue found. Fix the above before pushing.
```

---

## cpi-auth validate

Run all validations: templates, design tokens, and language strings.

```bash
cpi-auth validate
```

### Output

```
Running full validation...

Templates:
  ✓ login.html - valid
  ✓ signup.html - valid
  ✓ verification.html - valid
  ✓ password_reset.html - valid
  ✓ mfa_challenge.html - valid
  ✓ error.html - valid
  ✓ consent.html - valid
  ✓ profile.html - valid

Design Tokens:
  ✓ 18 tokens valid
  ✓ WCAG contrast ratios pass

Language Strings:
  ✓ en: 24 strings
  ✓ de: 24 strings
  ⚠ fr: 22 strings (2 missing: mfa.title, mfa.verify_button)
  ✓ es: 24 strings

Validation complete. 1 warning.
```

## cpi-auth setup

One-command setup: create application, roles, permissions, and users in a single step.

```bash
cpi-auth setup [options]
```

### Options

| Option | Default | Description |
|--------|---------|-------------|
| `-s, --server <url>` | env `CPI_AUTH_SERVER` | Server URL |
| `-t, --token <token>` | saved token | Access token |
| `--app-name <name>` | `My Application` | Application name |
| `--app-type <type>` | `spa` | Type: spa, web, native, m2m |
| `--redirect-uri <uri>` | | OAuth redirect URI (comma-separated) |
| `--allowed-origin <origin>` | | CORS origin (comma-separated) |
| `--logout-url <url>` | | Post-logout URL (comma-separated) |
| `--grant-types <types>` | `authorization_code,refresh_token` | Grant types |
| `--create-role <name>` | | Create role (repeatable) |
| `--create-permission <name>` | | Create permission (repeatable) |
| `--create-user <email>` | | Create user (repeatable) |
| `--user-password <pw>` | auto-generated | Password for created users |
| `--user-role <role>` | | Assign role to created users |
| `--output <format>` | `env` | Output: env, json, yaml |

### Examples

```bash
# Full setup with application, roles, and user
cpi-auth setup \
  --server https://auth.example.com \
  --app-name "My SPA" \
  --app-type spa \
  --redirect-uri "https://app.example.com/callback" \
  --allowed-origin "https://app.example.com" \
  --create-permission "posts:read" \
  --create-permission "posts:write" \
  --create-role editor \
  --create-user dev@example.com \
  --user-password "SecurePass123" \
  --user-role editor

# Minimal: just create an application
cpi-auth setup --app-name "API Backend" --app-type m2m --output json

# Output as JSON (for CI/CD pipelines)
cpi-auth setup --app-name "My App" --output json
```

## cpi-auth apps

Manage OAuth applications.

```bash
cpi-auth apps list [--json]
cpi-auth apps create --name <name> [--type <type>] [--redirect-uri <uri>]
cpi-auth apps delete <id>
cpi-auth apps rotate-secret <id>
```

### Examples

```bash
# List all applications
cpi-auth apps list

# Create a web application
cpi-auth apps create \
  --name "Backend API" \
  --type m2m \
  --grant-types client_credentials

# Rotate client secret
cpi-auth apps rotate-secret a1b2c3d4-...
```

## cpi-auth users

Manage users.

```bash
cpi-auth users list [--search <query>] [--json]
cpi-auth users create --email <email> --password <pw> [--name <name>] [--role <role>]
cpi-auth users delete <id>
cpi-auth users block <id>
cpi-auth users unblock <id>
```

### Examples

```bash
# Search for users
cpi-auth users list --search "john"

# Create user with role
cpi-auth users create \
  --email john@example.com \
  --password "SecurePass123" \
  --name "John Doe" \
  --role admin

# Block a user
cpi-auth users block a1b2c3d4-...
```

## cpi-auth roles

Manage roles and permissions.

```bash
cpi-auth roles list [--json]
cpi-auth roles create --name <name> [--permissions <perms>]
cpi-auth roles permissions [--json]
cpi-auth roles create-permission --name <name> [--display-name <name>]
```

### Examples

```bash
# List all roles
cpi-auth roles list

# Create role with permissions
cpi-auth roles create \
  --name editor \
  --permissions "posts:read,posts:write,comments:moderate"

# Create a new permission
cpi-auth roles create-permission --name "billing:manage" --display-name "Manage Billing"
```
