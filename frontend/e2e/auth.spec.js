import { test, expect, mintToken, API_URL } from './fixtures.js'
import { randomUUID } from 'node:crypto'

test.describe('Registration', () => {
  test('register → dashboard', async ({ page }) => {
    await page.goto('/register')

    const email = `e2e-${randomUUID()}@test.alaya-archive.com`
    const username = `e2e${Date.now().toString(36)}`

    await page.getByLabel('Email').fill(email)
    await page.getByLabel('Username').fill(username)
    await page.getByLabel('Display Name').fill(username)
    await page.getByLabel('Password').fill('TestPassword123!')
    await page.getByRole('button', { name: 'Create Account' }).click()

    await expect(page).toHaveURL('/dashboard')

    // Cleanup manually since this test doesn't use the createUser fixture.
    const tokens = {
      access_token: await page.evaluate(() => localStorage.getItem('access_token')),
    }
    await page.request.post(`${API_URL}/api/v1/auth/delete-account`, {
      data: { confirmation: 'DELETE' },
      headers: { Authorization: `Bearer ${tokens.access_token}` },
    })
  })
})

test.describe('Login', () => {
  test('existing user can sign in', async ({ page, createUser }) => {
    const user = await createUser()

    await page.goto('/login')
    await page.getByLabel('Email or Username').fill(user.username)
    await page.getByLabel('Password').fill(user.password)
    await page.getByRole('button', { name: 'Sign In' }).click()

    await expect(page).toHaveURL('/dashboard')
  })

  test('wrong password shows error', async ({ page, createUser }) => {
    const user = await createUser()

    await page.goto('/login')
    await page.getByLabel('Email or Username').fill(user.username)
    await page.getByLabel('Password').fill('wrong-password')
    await page.getByRole('button', { name: 'Sign In' }).click()

    await expect(page.getByText(/invalid credentials/i)).toBeVisible()
    await expect(page).toHaveURL('/login')
  })
})

test.describe('Forgot password', () => {
  test('submits silently and shows confirmation', async ({ page, createUser }) => {
    const user = await createUser()

    await page.goto('/login')
    await page.getByRole('link', { name: /forgot your password/i }).click()
    await expect(page).toHaveURL('/forgot-password')

    await page.getByLabel('Email').fill(user.email)
    await page.getByRole('button', { name: /send reset link/i }).click()

    await expect(page.getByText(/if that email is registered/i)).toBeVisible()
  })

  test('confirmation also shows for unknown email (anti-enumeration)', async ({ page }) => {
    await page.goto('/forgot-password')
    await page.getByLabel('Email').fill(`ghost-${randomUUID()}@test.alaya-archive.com`)
    await page.getByRole('button', { name: /send reset link/i }).click()

    await expect(page.getByText(/if that email is registered/i)).toBeVisible()
  })
})

test.describe('Reset password', () => {
  test('missing token shows error state', async ({ page }) => {
    await page.goto('/reset-password')
    await expect(page.getByRole('heading', { name: /missing reset token/i })).toBeVisible()
  })

  test('valid token resets password end-to-end', async ({ page, createUser, request }) => {
    const user = await createUser()

    // Look up the user id from the public profile endpoint via login tokens,
    // then mint a password-reset token ourselves.
    const me = await request.get(`${API_URL}/api/v1/users/me`, {
      headers: { Authorization: `Bearer ${user.tokens.access_token}` },
    })
    const meBody = await me.json()
    const token = mintToken(meBody.id, 'password_reset')

    await page.goto(`/reset-password?token=${encodeURIComponent(token)}`)

    const newPassword = 'BrandNewPassword9!'
    await page.getByLabel('New password', { exact: true }).fill(newPassword)
    await page.getByLabel('Confirm new password').fill(newPassword)
    await page.getByRole('button', { name: /reset password/i }).click()

    await expect(page.getByRole('heading', { name: /password reset/i })).toBeVisible()

    // New password works on login.
    await page.goto('/login')
    await page.getByLabel('Email or Username').fill(user.username)
    await page.getByLabel('Password').fill(newPassword)
    await page.getByRole('button', { name: 'Sign In' }).click()
    await expect(page).toHaveURL('/dashboard')

    // Old password no longer works.
    await page.evaluate(() => {
      localStorage.removeItem('access_token')
      localStorage.removeItem('refresh_token')
    })
    await page.goto('/login')
    await page.getByLabel('Email or Username').fill(user.username)
    await page.getByLabel('Password').fill(user.password)
    await page.getByRole('button', { name: 'Sign In' }).click()
    await expect(page.getByText(/invalid credentials/i)).toBeVisible()

    // Update the fixture-tracked password so cleanup uses a working login if needed.
    user.password = newPassword
  })
})

test.describe('Verify email', () => {
  test('missing token shows error state', async ({ page }) => {
    await page.goto('/verify-email')
    await expect(page.getByRole('heading', { name: /missing verification token/i })).toBeVisible()
  })

  test('valid token verifies the account', async ({ page, createUser, request }) => {
    const user = await createUser()

    const me = await request.get(`${API_URL}/api/v1/users/me`, {
      headers: { Authorization: `Bearer ${user.tokens.access_token}` },
    })
    const meBody = await me.json()
    expect(meBody.email_verified).toBe(false)

    const token = mintToken(meBody.id, 'email_verification', 24 * 3600)
    await page.goto(`/verify-email?token=${encodeURIComponent(token)}`)
    await expect(page.getByRole('heading', { name: /email verified/i })).toBeVisible()

    // Confirm server-side state changed.
    const meAfter = await request.get(`${API_URL}/api/v1/users/me`, {
      headers: { Authorization: `Bearer ${user.tokens.access_token}` },
    })
    const meAfterBody = await meAfter.json()
    expect(meAfterBody.email_verified).toBe(true)
  })
})

test.describe('Delete account', () => {
  test('from profile settings, deletes and logs out', async ({ page, createUser }) => {
    const user = await createUser()

    // Sign in via API-minted tokens, then visit profile.
    await page.goto('/login')
    await page.getByLabel('Email or Username').fill(user.username)
    await page.getByLabel('Password').fill(user.password)
    await page.getByRole('button', { name: 'Sign In' }).click()
    await expect(page).toHaveURL('/dashboard')

    await page.goto('/profile')

    // Accept the window.confirm dialog automatically.
    page.once('dialog', (dialog) => dialog.accept())

    await page.getByPlaceholder('DELETE').fill('DELETE')
    await page.getByRole('button', { name: /delete my account/i }).click()

    await expect(page).toHaveURL('/')

    // Logging in with the deleted credentials should fail.
    await page.goto('/login')
    await page.getByLabel('Email or Username').fill(user.username)
    await page.getByLabel('Password').fill(user.password)
    await page.getByRole('button', { name: 'Sign In' }).click()
    await expect(page.getByText(/invalid credentials/i)).toBeVisible()

    // Fixture teardown would try to delete again — mark tokens as invalid so
    // the cleanup call is a no-op (401 is handled silently in the fixture).
    user.tokens.access_token = 'already-deleted'
  })
})
