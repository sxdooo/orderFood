import { View, Input, Button } from '@tarojs/components'
import Taro, { useDidShow } from '@tarojs/taro'
import { useState } from 'react'
import { request } from '../../../utils/request'
import { formatPrice, ORDER_STATUS_LABEL } from '../../../utils/format'
import './index.scss'

interface Order {
  id: number
  orderNo: string
  contactName: string
  address: string
  status: string
  totalAmount: number
}

export default function SellerOrdersPage() {
  const [date, setDate] = useState('')
  const [orders, setOrders] = useState<Order[]>([])
  const [cutoffTime, setCutoffTime] = useState('17:00')

  useDidShow(() => {
    Taro.setNavigationBarTitle({ title: '卖家订单' })
    const tomorrow = new Date()
    tomorrow.setDate(tomorrow.getDate() + 1)
    const ds = tomorrow.toISOString().slice(0, 10)
    setDate(ds)
    load(ds)
  })

  const load = async (d: string) => {
    const data = await request<Order[]>({ url: `/seller/orders?deliveryDate=${d}` })
    setOrders(data || [])
  }

  const setCutoff = async () => {
    await request({ url: '/seller/cutoff', method: 'PUT', data: { cutoffTime } })
    Taro.showToast({ title: '截单时间已设置', icon: 'success' })
  }

  const genRoute = async () => {
    await request({ url: '/seller/routes', method: 'POST', data: { deliveryDate: date } })
    Taro.navigateTo({ url: `/pages/seller/route/index?deliveryDate=${date}` })
  }

  return (
    <View className='container'>
      <View className='card'>
        <Input value={date} onInput={(e) => setDate(e.detail.value)} placeholder='配送日期' />
        <Button size='mini' onClick={() => load(date)}>查询</Button>
        <View className='row'>
          <Input value={cutoffTime} onInput={(e) => setCutoffTime(e.detail.value)} placeholder='截单时间 HH:MM' />
          <Button size='mini' onClick={setCutoff}>设截单</Button>
        </View>
        <Button className='btn-primary' onClick={genRoute}>生成配送路线</Button>
      </View>
      {orders.map((o) => (
        <View key={o.id} className='card' onClick={() => Taro.navigateTo({ url: `/pages/seller/order-detail/index?id=${o.id}` })}>
          <View>{o.contactName} · {ORDER_STATUS_LABEL[o.status]}</View>
          <View className='addr'>{o.address}</View>
          <View>¥{formatPrice(o.totalAmount)}</View>
        </View>
      ))}
    </View>
  )
}
