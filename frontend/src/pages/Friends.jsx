import { Link } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getFriends, getFriendRequests, acceptFriendRequest, rejectFriendRequest, removeFriend } from '../api/social'
import Avatar from '../components/Avatar'

export default function Friends() {
  const queryClient = useQueryClient()

  const { data: friends = [] } = useQuery({ queryKey: ['friends'], queryFn: getFriends })
  const { data: requests = [] } = useQuery({ queryKey: ['friend-requests'], queryFn: getFriendRequests })

  const acceptMutation = useMutation({
    mutationFn: acceptFriendRequest,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['friends'] })
      queryClient.invalidateQueries({ queryKey: ['friend-requests'] })
    },
  })

  const rejectMutation = useMutation({
    mutationFn: rejectFriendRequest,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['friend-requests'] }),
  })

  const removeMutation = useMutation({
    mutationFn: removeFriend,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['friends'] }),
  })

  return (
    <div className="max-w-2xl mx-auto">
      <h1 className="text-2xl font-bold dark:text-white mb-6">Friends</h1>

      {requests.length > 0 && (
        <div className="mb-8">
          <h2 className="text-lg font-semibold dark:text-white mb-3">Friend Requests</h2>
          <div className="space-y-2">
            {requests.map((req) => (
              <div key={req.id} className="flex items-center justify-between bg-white dark:bg-slate-800 rounded-xl p-4 border border-slate-200 dark:border-slate-700">
                <Link to={`/user/${req.from_username}`} className="flex items-center gap-3 min-w-0">
                  <Avatar username={req.from_username} displayName={req.from_display_name} size="md" />
                  <div className="min-w-0">
                    <p className="font-medium dark:text-white truncate">{req.from_display_name || req.from_username}</p>
                    <p className="text-sm text-slate-500 truncate">@{req.from_username}</p>
                  </div>
                </Link>
                <div className="flex gap-2">
                  <button
                    onClick={() => acceptMutation.mutate(req.id)}
                    className="px-3 py-1.5 bg-green-600 hover:bg-green-500 text-white text-sm rounded-lg"
                  >
                    Accept
                  </button>
                  <button
                    onClick={() => rejectMutation.mutate(req.id)}
                    className="px-3 py-1.5 bg-slate-600 hover:bg-slate-500 text-white text-sm rounded-lg"
                  >
                    Decline
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      <h2 className="text-lg font-semibold dark:text-white mb-3">
        My Friends ({friends.length})
      </h2>

      {friends.length === 0 ? (
        <p className="text-slate-400">No friends yet. Search for users to add friends.</p>
      ) : (
        <div className="space-y-2">
          {friends.map((friend) => (
            <div key={friend.id} className="flex items-center justify-between bg-white dark:bg-slate-800 rounded-xl p-4 border border-slate-200 dark:border-slate-700">
              <Link to={`/user/${friend.username}`} className="flex items-center gap-3 min-w-0">
                <Avatar username={friend.username} displayName={friend.display_name} size="md" version={friend.updated_at} />
                <div className="min-w-0">
                  <p className="font-medium dark:text-white truncate">{friend.display_name || friend.username}</p>
                  <p className="text-sm text-slate-500 truncate">@{friend.username}</p>
                </div>
              </Link>
              <button
                onClick={() => {
                  if (window.confirm(`Remove ${friend.username} as a friend?`)) {
                    removeMutation.mutate(friend.id)
                  }
                }}
                className="text-sm text-red-400 hover:text-red-300"
              >
                Remove
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
