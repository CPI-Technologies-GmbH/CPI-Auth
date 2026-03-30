# Custom Fields

Custom fields let you extend the user profile with tenant-specific data. Define fields once, and they automatically appear on registration forms, profile pages, and in page templates.

## Field Types

| Type | Code | Description |
|------|------|-------------|
| Text | `text` | Single-line text input |
| Number | `number` | Numeric input |
| Email | `email` | Email address with validation |
| Telephone | `tel` | Phone number input |
| URL | `url` | URL with validation |
| Date | `date` | Date picker |
| Checkbox | `checkbox` | Boolean toggle |
| Select | `select` | Dropdown with predefined options |
| Textarea | `textarea` | Multi-line text input |

## Field Definition Model

```json
{
  "id": "f1a2b3c4-d5e6-7890-fghi-j12345678901",
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "company_name",
  "label": "Company Name",
  "field_type": "text",
  "placeholder": "Enter your company name",
  "description": "The user's employer or organization",
  "options": null,
  "required": true,
  "visible_on": "both",
  "position": 1,
  "validation_rules": {
    "min_length": 2,
    "max_length": 100
  },
  "is_active": true,
  "created_at": "2025-01-15T10:30:00Z",
  "updated_at": "2025-01-15T10:30:00Z"
}
```

### Field Reference

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Machine-readable identifier (snake_case, unique per tenant) |
| `label` | string | Human-readable label displayed in forms |
| `field_type` | string | One of the supported field types |
| `placeholder` | string | Placeholder text for the input |
| `description` | string | Help text shown below the field |
| `options` | JSON | Options for `select` fields |
| `required` | bool | Whether the field must be filled |
| `visible_on` | string | Where the field appears |
| `position` | int | Display order (lower = higher) |
| `validation_rules` | JSON | Additional validation constraints |
| `is_active` | bool | Whether the field is currently active |

## Visibility

The `visible_on` property controls where the field is shown:

| Value | Description |
|-------|-------------|
| `registration` | Only on the signup form |
| `profile` | Only on the user profile / account page |
| `both` | On both registration and profile pages |

## Validation Rules

The `validation_rules` JSON object supports the following properties depending on field type:

```json
// Text fields
{
  "min_length": 2,
  "max_length": 100,
  "pattern": "^[a-zA-Z\\s]+$"
}

// Number fields
{
  "min": 0,
  "max": 999
}

// Select fields (options defined separately)
{
  "allow_multiple": false
}
```

## Select Field Options

For `select` fields, define the available options in the `options` property:

```json
{
  "options": [
    { "value": "startup", "label": "Startup (1-10)" },
    { "value": "small", "label": "Small Business (11-50)" },
    { "value": "medium", "label": "Medium (51-200)" },
    { "value": "large", "label": "Large (201-1000)" },
    { "value": "enterprise", "label": "Enterprise (1000+)" }
  ]
}
```

## Managing Custom Fields via API

### Create a Field

```bash
curl -X POST http://localhost:5050/api/v1/custom-fields \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "company_name",
    "label": "Company Name",
    "field_type": "text",
    "placeholder": "Enter your company name",
    "required": true,
    "visible_on": "both",
    "position": 1,
    "validation_rules": {
      "min_length": 2,
      "max_length": 100
    }
  }'
```

### Create a Select Field

```bash
curl -X POST http://localhost:5050/api/v1/custom-fields \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "company_size",
    "label": "Company Size",
    "field_type": "select",
    "required": false,
    "visible_on": "registration",
    "position": 2,
    "options": [
      { "value": "startup", "label": "Startup (1-10)" },
      { "value": "small", "label": "Small Business (11-50)" },
      { "value": "medium", "label": "Medium (51-200)" },
      { "value": "enterprise", "label": "Enterprise (200+)" }
    ]
  }'
```

### List Fields

```bash
curl http://localhost:5050/api/v1/custom-fields \
  -H "Authorization: Bearer $TOKEN"
```

### Update a Field

```bash
curl -X PATCH http://localhost:5050/api/v1/custom-fields/{field_id} \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "label": "Organization Name",
    "required": false
  }'
```

### Delete a Field

```bash
curl -X DELETE http://localhost:5050/api/v1/custom-fields/{field_id} \
  -H "Authorization: Bearer $TOKEN"
```

## Custom Fields in Page Templates

Custom fields are rendered in page templates using two template variables:

### `&#123;&#123;custom_fields&#125;&#125;`

Renders all fields with `visible_on` set to `registration` or `both`. Use this in signup page templates.

```html v-pre
<form>
  <input type="email" name="email" placeholder="Email" required>
  <input type="password" name="password" placeholder="Password" required>

  <!-- Custom registration fields -->
  &#123;&#123;custom_fields&#125;&#125;

  <button type="submit">Sign Up</button>
</form>
```

### `&#123;&#123;profile_fields&#125;&#125;`

Renders all fields with `visible_on` set to `profile` or `both`. Use this in profile page templates.

```html v-pre
<form>
  <input type="text" name="name" value="&#123;&#123;user.name&#125;&#125;">
  <input type="email" name="email" value="&#123;&#123;user.email&#125;&#125;" readonly>

  <!-- Custom profile fields -->
  &#123;&#123;profile_fields&#125;&#125;

  <button type="submit">Save</button>
</form>
```

Each field is rendered as an appropriate HTML input element with its label, placeholder, validation attributes, and current value (if editing a profile).

## Custom Field Data Storage

Custom field values are stored in the user's `metadata` JSON column. The field `name` is used as the JSON key:

```json
{
  "company_name": "Acme Corp",
  "company_size": "medium",
  "job_title": "Senior Engineer"
}
```

This data is accessible through the user API and is included in the user object when fetched.

## Next Steps

- [Page Templates](./page-templates) -- Customize how custom fields are rendered
- [Users](./users) -- User metadata and custom data
- [Email Templates](./email-templates) -- Include custom field data in emails
