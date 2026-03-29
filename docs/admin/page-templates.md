# Page Templates Management

The Page Templates section allows you to customize the HTML and CSS of every user-facing authentication page. CPI Auth ships with default templates for each page type, and you can create custom templates or duplicate defaults to build your own.

## Template List

The main page displays templates organized into two sections.

### Default Templates

Built-in templates provided by CPI Auth. These are marked with a **lock icon** indicating they are read-only.

- Cannot be edited or deleted (returns 403)
- Can be **duplicated** to create an editable copy
- Each page type has exactly one default template
- Updated automatically with CPI Auth upgrades

### Custom Templates

User-created templates. These can be freely edited, duplicated, or deleted.

- Marked with a pencil icon
- One custom template per page type can be set as **active**
- Inactive custom templates are stored but not rendered to users

### List Columns

| Column | Description |
|--------|-------------|
| **Name** | Template display name |
| **Type** | Page type badge (login, signup, etc.) |
| **Status** | Active (green) / Inactive (gray) |
| **Default** | Lock icon if read-only default |
| **Modified** | Last update timestamp |
| **Actions** | Edit, Duplicate, Delete |

### Search and Filter

- **Search**: Filter templates by name
- **Type filter**: Dropdown to show only specific page types
- **Status filter**: Show active, inactive, or all templates

---

## Creating a Custom Template

Click **"Create Template"** to open the creation dialog:

1. **Name**: Give the template a descriptive name
2. **Type**: Select the page type (login, signup, verification, etc.)
3. **HTML**: Enter the initial HTML content
4. **CSS**: Enter the initial CSS styles

The new template is created as inactive. Set it to active to replace the current template for that page type.

---

## Duplicating a Default Template

Since default templates cannot be edited, the recommended workflow is:

1. Find the default template you want to customize
2. Click the **"Duplicate"** button
3. Enter a name for the copy (e.g., "Custom Login - Dark Theme")
4. The duplicate is created as an inactive custom template
5. Edit the copy as needed
6. Set it to active when ready

---

## Template Editor

Click any custom template to open the full editor. The editor provides three tabs.

### HTML Tab

A code editor with syntax highlighting for writing the template HTML. Features:

- Syntax highlighting for HTML and Handlebars template syntax
- Line numbers
- Auto-indentation
- Bracket matching
- Find and replace

Example HTML structure:

```html v-pre
<!DOCTYPE html>
<html lang="&#123;&#123;user.locale&#125;&#125;">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>&#123;&#123;t.login_title&#125;&#125; - &#123;&#123;application.name&#125;&#125;</title>
  <style>&#123;&#123;css&#125;&#125;</style>
</head>
<body>
  <div class="container">
    <div class="logo">
      <img src="&#123;&#123;branding.logo_url&#125;&#125;" alt="Logo">
    </div>

    <h1>&#123;&#123;t.login_title&#125;&#125;</h1>

    &#123;&#123;#if error&#125;&#125;
      <div class="error">&#123;&#123;error&#125;&#125;</div>
    &#123;&#123;/if&#125;&#125;

    <form method="POST">
      <input type="hidden" name="csrf_token" value="&#123;&#123;csrf_token&#125;&#125;">

      <div class="field">
        <label for="email">&#123;&#123;t.login_email_placeholder&#125;&#125;</label>
        <input type="email" id="email" name="email" required>
      </div>

      <div class="field">
        <label for="password">&#123;&#123;t.login_password_placeholder&#125;&#125;</label>
        <input type="password" id="password" name="password" required>
      </div>

      &#123;&#123;custom_fields&#125;&#125;

      <button type="submit">&#123;&#123;t.login_submit_button&#125;&#125;</button>
    </form>

    <p class="link">
      <a href="/forgot-password">&#123;&#123;t.login_forgot_password&#125;&#125;</a>
    </p>
    <p class="link">&#123;&#123;t.login_signup_link&#125;&#125;</p>
  </div>
</body>
</html>
```

### CSS Tab

A CSS editor for styling the template. Branding design tokens are available as CSS custom properties.

```css
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: var(--af-font-family);
  background-color: var(--af-color-background);
  color: var(--af-color-text);
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
}

.container {
  max-width: 420px;
  width: 100%;
  padding: 2rem;
}

.logo img {
  max-height: 48px;
  margin-bottom: 2rem;
}

h1 {
  font-size: 1.5rem;
  margin-bottom: 1.5rem;
}

.field {
  margin-bottom: 1rem;
}

.field label {
  display: block;
  font-size: 0.875rem;
  margin-bottom: 0.25rem;
}

.field input {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid #d1d5db;
  border-radius: var(--af-border-radius);
  font-size: 1rem;
}

button[type="submit"] {
  width: 100%;
  padding: 0.75rem;
  background-color: var(--af-color-primary);
  color: white;
  border: none;
  border-radius: var(--af-border-radius);
  font-size: 1rem;
  cursor: pointer;
  margin-top: 0.5rem;
}

button[type="submit"]:hover {
  opacity: 0.9;
}

.error {
  background-color: #fef2f2;
  color: #dc2626;
  padding: 0.75rem;
  border-radius: var(--af-border-radius);
  margin-bottom: 1rem;
}

.link {
  text-align: center;
  margin-top: 1rem;
  font-size: 0.875rem;
}

.link a {
  color: var(--af-color-primary);
  text-decoration: none;
}
```

### Preview Tab

A live preview of the template rendered with sample data. The preview applies:

- Current branding settings (colors, logo, fonts, border radius)
- Sample user data for template variables
- Selected locale for language string resolution
- Custom fields with placeholder values

---

## Variable Toolbar

Above the HTML editor, a toolbar provides quick-insert buttons for common template variables.

### Variable Categories

**User**: `&#123;&#123;user.name&#125;&#125;`, `&#123;&#123;user.email&#125;&#125;`, `&#123;&#123;user.locale&#125;&#125;`

**Branding**: `&#123;&#123;branding.logo_url&#125;&#125;`, `&#123;&#123;branding.primary_color&#125;&#125;`

**Application**: `&#123;&#123;application.name&#125;&#125;`

**Page**: `&#123;&#123;csrf_token&#125;&#125;`, `&#123;&#123;error&#125;&#125;`, `&#123;&#123;redirect_uri&#125;&#125;`, `&#123;&#123;css&#125;&#125;`, `&#123;&#123;custom_fields&#125;&#125;`

**Language Strings**: `&#123;&#123;t.login_title&#125;&#125;`, `&#123;&#123;t.login_submit_button&#125;&#125;`, etc.

Click any variable button to insert it at the current cursor position in the HTML editor.

---

## Language Strings Dialog

A **"Language Strings"** button in the toolbar opens a dialog for managing the localized text used by `&#123;&#123;t.key&#125;&#125;` variables.

The dialog shows:

| Key | English | German | French | Spanish |
|-----|---------|--------|--------|---------|
| `login.title` | Welcome back | Willkommen | Bon retour | Bienvenido |
| `login.submit_button` | Sign in | Anmelden | Se connecter | Iniciar sesion |

- Edit values inline by clicking a cell
- Add new keys with the "Add String" button
- Changes are saved to the language strings API

See [Language Strings](/api/language-strings) for the full API reference.

---

## Page Types

| Type | Description | Key Variables |
|------|-------------|---------------|
| `login` | User login form | `&#123;&#123;t.login_*&#125;&#125;`, `&#123;&#123;error&#125;&#125;` |
| `signup` | Registration form | `&#123;&#123;t.signup_*&#125;&#125;`, `&#123;&#123;custom_fields&#125;&#125;` |
| `verification` | Email verification | `&#123;&#123;t.verification_*&#125;&#125;` |
| `password_reset` | Password reset flow | `&#123;&#123;t.password_reset_*&#125;&#125;` |
| `mfa_challenge` | MFA code entry | `&#123;&#123;t.mfa_*&#125;&#125;` |
| `error` | Error display page | `&#123;&#123;error&#125;&#125;`, `&#123;&#123;t.error_*&#125;&#125;` |
| `consent` | OAuth consent screen | `&#123;&#123;t.consent_*&#125;&#125;`, `&#123;&#123;application.name&#125;&#125;` |
| `profile` | User profile page | `&#123;&#123;user.*&#125;&#125;`, `&#123;&#123;custom_fields&#125;&#125;` |
| `custom` | Custom page | All variables available |

---

## Workflow

1. Browse default templates to find a starting point
2. Duplicate the default template for the page you want to customize
3. Edit the HTML and CSS in the editor
4. Use the Preview tab to verify rendering
5. Test with different locales and viewport sizes
6. Set the template to active when satisfied
7. The custom template now renders for all users on that page type
