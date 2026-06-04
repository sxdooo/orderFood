import { View, Text, Button, Picker, ScrollView } from '@tarojs/components'
import Taro, { useDidShow } from '@tarojs/taro'
import { useState } from 'react'
import { request } from '../../../utils/request'
import { formatPrice, ORDER_STATUS_LABEL } from '../../../utils/format'
import './index.scss'

interface Order {
  id: number
  orderNo: string
  contactName: string
  contactPhone: string
  address: string
  status: string
  totalAmount: number
}

interface Summary {
  total: number
  totalAmount: number
  byStatus: Record<string, number>
}

const STATUS_FILTERS = [
  { key: 'all',        label: '全部' },
  { key: 'pending',    label: '待确认' },
  { key: 'confirmed',  label: '已确认' },
  { key: 'delivering', label: '配送中' },
  { key: 'completed',  label: '已完成' },
  { key: 'refunded',   label: '已退单' },
  { key: 'cancelled',  label: '已取消' },
]

const STATUS_CLASS: Record<string, string> = {
  pending:    's-pending',
  confirmed:  's-confirmed',
  delivering: 's-delivering',
  completed:  's-completed',
  refunded:   's-refunded',
  cancelled:  's-cancelled',
}

const getTomorrow = () => {
  const d = new Date()
  d.setDate(d.getDate() + 1)
  return d.toISOString().slice(0, 10)
}

function buildSummary(orders: Order[]): Summary {
  const byStatus: Record<string, number> = {}
  let totalAmount = 0
  for (const o of orders) {
    byStatus[o.status] = (byStatus[o.status] || 0) + 1
    totalAmount += o.totalAmount
  }
  return { total: orders.length, totalAmount, byStatus }
}

export default function SellerOrdersPage() {
  const [date, setDate] = useState(getTomorrow())
  const [orders, setOrders] = useState<Order[]>([])
  const [cutoffTime, setCutoffTimeState] = useState('17:00')
  const [filterStatus, setFilterStatus] = useState('all')
  const [summary, setSummary] = useState<Summary>({ total: 0, totalAmount: 0, byStatus: {} })

  const load = async (d: string) => {
    try {
      const data = await request<Order[]>({ url: `/seller/orders?deliveryDate=${d}` })
      const list = data || []
      setOrders(list)
      setSummary(buildSummary(list))
    } catch (err) {
      Taro.showToast({ title: err instanceof Error ? err.message : '加载失败', icon: 'none' })
    }
  }

  useDidShow(() => {
    Taro.setNavigationBarTitle({ title: '订单管理' })
    load(date)
  })

  const handleSetCutoff = async () => {
    try {
      await request({ url: '/seller/cutoff', method: 'PUT', data: { cutoffTime } })
      Taro.showToast({ title: `截单时间已设为 ${cutoffTime}`, icon: 'success' })
    } catch (err) {
      Taro.showToast({ title: err instanceof Error ? err.message : '设置失败', icon: 'none' })
    }
  }

  const filtered = filterStatus === 'all'
    ? orders
    : orders.filter((o) => o.status === filterStatus)

  return (
    <View className='container seller-orders-page'>

      {/* ── 工具栏 ─────────────────────────────────────── */}
      <View className='card toolbar-card'>
        <View className='toolbar-row'>
          <View className='toolbar-item'>
            <Text className='toolbar-label'>配送日期</Text>
            <Picker
              mode='date'
              value={date}
              onChange={(e) => {
                const d = String(e.detail.value)
                setDate(d)
                load(d)
              }}
            >
              <View className='picker-val'>
                <Text>{date}</Text>
                <Text className='picker-arrow'>›</Text>
              </View>
            </Picker>
          </View>
          <Button className='btn-outline route-btn' onClick={() => Taro.navigateTo({ url: `/pages/seller/route/index?deliveryDate=${date}` })}>
            路线规划
          </Button>
        </View>

        <View className='toolbar-row cutoff-row'>
          <Text className='toolbar-label'>今日截单</Text>
          <View className='cutoff-time-input'>
            <Picker
              mode='time'
              value={cutoffTime}
              onChange={(e) => setCutoffTimeState(String(e.detail.value))}
            >
              <View className='picker-val small'>
                <Text>{cutoffTime}</Text>
                <Text className='picker-arrow'>›</Text>
              </View>
            </Picker>
          </View>
          <Button className='btn-primary cutoff-btn' onClick={handleSetCutoff}>设置</Button>
        </View>
      </View>

      {/* ── 汇总卡片 ───────────────────────────────────── */}
      {summary.total > 0 && (
        <View className='card summary-card'>
          <View className='summary-top'>
            <View className='summary-main'>
              <Text className='summary-count'>{summary.total}</Text>
              <Text className='summary-count-label'>笔订单</Text>
            </View>
            <View className='summary-amount'>
              <Text className='summary-amount-label'>总金额</Text>
              <Text className='summary-amount-val'>¥{formatPrice(summary.totalAmount)}</Text>
            </View>
          </View>
          <View className='summary-badges'>
            {Object.entries(summary.byStatus).map(([s, cnt]) => (
              <View key={s} className={`summary-badge ${STATUS_CLASS[s] || ''}`} onClick={() => setFilterStatus(s)}>
                <Text>{ORDER_STATUS_LABEL[s] || s}</Text>
                <Text className='summary-badge-cnt'>{cnt}</Text>
              </View>
            ))}
          </View>
        </View>
      )}

      {/* ── 状态筛选 Tab ───────────────────────────────── */}
      <ScrollView scrollX className='filter-tabs' enableFlex>
        {STATUS_FILTERS.map((f) => {
          const cnt = f.key === 'all' ? summary.total : (summary.byStatus[f.key] || 0)
          return (
            <View
              key={f.key}
              className={`filter-tab ${filterStatus === f.key ? 'active' : ''}`}
              onClick={() => setFilterStatus(f.key)}
            >
              <Text>{f.label}</Text>
              {cnt > 0 && <Text className='filter-tab-cnt'>{cnt}</Text>}
            </View>
          )
        })}
      </ScrollView>

      {/* ── 订单列表 ───────────────────────────────────── */}
      {filtered.length === 0 ? (
        <View className='empty-tip'>
          {orders.length === 0 ? '该日期暂无订单' : '该状态暂无订单'}
        </View>
      ) : (
        filtered.map((o) => (
          <View
            key={o.id}
            className='card order-card'
            onClick={() => Taro.navigateTo({ url: `/pages/seller/order-detail/index?id=${o.id}` })}
          >
            <View className='order-card-head'>
              <Text className='order-buyer'>{o.contactName}</Text>
              <Text className={`status-badge ${STATUS_CLASS[o.status] || ''}`}>
                {ORDER_STATUS_LABEL[o.status] || o.status}
              </Text>
            </View>
            <Text className='order-addr'>{o.address}</Text>
            <View className='order-card-foot'>
              <Text className='order-phone'>{o.contactPhone}</Text>
              <Text className='order-amount'>¥{formatPrice(o.totalAmount)}</Text>
            </View>
          </View>
        ))
      )}
    </View>
  )
}
