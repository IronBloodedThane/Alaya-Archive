import { Link } from 'react-router-dom'

export default function NotFound() {
  return (
    <>
      <title>Page Not Found — Alaya Archive</title>
      <meta name="robots" content="noindex,follow" />
      <div className="min-h-screen bg-slate-900 flex items-center justify-center">
        <div className="text-center">
          <h1 className="text-6xl font-bold text-indigo-400 mb-4">404</h1>
          <p className="text-xl text-slate-300 mb-6">Page not found</p>
          <Link to="/" className="px-6 py-2 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg">
            Go Home
          </Link>
        </div>
      </div>
    </>
  )
}
