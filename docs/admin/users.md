# Users Management

The Users section of the Admin Console provides a complete interface for managing user accounts, viewing sessions, assigning roles, and performing administrative actions.

## User List

The main users page displays a searchable, paginated table of all users in the current tenant.

### Table Columns

| Column | Description |
|--------|-------------|
| **Email** | User's email address (primary identifier) |
| **Name** | Display name |
| **Status** | Badge: `Active` (green), `Blocked` (red), `Unverified` (yellow) |
| **Last Login** | Timestamp of most recent login |
| **Created** | Account creation date |
| **Actions** | Quick-action menu |

### Search and Filtering

- **Search bar**: Filter by email or name (real-time as you type)
- **Status filter**: Dropdown to show only Active, Blocked, or Unverified users
- **Pagination**: Navigate pages with configurable page size (10, 20, 50, 100)

### Sorting

Click any column header to sort. Click again to reverse the order. Default sort is by creation date (newest first).

---

## Create User Dialog

Click the **"Create User"** button in the top-right corner to open the creation dialog.

### Fields

| Field | Required | Description |
|-------|----------|-------------|
| Email | Yes | Must be unique within the tenant |
| Password | Yes | Must meet the tenant's password policy |
| Name | No | User's display name |
| Phone | No | Phone number |
| Locale | No | Preferred language (en, de, fr, es) |

Custom fields configured in the tenant also appear in this dialog if their visibility is set to `registration` or `both`.

After creation, the user receives a status of `Unverified` unless email verification is disabled in tenant settings.

---

## User Detail Page

Click any user in the list to open their detail page. The detail page is organized into tabs.

### Profile Tab

Displays and allows editing of core user fields:

- **Email** (read-only after creation)
- **Name**
- **Phone**
- **Locale**
- **Email Verified** toggle
- **Custom field values** (metadata)

A **Save** button persists changes. A **Reset Password** button opens a dialog to set a new password or send a reset email.

### Roles Tab

Shows all roles assigned to the user with a checkbox list of available roles.

- Check a role to assign it
- Uncheck to remove it
- Changes are saved immediately
- The user's effective permissions (union of all role permissions) are displayed below

### Sessions Tab

Lists all active sessions for the user:

| Column | Description |
|--------|-------------|
| IP Address | Source IP of the session |
| User Agent | Browser and OS information |
| Last Active | When the session was last used |
| Created | When the session was initiated |
| Actions | Revoke button for individual session |

A **"Revoke All Sessions"** button at the top force-logs out the user from all devices.

### MFA Tab

Displays the user's multi-factor authentication status:

- **MFA Enabled**: Whether MFA is currently active
- **Methods**: Which MFA methods are configured (TOTP, Email)
- **Recovery Codes**: Whether recovery codes have been generated
- **Disable MFA** button: Removes MFA configuration (requires confirmation)

### Identities Tab

Shows linked authentication identities:

| Provider | Email | Linked At |
|----------|-------|-----------|
| Password | jane@example.com | 2025-06-15 |
| Google | jane@gmail.com | 2025-08-20 |

### Audit Log Tab

A filtered view of the audit log showing only events related to this user. Displays the same format as the main audit log page but pre-filtered by user ID.

### Metadata Tab

A JSON editor for the user's metadata object. This includes custom field values and any additional key-value data stored on the user profile.

```json
{
  "company": "Acme Corp",
  "department": "engineering",
  "job_title": "Senior Engineer",
  "newsletter_opt_in": true
}
```

---

## Administrative Actions

The following actions are available from the user detail page or the quick-action menu in the user list.

### Impersonation

Click **"Impersonate"** to generate a temporary access token that acts as the selected user. This opens a new browser tab with the user's session. Impersonation events are recorded in the audit log.

### Password Reset

Two options:

1. **Set password directly**: Enter a new password in the dialog
2. **Send reset email**: Triggers a password reset email to the user

### Force Logout

Revokes all active sessions for the user. The user must log in again on all devices.

### Block / Unblock

- **Block**: Prevents the user from logging in. Active sessions are revoked immediately. The user sees a "blocked" error message on login attempt.
- **Unblock**: Restores login access. The user can authenticate normally.

### Delete

Permanently removes the user account and all associated data. This requires confirmation and cannot be undone.

---

## Bulk Operations

Select multiple users using the checkboxes in the user list. A bulk actions toolbar appears at the top of the table.

### Available Bulk Actions

| Action | Description |
|--------|-------------|
| **Block Selected** | Block all selected users |
| **Unblock Selected** | Unblock all selected users |
| **Delete Selected** | Permanently delete all selected users |
| **Export Selected** | Download selected users as CSV |

All bulk actions require confirmation before execution.

### Select All

The header checkbox toggles selection for the current page. A "Select all N users" link appears to extend the selection across all pages matching the current filter.

---

## User Export

Click the **"Export"** button in the top-right to download all users (or filtered results) as a CSV file. The export includes:

- User ID, email, name, phone, locale
- Status (active, blocked, unverified)
- Login count, last login date
- Creation date
- Custom field values

## User Import

Click the **"Import"** button to upload a CSV file of users. The import dialog shows:

1. **File upload** area (drag-and-drop or file picker)
2. **Column mapping** preview showing how CSV columns map to user fields
3. **Validation results** before committing the import
4. **Import summary** after completion (imported, skipped, errors)

### Expected CSV Format

```csv
email,password,name,phone,locale
john@example.com,TempPass123!,John Doe,+1234567890,en
jane@example.com,TempPass456!,Jane Doe,+0987654321,de
```

Duplicate emails are skipped. Passwords must meet the tenant's password policy.
