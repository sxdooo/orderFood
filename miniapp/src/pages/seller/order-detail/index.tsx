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
    items: { nameSnapshot: string; quantity: number; priceSnapshot: number }[]
  }
  distanceKm?: number
  sellerLat: number
  sellerLng: number
  buyerLat?: number
  buyerLng?: number
}

const STATUS_CLASS: Record<string, string> = {
  pending: 's-pending', confirmed: 's-confirmed', delivering: 's-delivering',
  completed: 's-completed', refunded: 's-refunded', cancelled: 's-cancelled',
}

const NEXT_STATUS: Record<string, { label: string; value: string }[]> = {
  pending:    [{ label: '确认订单', value: 'confirmed' }],
  confirmed:  [{ label: '开始配送', value: 'delivering' }],
  delivering: [{ label: '标记完成', value: 'completed' }],
}

export default function SellerOrderDetailPage() {
  const [detail, setDetail] = useState<SellerDetail | null>(null)
  const [reason, setReason] = useState('')
  const [showRefund, setShowRefund] = useState(false)
  const orderId = useRef(0)

  const reload = () =>
    request<SellerDetail>({ url: `/seller/orders/${orderId.current}` }).then(setDetail)

  useLoad((opts) => {
    orderId.current = Number(opts.id)
    Taro.setNavigationBarTitle({ title: '订单详情' })
    reload()
  })

  const updateStatus = async (status: string) => {
    try {
      await request({ url: `/seller/orders/${orderId.current}/status`, method: 'PUT', data: { status } })
      Taro.showToast({ title: '状态已更新', icon: 'success' })
      reload()
    } catch (err) {
      Taro.showToast({ title: err instanceof Error ? err.message : '操作失败', icon: 'none' })
    }
  }

  const refund = async () => {
    if (!reason.trim()) {
      Taro.showToast({ title: '请填写退单原因', icon: 'none' })
      return
    }
    try {
      await request({ url: `/seller/orders/${orderId.current}/refund`, method: 'POST', data: { reason } })
      Taro.showToast({ title: '已退单', icon: 'success' })
      setShowRefund(false)
      reload()
    } catch (err) {
      Taro.showToast({ title: err instanceof Error ? err.message : '退单失败', icon: 'none' })
    }
  }

  // Open WeChat built-in navigation (user can choose Amap / Tencent / Baidu)
  const openNavigation = () => {
    if (!detail?.buyerLat || !detail?.buyerLng) {
      Taro.showToast({ title: '买家地址暂无坐标', icon: 'none' })
      return
    }
    Taro.openLocation({
      latitude: detail.buyerLat,
      longitude: detail.buyerLng,
      name: detail.order.contactName,
      address: detail.order.address,
      scale: 16,
    })
  }

  if (!detail) return <View className='container'><Text className='loading-text'>加载中...</Text></View>

  const { order } = detail
  const markers = []
  if (detail.sellerLat && detail.sellerLng) {
    markers.push({
      id: 1,
      latitude: detail.sellerLat,
      longitude: detail.sellerLng,
      title: '店铺',
      iconPath: '',
      label: { content: '店铺', color: '#07c160', fontSize: 12, borderWidth: 1, borderColor: '#fff', bgColor: '#fff' },
    })
  }
  if (detail.buyerLat && detail.buyerLng) {
    markers.push({
      id: 2,
      latitude: detail.buyerLat,
      longitude: detail.buyerLng,
      title: '买家',
      label: { content: '买家', color: '#ff6b35', fontSize: 12, borderWidth: 1, borderColor: '#fff', bgColor: '#fff' },
    })
  }
  const center = markers.length > 0
    ? {
        latitude:  (markers.reduce((s, m) => s + m.latitude, 0))  / markers.length,
        longitude: (markers.reduce((s, m) => s + m.longitude, 0)) / markers.length,
      }
    : { latitude: 39.4387, longitude: 116.2987 }

  const nextActions = NEXT_STATUS[order.status] || []

  return (
    <View className='container detail-page'>

      {/* ── 买家信息 ──────────────────────────────────── */}
      <View className='card buyer-card'>
        <View className='buyer-head'>
          <View>
            <Text className='buyer-name'>{order.contactName}</Text>
            <Text className='buyer-phone'>{order.contactPhone}</Text>
          </View>
          <Text className={`status-badge ${STATUS_CLASS[order.status] || ''}`}>
            {ORDER_STATUS_LABEL[order.status] || order.status}
          </Text>
        </View>
        <Text className='buyer-addr'>{order.address}</Text>
        {detail.distanceKm != null && (
          <Text className='buyer-distance'>距离约 {detail.distanceKm.toFixed(1)} km</Text>
        )}
      </View>

      {/* ── 地图 + 导航 ───────────────────────────────── */}
      {markers.length > 0 && (
        <View className='map-section'>
          <Map
            className='map'
            latitude={center.latitude}
            longitude={center.longitude}
            markers={markers}
            scale={14}
            showLocation
          />
          {detail.buyerLat && detail.buyerLng && (
            <Button className='nav-btn' onClick={openNavigation}>
              导航到买家
            </Button>
          )}
        </View>
      )}

      {/* ── 菜品明细 ──────────────────────────────────── */}
      <View className='card'>
        <Text className='section-title'>订单菜品</Text>
        {order.items?.map((it, i) => (
          <View key={i} className='item-row'>
            <Text className='item-name'>{it.nameSnapshot}</Text>
            <Text className='item-qty'>×{it.quantity}</Text>
            <Text className='item-price'>¥{formatPrice(it.priceSnapshot * it.quantity)}</Text>
          </View>
        ))}
        <View className='total-row'>
          <Text className='total-label'>合计</Text>
          <Text className='total-amount'>¥{formatPrice(order.totalAmount)}</Text>
        </View>
      </View>

      {/* ── 操作按钮 ──────────────────────────────────── */}
      <View className='action-bar'>
        {nextActions.map((a) => (
          <Button key={a.value} className='btn-primary action-btn' onClick={() => updateStatus(a.value)}>
            {a.label}
          </Button>
        ))}
        <Button
          className='action-btn-chat'
          onClick={() => Taro.navigateTo({
            url: `/pages/chat/index?id=${orderId.current}&peerName=${encodeURIComponent(order.contactName || '买家')}`,
          })}
        >
          与买家对话
        </Button>
        {order.status !== 'refunded' && order.status !== 'cancelled' && (
          <Button className='action-btn-refund' onClick={() => setShowRefund(!showRefund)}>
            {showRefund ? '取消退单' : '退单'}
          </Button>
        )}
      </View>

      {/* ── 退单面板 ──────────────────────────────────── */}
      {showRefund && (
        <View className='card refund-panel'>
          <Text className='section-title'>退单原因</Text>
          <Input
            className='refund-input'
            placeholder='请输入退单原因（必填）'
            value={reason}
            onInput={(e) => setReason(e.detail.value)}
          />
          <Button className='btn-danger' onClick={refund}>确认退单</Button>
        </View>
      )}
    </View>
  )
}
