import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { useAuth } from '../hooks/useAuth'
import { updateProfile } from '../api/auth'

export default function Profile() {
  const { user, updateUser } = useAuth()
  const [form, setForm] = useState({
    display_name: user?.display_name || '',
    bio: user?.bio || '',
    profile_public: user?.profile_public || false,
  })
  const [saved, setSaved] = useState(false)

  const mutation = useMutation({
    mutationFn: updateProfile,
    onSuccess: (data) => {
      updateUser(data)
      setSaved(true)
      setTimeout(() => setSaved(false), 2000)
    },
  })

  const handleSubmit = (e) => {
    e.preventDefault()
    mutation.mutate(form)
  }

  const inputClass = 'w-full px-4 py-2 bg-white dark:bg-slate-700 border border-slate-200 dark:border-slate-600 rounded-lg dark:text-white focus:outline-none focus:border-indigo-500'

  return (
    <div className="max-w-xl mx-auto">
      <h1 className="text-2xl font-bold dark:text-white mb-6">Profile Settings</h1>

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
    </div>
  )
}
