import Taro from '@tarojs/taro'
import { request } from './request'
import { getToken } from './auth'

// Index of the 消息 tab in app.config.ts tabBar list.
export const MESSAGE_TAB_INDEX = 2

// Fetch the unread message count and reflect it on the message tab badge.
// Safe to call from any tab page's useDidShow; silently no-ops when logged out.
export async function refreshMessageBadge(): Promise<number> {
  if (!getToken()) return 0
  try {
    const res = await request<{ count: number }>({ url: '/messages/unread-count' })
    const count = res?.count ?? 0
    if (count > 0) {
      Taro.setTabBarBadge({ index: MESSAGE_TAB_INDEX, text: count > 99 ? '99+' : String(count) })
    } else {
      Taro.removeTabBarBadge({ index: MESSAGE_TAB_INDEX }).catch(() => {})
    }
    return count
  } catch {
    return 0
  }
}
