import { View, Text, Button, Input, ScrollView } from '@tarojs/components'
import Taro, { useLoad, useUnload } from '@tarojs/taro'
import { useState, useRef } from 'react'
import { request } from '../../utils/request'
import { formatPrice, ORDER_STATUS_LABEL } from '../../utils/format'
import { getUser } from '../../utils/auth'
import './index.scss'

interface OrderItem {
  nameSnapshot: string
  quantity: number
  priceSnapshot: number
}

interface Order {
  id: number
  orderNo: string
  status: string
  contactName: string
  contactPhone: string
  address: string
  totalAmount: number
  items: OrderItem[]
}

interface Message {
  id: number
  senderRole: string
  content: string
  createdAt: string
}

export default function OrderDetailPage() {
  const [order, setOrder] = useState<Order | null>(null)
  const [messages, setMessages] = useState<Message[]>([])
  const [text, setText] = useState('')
  const orderId = useRef(0)
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const lastSince = useRef(0)

  const loadOrder = async (id: number) => {
    const data = await request<Order>({ url: `/orders/${id}` })
    setOrder(data)
  }

  const loadMessages = async (id: number) => {
    const since = lastSince.current
    const data = await request<Message[]>({ url: `/orders/${id}/messages?since=${since}` })
    if (data?.length) {
      setMessages((prev) => [...prev, ...data])
      lastSince.current = new Date(data[data.length - 1].createdAt).getTime()
    }
  }

  useLoad((opts) => {
    const id = Number(opts.id)
    orderId.current = id
    Taro.setNavigationBarTitle({ title: '订单详情' })
    loadOrder(id).catch(() => {})
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
  })

  const send = async () => {
    if (!text.trim()) return
    await request({
      url: `/orders/${orderId.current}/messages`,
      method: 'POST',
      data: { type: 'text', content: text }
    })
    setText('')
    await loadMessages(orderId.current)
  }

  const cancelOrder = async () => {
    await request({ url: `/orders/${orderId.current}/cancel`, method: 'POST' })
    Taro.showToast({ title: '已取消', icon: 'success' })
    loadOrder(orderId.current)
  }

  if (!order) return <View className='container'>加载中...</View>

  const me = getUser()

  return (
    <View className='container detail-page'>
      <View className='card'>
        <View className='status'>{ORDER_STATUS_LABEL[order.status] || order.status}</View>
        <View>单号 {order.orderNo}</View>
        <View>{order.contactName} {order.contactPhone}</View>
        <View>{order.address}</View>
        {order.items?.map((it, i) => (
          <View key={i}>{it.nameSnapshot} x{it.quantity}</View>
        ))}
        <View className='amount'>¥{formatPrice(order.totalAmount)}</View>
        {order.status === 'pending' && me?.role === 'buyer' && (
          <Button size='mini' onClick={cancelOrder}>取消订单</Button>
        )}
      </View>
      <View className='card chat'>
        <View className='section-title'>订单对话</View>
        <ScrollView scrollY className='msg-list'>
          {messages.map((m) => (
            <View key={m.id} className={`msg msg-${m.senderRole}`}>
              <Text>{m.content}</Text>
            </View>
          ))}
        </ScrollView>
        <View className='msg-input'>
          <Input value={text} onInput={(e) => setText(e.detail.value)} placeholder='输入消息' />
          <Button size='mini' onClick={send}>发送</Button>
        </View>
      </View>
    </View>
  )
}
