# Custom Fields API

Custom fields extend the user profile with additional data specific to your application. Fields can appear on registration forms, profile pages, or both.

## Base URL

```
http://localhost:5054/admin/custom-fields
```

---

## List Custom Fields

### GET /admin/custom-fields

```bash
curl http://localhost:5054/admin/custom-fields \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}"
```

**Response 200 OK:**

```json
{
  "data": [
    {
      "id": "cf-uuid-1",
      "name": "company",
      "label": "Company Name",
      "type": "text",
      "required": true,
      "visibility": "both",
      "placeholder": "Enter your company name",
      "description": "The name of your organization",
      "validation": {
        "min_length": 2,
        "max_length": 100
      },
      "options": null,
      "sort_order": 1,
      "created_at": "2025-06-15T10:00:00Z",
      "updated_at": "2026-03-28T08:00:00Z"
    },
    {
      "id": "cf-uuid-2",
      "name": "department",
      "label": "Department",
      "type": "select",
      "required": false,
      "visibility": "profile",
      "placeholder": "Select department",
      "description": null,
      "validation": null,
      "options": [
        { "label": "Engineering", "value": "engineering" },
        { "label": "Marketing", "value": "marketing" },
        { "label": "Sales", "value": "sales" },
        { "label": "Support", "value": "support" }
      ],
      "sort_order": 2,
      "created_at": "2025-08-01T10:00:00Z",
      "updated_at": "2026-01-15T14:00:00Z"
    }
  ],
  "total": 2
}
```

---

## Create Custom Field

### POST /admin/custom-fields

### Text Field Example

```bash
curl -X POST http://localhost:5054/admin/custom-fields \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: {tenant_id}" \
  -d '{
    "name": "job_title",
    "label": "Job Title",
    "type": "text",
    "required": false,
    "visibility": "both",
    "placeholder": "e.g. Software Engineer",
    "description": "Your current role or position",
    "validation": {
      "min_length": 2,
      "max_length": 50
    },
    "sort_order": 3
  }'
```

### Select Field Example

```bash
curl -X POST http://localhost:5054/admin/custom-fields \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: {tenant_id}" \
  -d '{
    "name": "plan_type",
    "label": "Plan Type",
    "type": "select",
    "required": true,
    "visibility": "registration",
    "options": [
      { "label": "Free", "value": "free" },
      { "label": "Pro", "value": "pro" },
      { "label": "Enterprise", "value": "enterprise" }
    ],
    "sort_order": 4
  }'
```

### Checkbox Field Example

```bash
curl -X POST http://localhost:5054/admin/custom-fields \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: {tenant_id}" \
  -d '{
    "name": "newsletter_opt_in",
    "label": "Subscribe to newsletter",
    "type": "checkbox",
    "required": false,
    "visibility": "registration",
    "description": "Receive product updates and announcements",
    "sort_order": 5
  }'
```

**Response 201 Created:**

```json
{
  "id": "cf-uuid-3",
  "name": "job_title",
  "label": "Job Title",
  "type": "text",
  "required": false,
  "visibility": "both",
  "placeholder": "e.g. Software Engineer",
  "description": "Your current role or position",
  "validation": {
    "min_length": 2,
    "max_length": 50
  },
  "options": null,
  "sort_order": 3,
  "created_at": "2026-03-28T12:00:00Z",
  "updated_at": "2026-03-28T12:00:00Z"
}
```

---

## Get Custom Field

### GET /admin/custom-fields/:id

```bash
curl http://localhost:5054/admin/custom-fields/cf-uuid-1 \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}"
```

---

## Update Custom Field

### PATCH /admin/custom-fields/:id

```bash
curl -X PATCH http://localhost:5054/admin/custom-fields/cf-uuid-2 \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: {tenant_id}" \
  -d '{
    "label": "Department / Team",
    "options": [
      { "label": "Engineering", "value": "engineering" },
      { "label": "Marketing", "value": "marketing" },
      { "label": "Sales", "value": "sales" },
      { "label": "Support", "value": "support" },
      { "label": "Design", "value": "design" }
    ]
  }'
```

**Response 200 OK:**

Returns the full updated custom field object.

---

## Delete Custom Field

### DELETE /admin/custom-fields/:id

Permanently removes the field definition and all stored user values for this field.

```bash
curl -X DELETE http://localhost:5054/admin/custom-fields/cf-uuid-3 \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}"
```

**Response 204 No Content:**

::: warning
Deleting a custom field permanently removes all user data stored for that field. This action cannot be undone.
:::

---

## Field Types

| Type | Description | Validation Options |
|------|-------------|-------------------|
| `text` | Single-line text input | `min_length`, `max_length`, `pattern` |
| `textarea` | Multi-line text input | `min_length`, `max_length` |
| `number` | Numeric input | `min`, `max` |
| `email` | Email address input | Auto-validated email format |
| `tel` | Phone number input | `pattern` |
| `url` | URL input | Auto-validated URL format |
| `date` | Date picker | `min_date`, `max_date` |
| `checkbox` | Boolean toggle | None |
| `select` | Dropdown selection | Requires `options` array |

---

## Visibility Options

| Value | Description |
|-------|-------------|
| `registration` | Shown only on the signup form |
| `profile` | Shown only on the user profile page |
| `both` | Shown on both registration and profile |

---

## Validation Rules

Validation rules vary by field type. All rules are optional.

### Text / Textarea

```json
{
  "validation": {
    "min_length": 2,
    "max_length": 500,
    "pattern": "^[a-zA-Z\\s]+$"
  }
}
```

### Number

```json
{
  "validation": {
    "min": 0,
    "max": 1000
  }
}
```

### Date

```json
{
  "validation": {
    "min_date": "2000-01-01",
    "max_date": "2030-12-31"
  }
}
```

### Tel

```json
{
  "validation": {
    "pattern": "^\\+?[1-9]\\d{1,14}$"
  }
}
```

---

## Custom Field Data in User Profiles

Custom field values are stored in the user's `metadata` object and returned when fetching user details:

```bash
curl http://localhost:5054/admin/users/user-uuid-1 \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}"
```

```json
{
  "id": "user-uuid-1",
  "email": "jane@example.com",
  "name": "Jane Doe",
  "metadata": {
    "company": "Acme Corp",
    "department": "engineering",
    "job_title": "Senior Engineer",
    "newsletter_opt_in": true
  }
}
```

## Fields Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Unique machine-readable identifier (snake_case) |
| `label` | string | Yes | Human-readable display label |
| `type` | string | Yes | One of the supported field types |
| `required` | boolean | No | Whether the field is required (default: false) |
| `visibility` | string | Yes | `registration`, `profile`, or `both` |
| `placeholder` | string | No | Input placeholder text |
| `description` | string | No | Help text shown below the field |
| `validation` | object | No | Type-specific validation rules |
| `options` | array | Conditional | Required for `select` type fields |
| `sort_order` | integer | No | Display order (lower = first) |
