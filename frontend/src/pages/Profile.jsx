import { useRef, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useMutation } from '@tanstack/react-query'
import { useAuth } from '../hooks/useAuth'
import { updateProfile, uploadAvatar, deleteAvatar, deleteAccount } from '../api/auth'
import Avatar from '../components/Avatar'

export default function Profile() {
  const { user, updateUser, logoutUser } = useAuth()
  const navigate = useNavigate()
  const fileInputRef = useRef(null)
  const [form, setForm] = useState({
    display_name: user?.display_name || '',
    bio: user?.bio || '',
    profile_public: user?.profile_public || false,
  })
  const [saved, setSaved] = useState(false)
  const [avatarError, setAvatarError] = useState('')
  const [copied, setCopied] = useState(false)
  const [deleteConfirmation, setDeleteConfirmation] = useState('')
  const [deleteError, setDeleteError] = useState('')

  const publicPath = user?.username ? `/user/${user.username}` : ''
  const publicUrl = user?.username && typeof window !== 'undefined' ? `${window.location.origin}${publicPath}` : ''

  const copyPublicLink = async () => {
    try {
      await navigator.clipboard.writeText(publicUrl)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      // Clipboard not available (e.g. insecure context) — fall back to selecting the text
    }
  }

  const mutation = useMutation({
    mutationFn: updateProfile,
    onSuccess: (data) => {
      updateUser(data)
      setSaved(true)
      setTimeout(() => setSaved(false), 2000)
    },
  })

  const avatarUpload = useMutation({
    mutationFn: uploadAvatar,
    onMutate: () => setAvatarError(''),
    onSuccess: (data) => updateUser(data),
    onError: (err) => setAvatarError(err?.response?.data?.error || 'upload failed'),
  })

  const avatarDelete = useMutation({
    mutationFn: deleteAvatar,
    onSuccess: (data) => updateUser(data),
  })

  const accountDelete = useMutation({
    mutationFn: () => deleteAccount('DELETE'),
    onMutate: () => setDeleteError(''),
    onSuccess: () => {
      logoutUser()
      navigate('/', { replace: true })
    },
    onError: (err) => {
      setDeleteError(err?.response?.data?.error || 'failed to delete account')
    },
  })

  const handleDelete = () => {
    if (deleteConfirmation !== 'DELETE') return
    const ok = window.confirm(
      'This permanently deletes your account, collection, tags, friends, and all activity. It cannot be undone. Continue?'
    )
    if (!ok) return
    accountDelete.mutate()
  }

  const handleSubmit = (e) => {
    e.preventDefault()
    mutation.mutate(form)
  }

  const handleFilePick = (e) => {
    const file = e.target.files?.[0]
    if (!file) return
    if (file.size > 5 * 1024 * 1024) {
      setAvatarError('file too large (max 5MB)')
      return
    }
    avatarUpload.mutate(file)
    e.target.value = ''
  }

  const inputClass = 'w-full px-4 py-2 bg-white dark:bg-slate-700 border border-slate-200 dark:border-slate-600 rounded-lg dark:text-white focus:outline-none focus:border-indigo-500'

  return (
    <div className="max-w-xl mx-auto">
      <h1 className="text-2xl font-bold dark:text-white mb-6">Profile Settings</h1>

      <div className="bg-white dark:bg-slate-800 rounded-xl p-6 border border-slate-200 dark:border-slate-700 mb-6">
        <div className="flex items-center gap-4">
          <Avatar username={user?.username} displayName={user?.display_name} size="xl" version={user?.updated_at} />
          <div className="flex-1">
            <div className="flex gap-2 flex-wrap">
              <input
                ref={fileInputRef}
                type="file"
                accept="image/jpeg,image/png,image/gif,image/webp"
                onChange={handleFilePick}
                className="hidden"
              />
              <button
                type="button"
                onClick={() => fileInputRef.current?.click()}
                disabled={avatarUpload.isPending}
                className="px-4 py-2 bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 text-white text-sm font-medium rounded-lg"
              >
                {avatarUpload.isPending ? 'Uploading...' : user?.has_avatar ? 'Replace photo' : 'Upload photo'}
              </button>
              {user?.has_avatar && (
                <button
                  type="button"
                  onClick={() => avatarDelete.mutate()}
                  disabled={avatarDelete.isPending}
                  className="px-4 py-2 bg-slate-200 dark:bg-slate-700 hover:bg-slate-300 dark:hover:bg-slate-600 disabled:opacity-50 text-slate-700 dark:text-slate-200 text-sm font-medium rounded-lg"
                >
                  Remove
                </button>
              )}
            </div>
            <p className="text-xs text-slate-500 mt-2">JPEG, PNG, GIF, or WebP. Max 5MB.</p>
            {avatarError && <p className="text-xs text-red-500 mt-1">{avatarError}</p>}
          </div>
        </div>
      </div>

      <div className="bg-white dark:bg-slate-800 rounded-xl p-6 border border-slate-200 dark:border-slate-700 mb-6">
        <h2 className="text-sm font-medium text-slate-600 dark:text-slate-300 mb-2">Public profile</h2>
        {user?.profile_public ? (
          <>
            <p className="text-xs text-slate-500 mb-2">Anyone with this link can view your profile and public collection.</p>
            <div className="flex flex-wrap gap-2">
              <input
                type="text"
                readOnly
                value={publicUrl}
                onFocus={(e) => e.target.select()}
                className="flex-1 min-w-0 px-3 py-2 text-sm bg-slate-50 dark:bg-slate-700 border border-slate-200 dark:border-slate-600 rounded-lg dark:text-white"
              />
              <button
                type="button"
                onClick={copyPublicLink}
                className="px-3 py-2 bg-slate-200 dark:bg-slate-700 hover:bg-slate-300 dark:hover:bg-slate-600 text-slate-700 dark:text-slate-200 text-sm font-medium rounded-lg"
              >
                {copied ? 'Copied!' : 'Copy link'}
              </button>
              <Link
                to={publicPath}
                className="px-3 py-2 bg-indigo-600 hover:bg-indigo-500 text-white text-sm font-medium rounded-lg"
              >
                View
              </Link>
            </div>
          </>
        ) : (
          <p className="text-xs text-slate-500">
            Your profile is private. Enable <span className="font-medium">Make my profile and collection public</span> below to get a shareable link.
          </p>
        )}
      </div>

      <form onSubmit={handleSubmit} className="bg-white dark:bg-slate-800 rounded-xl p-6 border border-slate-200 dark:border-slate-700">
        <div className="mb-4">
          <label className="block text-sm font-medium text-slate-600 dark:text-slate-300 mb-1">Email</label>
          <p className="text-slate-400">{user?.email}</p>
        </div>

        <div className="mb-4">
          <label className="block text-sm font-medium text-slate-600 dark:text-slate-300 mb-1">Username</label>
          <p className="text-slate-400">@{user?.username}</p>
        </div>

        <div className="mb-4">
          <label className="block text-sm font-medium text-slate-600 dark:text-slate-300 mb-1">Display Name</label>
          <input
            type="text"
            value={form.display_name}
            onChange={(e) => setForm({ ...form, display_name: e.target.value })}
            className={inputClass}
          />
        </div>

        <div className="mb-4">
          <label className="block text-sm font-medium text-slate-600 dark:text-slate-300 mb-1">Bio</label>
          <textarea
            value={form.bio}
            onChange={(e) => setForm({ ...form, bio: e.target.value })}
            rows={3}
            className={inputClass}
          />
        </div>

        <div className="mb-6">
          <label className="flex items-center gap-2 cursor-pointer">
            <input
              type="checkbox"
              checked={form.profile_public}
              onChange={(e) => setForm({ ...form, profile_public: e.target.checked })}
              className="rounded"
            />
            <span className="text-sm dark:text-slate-300">Make my profile and collection public</span>
          </label>
        </div>

        <button
          type="submit"
          disabled={mutation.isPending}
          className="px-6 py-2 bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 text-white font-medium rounded-lg"
        >
          {mutation.isPending ? 'Saving...' : saved ? 'Saved!' : 'Save Changes'}
        </button>
      </form>

      <section className="mt-8 bg-white dark:bg-slate-800 rounded-xl p-6 border border-red-300 dark:border-red-900/60">
        <h2 className="text-lg font-semibold text-red-600 dark:text-red-400 mb-2">Delete account</h2>
        <p className="text-sm text-slate-600 dark:text-slate-300 mb-4">
          Permanently remove your account and every piece of data tied to it — your
          collection, tags, ratings, notes, profile, avatar, friend connections, and
          activity. This cannot be undone.
        </p>
        <label className="block text-sm font-medium text-slate-600 dark:text-slate-300 mb-1">
          Type <span className="font-mono text-red-600 dark:text-red-400">DELETE</span> to confirm
        </label>
        <input
          type="text"
          value={deleteConfirmation}
          onChange={(e) => setDeleteConfirmation(e.target.value)}
          placeholder="DELETE"
          autoComplete="off"
          className={inputClass + ' mb-3'}
        />
        {deleteError && (
          <p className="text-sm text-red-600 dark:text-red-400 mb-3">{deleteError}</p>
        )}
        <button
          type="button"
          onClick={handleDelete}
          disabled={deleteConfirmation !== 'DELETE' || accountDelete.isPending}
          className="px-6 py-2 bg-red-600 hover:bg-red-500 disabled:opacity-50 disabled:cursor-not-allowed text-white font-medium rounded-lg"
        >
          {accountDelete.isPending ? 'Deleting...' : 'Delete my account'}
        </button>
      </section>
    </div>
  )
}
