import { View, Map, Button, Input, Text } from '@tarojs/components'
import Taro, { useLoad } from '@tarojs/taro'
import { useState, useRef } from 'react'
import { request } from '../../../utils/request'
import { formatPrice, ORDER_STATUS_LABEL } from '../../../utils/format'
import './index.scss'

interface SellerDetail {
  order: {
    id: number
    status: string
    contactName: string
    contactPhone: string
    address: string
    totalAmount: number
    items: { nameSnapshot: string; quantity: number }[]
  }
  distanceKm?: number
  sellerLat: number
  sellerLng: number
  buyerLat?: number
  buyerLng?: number
}

export default function SellerOrderDetailPage() {
  const [detail, setDetail] = useState<SellerDetail | null>(null)
  const [reason, setReason] = useState('')
  const orderId = useRef(0)

  useLoad((opts) => {
    orderId.current = Number(opts.id)
    Taro.setNavigationBarTitle({ title: '订单详情' })
    request<SellerDetail>({ url: `/seller/orders/${orderId.current}` }).then(setDetail)
  })

  const updateStatus = async (status: string) => {
    await request({ url: `/seller/orders/${orderId.current}/status`, method: 'PUT', data: { status } })
    Taro.showToast({ title: '已更新', icon: 'success' })
    const d = await request<SellerDetail>({ url: `/seller/orders/${orderId.current}` })
    setDetail(d)
  }

  const refund = async () => {
    if (!reason) {
      Taro.showToast({ title: '请填写退单原因', icon: 'none' })
      return
    }
    await request({ url: `/seller/orders/${orderId.current}/refund`, method: 'POST', data: { reason } })
    Taro.showToast({ title: '已退单', icon: 'success' })
    const d = await request<SellerDetail>({ url: `/seller/orders/${orderId.current}` })
    setDetail(d)
  }

  if (!detail) return <View className='container'>加载中...</View>

  const markers = []
  if (detail.sellerLat && detail.sellerLng) {
    markers.push({ id: 1, latitude: detail.sellerLat, longitude: detail.sellerLng, title: '店铺' })
  }
  if (detail.buyerLat && detail.buyerLng) {
    markers.push({ id: 2, latitude: detail.buyerLat, longitude: detail.buyerLng, title: '买家' })
  }
  const center = markers[0] || { latitude: 31.2, longitude: 121.5 }

  return (
    <View className='container'>
      <View className='card'>
        <View className='status'>{ORDER_STATUS_LABEL[detail.order.status]}</View>
        <View>{detail.order.contactName} {detail.order.contactPhone}</View>
        <View>{detail.order.address}</View>
        {detail.distanceKm != null && <View>距离约 {detail.distanceKm.toFixed(1)} km</View>}
        {detail.order.items?.map((it, i) => <View key={i}>{it.nameSnapshot} x{it.quantity}</View>)}
        <View className='amount'>¥{formatPrice(detail.order.totalAmount)}</View>
      </View>
      {markers.length > 0 && (
        <Map
          className='map'
          latitude={center.latitude}
          longitude={center.longitude}
          markers={markers}
          scale={14}
        />
      )}
      <View className='actions'>
        <Button size='mini' onClick={() => updateStatus('confirmed')}>确认</Button>
        <Button size='mini' onClick={() => updateStatus('delivering')}>配送中</Button>
        <Button size='mini' onClick={() => updateStatus('completed')}>完成</Button>
        <Button size='mini' onClick={() => Taro.navigateTo({ url: `/pages/order-detail/index?id=${orderId.current}` })}>对话</Button>
      </View>
      <View className='card'>
        <Input placeholder='退单原因（必填）' value={reason} onInput={(e) => setReason(e.detail.value)} />
        <Button className='refund-btn' onClick={refund}>退单</Button>
      </View>
    </View>
  )
}
