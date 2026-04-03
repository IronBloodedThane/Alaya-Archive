import { useParams } from 'react-router-dom'

export default function PublicProfile() {
  const { username } = useParams()

  return (
    <div className="min-h-screen bg-slate-900 flex items-center justify-center">
      <div className="text-center">
        <h1 className="text-2xl font-bold text-white mb-2">@{username}</h1>
        <p className="text-slate-400">Public profile coming soon</p>
      </div>
    </div>
  )
}
