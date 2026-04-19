import client from './client'

export const register = (data) => client.post('/auth/register', data).then((r) => r.data)
export const login = (data) => client.post('/auth/login', data).then((r) => r.data)
export const refreshToken = (refresh_token) => client.post('/auth/refresh', { refresh_token }).then((r) => r.data)
export const verifyEmail = (token) => client.post('/auth/verify-email', { token }).then((r) => r.data)
export const forgotPassword = (email) => client.post('/auth/forgot-password', { email }).then((r) => r.data)
export const resetPassword = (token, new_password) => client.post('/auth/reset-password', { token, new_password }).then((r) => r.data)
export const changePassword = (current_password, new_password) => client.post('/auth/change-password', { current_password, new_password }).then((r) => r.data)
export const deleteAccount = (confirmation) => client.post('/auth/delete-account', { confirmation }).then((r) => r.data)
export const getCurrentUser = () => client.get('/users/me').then((r) => r.data)
export const updateProfile = (data) => client.patch('/users/me', data).then((r) => r.data)
export const uploadAvatar = (file) => {
  const fd = new FormData()
  fd.append('avatar', file)
  return client.post('/users/me/avatar', fd, { headers: { 'Content-Type': 'multipart/form-data' } }).then((r) => r.data)
}
export const deleteAvatar = () => client.delete('/users/me/avatar').then((r) => r.data)
