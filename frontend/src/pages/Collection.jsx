import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Link, useSearchParams } from 'react-router-dom'
import { listMedia } from '../api/media'

const MEDIA_TYPES = ['all', 'manga', 'anime', 'movie', 'book', 'game', 'tv_show', 'music', 'other']
const STATUSES = ['all', 'planned', 'in_progress', 'completed', 'on_hold', 'dropped']

export default function Collection() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [search, setSearch] = useState('')

  const type = searchParams.get('type') || 'all'
  const status = searchParams.get('status') || 'all'

  const params = {
    ...(type !== 'all' && { type }),
    ...(status !== 'all' && { status }),
    ...(search && { search }),
    limit: 50,
    offset: 0,
  }

  const { data, isLoading } = useQuery({
    queryKey: ['media', params],
    queryFn: () => listMedia(params),
  })

  const setFilter = (key, value) => {
    const next = new URLSearchParams(searchParams)
    if (value === 'all') next.delete(key)
    else next.set(key, value)
    setSearchParams(next)
  }

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

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold dark:text-white">My Collection</h1>
        <Link
          to="/collection/add"
          className="px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-white font-medium rounded-lg"
        >
          + Add Media
        </Link>
      </div>

      {/* Filters */}
      <div className="flex flex-wrap gap-3 mb-6">
        <input
          type="text"
          placeholder="Search..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="px-4 py-2 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg dark:text-white focus:outline-none focus:border-indigo-500"
        />
        <select
          value={type}
          onChange={(e) => setFilter('type', e.target.value)}
          className="px-3 py-2 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg dark:text-white"
        >
          {MEDIA_TYPES.map((t) => (
            <option key={t} value={t}>{t === 'all' ? 'All Types' : t.replace('_', ' ')}</option>
          ))}
        </select>
        <select
          value={status}
          onChange={(e) => setFilter('status', e.target.value)}
          className="px-3 py-2 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg dark:text-white"
        >
          {STATUSES.map((s) => (
            <option key={s} value={s}>{s === 'all' ? 'All Statuses' : s.replace('_', ' ')}</option>
          ))}
        </select>
      </div>

      {isLoading ? (
        <div className="text-slate-400">Loading...</div>
      ) : (
        <>
          <p className="text-sm text-slate-500 mb-4">{data?.total || 0} items</p>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
            {data?.items?.map((item) => (
              <Link
                key={item.id}
                to={`/collection/${item.id}`}
                className="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 hover:border-indigo-500 transition-colors overflow-hidden"
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
              </Link>
            ))}
          </div>

          {(!data?.items || data.items.length === 0) && (
            <div className="text-center py-12">
              <p className="text-slate-400 mb-4">No items in your collection yet.</p>
              <Link
                to="/collection/add"
                className="px-6 py-2 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg"
              >
                Add your first item
              </Link>
            </div>
          )}
        </>
      )}
    </div>
  )
}
