import api from './api'

export interface UserInfo {
  id: string
  email: string
  username: string
  display_name: string
  role: string
}

export interface AuthResponse {
  user: UserInfo
  access_token: string
}

export interface RegisterInput {
  email: string
  username: string
  display_name: string
  password: string
}

export interface LoginInput {
  email: string
  password: string
}

export async function register(input: RegisterInput): Promise<AuthResponse> {
  const { data } = await api.post<AuthResponse>('/auth/register', input)
  return data
}

export async function login(input: LoginInput): Promise<AuthResponse> {
  const { data } = await api.post<AuthResponse>('/auth/login', input)
  return data
}

export async function refresh(): Promise<AuthResponse> {
  const { data } = await api.post<AuthResponse>('/auth/refresh')
  return data
}

export async function logout(): Promise<void> {
  await api.post('/auth/logout')
}
