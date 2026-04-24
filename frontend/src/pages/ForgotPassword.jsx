import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useMutation } from '@tanstack/react-query'
import { forgotPassword } from '../api/auth'

export default function ForgotPassword() {
  const [email, setEmail] = useState('')
  const [submitted, setSubmitted] = useState(false)

  const mutation = useMutation({
    mutationFn: forgotPassword,
    onSuccess: () => setSubmitted(true),
  })

  const handleSubmit = (e) => {
    e.preventDefault()
    if (!email.trim()) return
    mutation.mutate(email.trim())
  }

  return (
    <>
      <title>Forgot Password — Alaya Archive</title>
      <meta name="robots" content="noindex,follow" />
      <div className="min-h-screen bg-slate-900 flex items-center justify-center px-4">
        <div className="w-full max-w-md">
          <h1 className="text-3xl font-bold text-white text-center mb-8">Forgot Password</h1>

          <div className="bg-slate-800 rounded-xl p-6 shadow-xl border border-slate-700">
            {submitted ? (
              <>
                <p className="text-slate-300 mb-6">
                  If that email is registered, a reset link is on its way. Check your inbox
                  (and spam folder, just in case). The link expires in one hour.
                </p>
                <Link
                  to="/login"
                  className="block text-center w-full py-2.5 bg-indigo-600 hover:bg-indigo-500 text-white font-semibold rounded-lg"
                >
                  Back to sign in
                </Link>
              </>
            ) : (
              <form onSubmit={handleSubmit}>
                <p className="text-slate-300 text-sm mb-4">
                  Enter the email address on your account. We'll send a link to reset your
                  password.
                </p>

                <div className="mb-6">
                  <label htmlFor="forgot-email" className="block text-sm font-medium text-slate-300 mb-1">Email</label>
                  <input
                    id="forgot-email"
                    type="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    className="w-full px-4 py-2 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:border-indigo-500"
                    required
                    autoFocus
                  />
                </div>

                <button
                  type="submit"
                  disabled={mutation.isPending}
                  className="w-full py-2.5 bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 text-white font-semibold rounded-lg transition-colors"
                >
                  {mutation.isPending ? 'Sending...' : 'Send reset link'}
                </button>

                <p className="mt-4 text-center text-sm text-slate-400">
                  Remembered it?{' '}
                  <Link to="/login" className="text-indigo-400 hover:text-indigo-300">Sign In</Link>
                </p>
              </form>
            )}
          </div>
        </div>
      </div>
    </>
  )
}
