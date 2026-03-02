import api from './client'

export interface ParentalStatus {
  adult_content_enabled: boolean
  pin_set: boolean
}

export interface AdminParentalStatus {
  pin_set: boolean
  restricted_genre_codes: string[]
}

export interface UserAdultStatus {
  user_id: string
  username: string
  display_name: string
  role: string
  adult_content_enabled: boolean
}

// --- User endpoints ---

export async function getMyParentalStatus(): Promise<ParentalStatus> {
  const { data } = await api.get<ParentalStatus>('/me/parental/status')
  return data
}

export async function unlockAdultContent(pin: string): Promise<void> {
  await api.post('/me/parental/unlock', { pin })
}

export async function lockAdultContent(): Promise<void> {
  await api.post('/me/parental/lock')
}

// --- Admin endpoints ---

export async function getAdminParentalStatus(): Promise<AdminParentalStatus> {
  const { data } = await api.get<AdminParentalStatus>('/admin/parental/status')
  return data
}

export async function getRestrictedGenres(): Promise<{ codes: string[] }> {
  const { data } = await api.get<{ codes: string[] }>('/admin/parental/genres')
  return data
}

export async function updateRestrictedGenres(codes: string[]): Promise<void> {
  await api.put('/admin/parental/genres', { codes })
}

export async function setParentalPin(pin: string): Promise<void> {
  await api.post('/admin/parental/pin', { pin })
}

export async function removeParentalPin(): Promise<void> {
  await api.delete('/admin/parental/pin')
}

export async function listUsersAdultStatus(): Promise<UserAdultStatus[]> {
  const { data } = await api.get<UserAdultStatus[]>('/admin/parental/users')
  return data
}

export async function setUserAdultContent(userId: string, enabled: boolean): Promise<void> {
  await api.put(`/admin/parental/users/${userId}`, { adult_content_enabled: enabled })
}
