import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { getStats } from '../api/media'
import { useAuth } from '../hooks/useAuth'

export default function Dashboard() {
  const { user } = useAuth()
  const { data: stats, isLoading } = useQuery({ queryKey: ['media-stats'], queryFn: getStats })

  const mediaTypes = [
    { key: 'manga', label: 'Manga', color: 'bg-pink-500' },
    { key: 'anime', label: 'Anime', color: 'bg-purple-500' },
    { key: 'movie', label: 'Movies', color: 'bg-blue-500' },
    { key: 'book', label: 'Books', color: 'bg-green-500' },
    { key: 'game', label: 'Games', color: 'bg-orange-500' },
    { key: 'tv_show', label: 'TV Shows', color: 'bg-cyan-500' },
    { key: 'music', label: 'Music', color: 'bg-yellow-500' },
    { key: 'other', label: 'Other', color: 'bg-slate-500' },
  ]

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold dark:text-white">
          Welcome back, {user?.display_name || user?.username}
        </h1>
        <Link
          to="/collection/add"
          className="px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-white font-medium rounded-lg transition-colors"
        >
          + Add Media
        </Link>
      </div>

      {isLoading ? (
        <div className="text-slate-400">Loading stats...</div>
      ) : (
        <>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
            <div className="bg-white dark:bg-slate-800 rounded-xl p-4 border border-slate-200 dark:border-slate-700">
              <p className="text-sm text-slate-500 dark:text-slate-400">Total Items</p>
              <p className="text-3xl font-bold text-indigo-500">{stats?.total || 0}</p>
            </div>
            {mediaTypes.slice(0, 3).map((type) => (
              <div key={type.key} className="bg-white dark:bg-slate-800 rounded-xl p-4 border border-slate-200 dark:border-slate-700">
                <p className="text-sm text-slate-500 dark:text-slate-400">{type.label}</p>
                <p className="text-3xl font-bold dark:text-white">{stats?.by_type?.[type.key] || 0}</p>
              </div>
            ))}
          </div>

          <h2 className="text-lg font-semibold dark:text-white mb-4">Collection by Type</h2>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-3 mb-8">
            {mediaTypes.map((type) => (
              <Link
                key={type.key}
                to={`/collection?type=${type.key}`}
                className="bg-white dark:bg-slate-800 rounded-xl p-4 border border-slate-200 dark:border-slate-700 hover:border-indigo-500 transition-colors"
              >
                <div className={`w-3 h-3 rounded-full ${type.color} mb-2`} />
                <p className="font-medium dark:text-white">{type.label}</p>
                <p className="text-sm text-slate-500">{stats?.by_type?.[type.key] || 0} items</p>
              </Link>
            ))}
          </div>

          <h2 className="text-lg font-semibold dark:text-white mb-4">By Status</h2>
          <div className="grid grid-cols-2 md:grid-cols-5 gap-3">
            {['planned', 'in_progress', 'completed', 'on_hold', 'dropped'].map((status) => (
              <Link
                key={status}
                to={`/collection?status=${status}`}
                className="bg-white dark:bg-slate-800 rounded-xl p-3 border border-slate-200 dark:border-slate-700 hover:border-indigo-500 transition-colors text-center"
              >
                <p className="font-medium dark:text-white capitalize">{status.replace('_', ' ')}</p>
                <p className="text-sm text-slate-500">{stats?.by_status?.[status] || 0}</p>
              </Link>
            ))}
          </div>
        </>
      )}
    </div>
  )
}
