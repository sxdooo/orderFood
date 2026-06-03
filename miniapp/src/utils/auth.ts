import Taro from '@tarojs/taro'

const TOKEN_KEY = 'auth_token'
const USER_KEY = 'auth_user'

export interface AuthUser {
  id: number
  role: 'buyer' | 'seller'
}

export function getToken(): string {
  return Taro.getStorageSync(TOKEN_KEY) || ''
}

export function setToken(token: string): void {
  Taro.setStorageSync(TOKEN_KEY, token)
}

export function clearAuth(): void {
  Taro.removeStorageSync(TOKEN_KEY)
  Taro.removeStorageSync(USER_KEY)
}

export function setUser(user: AuthUser): void {
  Taro.setStorageSync(USER_KEY, user)
}

export function getUser(): AuthUser | null {
  return Taro.getStorageSync(USER_KEY) || null
}
