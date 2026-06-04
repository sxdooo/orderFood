import { View, Text, Input, Button, ScrollView } from '@tarojs/components'
import Taro, { useLoad, useUnload } from '@tarojs/taro'
import { useState, useRef } from 'react'
import { request } from '../../utils/request'
import { getUser } from '../../utils/auth'
import { refreshMessageBadge } from '../../utils/message'
import './index.scss'

interface Message {
  id: number
  senderRole: 'buyer' | 'seller'
  content: string
  createdAt: string
}

function formatTime(iso: string): string {
  const d = new Date(iso)
  if (isNaN(d.getTime())) return ''
  const hh = String(d.getHours()).padStart(2, '0')
  const mm = String(d.getMinutes()).padStart(2, '0')
  return `${hh}:${mm}`
}

export default function ChatPage() {
  const [messages, setMessages] = useState<Message[]>([])
  const [text, setText] = useState('')
  const [sending, setSending] = useState(false)
  const orderId = useRef(0)
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const lastSince = useRef(0)
  // Drives ScrollView auto-scroll to the latest bubble
  const [scrollTarget, setScrollTarget] = useState('')

  const myRole = getUser()?.role ?? 'buyer'

  const scrollToBottom = (list: Message[]) => {
    if (list.length > 0) setScrollTarget(`msg-${list[list.length - 1].id}`)
  }

  const loadMessages = async (id: number) => {
    const since = lastSince.current
    const data = await request<Message[]>({ url: `/orders/${id}/messages?since=${since}` })
    if (data?.length) {
      setMessages((prev) => {
        const merged = [...prev, ...data]
        scrollToBottom(merged)
        return merged
      })
      lastSince.current = new Date(data[data.length - 1].createdAt).getTime()
    }
  }

  useLoad((opts) => {
    const id = Number(opts.id)
    orderId.current = id
    const peer = opts.peerName ? decodeURIComponent(opts.peerName as string) : ''
    Taro.setNavigationBarTitle({ title: peer ? `与${peer}对话` : '订单对话' })
    loadMessages(id).catch(() => {})
    pollRef.current = setInterval(() => {
      loadMessages(id).catch(() => {})
    }, 5000)
  })

  useUnload(() => {
    if (pollRef.current) {
      clearInterval(pollRef.current)
      pollRef.current = null
    }
    // Messages were marked read while open; sync the tab badge on the way out.
    refreshMessageBadge()
  })

  const send = async () => {
    const content = text.trim()
    if (!content || sending) return
    setSending(true)
    try {
      await request({
        url: `/orders/${orderId.current}/messages`,
        method: 'POST',
        data: { type: 'text', content },
      })
      setText('')
      await loadMessages(orderId.current)
    } catch (e: any) {
      Taro.showToast({ title: e?.message || '发送失败', icon: 'none' })
    } finally {
      setSending(false)
    }
  }

  return (
    <View className='chat-page'>
      <ScrollView
        scrollY
        className='msg-list'
        scrollIntoView={scrollTarget}
        scrollWithAnimation
      >
        <View className='msg-inner'>
          {messages.length === 0 && (
            <View className='empty-hint'>还没有消息，发条消息开始沟通吧</View>
          )}
          {messages.map((m) => {
            const mine = m.senderRole === myRole
            return (
              <View
                id={`msg-${m.id}`}
                key={m.id}
                className={`msg-row ${mine ? 'mine' : 'peer'}`}
              >
                <View className='bubble'>
                  <Text className='bubble-text'>{m.content}</Text>
                </View>
                <Text className='msg-time'>{formatTime(m.createdAt)}</Text>
              </View>
            )
          })}
        </View>
      </ScrollView>

      <View className='input-bar'>
        <Input
          className='input-field'
          value={text}
          confirmType='send'
          placeholder='输入消息'
          onInput={(e) => setText(e.detail.value)}
          onConfirm={send}
        />
        <Button
          className={`send-btn ${text.trim() ? 'active' : ''}`}
          size='mini'
          loading={sending}
          onClick={send}
        >
          发送
        </Button>
      </View>
    </View>
  )
}
