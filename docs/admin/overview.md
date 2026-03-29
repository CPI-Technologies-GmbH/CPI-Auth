# Admin Console Overview

The CPI Auth Admin Console is a web-based management interface for configuring and monitoring your authentication system. It provides full control over users, applications, branding, templates, and security settings.

## Accessing the Console

The admin console runs on **port 5054** by default:

```
http://localhost:5054
```

Log in with your admin credentials (email and password) created during initial setup.

## Navigation

The admin console is organized into the following sections, accessible from the left sidebar:

### Core

| Page | Description |
|------|-------------|
| **Dashboard** | Metrics, charts, and recent activity feed |
| **Users** | Manage user accounts, sessions, roles, and profiles |
| **Applications** | Configure OAuth applications and API clients |
| **Organizations** | Manage B2B organization grouping |

### Access Control

| Page | Description |
|------|-------------|
| **Roles** | Define and manage user roles |
| **Permissions** | Create and organize granular permissions |

### Customization

| Page | Description |
|------|-------------|
| **Branding** | Colors, logos, fonts, and layout settings |
| **Page Templates** | HTML/CSS templates for login, signup, and other pages |
| **Language Strings** | Manage localized text for all page templates |
| **Custom Fields** | Define additional user profile fields |

### Integration

| Page | Description |
|------|-------------|
| **Webhooks** | Configure event-driven HTTP callbacks |

### Monitoring

| Page | Description |
|------|-------------|
| **Audit Logs** | View and export the activity audit trail |

### Settings

| Page | Description |
|------|-------------|
| **Tenant Settings** | Session lifetime, password policy, MFA settings |
| **Tenants** | Manage multi-tenant environments |

---

## Tenant Selector

If you manage multiple tenants, a **tenant selector** dropdown appears in the top navigation bar. Switching tenants changes the context for all admin operations. All data displayed (users, applications, settings) belongs to the selected tenant.

```
[Production v] | [English v] | admin@example.com | [Logout]
```

## Language Switcher

The admin console interface supports multiple languages. The language switcher in the top navigation bar allows administrators to change the console display language. This does not affect the language of user-facing pages, which is controlled by tenant settings and user locale preferences.

## Session Management

Admin sessions have a configurable timeout (default: 1 hour). When your session expires, you are redirected to the login page. The admin console uses JWT tokens for authentication, with automatic token refresh when the session is still valid.

## Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `/` | Focus search bar |
| `Esc` | Close modal or dialog |
| `Ctrl+S` / `Cmd+S` | Save current form |

## Browser Support

The admin console supports the latest two versions of:

- Chrome
- Firefox
- Safari
- Edge

## Responsive Design

The admin console is optimized for desktop screens (1280px and wider). Tablet and mobile layouts are functional but designed primarily for occasional use.

## Getting Started

1. Navigate to `http://localhost:5054`
2. Log in with your admin email and password
3. The **Dashboard** loads by default, showing key metrics
4. Use the sidebar to navigate between sections
5. Select a tenant from the dropdown if managing multiple tenants

For details on each section, see the dedicated pages:

- [Dashboard](/admin/dashboard)
- [Users](/admin/users)
- [Applications](/admin/applications)
- [Branding](/admin/branding)
- [Page Templates](/admin/page-templates)
