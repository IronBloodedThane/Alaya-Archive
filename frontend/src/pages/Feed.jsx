import { Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { getFeed } from '../api/social'
import Avatar from '../components/Avatar'

export default function Feed() {
  const { data: items = [], isLoading } = useQuery({
    queryKey: ['feed'],
    queryFn: () => getFeed({ limit: 50 }),
  })

  return (
    <div className="max-w-2xl mx-auto">
      <h1 className="text-2xl font-bold dark:text-white mb-6">Activity Feed</h1>

      {isLoading ? (
        <div className="text-slate-400">Loading...</div>
      ) : items.length === 0 ? (
        <div className="text-center py-12">
          <p className="text-slate-400 mb-2">No activity yet.</p>
          <p className="text-sm text-slate-500">Follow friends or add them to see their activity here.</p>
        </div>
      ) : (
        <div className="space-y-3">
          {items.map((item) => (
            <div key={item.id} className="bg-white dark:bg-slate-800 rounded-xl p-4 border border-slate-200 dark:border-slate-700">
              <div className="flex items-start gap-3">
                <Link to={`/user/${item.username}`}>
                  <Avatar username={item.username} displayName={item.display_name} size="md" />
                </Link>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-1 flex-wrap">
                    <Link to={`/user/${item.username}`} className="font-medium dark:text-white hover:underline">
                      {item.display_name || item.username}
                    </Link>
                    <Link to={`/user/${item.username}`} className="text-sm text-slate-500 hover:underline">
                      @{item.username}
                    </Link>
                  </div>
                  <p className="text-sm text-slate-400">{item.item_type.replace('_', ' ')}</p>
                  <p className="text-xs text-slate-500 mt-1">
                    {new Date(item.created_at).toLocaleDateString()}
                  </p>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
