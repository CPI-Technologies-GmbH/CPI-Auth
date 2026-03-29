import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './tests',
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: 1,
  reporter: [['html', { open: 'never' }], ['list']],
  timeout: 60_000,
  use: {
    baseURL: 'http://localhost:5054',
    trace: 'on-first-retry',
    video: 'on',
    screenshot: 'on',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
  webServer: [
    {
      command: 'npm run dev',
      cwd: '../../login-ui',
      url: 'http://localhost:5053',
      reuseExistingServer: true,
      timeout: 30_000,
      env: {
        PUBLIC_API_URL: 'http://localhost:5050',
        PORT: '5053',
        ORIGIN: 'http://localhost:5053',
      },
    },
    {
      command: 'VITE_API_URL=http://localhost:5050 npx vite --port 5054',
      cwd: '../../admin-ui',
      url: 'http://localhost:5054',
      reuseExistingServer: true,
      timeout: 30_000,
    },
  ],
});
