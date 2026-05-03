import client from './client'

export const listMedia = (params) => client.get('/media', { params }).then((r) => r.data)
export const getMedia = (id) => client.get(`/media/${id}`).then((r) => r.data)
// createMedia accepts an optional onDuplicate flag: 'error' | 'overwrite' |
// 'skip' | 'allow'. When omitted the backend defaults to 'error', which
// returns 409 + the existing record so the caller can prompt the user.
export const createMedia = (data, onDuplicate) =>
  client.post('/media', onDuplicate ? { ...data, on_duplicate: onDuplicate } : data).then((r) => r.data)
export const updateMedia = (id, data) => client.patch(`/media/${id}`, data).then((r) => r.data)
// checkDuplicate returns existing items with the same (media_type, isbn).
// Empty isbn returns an empty list — used to pre-warn the user before submit.
export const checkDuplicate = (mediaType, isbn) =>
  client.get('/media/check', { params: { type: mediaType, isbn } }).then((r) => r.data)
// lookupByIsbn proxies to the backend's /lookup endpoint, which calls Google
// Books today. Response: { provider, result: {...metadata...} }.
export const lookupByIsbn = (mediaType, isbn) =>
  client.get('/lookup', { params: { type: mediaType, isbn } }).then((r) => r.data)
export const deleteMedia = (id) => client.delete(`/media/${id}`).then((r) => r.data)
export const rateMedia = (id, rating) => client.post(`/media/${id}/rating`, { rating }).then((r) => r.data)
export const setTags = (id, tags) => client.post(`/media/${id}/tags`, { tags }).then((r) => r.data)
export const getStats = () => client.get('/media/stats').then((r) => r.data)
export const searchMedia = (q, params) => client.get('/media/search', { params: { q, ...params } }).then((r) => r.data)
export const getPublicProfile = (username) => client.get(`/users/${encodeURIComponent(username)}`).then((r) => r.data)
export const getPublicCollection = (username, params) => client.get(`/users/${encodeURIComponent(username)}/collection`, { params }).then((r) => r.data)
