import { Link } from 'react-router-dom'
import { SITE_URL, SITE_NAME, DEFAULT_DESCRIPTION } from '../seo'

export default function Landing() {
  const title = `${SITE_NAME} — Catalog Your Manga, Movies & Media Collection`
  return (
    <>
      <title>{title}</title>
      <meta name="description" content={DEFAULT_DESCRIPTION} />
      <link rel="canonical" href={`${SITE_URL}/`} />
      <meta property="og:title" content={title} />
      <meta property="og:description" content={DEFAULT_DESCRIPTION} />
      <meta property="og:url" content={`${SITE_URL}/`} />
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-indigo-950 to-slate-900 flex items-center justify-center px-4">
      <div className="text-center max-w-2xl">
        <h1 className="text-5xl md:text-7xl font-bold text-white mb-4">
          Alaya <span className="text-indigo-400">Archive</span>
        </h1>
        <p className="text-xl text-slate-300 mb-8">
          Catalog your manga, movies, anime, and more. Track what you own, what you've watched,
          and share your collection with friends.
        </p>
        <div className="flex gap-4 justify-center">
          <Link
            to="/register"
            className="px-8 py-3 bg-indigo-600 hover:bg-indigo-500 text-white font-semibold rounded-xl transition-colors text-lg"
          >
            Get Started
          </Link>
          <Link
            to="/login"
            className="px-8 py-3 border border-slate-600 hover:border-slate-400 text-slate-300 font-semibold rounded-xl transition-colors text-lg"
          >
            Sign In
          </Link>
        </div>

        <div className="mt-16 grid grid-cols-1 md:grid-cols-3 gap-6 text-left">
          <div className="bg-slate-800/50 rounded-xl p-6 border border-slate-700">
            <h3 className="text-lg font-semibold text-indigo-400 mb-2">Track Everything</h3>
            <p className="text-slate-400 text-sm">Manga volumes, movies, anime series, books, games - all in one place.</p>
          </div>
          <div className="bg-slate-800/50 rounded-xl p-6 border border-slate-700">
            <h3 className="text-lg font-semibold text-indigo-400 mb-2">Rate & Tag</h3>
            <p className="text-slate-400 text-sm">Rate your collection, add custom tags, and track your progress.</p>
          </div>
          <div className="bg-slate-800/50 rounded-xl p-6 border border-slate-700">
            <h3 className="text-lg font-semibold text-indigo-400 mb-2">Share with Friends</h3>
            <p className="text-slate-400 text-sm">Connect with friends, see their collections, and discover new media.</p>
          </div>
        </div>
      </div>
    </div>
    </>
  )
}
