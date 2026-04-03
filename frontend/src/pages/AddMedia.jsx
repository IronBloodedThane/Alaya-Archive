import { useState, useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { createMedia, updateMedia, getMedia } from '../api/media'

const MEDIA_TYPES = ['manga', 'anime', 'movie', 'book', 'game', 'tv_show', 'music', 'other']
const STATUSES = ['planned', 'in_progress', 'completed', 'on_hold', 'dropped']

export default function AddMedia() {
  const { mediaId } = useParams()
  const isEditing = Boolean(mediaId)
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  const [form, setForm] = useState({
    media_type: 'manga',
    title: '',
    title_original: '',
    description: '',
    cover_image: '',
    status: 'planned',
    rating: '',
    notes: '',
    year_released: '',
    creator: '',
    genre: '',
    volumes_total: '',
    volumes_owned: '',
    episodes_total: '',
    episodes_watched: '',
    chapters_total: '',
    chapters_read: '',
    is_public: true,
    tags: '',
  })
  const [error, setError] = useState('')

  const { data: existing } = useQuery({
    queryKey: ['media', mediaId],
    queryFn: () => getMedia(mediaId),
    enabled: isEditing,
  })

  useEffect(() => {
    if (existing) {
      setForm({
        media_type: existing.media_type,
        title: existing.title,
        title_original: existing.title_original || '',
        description: existing.description || '',
        cover_image: existing.cover_image || '',
        status: existing.status,
        rating: existing.rating ?? '',
        notes: existing.notes || '',
        year_released: existing.year_released ?? '',
        creator: existing.creator || '',
        genre: existing.genre || '',
        volumes_total: existing.volumes_total ?? '',
        volumes_owned: existing.volumes_owned ?? '',
        episodes_total: existing.episodes_total ?? '',
        episodes_watched: existing.episodes_watched ?? '',
        chapters_total: existing.chapters_total ?? '',
        chapters_read: existing.chapters_read ?? '',
        is_public: existing.is_public,
        tags: existing.tags?.join(', ') || '',
      })
    }
  }, [existing])

  const mutation = useMutation({
    mutationFn: (data) => isEditing ? updateMedia(mediaId, data) : createMedia(data),
    onSuccess: (result) => {
      queryClient.invalidateQueries({ queryKey: ['media'] })
      queryClient.invalidateQueries({ queryKey: ['media-stats'] })
      navigate(`/collection/${result.id}`)
    },
    onError: (err) => setError(err.response?.data?.error || 'Failed to save'),
  })

  const handleChange = (e) => {
    const { name, value, type, checked } = e.target
    setForm({ ...form, [name]: type === 'checkbox' ? checked : value })
  }

  const handleSubmit = (e) => {
    e.preventDefault()
    const data = {
      ...form,
      rating: form.rating ? parseInt(form.rating) : null,
      year_released: form.year_released ? parseInt(form.year_released) : null,
      volumes_total: form.volumes_total ? parseInt(form.volumes_total) : null,
      volumes_owned: form.volumes_owned ? parseInt(form.volumes_owned) : null,
      episodes_total: form.episodes_total ? parseInt(form.episodes_total) : null,
      episodes_watched: form.episodes_watched ? parseInt(form.episodes_watched) : null,
      chapters_total: form.chapters_total ? parseInt(form.chapters_total) : null,
      chapters_read: form.chapters_read ? parseInt(form.chapters_read) : null,
      tags: form.tags ? form.tags.split(',').map((t) => t.trim()).filter(Boolean) : [],
    }
    mutation.mutate(data)
  }

  const inputClass = 'w-full px-4 py-2 bg-white dark:bg-slate-700 border border-slate-200 dark:border-slate-600 rounded-lg dark:text-white focus:outline-none focus:border-indigo-500'
  const labelClass = 'block text-sm font-medium text-slate-600 dark:text-slate-300 mb-1'

  const showMangaFields = ['manga', 'book'].includes(form.media_type)
  const showAnimeFields = ['anime', 'tv_show'].includes(form.media_type)

  return (
    <div className="max-w-2xl mx-auto">
      <h1 className="text-2xl font-bold dark:text-white mb-6">
        {isEditing ? 'Edit Media' : 'Add to Collection'}
      </h1>

      <form onSubmit={handleSubmit} className="bg-white dark:bg-slate-800 rounded-xl p-6 border border-slate-200 dark:border-slate-700">
        {error && (
          <div className="mb-4 p-3 bg-red-500/10 border border-red-500/30 rounded-lg text-red-400 text-sm">{error}</div>
        )}

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
          <div>
            <label className={labelClass}>Type *</label>
            <select name="media_type" value={form.media_type} onChange={handleChange} className={inputClass}>
              {MEDIA_TYPES.map((t) => <option key={t} value={t}>{t.replace('_', ' ')}</option>)}
            </select>
          </div>
          <div>
            <label className={labelClass}>Status</label>
            <select name="status" value={form.status} onChange={handleChange} className={inputClass}>
              {STATUSES.map((s) => <option key={s} value={s}>{s.replace('_', ' ')}</option>)}
            </select>
          </div>
        </div>

        <div className="mb-4">
          <label className={labelClass}>Title *</label>
          <input type="text" name="title" value={form.title} onChange={handleChange} className={inputClass} required />
        </div>

        <div className="mb-4">
          <label className={labelClass}>Original Title</label>
          <input type="text" name="title_original" value={form.title_original} onChange={handleChange} className={inputClass} placeholder="Japanese/original language title" />
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
          <div>
            <label className={labelClass}>Creator / Author</label>
            <input type="text" name="creator" value={form.creator} onChange={handleChange} className={inputClass} />
          </div>
          <div>
            <label className={labelClass}>Year Released</label>
            <input type="number" name="year_released" value={form.year_released} onChange={handleChange} className={inputClass} />
          </div>
        </div>

        <div className="mb-4">
          <label className={labelClass}>Genre</label>
          <input type="text" name="genre" value={form.genre} onChange={handleChange} className={inputClass} placeholder="e.g. Action, Fantasy, Sci-Fi" />
        </div>

        <div className="mb-4">
          <label className={labelClass}>Description</label>
          <textarea name="description" value={form.description} onChange={handleChange} rows={3} className={inputClass} />
        </div>

        <div className="mb-4">
          <label className={labelClass}>Cover Image URL</label>
          <input type="text" name="cover_image" value={form.cover_image} onChange={handleChange} className={inputClass} placeholder="https://..." />
        </div>

        {showMangaFields && (
          <div className="grid grid-cols-2 gap-4 mb-4">
            <div>
              <label className={labelClass}>Volumes Total</label>
              <input type="number" name="volumes_total" value={form.volumes_total} onChange={handleChange} className={inputClass} />
            </div>
            <div>
              <label className={labelClass}>Volumes Owned</label>
              <input type="number" name="volumes_owned" value={form.volumes_owned} onChange={handleChange} className={inputClass} />
            </div>
            <div>
              <label className={labelClass}>Chapters Total</label>
              <input type="number" name="chapters_total" value={form.chapters_total} onChange={handleChange} className={inputClass} />
            </div>
            <div>
              <label className={labelClass}>Chapters Read</label>
              <input type="number" name="chapters_read" value={form.chapters_read} onChange={handleChange} className={inputClass} />
            </div>
          </div>
        )}

        {showAnimeFields && (
          <div className="grid grid-cols-2 gap-4 mb-4">
            <div>
              <label className={labelClass}>Episodes Total</label>
              <input type="number" name="episodes_total" value={form.episodes_total} onChange={handleChange} className={inputClass} />
            </div>
            <div>
              <label className={labelClass}>Episodes Watched</label>
              <input type="number" name="episodes_watched" value={form.episodes_watched} onChange={handleChange} className={inputClass} />
            </div>
          </div>
        )}

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
          <div>
            <label className={labelClass}>Rating (1-10)</label>
            <input type="number" name="rating" value={form.rating} onChange={handleChange} min="1" max="10" className={inputClass} />
          </div>
          <div>
            <label className={labelClass}>Tags</label>
            <input type="text" name="tags" value={form.tags} onChange={handleChange} className={inputClass} placeholder="tag1, tag2, tag3" />
          </div>
        </div>

        <div className="mb-4">
          <label className={labelClass}>Notes</label>
          <textarea name="notes" value={form.notes} onChange={handleChange} rows={2} className={inputClass} />
        </div>

        <div className="mb-6">
          <label className="flex items-center gap-2 cursor-pointer">
            <input type="checkbox" name="is_public" checked={form.is_public} onChange={handleChange} className="rounded" />
            <span className="text-sm dark:text-slate-300">Show in public collection</span>
          </label>
        </div>

        <div className="flex gap-3">
          <button
            type="submit"
            disabled={mutation.isPending}
            className="px-6 py-2 bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 text-white font-medium rounded-lg"
          >
            {mutation.isPending ? 'Saving...' : isEditing ? 'Save Changes' : 'Add to Collection'}
          </button>
          <button
            type="button"
            onClick={() => navigate(-1)}
            className="px-6 py-2 border border-slate-300 dark:border-slate-600 text-slate-600 dark:text-slate-300 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-700"
          >
            Cancel
          </button>
        </div>
      </form>
    </div>
  )
}
