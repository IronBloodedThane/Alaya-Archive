import { useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { getPublicCollection } from '../api/media'
import Avatar from '../components/Avatar'
import { useAuth } from '../hooks/useAuth'

const MEDIA_TYPES = ['all', 'manga', 'anime', 'movie', 'book', 'game', 'tv_show', 'music', 'other']

const statusColor = (s) => {
  const colors = {
    planned: 'bg-blue-500/20 text-blue-400',
    in_progress: 'bg-yellow-500/20 text-yellow-400',
    completed: 'bg-green-500/20 text-green-400',
    on_hold: 'bg-orange-500/20 text-orange-400',
    dropped: 'bg-red-500/20 text-red-400',
  }
  return colors[s] || 'bg-slate-500/20 text-slate-400'
}

export default function PublicProfile() {
  const { username } = useParams()
  const { user } = useAuth()
  const [type, setType] = useState('all')
  const [copied, setCopied] = useState(false)

  const copyLink = async () => {
    try {
      await navigator.clipboard.writeText(window.location.href)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      // Clipboard not available
    }
  }

  const params = { limit: 100, ...(type !== 'all' && { type }) }
  const { data, isLoading, isError, error } = useQuery({
    queryKey: ['public-collection', username, params],
    queryFn: () => getPublicCollection(username, params),
    retry: false,
  })

  const notFound = isError && error?.response?.status === 404

  return (
    <div className="min-h-screen bg-slate-50 dark:bg-slate-900">
      <header className="bg-white dark:bg-slate-800 shadow-lg">
        <div className="max-w-5xl mx-auto px-4 py-3 flex items-center justify-between">
          <Link to="/" className="text-xl font-bold text-indigo-600 dark:text-indigo-400">
            Alaya Archive
          </Link>
          <div className="flex items-center gap-3 text-sm">
            {user ? (
              <Link to="/dashboard" className="text-slate-600 hover:text-slate-900 dark:text-slate-300 dark:hover:text-white">
                Back to app
              </Link>
            ) : (
              <>
                <Link to="/login" className="text-slate-600 hover:text-slate-900 dark:text-slate-300 dark:hover:text-white">
                  Log in
                </Link>
                <Link
                  to="/register"
                  className="px-3 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg"
                >
                  Sign up
                </Link>
              </>
            )}
          </div>
        </div>
      </header>

      <main className="max-w-5xl mx-auto px-4 py-6">
        {notFound ? (
          <div className="text-center py-20">
            <h1 className="text-2xl font-bold dark:text-white mb-2">@{username}</h1>
            <p className="text-slate-500">This profile doesn't exist or isn't public.</p>
          </div>
        ) : isLoading ? (
          <div className="text-slate-400">Loading...</div>
        ) : isError ? (
          <div className="text-slate-400">Couldn't load profile.</div>
        ) : (
          <>
            <div className="bg-white dark:bg-slate-800 rounded-xl p-6 border border-slate-200 dark:border-slate-700 mb-6">
              <div className="flex items-start gap-4">
                <Avatar username={data.user.username} displayName={data.user.display_name} size="xl" />
                <div className="flex-1 min-w-0">
                  <h1 className="text-2xl font-bold dark:text-white">
                    {data.user.display_name || data.user.username}
                  </h1>
                  <p className="text-slate-500">@{data.user.username}</p>
                  {data.user.bio && (
                    <p className="text-slate-600 dark:text-slate-300 mt-2 whitespace-pre-wrap">{data.user.bio}</p>
                  )}
                </div>
                <button
                  type="button"
                  onClick={copyLink}
                  className="px-3 py-2 bg-slate-200 dark:bg-slate-700 hover:bg-slate-300 dark:hover:bg-slate-600 text-slate-700 dark:text-slate-200 text-sm font-medium rounded-lg shrink-0"
                >
                  {copied ? 'Copied!' : 'Copy link'}
                </button>
              </div>
            </div>

            <div className="flex items-center justify-between mb-4 flex-wrap gap-3">
              <h2 className="text-lg font-semibold dark:text-white">
                Collection ({data.total})
              </h2>
              <select
                value={type}
                onChange={(e) => setType(e.target.value)}
                className="px-3 py-2 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg dark:text-white text-sm"
              >
                {MEDIA_TYPES.map((t) => (
                  <option key={t} value={t}>{t === 'all' ? 'All Types' : t.replace('_', ' ')}</option>
                ))}
              </select>
            </div>

            {data.items.length === 0 ? (
              <div className="text-center py-12">
                <p className="text-slate-500">Nothing public to show yet.</p>
              </div>
            ) : (
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
                {data.items.map((item) => (
                  <div
                    key={item.id}
                    className="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 overflow-hidden"
                  >
                    {item.cover_image && (
                      <img src={item.cover_image} alt={item.title} className="w-full h-48 object-cover" />
                    )}
                    <div className="p-4">
                      <p className="text-xs text-indigo-400 uppercase font-medium mb-1">
                        {item.media_type.replace('_', ' ')}
                      </p>
                      <h3 className="font-semibold dark:text-white truncate">{item.title}</h3>
                      {item.creator && <p className="text-sm text-slate-500 truncate">{item.creator}</p>}
                      <div className="flex items-center justify-between mt-2">
                        <span className={`text-xs px-2 py-0.5 rounded-full ${statusColor(item.status)}`}>
                          {item.status.replace('_', ' ')}
                        </span>
                        {item.rating && (
                          <span className="text-sm text-yellow-500">{item.rating}/10</span>
                        )}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </>
        )}
      </main>
    </div>
  )
}
