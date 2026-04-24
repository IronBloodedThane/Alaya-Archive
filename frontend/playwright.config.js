import { defineConfig, devices } from '@playwright/test'

// Shared across backend (via webServer env) and test code (via process.env)
// so Playwright can mint JWTs that the Go backend will accept.
const E2E_SECRET_KEY = 'e2e-fixed-secret-key-for-local-tests-only'

export default defineConfig({
  testDir: './e2e',
  fullyParallel: false, // one worker — shared SQLite test DB
  workers: 1,
  retries: 0,
  reporter: [['list']],

  use: {
    baseURL: 'http://localhost:5173',
    trace: 'retain-on-failure',
  },

  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],

  webServer: [
    {
      // Backend on :8080 with an isolated test DB and known SECRET_KEY.
      command: 'cd ../backend-go && go run ./cmd/api',
      url: 'http://localhost:8080/healthz',
      reuseExistingServer: false,
      timeout: 60_000,
      env: {
        PORT: '8080',
        DATABASE_PATH: './data/e2e.db',
        DB_JOURNAL_MODE: 'WAL',
        SECRET_KEY: E2E_SECRET_KEY,
        CORS_ORIGINS: 'http://localhost:5173',
        FRONTEND_URL: 'http://localhost:5173',
        // Leave Resend unset — email.Mailer logs instead of sending when API
        // key or from address is empty.
        EMAIL_API_KEY: '',
        EMAIL_FROM: '',
        ACCESS_TOKEN_EXPIRE_MINUTES: '15',
        REFRESH_TOKEN_EXPIRE_DAYS: '30',
      },
    },
    {
      command: 'npm run dev',
      url: 'http://localhost:5173',
      reuseExistingServer: false,
      timeout: 60_000,
    },
  ],
})

export { E2E_SECRET_KEY }
