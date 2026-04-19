import client from './client'

export const listMedia = (params) => client.get('/media', { params }).then((r) => r.data)
export const getMedia = (id) => client.get(`/media/${id}`).then((r) => r.data)
export const createMedia = (data) => client.post('/media', data).then((r) => r.data)
export const updateMedia = (id, data) => client.patch(`/media/${id}`, data).then((r) => r.data)
export const deleteMedia = (id) => client.delete(`/media/${id}`).then((r) => r.data)
export const rateMedia = (id, rating) => client.post(`/media/${id}/rating`, { rating }).then((r) => r.data)
export const setTags = (id, tags) => client.post(`/media/${id}/tags`, { tags }).then((r) => r.data)
export const getStats = () => client.get('/media/stats').then((r) => r.data)
export const searchMedia = (q, params) => client.get('/media/search', { params: { q, ...params } }).then((r) => r.data)
export const getPublicProfile = (username) => client.get(`/users/${encodeURIComponent(username)}`).then((r) => r.data)
export const getPublicCollection = (username, params) => client.get(`/users/${encodeURIComponent(username)}/collection`, { params }).then((r) => r.data)
