import { useState, useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { createMedia, updateMedia, getMedia, lookupByIsbn } from '../api/media'

const MEDIA_TYPES = ['manga', 'anime', 'movie', 'book', 'game', 'tv_show', 'music', 'other']
const STATUSES = ['planned', 'in_progress', 'completed', 'on_hold', 'dropped']
// list_type fingerprints "do I own this" vs "do I want this" — orthogonal
// to status (which is consumption progress). Mobile batch scanner sets
// this; the web form lets the user pick when adding manually.
const LIST_TYPES = [
  { value: 'owned', label: 'Collection (I have it)' },
  { value: 'wishlist', label: 'Wishlist (I want it)' },
]
// Which media types support an ISBN. Anything else hides the ISBN field.
const ISBN_TYPES = new Set(['book', 'manga'])

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
    isbn: '',
    series: '',
    series_position: '',
    list_type: 'owned',
    is_public: true,
    tags: '',
  })
  const [error, setError] = useState('')
  // conflict holds the existing items returned by a 409 from the create
  // endpoint. While set, the form shows the duplicate-resolution banner
  // instead of the normal submit button.
  const [conflict, setConflict] = useState(null)
  // ISBN auto-fill state: lookupBusy disables the button mid-request,
  // lookupNote surfaces the result ("filled from Google Books" or an error).
  const [lookupBusy, setLookupBusy] = useState(false)
  const [lookupNote, setLookupNote] = useState(null)

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
        isbn: existing.isbn || '',
        series: existing.series || '',
        series_position: existing.series_position ?? '',
        list_type: existing.list_type || 'owned',
        is_public: existing.is_public,
        tags: existing.tags?.join(', ') || '',
      })
    }
  }, [existing])

  const mutation = useMutation({
    mutationFn: ({ data, onDuplicate }) =>
      isEditing ? updateMedia(mediaId, data) : createMedia(data, onDuplicate),
    onSuccess: (result) => {
      queryClient.invalidateQueries({ queryKey: ['media'] })
      queryClient.invalidateQueries({ queryKey: ['media-stats'] })
      navigate(`/collection/${result.id}`)
    },
    onError: (err) => {
      // 409 means the backend found an existing item with the same
      // (media_type, isbn). Surface the conflict UI; the user picks
      // overwrite/skip/add-as-copy and we resubmit.
      if (err.response?.status === 409 && err.response?.data?.existing) {
        setConflict(err.response.data.existing)
        setError('')
        return
      }
      setError(err.response?.data?.error || 'Failed to save')
    },
  })

  const handleChange = (e) => {
    const { name, value, type, checked } = e.target
    setForm({ ...form, [name]: type === 'checkbox' ? checked : value })
  }

  const buildPayload = () => ({
    ...form,
    rating: form.rating ? parseInt(form.rating) : null,
    year_released: form.year_released ? parseInt(form.year_released) : null,
    volumes_total: form.volumes_total ? parseInt(form.volumes_total) : null,
    volumes_owned: form.volumes_owned ? parseInt(form.volumes_owned) : null,
    episodes_total: form.episodes_total ? parseInt(form.episodes_total) : null,
    episodes_watched: form.episodes_watched ? parseInt(form.episodes_watched) : null,
    chapters_total: form.chapters_total ? parseInt(form.chapters_total) : null,
    chapters_read: form.chapters_read ? parseInt(form.chapters_read) : null,
    series_position: form.series_position ? parseInt(form.series_position) : null,
    tags: form.tags ? form.tags.split(',').map((t) => t.trim()).filter(Boolean) : [],
  })

  const handleSubmit = (e) => {
    e.preventDefault()
    setConflict(null)
    mutation.mutate({ data: buildPayload() })
  }

  const resolveConflict = (onDuplicate) => {
    setConflict(null)
    mutation.mutate({ data: buildPayload(), onDuplicate })
  }

  // Strip hyphens/spaces and require a 10- or 13-digit ISBN before allowing
  // the lookup button to fire — saves a wasted round-trip and a generic 404.
  const normalizedIsbn = form.isbn.replace(/[\s-]/g, '')
  const isbnLooksValid =
    normalizedIsbn.length === 10 || normalizedIsbn.length === 13

  const handleLookup = async () => {
    if (!isbnLooksValid || lookupBusy) return
    setLookupBusy(true)
    setLookupNote(null)
    try {
      const { provider, result } = await lookupByIsbn(form.media_type, normalizedIsbn)
      // Fill from the result; the user can still edit anything before save.
      // We prefer the canonical isbn_13 from the provider over whatever the
      // user typed (handles isbn-10 → isbn-13 normalization).
      setForm((f) => ({
        ...f,
        title: result.title || f.title,
        creator: (result.authors && result.authors.join(', ')) || f.creator,
        year_released: result.year ? String(result.year) : f.year_released,
        description: result.description || f.description,
        cover_image: result.cover_image || f.cover_image,
        isbn: result.isbn_13 || result.isbn_10 || f.isbn,
        series: result.series || f.series,
        series_position:
          result.series_position ? String(result.series_position) : f.series_position,
      }))
      setLookupNote({ kind: 'ok', text: `Filled from ${provider}` })
    } catch (err) {
      const status = err.response?.status
      const msg =
        status === 404
          ? 'No book found for that ISBN.'
          : err.response?.data?.error || 'Lookup failed.'
      setLookupNote({ kind: 'error', text: msg })
    } finally {
      setLookupBusy(false)
    }
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

        {conflict && (
          <ConflictBanner
            existing={conflict}
            onResolve={resolveConflict}
            onDismiss={() => setConflict(null)}
          />
        )}

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
          <div>
            <label className={labelClass}>Type *</label>
            <select name="media_type" value={form.media_type} onChange={handleChange} className={inputClass}>
              {MEDIA_TYPES.map((t) => <option key={t} value={t}>{t.replace('_', ' ')}</option>)}
            </select>
          </div>
          <div>
            <label className={labelClass}>List</label>
            <select name="list_type" value={form.list_type} onChange={handleChange} className={inputClass}>
              {LIST_TYPES.map((l) => <option key={l.value} value={l.value}>{l.label}</option>)}
            </select>
          </div>
        </div>

        <div className="mb-4">
          <label className={labelClass}>Status</label>
          <select name="status" value={form.status} onChange={handleChange} className={inputClass}>
            {STATUSES.map((s) => <option key={s} value={s}>{s.replace('_', ' ')}</option>)}
          </select>
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

        {ISBN_TYPES.has(form.media_type) && (
          <div className="mb-4">
            <label className={labelClass}>ISBN</label>
            <div className="flex gap-2">
              <input
                type="text"
                name="isbn"
                value={form.isbn}
                onChange={handleChange}
                className={inputClass}
                placeholder="978..."
                inputMode="numeric"
              />
              <button
                type="button"
                onClick={handleLookup}
                disabled={!isbnLooksValid || lookupBusy}
                title={isbnLooksValid ? 'Fill the form from this ISBN' : 'Enter a 10- or 13-digit ISBN'}
                className="shrink-0 px-4 py-2 bg-indigo-600 hover:bg-indigo-500 disabled:opacity-40 disabled:cursor-not-allowed text-white text-sm font-medium rounded-lg"
              >
                {lookupBusy ? 'Looking up…' : 'Look up'}
              </button>
            </div>
            {lookupNote ? (
              <p
                className={`mt-1 text-xs ${
                  lookupNote.kind === 'ok' ? 'text-green-500' : 'text-red-500'
                }`}
              >
                {lookupNote.text}
              </p>
            ) : (
              <p className="mt-1 text-xs text-slate-500">
                Used to detect duplicates and share entries with the mobile app's barcode scanner.
              </p>
            )}
          </div>
        )}

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-4">
          <div className="md:col-span-2">
            <label className={labelClass}>Series</label>
            <input
              type="text"
              name="series"
              value={form.series}
              onChange={handleChange}
              className={inputClass}
              placeholder="e.g. Berserk, Dune"
            />
          </div>
          <div>
            <label className={labelClass}>Volume #</label>
            <input
              type="number"
              name="series_position"
              value={form.series_position}
              onChange={handleChange}
              className={inputClass}
              min="1"
            />
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
            {mutation.isPending
              ? 'Saving...'
              : isEditing
                ? 'Save Changes'
                : form.list_type === 'wishlist'
                  ? 'Add to Wishlist'
                  : 'Add to Collection'}
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

// ConflictBanner is shown when POST /media returns 409 because an item with
// the same (media_type, isbn) already exists in the user's collection. It
// surfaces the existing item and offers the same three policies the mobile
// app uses: skip, overwrite, or add as a second copy.
function ConflictBanner({ existing, onResolve, onDismiss }) {
  const first = existing[0] || {}
  const n = existing.length
  const subtitle = [first.title, first.creator].filter(Boolean).join(' — ')
  return (
    <div className="mb-6 p-4 bg-amber-500/10 border border-amber-500/40 rounded-lg">
      <div className="flex items-start justify-between gap-4 mb-3">
        <div>
          <h3 className="font-semibold text-amber-700 dark:text-amber-300">
            Already in your collection{n > 1 ? ` (${n} copies)` : ''}
          </h3>
          {subtitle && (
            <p className="text-sm text-amber-700/90 dark:text-amber-200/90 mt-1">
              {subtitle}
            </p>
          )}
          <p className="text-sm text-slate-600 dark:text-slate-300 mt-3 font-medium">
            What do you want to do?
          </p>
        </div>
        <button
          type="button"
          onClick={onDismiss}
          className="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 text-sm"
          aria-label="Dismiss"
        >
          ✕
        </button>
      </div>
      <div className="flex flex-wrap gap-2">
        <ConflictAction
          label="Skip"
          description="Keep the existing entry as-is"
          onClick={() => onResolve('skip')}
        />
        <ConflictAction
          label="Overwrite"
          description="Replace the existing entry with this form"
          onClick={() => onResolve('overwrite')}
          primary
        />
        <ConflictAction
          label="Add as second copy"
          description="Create a new entry alongside it"
          onClick={() => onResolve('allow')}
        />
      </div>
    </div>
  )
}

function ConflictAction({ label, description, onClick, primary }) {
  const base = 'flex-1 min-w-[140px] text-left p-3 rounded-lg border transition-colors'
  const tone = primary
    ? 'bg-indigo-600 hover:bg-indigo-500 border-indigo-600 text-white'
    : 'bg-white dark:bg-slate-800 hover:bg-slate-50 dark:hover:bg-slate-700 border-slate-300 dark:border-slate-600 text-slate-700 dark:text-slate-200'
  return (
    <button type="button" onClick={onClick} className={`${base} ${tone}`}>
      <div className="font-semibold text-sm">{label}</div>
      <div className={`text-xs mt-0.5 ${primary ? 'text-indigo-100' : 'text-slate-500 dark:text-slate-400'}`}>
        {description}
      </div>
    </button>
  )
}
