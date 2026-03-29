# Dev Server

The CPI Auth dev server provides a local development environment for previewing and iterating on authentication page templates with hot reload.

## Starting the Dev Server

```bash
cpi-auth dev [--port <port>]
```

| Option | Default | Description |
|--------|---------|-------------|
| `--port` | `4400` | Port number for the dev server |

```bash
cpi-auth dev --port 4400
```

Output:

```
CPI Auth dev server running at http://localhost:4400
Watching for changes in:
  templates/  (HTML templates)
  styles/     (CSS stylesheets)
  tokens/     (design tokens YAML)
  strings/    (language string JSON files)

Press Ctrl+C to stop.
```

Open `http://localhost:4400` in your browser to access the dev server UI.

---

## Dev Server UI

The dev server renders a preview environment with a sidebar for navigation and controls.

### Layout

```
┌──────────────┬──────────────────────────────────────┐
│   Sidebar    │                                      │
│              │                                      │
│  Pages       │          Template Preview             │
│  ├ Login     │                                      │
│  ├ Signup    │     (rendered HTML with live          │
│  ├ Verify    │      CSS and sample data)             │
│  ├ Reset     │                                      │
│  ├ MFA       │                                      │
│  ├ Error     │                                      │
│  ├ Consent   │                                      │
│  └ Profile   │                                      │
│              │                                      │
│  Controls    │                                      │
│  [Locale  v] ��                                      │
│  [Viewport v]│                                      │
│              │                                      │
└───────���──────┴───────────────��──────────────────────┘
```

### Sidebar: Page List

The sidebar lists all page templates found in the `templates/` directory. Click a page name to preview it in the main area.

| Page | Template File |
|------|---------------|
| Login | `templates/login.html` |
| Signup | `templates/signup.html` |
| Verification | `templates/verification.html` |
| Password Reset | `templates/password_reset.html` |
| MFA Challenge | `templates/mfa_challenge.html` |
| Error | `templates/error.html` |
| Consent | `templates/consent.html` |
| Profile | `templates/profile.html` |

Custom templates (any `.html` file in `templates/`) also appear in the list.

### Locale Switcher

A dropdown to switch the preview locale. Language strings are resolved based on the selected locale.

| Locale | Language |
|--------|----------|
| `en` | English |
| `de` | German |
| `fr` | French |
| `es` | Spanish |

Switching the locale immediately re-renders the preview with the corresponding language strings from the `strings/` directory.

### Viewport Toggle

Switch between viewport sizes to test responsive behavior:

| Viewport | Width |
|----------|-------|
| Desktop | 1280px |
| Tablet | 768px |
| Mobile | 375px |

The preview area resizes to match the selected viewport. A visible frame indicates the viewport boundaries.

---

## Hot Reload

The dev server watches all project files and automatically reloads the preview when changes are detected.

### Watched Directories

| Directory | File Types | Action on Change |
|-----------|------------|-----------------|
| `templates/` | `.html` | Re-render preview |
| `styles/` | `.css` | Inject updated styles |
| `tokens/` | `.yaml` | Rebuild CSS tokens, inject styles |
| `strings/` | `.json` | Re-render with new strings |

### Behavior

- **CSS changes** are injected without a full page reload (style-only hot reload)
- **HTML changes** trigger a full preview re-render
- **Token changes** recompile to CSS, then inject the new styles
- **String changes** re-render the preview with updated localization

Changes are reflected within approximately 100ms of saving the file.

---

## Preview Rendering

The preview renders templates with sample data to simulate the production environment.

### Sample Data

The dev server injects default sample data for template variables:

```json
{
  "user": {
    "name": "Jane Doe",
    "email": "jane@example.com",
    "locale": "en",
    "email_verified": true
  },
  "branding": {
    "logo_url": "/assets/sample-logo.svg",
    "logo_dark_url": "/assets/sample-logo-dark.svg",
    "primary_color": "#4F46E5",
    "secondary_color": "#7C3AED",
    "font_family": "Inter"
  },
  "application": {
    "name": "Sample Application",
    "logo_url": null
  },
  "csrf_token": "dev-csrf-token-placeholder",
  "error": null,
  "redirect_uri": "http://localhost:3000/callback"
}
```

### Error State Preview

Toggle the error state to test how templates render error messages:

- Click the **"Show Error"** toggle in the sidebar
- The `&#123;&#123;error&#125;&#125;` variable is populated with a sample error message
- Useful for verifying error styling and placement

---

## Custom Fields Simulation

The dev server simulates custom fields so you can test how `&#123;&#123;custom_fields&#125;&#125;` renders in templates.

### Configuration

Add sample custom fields in `cpi-auth.config.yaml`:

```yaml
dev:
  custom_fields:
    - name: company
      label: Company Name
      type: text
      required: true
      placeholder: "Enter your company"
    - name: department
      label: Department
      type: select
      required: false
      options:
        - { label: "Engineering", value: "engineering" }
        - { label: "Marketing", value: "marketing" }
        - { label: "Sales", value: "sales" }
    - name: newsletter
      label: Subscribe to newsletter
      type: checkbox
      required: false
```

These fields render as HTML form elements wherever `&#123;&#123;custom_fields&#125;&#125;` appears in a template.

---

## Language String Resolution

Language strings referenced with `&#123;&#123;t.key&#125;&#125;` are resolved from the local `strings/` directory JSON files.

### Resolution Order

1. Load the JSON file for the selected locale (e.g., `strings/de.json`)
2. For any missing keys, fall back to the default locale (`strings/en.json`)
3. If still missing, render the raw key name as `[missing: key]`

### Example

`strings/en.json`:

```json
{
  "login.title": "Welcome back",
  "login.email_placeholder": "Email address",
  "login.password_placeholder": "Password",
  "login.submit_button": "Sign in",
  "login.forgot_password": "Forgot password?",
  "login.signup_link": "Don't have an account? Sign up"
}
```

`strings/de.json`:

```json
{
  "login.title": "Willkommen zurueck",
  "login.email_placeholder": "E-Mail-Adresse",
  "login.password_placeholder": "Passwort",
  "login.submit_button": "Anmelden",
  "login.forgot_password": "Passwort vergessen?",
  "login.signup_link": "Noch kein Konto? Registrieren"
}
```

Switching the locale in the sidebar toggles between these string sets.

### Missing String Indicator

If a key is missing from the selected locale and the fallback locale, the preview renders:

```html
<h1>[missing: login.subtitle]</h1>
```

This makes it easy to identify untranslated strings during development.

---

## Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl+R` / `Cmd+R` | Force reload preview |
| `1` - `4` | Switch locale (1=en, 2=de, 3=fr, 4=es) |
| `D` | Toggle desktop viewport |
| `T` | Toggle tablet viewport |
| `M` | Toggle mobile viewport |
| `E` | Toggle error state |

---

## Stopping the Server

Press `Ctrl+C` in the terminal to stop the dev server.

```
^C
Dev server stopped.
```
