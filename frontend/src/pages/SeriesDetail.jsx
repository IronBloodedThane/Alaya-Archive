import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Link, useParams, useSearchParams } from 'react-router-dom'
import { listMedia } from '../api/media'

// SeriesDetail lists every volume the user has in a single series.
// The series name comes from the URL path; creator + media_type are
// optional filters (since two different "Berserk"s could exist for
// different creators or media types — manga vs anime).
//
// We ask the backend for all rows matching the series name (cheap, since
// users typically have <100 volumes in a series), then narrow further on
// the client by creator + media_type if those query params are present.
export default function SeriesDetail() {
  const { name } = useParams()
  const [searchParams] = useSearchParams()
  const seriesName = decodeURIComponent(name || '')
  const creatorFilter = searchParams.get('creator') || ''
  const typeFilter = searchParams.get('type') || ''

  const { data, isLoading } = useQuery({
    queryKey: ['series', seriesName],
    queryFn: () => listMedia({ series: seriesName, limit: 200 }),
    enabled: seriesName.length > 0,
  })

  const items = useMemo(() => {
    const rows = data?.items || []
    const filtered = rows.filter(
      (r) =>
        (!creatorFilter || (r.creator || '') === creatorFilter) &&
        (!typeFilter || r.media_type === typeFilter),
    )
    // Earliest volume first; nulls sort to the end.
    return [...filtered].sort((a, b) => {
      const ap = a.series_position ?? Infinity
      const bp = b.series_position ?? Infinity
      return ap - bp
    })
  }, [data?.items, creatorFilter, typeFilter])

  if (isLoading) return <div className="text-slate-400">Loading...</div>

  if (!items.length) {
    return (
      <div>
        <Link to="/collection" className="text-sm text-indigo-400 hover:text-indigo-300">
          ← Back to collection
        </Link>
        <h1 className="text-2xl font-bold dark:text-white mt-4 mb-2">{seriesName}</h1>
        <p className="text-slate-400">
          No volumes match this series in your collection.
        </p>
      </div>
    )
  }

  // The header creator/cover come from the first volume — that's also
  // the visual you saw on the Collection grid card, so the transition
  // feels continuous.
  const firstWithCover = items.find((i) => i.cover_image)
  const ownedCount = items.filter((i) => i.list_type !== 'wishlist').length
  const wishlistCount = items.length - ownedCount

  return (
    <div>
      <Link to="/collection" className="text-sm text-indigo-400 hover:text-indigo-300">
        ← Back to collection
      </Link>

      <div className="flex items-start gap-6 mt-4 mb-6">
        {firstWithCover?.cover_image && (
          <img
            src={firstWithCover.cover_image}
            alt={seriesName}
            className="w-32 rounded-xl shadow-lg object-cover"
          />
        )}
        <div className="flex-1">
          <p className="text-sm text-indigo-400 uppercase font-medium mb-1">
            Series · {items[0].media_type.replace('_', ' ')}
          </p>
          <h1 className="text-3xl font-bold dark:text-white mb-1">{seriesName}</h1>
          {items[0].creator && (
            <p className="text-slate-400 mb-3">by {items[0].creator}</p>
          )}
          <div className="flex flex-wrap gap-2">
            <span className="px-3 py-1 bg-indigo-500/20 text-indigo-400 rounded-full text-sm">
              {items.length} {items.length === 1 ? 'volume' : 'volumes'}
            </span>
            {ownedCount > 0 && (
              <span className="px-3 py-1 bg-green-500/20 text-green-400 rounded-full text-sm">
                {ownedCount} owned
              </span>
            )}
            {wishlistCount > 0 && (
              <span className="px-3 py-1 bg-amber-500/20 text-amber-400 rounded-full text-sm">
                {wishlistCount} wishlist
              </span>
            )}
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
        {items.map((item) => (
          <VolumeCard key={item.id} item={item} />
        ))}
      </div>
    </div>
  )
}

function VolumeCard({ item }) {
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
      {item.series_position != null && (
        <span className="absolute top-2 left-2 z-10 px-2 py-0.5 bg-slate-900/80 text-white text-xs font-semibold rounded-full">
          Vol. {item.series_position}
        </span>
      )}
      {item.cover_image && (
        <img src={item.cover_image} alt={item.title} className="w-full h-48 object-cover" />
      )}
      <div className="p-4">
        <h3 className="font-semibold dark:text-white truncate">{item.title}</h3>
        <div className="flex items-center justify-between mt-2">
          <span className="text-xs px-2 py-0.5 rounded-full bg-slate-500/20 text-slate-400">
            {item.status.replace('_', ' ')}
          </span>
          {item.rating && (
            <span className="text-sm text-yellow-500">{item.rating}/10</span>
          )}
        </div>
      </div>
    </Link>
  )
}
