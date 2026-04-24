import { test as base, expect } from '@playwright/test'
import jwt from 'jsonwebtoken'
import { randomUUID } from 'node:crypto'

// Same value the playwright webServer passes to the Go backend as SECRET_KEY.
// Keep them in sync; changing one without the other breaks token minting.
export const E2E_SECRET_KEY = 'e2e-fixed-secret-key-for-local-tests-only'

export const API_URL = 'http://localhost:8080'

// Mints a JWT the Go backend will accept for the given token type.
// Mirrors auth.CreateToken in backend-go/internal/auth/jwt.go.
export function mintToken(userId, type, expiresInSec = 3600) {
  const now = Math.floor(Date.now() / 1000)
  return jwt.sign(
    { sub: userId, type, iat: now, exp: now + expiresInSec },
    E2E_SECRET_KEY,
    { algorithm: 'HS256' }
  )
}

function makeEmail() {
  return `e2e-${randomUUID()}@test.alaya-archive.com`
}

function makeUsername() {
  // Usernames must be >=3 chars, lowercase.
  return `e2e${Date.now().toString(36)}${Math.random().toString(36).slice(2, 6)}`
}

export const test = base.extend({
  // Per-test fixture: returns a function that creates a user via the API
  // and tracks them for teardown. Each test gets a fresh list; teardown
  // deletes every tracked user via /auth/delete-account.
  createUser: async ({ request }, use) => {
    const created = []

    const fn = async (overrides = {}) => {
      const email = overrides.email ?? makeEmail()
      const username = overrides.username ?? makeUsername()
      const password = overrides.password ?? 'TestPassword123!'
      const displayName = overrides.display_name ?? username

      const res = await request.post(`${API_URL}/api/v1/auth/register`, {
        data: { email, username, password, display_name: displayName },
      })
      if (!res.ok()) {
        const body = await res.text()
        throw new Error(`register failed (${res.status()}): ${body}`)
      }
      const tokens = await res.json()
      const user = { email, username, password, display_name: displayName, tokens }
      created.push(user)
      return user
    }

    await use(fn)

    // Teardown: delete every user we created, best-effort.
    for (const user of created) {
      try {
        const res = await request.post(`${API_URL}/api/v1/auth/delete-account`, {
          data: { confirmation: 'DELETE' },
          headers: { Authorization: `Bearer ${user.tokens.access_token}` },
        })
        if (!res.ok() && res.status() !== 401) {
          console.warn(`[e2e] cleanup failed for ${user.email}: ${res.status()}`)
        }
      } catch (err) {
        console.warn(`[e2e] cleanup exception for ${user.email}:`, err.message)
      }
    }
  },
})

export { expect }
