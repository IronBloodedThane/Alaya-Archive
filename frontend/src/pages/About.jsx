import { Link } from 'react-router-dom'
import { SITE_URL, SITE_NAME } from '../seo'

export default function About() {
  const title = `About — ${SITE_NAME}`
  const description = `${SITE_NAME} is a privacy-focused community for cataloging and sharing manga and media collections, built by James "Thane" DeWees.`
  const url = `${SITE_URL}/about`

  return (
    <>
      <title>{title}</title>
      <meta name="description" content={description} />
      <link rel="canonical" href={url} />
      <meta property="og:title" content={title} />
      <meta property="og:description" content={description} />
      <meta property="og:url" content={url} />

      <div className="min-h-screen bg-gradient-to-br from-slate-900 via-indigo-950 to-slate-900 text-slate-200 px-4 py-16">
        <main className="max-w-2xl mx-auto">
          <nav aria-label="Breadcrumb" className="text-sm text-slate-400 mb-8">
            <Link to="/" className="hover:text-slate-200">Home</Link>
            <span aria-hidden="true"> / </span>
            <span className="text-slate-300">About</span>
          </nav>

          <h1 className="text-4xl md:text-5xl font-bold text-white mb-6">
            About Alaya Archive
          </h1>

          <p className="text-lg text-slate-300 mb-8">
            Alaya Archive is a privacy-focused community for cataloging, rating, and
            sharing your manga and wider media collection.
          </p>

          <section className="mb-10">
            <h2 className="text-2xl font-semibold text-indigo-400 mb-3">Who built this</h2>
            <p className="text-slate-300">
              The site was built by James "Thane" DeWees as a way to manage his own
              growing manga collection — and to give other collectors a place to do the
              same without handing their data to an ad-driven platform.
            </p>
          </section>

          <section className="mb-10">
            <h2 className="text-2xl font-semibold text-indigo-400 mb-3">Why privacy</h2>
            <p className="text-slate-300">
              Your collection is yours. Alaya Archive is designed so that what you track
              stays private by default. Only what you explicitly choose to make public —
              your profile, specific collection entries, or shared lists — is visible to
              anyone else.
            </p>
          </section>

          <section className="mb-10">
            <h2 className="text-2xl font-semibold text-indigo-400 mb-3">Leaving any time</h2>
            <p className="text-slate-300 mb-3">
              We don't keep your data around once you're done with it. You can delete your
              account from your profile settings whenever you want, and everything tied to
              it is wiped: your collection, tags, ratings, notes, friends, followers,
              activity, and profile itself.
            </p>
            <p className="text-slate-300">
              No waiting period, no "just in case" archive, no retention for us to mine
              later. When you say delete, it's gone.
            </p>
          </section>

          <section className="mb-10">
            <h2 className="text-2xl font-semibold text-indigo-400 mb-3">What it supports</h2>
            <p className="text-slate-300">
              While manga is the heart of the project, Alaya Archive also tracks movies,
              anime, books, games, TV shows, and music — the whole collection in one
              place, with ratings, tags, progress, and notes on every entry.
            </p>
          </section>

          <section className="mb-12">
            <h2 className="text-2xl font-semibold text-indigo-400 mb-3">Get started</h2>
            <p className="text-slate-300 mb-4">
              New accounts are free. Privacy defaults are on from the moment you sign up.
            </p>
            <div className="flex gap-3 flex-wrap">
              <Link
                to="/register"
                className="px-6 py-2.5 bg-indigo-600 hover:bg-indigo-500 text-white font-semibold rounded-xl transition-colors"
              >
                Create account
              </Link>
              <Link
                to="/login"
                className="px-6 py-2.5 border border-slate-600 hover:border-slate-400 text-slate-300 font-semibold rounded-xl transition-colors"
              >
                Sign in
              </Link>
            </div>
          </section>

          <footer className="pt-8 border-t border-slate-800">
            <nav aria-label="Footer" className="flex gap-6 justify-center flex-wrap text-sm text-slate-400">
              <Link to="/" className="hover:text-slate-200">Home</Link>
              <Link to="/register" className="hover:text-slate-200">Create account</Link>
              <Link to="/login" className="hover:text-slate-200">Sign in</Link>
              <a href="/sitemap.xml" className="hover:text-slate-200">Sitemap</a>
            </nav>
          </footer>
        </main>
      </div>
    </>
  )
}
