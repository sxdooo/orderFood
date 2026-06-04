import { View, Text, ScrollView } from '@tarojs/components'
import Taro, { useDidShow } from '@tarojs/taro'
import { useState } from 'react'
import { request } from '../../utils/request'
import { getToken } from '../../utils/auth'
import { refreshMessageBadge } from '../../utils/message'
import './index.scss'

interface Conversation {
  orderId: number
  orderNo: string
  orderStatus: string
  peerName: string
  lastContent: string
  lastAt: string
  unread: number
}

const STATUS_LABEL: Record<string, string> = {
  pending_payment: '待支付',
  pending: '待确认',
  confirmed: '已确认',
  delivering: '配送中',
  completed: '已完成',
  refunded: '已退单',
  cancelled: '已取消',
}

// Friendly relative time, e.g. 刚刚 / 14:05 / 昨天 / 06-01
function formatChatTime(iso: string): string {
  const d = new Date(iso)
  if (isNaN(d.getTime())) return ''
  const now = new Date()
  const sameDay = d.toDateString() === now.toDateString()
  if (sameDay) {
    const diff = now.getTime() - d.getTime()
    if (diff < 60 * 1000) return '刚刚'
    return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
  }
  const yesterday = new Date(now)
  yesterday.setDate(now.getDate() - 1)
  if (d.toDateString() === yesterday.toDateString()) return '昨天'
  return `${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

export default function MessagesPage() {
  const [list, setList] = useState<Conversation[]>([])
  const [loaded, setLoaded] = useState(false)

  const load = async () => {
    if (!getToken()) { setLoaded(true); return }
    try {
      const data = await request<Conversation[]>({ url: '/messages/conversations' })
      setList(data || [])
    } catch { /* ignore */ } finally {
      setLoaded(true)
    }
  }

  useDidShow(() => {
    Taro.setNavigationBarTitle({ title: '消息' })
    load()
    refreshMessageBadge()
  })

  const openChat = (c: Conversation) => {
    Taro.navigateTo({
      url: `/pages/chat/index?id=${c.orderId}&peerName=${encodeURIComponent(c.peerName || '对方')}`,
    })
  }

  return (
    <View className='messages-page'>
      <ScrollView scrollY className='conv-list'>
        {list.map((c) => (
          <View key={c.orderId} className='conv-item' onClick={() => openChat(c)}>
            <View className='avatar'>
              <Text className='avatar-text'>{(c.peerName || '?').slice(0, 1)}</Text>
              {c.unread > 0 && (
                <View className='unread-dot'>
                  <Text className='unread-num'>{c.unread > 99 ? '99+' : c.unread}</Text>
                </View>
              )}
            </View>
            <View className='conv-main'>
              <View className='conv-row1'>
                <Text className='conv-name'>{c.peerName || '对方'}</Text>
                <Text className='conv-time'>{formatChatTime(c.lastAt)}</Text>
              </View>
              <View className='conv-row2'>
                <Text className='conv-last'>{c.lastContent || '（暂无消息）'}</Text>
              </View>
              <View className='conv-row3'>
                <Text className='conv-order'>订单 {c.orderNo}</Text>
                <Text className='conv-status'>{STATUS_LABEL[c.orderStatus] || c.orderStatus}</Text>
              </View>
            </View>
          </View>
        ))}

        {loaded && list.length === 0 && (
          <View className='empty'>
            <Text className='empty-icon'>💬</Text>
            <Text className='empty-text'>暂无消息</Text>
            <Text className='empty-sub'>下单或处理订单后即可在这里沟通</Text>
          </View>
        )}
      </ScrollView>
    </View>
  )
}
