import { useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { useMutation } from '@tanstack/react-query'
import { resetPassword } from '../api/auth'

export default function ResetPassword() {
  const [params] = useSearchParams()
  const token = params.get('token')

  const [password, setPassword] = useState('')
  const [confirm, setConfirm] = useState('')
  const [error, setError] = useState('')

  const mutation = useMutation({
    mutationFn: ({ token, newPassword }) => resetPassword(token, newPassword),
    onError: (err) => {
      setError(err?.response?.data?.error || 'Reset failed. The link may have expired.')
    },
  })

  const handleSubmit = (e) => {
    e.preventDefault()
    setError('')
    if (password.length < 8) {
      setError('Password must be at least 8 characters')
      return
    }
    if (password !== confirm) {
      setError('Passwords do not match')
      return
    }
    mutation.mutate({ token, newPassword: password })
  }

  if (!token) {
    return (
      <>
        <title>Reset Password — Alaya Archive</title>
        <meta name="robots" content="noindex,follow" />
        <div className="min-h-screen bg-slate-900 flex items-center justify-center px-4">
          <div className="w-full max-w-md bg-slate-800 rounded-xl p-8 shadow-xl border border-slate-700 text-center">
            <h1 className="text-2xl font-bold text-white mb-2">Missing reset token</h1>
            <p className="text-slate-300 mb-6">This page needs a valid link from your email.</p>
            <Link
              to="/forgot-password"
              className="inline-block px-6 py-2 bg-indigo-600 hover:bg-indigo-500 text-white font-semibold rounded-lg"
            >
              Request a new link
            </Link>
          </div>
        </div>
      </>
    )
  }

  return (
    <>
      <title>Reset Password — Alaya Archive</title>
      <meta name="robots" content="noindex,follow" />
      <div className="min-h-screen bg-slate-900 flex items-center justify-center px-4">
        <div className="w-full max-w-md">
          <h1 className="text-3xl font-bold text-white text-center mb-8">Reset Password</h1>

          <div className="bg-slate-800 rounded-xl p-6 shadow-xl border border-slate-700">
            {mutation.isSuccess ? (
              <>
                <h2 className="text-xl font-semibold text-white mb-2">Password reset</h2>
                <p className="text-slate-300 mb-6">
                  Your password has been updated. You can sign in with your new password now.
                </p>
                <Link
                  to="/login"
                  className="block text-center w-full py-2.5 bg-indigo-600 hover:bg-indigo-500 text-white font-semibold rounded-lg"
                >
                  Sign in
                </Link>
              </>
            ) : (
              <form onSubmit={handleSubmit}>
                {error && (
                  <div className="mb-4 p-3 bg-red-500/10 border border-red-500/30 rounded-lg text-red-400 text-sm">
                    {error}
                  </div>
                )}

                <div className="mb-4">
                  <label htmlFor="reset-new-password" className="block text-sm font-medium text-slate-300 mb-1">New password</label>
                  <input
                    id="reset-new-password"
                    type="password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    className="w-full px-4 py-2 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:border-indigo-500"
                    required
                    minLength={8}
                    autoFocus
                  />
                  <p className="mt-1 text-xs text-slate-500">At least 8 characters</p>
                </div>

                <div className="mb-6">
                  <label htmlFor="reset-confirm-password" className="block text-sm font-medium text-slate-300 mb-1">Confirm new password</label>
                  <input
                    id="reset-confirm-password"
                    type="password"
                    value={confirm}
                    onChange={(e) => setConfirm(e.target.value)}
                    className="w-full px-4 py-2 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:border-indigo-500"
                    required
                    minLength={8}
                  />
                </div>

                <button
                  type="submit"
                  disabled={mutation.isPending}
                  className="w-full py-2.5 bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 text-white font-semibold rounded-lg transition-colors"
                >
                  {mutation.isPending ? 'Resetting...' : 'Reset password'}
                </button>
              </form>
            )}
          </div>
        </div>
      </div>
    </>
  )
}
