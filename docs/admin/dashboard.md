# Dashboard

The Dashboard is the landing page of the CPI Auth Admin Console. It provides a real-time overview of authentication activity, user metrics, and system health for the selected tenant.

## Metrics Cards

The top row displays five key metric cards, each showing the current value and a percentage change from the previous period:

### Active Users

Total number of users who have logged in within the last 30 days. The change indicator compares against the previous 30-day period.

```
Active Users
    1,247
    +12.3% vs last period
```

### Login Success Rate

Percentage of successful login attempts out of all login attempts. Helps identify brute-force attacks or user experience issues.

```
Login Success Rate
    96.8%
    +0.5% vs last period
```

### MFA Adoption

Percentage of active users who have enabled multi-factor authentication. Useful for tracking security posture improvements.

```
MFA Adoption
    34.2%
    +8.1% vs last period
```

### Active Sessions

Current number of active user sessions across all applications in the tenant.

```
Active Sessions
    892
    -3.2% vs last period
```

### Error Rate

Percentage of authentication requests that resulted in server errors (5xx). A rising error rate may indicate infrastructure issues.

```
Error Rate
    0.12%
    -0.03% vs last period
```

---

## Login Activity Chart

A line chart showing login activity over time, with separate lines for successful and failed logins.

### Period Toggle

Switch between time ranges using the toggle buttons above the chart:

| Period | Description |
|--------|-------------|
| **7D** | Last 7 days, hourly granularity |
| **30D** | Last 30 days, daily granularity |

The chart updates immediately when switching periods. Hovering over a data point shows the exact count and timestamp.

### Reading the Chart

- **Green line**: Successful logins
- **Red line**: Failed login attempts
- A spike in failed logins may indicate a brute-force attack
- A drop in successful logins may indicate an outage or configuration issue

---

## Auth Methods Chart

A donut or pie chart showing the distribution of authentication methods used during the selected period.

Typical segments include:

| Method | Description |
|--------|-------------|
| Password | Standard email + password login |
| MFA (TOTP) | Time-based one-time password |
| MFA (Email) | Email-based verification code |
| Social (Google) | Google OAuth login |
| Social (GitHub) | GitHub OAuth login |
| Client Credentials | M2M token grants |

The chart legend shows both the count and percentage for each method. Click a segment to filter the recent events feed to that method.

---

## Recent Events Feed

A real-time feed of the latest authentication events in the tenant, displayed below the charts.

### Event Entry Format

Each event shows:

```
[icon] [action]    [actor email]    [timestamp]    [IP address]

  ✓   login.success    jane@example.com    2 minutes ago    192.168.1.100
  ✗   login.failed     unknown@test.com    5 minutes ago    10.0.0.45
  →   user.created     admin@example.com   12 minutes ago   10.0.0.1
  ✓   mfa.challenge    bob@example.com     15 minutes ago   172.16.0.88
```

### Event Types Shown

The feed displays the most recent 20 events across these categories:

- `login.success` -- successful authentication
- `login.failed` -- failed login attempt
- `user.created` -- new user registered
- `user.blocked` -- user account blocked
- `user.deleted` -- user account deleted
- `mfa.challenge.success` -- MFA verification passed
- `mfa.challenge.failed` -- MFA verification failed
- `token.revoked` -- token manually revoked

### Interaction

- Click any event to navigate to the related entity (user profile, audit log detail)
- Events auto-refresh every 30 seconds
- A "View all" link at the bottom navigates to the full [Audit Logs](/admin/audit-logs) page

---

## Dashboard Refresh

The dashboard data refreshes automatically:

| Component | Refresh Interval |
|-----------|-----------------|
| Metrics cards | Every 60 seconds |
| Charts | Every 60 seconds |
| Recent events | Every 30 seconds |

A manual refresh button in the top-right corner forces an immediate data reload.

---

## Tenant Context

All dashboard data is scoped to the currently selected tenant. Switch tenants using the tenant selector in the top navigation bar to view metrics for a different environment. The dashboard data reloads immediately upon tenant change.
