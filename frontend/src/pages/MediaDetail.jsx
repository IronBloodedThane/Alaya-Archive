import { useParams, useNavigate, Link } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getMedia, deleteMedia } from '../api/media'

export default function MediaDetail() {
  const { mediaId } = useParams()
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  const { data: media, isLoading } = useQuery({
    queryKey: ['media', mediaId],
    queryFn: () => getMedia(mediaId),
  })

  const deleteMutation = useMutation({
    mutationFn: () => deleteMedia(mediaId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['media'] })
      navigate('/collection')
    },
  })

  if (isLoading) return <div className="text-slate-400">Loading...</div>
  if (!media) return <div className="text-slate-400">Not found</div>

  const handleDelete = () => {
    if (window.confirm('Delete this item from your collection?')) {
      deleteMutation.mutate()
    }
  }

  return (
    <div className="max-w-3xl mx-auto">
      <div className="flex items-start gap-6 mb-6">
        {media.cover_image && (
          <img src={media.cover_image} alt={media.title} className="w-48 rounded-xl shadow-lg object-cover" />
        )}
        <div className="flex-1">
          <p className="text-sm text-indigo-400 uppercase font-medium mb-1">
            {media.media_type.replace('_', ' ')}
          </p>
          <h1 className="text-3xl font-bold dark:text-white mb-1">{media.title}</h1>
          {media.title_original && (
            <p className="text-lg text-slate-500 mb-2">{media.title_original}</p>
          )}
          {media.creator && <p className="text-slate-400 mb-1">by {media.creator}</p>}
          {media.year_released && <p className="text-slate-500 text-sm mb-3">{media.year_released}</p>}

          <div className="flex flex-wrap gap-2 mb-4">
            <span className="px-3 py-1 bg-indigo-500/20 text-indigo-400 rounded-full text-sm">
              {media.status.replace('_', ' ')}
            </span>
            {media.list_type === 'wishlist' && (
              <span className="px-3 py-1 bg-amber-500/20 text-amber-400 rounded-full text-sm">
                Wishlist
              </span>
            )}
            {media.rating && (
              <span className="px-3 py-1 bg-yellow-500/20 text-yellow-400 rounded-full text-sm">
                {media.rating}/10
              </span>
            )}
          </div>

          <div className="flex gap-2">
            <Link
              to={`/collection/${mediaId}/edit`}
              className="px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-white text-sm rounded-lg"
            >
              Edit
            </Link>
            <button
              onClick={handleDelete}
              className="px-4 py-2 bg-red-600 hover:bg-red-500 text-white text-sm rounded-lg"
            >
              Delete
            </button>
          </div>
        </div>
      </div>

      <div className="bg-white dark:bg-slate-800 rounded-xl p-6 border border-slate-200 dark:border-slate-700">
        {media.isbn && (
          <div className="mb-4">
            <h3 className="text-sm font-medium text-slate-500 mb-1">ISBN</h3>
            <p className="dark:text-white font-mono text-sm">{media.isbn}</p>
          </div>
        )}

        {media.genre && (
          <div className="mb-4">
            <h3 className="text-sm font-medium text-slate-500 mb-1">Genre</h3>
            <p className="dark:text-white">{media.genre}</p>
          </div>
        )}

        {media.description && (
          <div className="mb-4">
            <h3 className="text-sm font-medium text-slate-500 mb-1">Description</h3>
            <p className="dark:text-slate-300">{media.description}</p>
          </div>
        )}

        {(media.volumes_total || media.volumes_owned) && (
          <div className="mb-4">
            <h3 className="text-sm font-medium text-slate-500 mb-1">Volumes</h3>
            <p className="dark:text-white">{media.volumes_owned || 0} / {media.volumes_total || '?'} owned</p>
          </div>
        )}

        {(media.chapters_total || media.chapters_read) && (
          <div className="mb-4">
            <h3 className="text-sm font-medium text-slate-500 mb-1">Chapters</h3>
            <p className="dark:text-white">{media.chapters_read || 0} / {media.chapters_total || '?'} read</p>
          </div>
        )}

        {(media.episodes_total || media.episodes_watched) && (
          <div className="mb-4">
            <h3 className="text-sm font-medium text-slate-500 mb-1">Episodes</h3>
            <p className="dark:text-white">{media.episodes_watched || 0} / {media.episodes_total || '?'} watched</p>
          </div>
        )}

        {media.tags?.length > 0 && (
          <div className="mb-4">
            <h3 className="text-sm font-medium text-slate-500 mb-1">Tags</h3>
            <div className="flex flex-wrap gap-2">
              {media.tags.map((tag) => (
                <span key={tag} className="px-2 py-1 bg-slate-100 dark:bg-slate-700 text-sm rounded-lg dark:text-slate-300">
                  {tag}
                </span>
              ))}
            </div>
          </div>
        )}

        {media.notes && (
          <div>
            <h3 className="text-sm font-medium text-slate-500 mb-1">Notes</h3>
            <p className="dark:text-slate-300">{media.notes}</p>
          </div>
        )}
      </div>
    </div>
  )
}
