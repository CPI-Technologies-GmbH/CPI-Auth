# TypeScript SDK Reference

The `@cpi-auth/sdk` package provides a TypeScript API for programmatic access to CPI Auth resources. Use it to build custom tooling, automation scripts, or CI/CD integrations.

## Installation

```bash
npm install @cpi-auth/sdk
```

## CPI Auth Class

The main entry point. Create an instance with your server URL and authentication credentials.

### Constructor

```typescript
import { CPI Auth } from '@cpi-auth/sdk';

// Authenticate with email and password
const client = new CPI Auth({
  server: 'http://localhost:5054',
  credentials: {
    email: 'admin@example.com',
    password: 'your-password'
  },
  tenantId: 'your-tenant-uuid'
});

// Or authenticate with an existing token
const client = new CPI Auth({
  server: 'http://localhost:5054',
  token: 'eyJhbGciOiJSUzI1NiIs...',
  tenantId: 'your-tenant-uuid'
});
```

### Constructor Options

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `server` | string | Yes | CPI Auth server URL |
| `credentials` | object | One of | `{ email, password }` for login-based auth |
| `token` | string | One of | Pre-existing access token |
| `tenantId` | string | Yes | Target tenant UUID |

---

## templates

Manage page templates.

### templates.list()

Retrieve all templates for the current tenant.

```typescript
const templates = await client.templates.list();

console.log(templates);
// [
//   { id: 'tpl-1', name: 'Default Login', type: 'login', is_default: true, ... },
//   { id: 'tpl-2', name: 'Custom Login', type: 'login', is_default: false, ... }
// ]
```

### templates.get(id)

Get a single template by ID.

```typescript
const template = await client.templates.get('tpl-uuid-1');

console.log(template.html);
console.log(template.css);
```

### templates.create(data)

Create a new custom template.

```typescript v-pre
const template = await client.templates.create({
  name: 'Dark Login',
  type: 'login',
  html: `<!DOCTYPE html>
<html>
<head><style>&#123;&#123;css&#125;&#125;</style></head>
<body>
  <div class="container">
    <h1>&#123;&#123;t.login_title&#125;&#125;</h1>
    <form method="POST">
      <input type="email" name="email" placeholder="&#123;&#123;t.login_email_placeholder&#125;&#125;">
      <input type="password" name="password" placeholder="&#123;&#123;t.login_password_placeholder&#125;&#125;">
      <button type="submit">&#123;&#123;t.login_submit_button&#125;&#125;</button>
    </form>
  </div>
</body>
</html>`,
  css: `body { background: #1a1a2e; color: #e0e0e0; }`
});

console.log(template.id); // 'tpl-uuid-new'
```

### templates.update(id, data)

Update an existing custom template.

```typescript
const updated = await client.templates.update('tpl-uuid-2', {
  name: 'Dark Login v2',
  css: 'body { background: #0a0a1a; color: #f0f0f0; }'
});
```

### templates.delete(id)

Delete a custom template.

```typescript
await client.templates.delete('tpl-uuid-2');
```

### templates.duplicate(id, name)

Duplicate a template (including defaults).

```typescript
const copy = await client.templates.duplicate('tpl-uuid-1', 'My Custom Login');

console.log(copy.id); // new template ID
console.log(copy.is_default); // false
```

---

## strings

Manage language strings for localization.

### strings.list(locale?)

List all language strings, optionally filtered by locale.

```typescript
// All strings
const allStrings = await client.strings.list();

// English only
const enStrings = await client.strings.list('en');

console.log(enStrings);
// [
//   { string_key: 'login.title', locale: 'en', value: 'Welcome back' },
//   { string_key: 'login.submit_button', locale: 'en', value: 'Sign in' }
// ]
```

### strings.set(key, locale, value)

Create or update a language string.

```typescript
await client.strings.set('login.subtitle', 'en', 'Sign in to continue');
await client.strings.set('login.subtitle', 'de', 'Melden Sie sich an, um fortzufahren');
await client.strings.set('login.subtitle', 'fr', 'Connectez-vous pour continuer');
await client.strings.set('login.subtitle', 'es', 'Inicia sesion para continuar');
```

### strings.delete(key, locale)

Delete a language string for a specific locale.

```typescript
await client.strings.delete('login.subtitle', 'fr');
```

---

## tokens

Build and validate design tokens.

### tokens.buildCSS(tokensPath?)

Compile design tokens from YAML into CSS custom properties.

```typescript
const css = await client.tokens.buildCSS('./tokens/design-tokens.yaml');

console.log(css);
// :root {
//   --af-color-primary: #4F46E5;
//   --af-color-secondary: #7C3AED;
//   ...
// }
```

### tokens.validate(tokensPath?)

Validate design tokens for correctness and WCAG contrast compliance.

```typescript
const result = await client.tokens.validate('./tokens/design-tokens.yaml');

console.log(result.valid); // true or false
console.log(result.issues);
// [
//   { type: 'warning', message: 'Primary/background contrast ratio: 3.1:1 (below WCAG AA 4.5:1)' }
// ]
```

---

## preview

Render template previews with sample data.

### preview.render(templateId, options?)

Render a template with sample data for preview purposes.

```typescript
const html = await client.preview.render('tpl-uuid-1', {
  locale: 'en',
  user: {
    name: 'Jane Doe',
    email: 'jane@example.com'
  },
  customFields: {
    company: 'Acme Corp'
  }
});

// html contains the fully rendered HTML string
```

---

## sync

Synchronize local files with the server.

### sync.diff()

Compare local files against the server state.

```typescript
const changes = await client.sync.diff();

console.log(changes);
// {
//   templates: [
//     { file: 'login.html', status: 'modified' },
//     { file: 'signup.html', status: 'unchanged' }
//   ],
//   strings: [
//     { locale: 'en', added: 2, removed: 0, modified: 1 }
//   ],
//   tokens: { status: 'modified', changes: ['colors.primary'] }
// }
```

### sync.push(options?)

Push local changes to the server.

```typescript
// Push everything
const result = await client.sync.push();
console.log(result.updated); // number of files updated

// Dry run
const preview = await client.sync.push({ dryRun: true });
console.log(preview.changes); // list of what would change

// Push specific template
const single = await client.sync.push({ template: 'login' });
```

### sync.pull()

Pull the latest state from the server to local files.

```typescript
const result = await client.sync.pull();
console.log(result.updated); // number of files updated
```

---

## Error Handling

All SDK methods throw `CPI AuthError` on failure.

```typescript
import { CPI Auth, CPI AuthError } from '@cpi-auth/sdk';

try {
  await client.templates.update('tpl-default', { name: 'Modified' });
} catch (error) {
  if (error instanceof CPI AuthError) {
    console.error(error.status);  // 403
    console.error(error.code);    // 'forbidden'
    console.error(error.message); // 'Default templates are read-only'
  }
}
```

---

## Full Example: CI/CD Deployment

```typescript
import { CPI Auth } from '@cpi-auth/sdk';

async function deploy() {
  const client = new CPI Auth({
    server: process.env.CPI_AUTH_SERVER!,
    credentials: {
      email: process.env.CPI_AUTH_EMAIL!,
      password: process.env.CPI_AUTH_PASSWORD!
    },
    tenantId: process.env.CPI_AUTH_TENANT_ID!
  });

  // Validate before pushing
  const validation = await client.tokens.validate();
  if (!validation.valid) {
    console.error('Token validation failed:', validation.issues);
    process.exit(1);
  }

  // Preview changes
  const diff = await client.sync.diff();
  console.log('Changes to deploy:', diff);

  // Push if there are changes
  const result = await client.sync.push();
  console.log(`Deployed ${result.updated} files successfully.`);
}

deploy().catch(console.error);
```

---

## Full Example: Translation Management

```typescript
import { CPI Auth } from '@cpi-auth/sdk';
import { readFileSync } from 'fs';
import { parse } from 'csv-parse/sync';

async function importTranslations(csvPath: string) {
  const client = new CPI Auth({
    server: 'http://localhost:5054',
    token: process.env.CPI_AUTH_TOKEN!,
    tenantId: 'tenant-uuid'
  });

  const csv = readFileSync(csvPath, 'utf-8');
  const records = parse(csv, { columns: true });

  for (const row of records) {
    const key = row.key;
    if (row.en) await client.strings.set(key, 'en', row.en);
    if (row.de) await client.strings.set(key, 'de', row.de);
    if (row.fr) await client.strings.set(key, 'fr', row.fr);
    if (row.es) await client.strings.set(key, 'es', row.es);
  }

  console.log(`Imported ${records.length} translation keys.`);
}

importTranslations('./translations.csv');
```
