import client from './client'

export const followUser = (userId) => client.post(`/social/follow/${userId}`).then((r) => r.data)
export const unfollowUser = (userId) => client.delete(`/social/follow/${userId}`).then((r) => r.data)
export const getFollowers = () => client.get('/social/followers').then((r) => r.data)
export const getFollowing = () => client.get('/social/following').then((r) => r.data)
export const getFeed = (params) => client.get('/social/feed', { params }).then((r) => r.data)

export const sendFriendRequest = (userId) => client.post(`/friends/request/${userId}`).then((r) => r.data)
export const acceptFriendRequest = (requestId) => client.post(`/friends/accept/${requestId}`).then((r) => r.data)
export const rejectFriendRequest = (requestId) => client.post(`/friends/reject/${requestId}`).then((r) => r.data)
export const getFriends = () => client.get('/friends').then((r) => r.data)
export const getFriendRequests = () => client.get('/friends/requests').then((r) => r.data)
export const removeFriend = (friendId) => client.delete(`/friends/${friendId}`).then((r) => r.data)
