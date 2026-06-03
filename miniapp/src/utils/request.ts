import Taro from '@tarojs/taro'
import { API_BASE_URL } from '../config/api'
import { clearAuth, getToken } from './auth'

const BASE_URL = API_BASE_URL

interface ApiBody<T> {
  code: number
  message: string
  data: T
}

let redirectingToLogin = false

export async function request<T>(options: {
  url: string
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE'
  data?: Record<string, unknown>
}): Promise<T> {
  const token = getToken()
  try {
    const res = await Taro.request<ApiBody<T>>({
      url: `${BASE_URL}${options.url}`,
      method: options.method || 'GET',
      data: options.data,
      timeout: 15000,
      header: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {})
      }
    })

    if (res.statusCode === 401) {
      clearAuth()
      if (!redirectingToLogin) {
        redirectingToLogin = true
        Taro.reLaunch({ url: '/pages/login/index' }).finally(() => {
          redirectingToLogin = false
        })
      }
      throw new Error('unauthorized')
    }

    const body = res.data
    if (!body || body.code !== 0) {
      throw new Error(body?.message || 'request failed')
    }
    return body.data
  } catch (err) {
    // Taro throws plain objects like { errMsg: 'request:fail ...' } on network errors.
    // Normalize to Error so callers can always use err.message.
    const errMsg: string =
      err instanceof Error
        ? err.message
        : typeof (err as { errMsg?: string }).errMsg === 'string'
        ? (err as { errMsg: string }).errMsg
        : JSON.stringify(err)

    if (errMsg.includes('timeout')) {
      throw new Error('网络超时，请检查 API 是否启动')
    }
    if (errMsg.includes('request:fail')) {
      throw new Error(`连接失败：${errMsg}`)
    }
    if (err instanceof Error) throw err
    throw new Error(errMsg)
  }
}
