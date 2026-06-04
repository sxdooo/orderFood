import { View, Text } from '@tarojs/components'
import Taro, { useDidShow } from '@tarojs/taro'
import { useState } from 'react'
import { request } from '../../utils/request'
import { getUser } from '../../utils/auth'
import { refreshMessageBadge } from '../../utils/message'
import { formatPrice, ORDER_STATUS_LABEL } from '../../utils/format'
import './index.scss'

interface Order {
  id: number
  orderNo: string
  deliveryDate: string
  totalAmount: number
  status: string
  contactName?: string
}

const STATUS_CLASS: Record<string, string> = {
  pending: 's-pending',
  confirmed: 's-confirmed',
  delivering: 's-delivering',
  completed: 's-completed',
  cancelled: 's-cancelled',
  refunded: 's-refunded',
}

export default function OrdersPage() {
  const [orders, setOrders] = useState<Order[]>([])
  const [isSeller, setIsSeller] = useState(false)

  useDidShow(async () => {
    const user = getUser()
    const seller = user?.role === 'seller'
    setIsSeller(seller)
    Taro.setNavigationBarTitle({ title: seller ? '订单管理' : '我的订单' })
    try {
      if (seller) {
        // Pre-order model: buyers order for next-day delivery, so the quick
        // view defaults to tomorrow's orders (matches the full管理 page).
        const d = new Date()
        d.setDate(d.getDate() + 1)
        const target = d.toISOString().slice(0, 10)
        const data = await request<Order[]>({ url: `/seller/orders?date=${target}` })
        setOrders(data || [])
      } else {
        const data = await request<Order[]>({ url: '/orders' })
        setOrders(data || [])
      }
    } catch { /* ignore */ }
    refreshMessageBadge()
  })

  const openDetail = (id: number) => {
    const url = isSeller
      ? `/pages/seller/order-detail/index?id=${id}`
      : `/pages/order-detail/index?id=${id}`
    Taro.navigateTo({ url })
  }

  return (
    <View className='container orders-page'>
      {isSeller && (
        <View className='seller-tip'>
          <Text>明日订单（点击进入详情）</Text>
          <Text
            className='seller-tip-link'
            onClick={() => Taro.navigateTo({ url: '/pages/seller/orders/index' })}
          >
            完整管理 ›
          </Text>
        </View>
      )}
      {orders.length === 0 ? (
        <View className='empty-tip'>{isSeller ? '明日暂无订单' : '暂无订单，快去订餐吧'}</View>
      ) : (
        orders.map((o) => (
          <View key={o.id} className='card order-card' onClick={() => openDetail(o.id)}>
            <View className='order-head'>
              <Text className='order-date'>{String(o.deliveryDate).slice(0, 10)} 配送</Text>
              <Text className={`status-badge ${STATUS_CLASS[o.status] || ''}`}>
                {ORDER_STATUS_LABEL[o.status] || o.status}
              </Text>
            </View>
            {isSeller && o.contactName && (
              <Text className='order-buyer'>{o.contactName}</Text>
            )}
            <View className='order-foot'>
              <Text className='order-no'>单号 {o.orderNo}</Text>
              <Text className='order-amount'>¥{formatPrice(o.totalAmount)}</Text>
            </View>
          </View>
        ))
      )}
    </View>
  )
}
