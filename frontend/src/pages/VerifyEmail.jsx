import { useEffect, useRef, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { verifyEmail } from '../api/auth'

export default function VerifyEmail() {
  const [params] = useSearchParams()
  const token = params.get('token')
  const [status, setStatus] = useState(token ? 'verifying' : 'missing')
  const [message, setMessage] = useState('')
  const ran = useRef(false)

  useEffect(() => {
    if (!token || ran.current) return
    ran.current = true
    verifyEmail(token)
      .then(() => setStatus('success'))
      .catch((err) => {
        setStatus('error')
        setMessage(err.response?.data?.error || 'Verification failed. The link may have expired.')
      })
  }, [token])

  return (
    <>
      <title>Verify Email — Alaya Archive</title>
      <meta name="robots" content="noindex,follow" />
      <div className="min-h-screen bg-slate-900 flex items-center justify-center px-4">
        <div className="w-full max-w-md bg-slate-800 rounded-xl p-8 shadow-xl border border-slate-700 text-center">
          {status === 'verifying' && (
            <>
              <h1 className="text-2xl font-bold text-white mb-2">Verifying your email…</h1>
              <p className="text-slate-400">Hang tight.</p>
            </>
          )}
          {status === 'success' && (
            <>
              <h1 className="text-2xl font-bold text-white mb-2">Email verified</h1>
              <p className="text-slate-300 mb-6">You're all set. Welcome to Alaya Archive.</p>
              <Link
                to="/dashboard"
                className="inline-block px-6 py-2 bg-indigo-600 hover:bg-indigo-500 text-white font-semibold rounded-lg"
              >
                Go to dashboard
              </Link>
            </>
          )}
          {status === 'error' && (
            <>
              <h1 className="text-2xl font-bold text-white mb-2">Verification failed</h1>
              <p className="text-slate-300 mb-6">{message}</p>
              <Link
                to="/login"
                className="inline-block px-6 py-2 bg-indigo-600 hover:bg-indigo-500 text-white font-semibold rounded-lg"
              >
                Back to sign in
              </Link>
            </>
          )}
          {status === 'missing' && (
            <>
              <h1 className="text-2xl font-bold text-white mb-2">Missing verification token</h1>
              <p className="text-slate-300 mb-6">This page needs a valid link from your email.</p>
              <Link
                to="/login"
                className="inline-block px-6 py-2 bg-indigo-600 hover:bg-indigo-500 text-white font-semibold rounded-lg"
              >
                Back to sign in
              </Link>
            </>
          )}
        </div>
      </div>
    </>
  )
}
