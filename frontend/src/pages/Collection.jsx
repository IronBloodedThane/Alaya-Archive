import { useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Link, useSearchParams } from 'react-router-dom'
import { listMedia } from '../api/media'

// groupBySeries collapses items sharing a (series, creator, media_type)
// into a single "series entry" so the Collection grid shows one card per
// series instead of N cards per volume. Items without a series stay
// standalone. Order is preserved: the first time a series is seen
// (server returns updated_at DESC), that's its slot in the output.
function groupBySeries(items) {
  if (!items) return []
  const out = []
  const groups = new Map() // key → index in out

  for (const item of items) {
    const series = (item.series || '').trim()
    if (!series) {
      out.push({ kind: 'item', item })
      continue
    }
    const key = `${series}|${item.creator || ''}|${item.media_type}`
    if (!groups.has(key)) {
      groups.set(key, out.length)
      out.push({
        kind: 'series',
        series,
        creator: item.creator || '',
        mediaType: item.media_type,
        items: [item],
      })
    } else {
      out[groups.get(key)].items.push(item)
    }
  }
  // Sort each group's items by series_position (nulls last) so the
  // displayed cover comes from the earliest volume.
  for (const entry of out) {
    if (entry.kind === 'series') {
      entry.items.sort((a, b) => {
        const ap = a.series_position ?? Infinity
        const bp = b.series_position ?? Infinity
        return ap - bp
      })
    }
  }
  return out
}

function seriesPath(entry) {
  // Series name goes in the path (encoded); creator/type in query params
  // keep the URL readable when you share it.
  const params = new URLSearchParams()
  if (entry.creator) params.set('creator', entry.creator)
  if (entry.mediaType) params.set('type', entry.mediaType)
  const qs = params.toString()
  return `/series/${encodeURIComponent(entry.series)}${qs ? `?${qs}` : ''}`
}

const MEDIA_TYPES = ['all', 'manga', 'anime', 'movie', 'book', 'game', 'tv_show', 'music', 'other']
const STATUSES = ['all', 'planned', 'in_progress', 'completed', 'on_hold', 'dropped']
const LIST_TYPES = [
  { value: 'all', label: 'All Lists' },
  { value: 'owned', label: 'Collection' },
  { value: 'wishlist', label: 'Wishlist' },
]

export default function Collection() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [search, setSearch] = useState('')

  const type = searchParams.get('type') || 'all'
  const status = searchParams.get('status') || 'all'
  const listType = searchParams.get('list_type') || 'all'

  const params = {
    ...(type !== 'all' && { type }),
    ...(status !== 'all' && { status }),
    ...(listType !== 'all' && { list_type: listType }),
    ...(search && { search }),
    limit: 50,
    offset: 0,
  }

  const { data, isLoading } = useQuery({
    queryKey: ['media', params],
    queryFn: () => listMedia(params),
  })

  const entries = useMemo(() => groupBySeries(data?.items), [data?.items])

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
        <select
          value={listType}
          onChange={(e) => setFilter('list_type', e.target.value)}
          className="px-3 py-2 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg dark:text-white"
        >
          {LIST_TYPES.map((l) => (
            <option key={l.value} value={l.value}>{l.label}</option>
          ))}
        </select>
      </div>

      {isLoading ? (
        <div className="text-slate-400">Loading...</div>
      ) : (
        <>
          <p className="text-sm text-slate-500 mb-4">
            {data?.total || 0} item{data?.total === 1 ? '' : 's'}
            {entries.length !== (data?.total || 0) &&
              ` · ${entries.length} card${entries.length === 1 ? '' : 's'}`}
          </p>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
            {entries.map((entry) =>
              entry.kind === 'series' ? (
                <SeriesCard key={`s:${entry.series}|${entry.creator}|${entry.mediaType}`} entry={entry} />
              ) : (
                <ItemCard key={entry.item.id} item={entry.item} statusColor={statusColor} />
              ),
            )}
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

function ItemCard({ item, statusColor }) {
  return (
    <Link
      to={`/collection/${item.id}`}
      className="relative bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 hover:border-indigo-500 transition-colors overflow-hidden"
    >
      {item.list_type === 'wishlist' && (
        <span className="absolute top-2 right-2 z-10 px-2 py-0.5 bg-amber-500/90 text-white text-xs font-semibold rounded-full">
          Wishlist
        </span>
      )}
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
          {item.rating && <span className="text-sm text-yellow-500">{item.rating}/10</span>}
        </div>
      </div>
    </Link>
  )
}

function SeriesCard({ entry }) {
  // Use the first volume's cover so the visual is recognizable; fall back
  // to any volume that has one if vol 1 doesn't.
  const cover =
    entry.items.find((i) => i.cover_image)?.cover_image || ''
  const count = entry.items.length
  const wishlistOnly =
    count > 0 && entry.items.every((i) => i.list_type === 'wishlist')
  return (
    <Link
      to={seriesPath(entry)}
      className="relative bg-white dark:bg-slate-800 rounded-xl border-2 border-indigo-500/30 hover:border-indigo-500 transition-colors overflow-hidden"
    >
      <span className="absolute top-2 left-2 z-10 px-2 py-0.5 bg-indigo-600 text-white text-xs font-semibold rounded-full">
        Series · {count}
      </span>
      {wishlistOnly && (
        <span className="absolute top-2 right-2 z-10 px-2 py-0.5 bg-amber-500/90 text-white text-xs font-semibold rounded-full">
          Wishlist
        </span>
      )}
      {cover && <img src={cover} alt={entry.series} className="w-full h-48 object-cover" />}
      <div className="p-4">
        <p className="text-xs text-indigo-400 uppercase font-medium mb-1">
          {entry.mediaType.replace('_', ' ')}
        </p>
        <h3 className="font-semibold dark:text-white truncate">{entry.series}</h3>
        {entry.creator && (
          <p className="text-sm text-slate-500 truncate">{entry.creator}</p>
        )}
        <p className="text-xs text-slate-500 mt-2">
          {count} {count === 1 ? 'volume' : 'volumes'} in your collection
        </p>
      </div>
    </Link>
  )
}
