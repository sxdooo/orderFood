import { View, Map, Text } from '@tarojs/components'
import Taro, { useLoad } from '@tarojs/taro'
import { useState } from 'react'
import { request } from '../../../utils/request'
import './index.scss'

interface RouteStop {
  orderId: number
  contactName: string
  address: string
  lat: number
  lng: number
}

export default function SellerRoutePage() {
  const [stops, setStops] = useState<RouteStop[]>([])
  const [date, setDate] = useState('')

  useLoad(async (opts) => {
    const d = opts.deliveryDate as string
    setDate(d)
    Taro.setNavigationBarTitle({ title: '配送路线' })
    const data = await request<{ stops: RouteStop[] }>({ url: `/seller/routes?deliveryDate=${d}` })
    setStops(data.stops || [])
  })

  const center = stops[0] || { lat: 31.2, lng: 121.5 }
  const markers = stops.map((s, i) => ({
    id: i + 1,
    latitude: s.lat,
    longitude: s.lng,
    title: `${i + 1}.${s.contactName}`
  }))
  const points = stops.map((s) => ({ latitude: s.lat, longitude: s.lng }))

  return (
    <View className='container'>
      <Map
        className='map'
        latitude={center.lat}
        longitude={center.lng}
        markers={markers}
        polyline={points.length > 1 ? [{ points, color: '#07c160', width: 4 }] : []}
        scale={12}
      />
      {stops.map((s, i) => (
        <View key={s.orderId} className='card stop'>
          <Text>{i + 1}. {s.contactName}</Text>
          <Text className='addr'>{s.address}</Text>
        </View>
      ))}
      {stops.length === 0 && <View className='card'>暂无配送点（订单需有坐标）</View>}
    </View>
  )
}
