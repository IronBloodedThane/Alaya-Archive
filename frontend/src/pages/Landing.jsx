import { Link } from 'react-router-dom'

export default function Landing() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-indigo-950 to-slate-900 flex flex-col items-center justify-center px-4 py-16">
      <div className="text-center max-w-2xl">
        <h1 className="text-5xl md:text-7xl font-bold text-white mb-4">
          Alaya <span className="text-indigo-400">Archive</span>
        </h1>
        <p className="text-xl text-slate-300 mb-8">
          Alaya Archive is a catalog for your manga, movies, anime, books, and games. Track
          what you own, what you've watched or read, rate and tag your collection, and
          share it with friends.
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

        <section aria-labelledby="features-heading" className="mt-16">
          <h2 id="features-heading" className="sr-only">Features</h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 text-left">
            <div className="bg-slate-800/50 rounded-xl p-6 border border-slate-700">
              <h3 className="text-lg font-semibold text-indigo-400 mb-2">Track Everything</h3>
              <p className="text-slate-400 text-sm">Manga volumes, movies, anime series, books, games — all in one place.</p>
            </div>
            <div className="bg-slate-800/50 rounded-xl p-6 border border-slate-700">
              <h3 className="text-lg font-semibold text-indigo-400 mb-2">Rate &amp; Tag</h3>
              <p className="text-slate-400 text-sm">Rate your collection, add custom tags, and track your progress.</p>
            </div>
            <div className="bg-slate-800/50 rounded-xl p-6 border border-slate-700">
              <h3 className="text-lg font-semibold text-indigo-400 mb-2">Share with Friends</h3>
              <p className="text-slate-400 text-sm">Connect with friends, see their collections, and discover new media.</p>
            </div>
          </div>
        </section>
      </div>

      <footer className="mt-16 pt-8 border-t border-slate-800 w-full max-w-2xl">
        <nav aria-label="Footer" className="flex gap-6 justify-center flex-wrap text-sm text-slate-400">
          <Link to="/about" className="hover:text-slate-200">About</Link>
          <Link to="/register" className="hover:text-slate-200">Create account</Link>
          <Link to="/login" className="hover:text-slate-200">Sign in</Link>
          <a href="/sitemap.xml" className="hover:text-slate-200">Sitemap</a>
        </nav>
      </footer>
    </div>
  )
}
