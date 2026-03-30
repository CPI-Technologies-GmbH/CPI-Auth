import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'CPI Auth',
  description: 'Open-source Identity & Access Management Platform',
  head: [['link', { rel: 'icon', href: '/favicon.ico' }]],
  ignoreDeadLinks: true,
  themeConfig: {
    logo: '/logo.svg',
    nav: [
      { text: 'Guide', link: '/guide/getting-started' },
      { text: 'API Reference', link: '/api/authentication' },
      { text: 'Admin UI', link: '/admin/overview' },
      { text: 'CLI & SDK', link: '/cli/overview' },
      { text: 'AI Integration', link: '/cli/agent-integration' },
    ],
    sidebar: {
      '/guide/': [
        {
          text: 'Introduction',
          items: [
            { text: 'What is CPI Auth?', link: '/guide/introduction' },
            { text: 'Getting Started', link: '/guide/getting-started' },
            { text: 'Architecture', link: '/guide/architecture' },
            { text: 'Configuration', link: '/guide/configuration' },
          ],
        },
        {
          text: 'Core Concepts',
          items: [
            { text: 'Tenants', link: '/guide/tenants' },
            { text: 'Users', link: '/guide/users' },
            { text: 'Applications', link: '/guide/applications' },
            { text: 'Authentication Flows', link: '/guide/auth-flows' },
            { text: 'Roles & Permissions', link: '/guide/rbac' },
            { text: 'Organizations', link: '/guide/organizations' },
          ],
        },
        {
          text: 'Features',
          items: [
            { text: 'Multi-Factor Authentication', link: '/guide/mfa' },
            { text: 'Social Login', link: '/guide/social-login' },
            { text: 'Custom Fields', link: '/guide/custom-fields' },
            { text: 'Page Templates', link: '/guide/page-templates' },
            { text: 'Email Templates', link: '/guide/email-templates' },
            { text: 'Webhooks & Actions', link: '/guide/webhooks-actions' },
            { text: 'Custom Domains', link: '/guide/custom-domains' },
            { text: 'Audit Logs', link: '/guide/audit-logs' },
          ],
        },
      ],
      '/api/': [
        {
          text: 'API Reference',
          items: [
            { text: 'Authentication', link: '/api/authentication' },
            { text: 'OAuth 2.0 / OIDC', link: '/api/oauth' },
            { text: 'Users', link: '/api/users' },
            { text: 'Applications', link: '/api/applications' },
            { text: 'Tenants', link: '/api/tenants' },
            { text: 'Organizations', link: '/api/organizations' },
            { text: 'Roles & Permissions', link: '/api/roles-permissions' },
            { text: 'Page Templates', link: '/api/page-templates' },
            { text: 'Language Strings', link: '/api/language-strings' },
            { text: 'Webhooks', link: '/api/webhooks' },
            { text: 'Custom Fields', link: '/api/custom-fields' },
            { text: 'Audit Logs', link: '/api/audit-logs' },
          ],
        },
      ],
      '/admin/': [
        {
          text: 'Admin Console',
          items: [
            { text: 'Overview', link: '/admin/overview' },
            { text: 'Dashboard', link: '/admin/dashboard' },
            { text: 'User Management', link: '/admin/users' },
            { text: 'Application Settings', link: '/admin/applications' },
            { text: 'Branding & Theming', link: '/admin/branding' },
            { text: 'Page Template Editor', link: '/admin/page-templates' },
          ],
        },
      ],
      '/cli/': [
        {
          text: 'CLI & SDK',
          items: [
            { text: 'Overview', link: '/cli/overview' },
            { text: 'Installation', link: '/cli/installation' },
            { text: 'CLI Commands', link: '/cli/commands' },
            { text: 'TypeScript SDK', link: '/cli/sdk' },
            { text: 'Design Tokens', link: '/cli/design-tokens' },
            { text: 'Dev Server', link: '/cli/dev-server' },
          ],
        },
        {
          text: 'Integration',
          items: [
            { text: 'AI Agent Integration', link: '/cli/agent-integration' },
          ],
        },
      ],
    },
    socialLinks: [
      { icon: 'github', link: 'https://github.com/cpi-auth/cpi-auth' },
    ],
    footer: {
      message: 'Released under the MIT License.',
      copyright: 'Copyright CPI Auth Contributors',
    },
    search: { provider: 'local' },
  },
})
