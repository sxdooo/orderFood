import { View, Map, Text, Input, Button, ScrollView } from '@tarojs/components'
import Taro, { useLoad } from '@tarojs/taro'
import { useState, useEffect, useCallback, useRef } from 'react'
import { request } from '../../../utils/request'
import './index.scss'

const SLICE = 16

interface RouteStop {
  orderId: number
  contactName: string
  address: string
  phone: string
  lat: number
  lng: number
}

interface DriverRoute {
  driverIndex: number
  color: string
  stops: RouteStop[]
  totalDistance: number
}

interface ClusterResult {
  deliveryDate: string
  driverCount: number
  sellerLat: number
  sellerLng: number
  drivers: DriverRoute[]
}

interface Directions {
  points: { lat: number; lng: number }[]
  distance: number
  duration: number
}

const SHOP_MARKER_ID = 999999

// Keep in sync with server/internal/service/order.go cluster color list
const DRIVER_COLORS = ['#07c160', '#1677ff', '#ff6b35', '#9254de', '#eb2f96', '#faad14']

// Open WeChat navigation to a single destination (current GPS → dest).
function navigateTo(stop: RouteStop) {
  if (!stop.lat || !stop.lng) {
    Taro.showToast({ title: '该配送点无坐标', icon: 'none' })
    return
  }
  Taro.openLocation({
    latitude: stop.lat, longitude: stop.lng,
    name: stop.contactName, address: stop.address, scale: 17,
  })
}

// Build an Amap multi-stop navigation deep link.
//
// IMPORTANT: the public web endpoint https://uri.amap.com/navigation supports
// only ONE waypoint (param `position`, NOT `via`). For multiple waypoints we
// MUST use the Amap app Schema `amapuri://route/plan/` with these params:
//   slat/slon/sname  → start
//   dlat/dlon/dname  → destination
//   vian             → number of waypoints
//   vialons/vialats/vianames → waypoint lng / lat / name, each '|'-separated,
//                              and ALL three counts must match `vian`.
//   t=0 (drive)  dev=0  sourceApplication=appname
// Coordinates are already GCJ-02 (gaode), which the scheme expects by default.
async function openAmapFullRoute(
  stops: RouteStop[],
  sellerLat?: number,
  sellerLng?: number,
) {
  const valid = stops.filter((s) => s.lat && s.lng)
  if (valid.length === 0) {
    Taro.showToast({ title: '配送点无坐标数据', icon: 'none' })
    return
  }

  // Amap app accepts up to 16 waypoints; keep destination + waypoints within that.
  const batch = valid.slice(0, 17)
  const dest  = batch[batch.length - 1]
  const rest  = batch.slice(0, batch.length - 1)

  const hasShop = !!(sellerLat && sellerLng)
  // Start point: the shop if known, otherwise the first stop.
  const start = hasShop
    ? { lat: sellerLat as number, lng: sellerLng as number, name: '店铺' }
    : { lat: rest[0].lat, lng: rest[0].lng, name: rest[0].contactName || '起点' }

  // Waypoints exclude whatever we used as the start point.
  const waypoints = hasShop ? rest : rest.slice(1)

  const enc = (s: string) => encodeURIComponent(s || '')

  let uri =
    `amapuri://route/plan/?sourceApplication=orderfood&dev=0&t=0` +
    `&slat=${start.lat}&slon=${start.lng}&sname=${enc(start.name)}` +
    `&dlat=${dest.lat}&dlon=${dest.lng}&dname=${enc(dest.contactName || '终点')}`

  if (waypoints.length > 0) {
    const vialons  = waypoints.map((s) => s.lng).join('|')
    const vialats  = waypoints.map((s) => s.lat).join('|')
    const vianames = waypoints.map((s, i) => enc(s.contactName || `途经点${i + 1}`)).join('|')
    uri += `&vian=${waypoints.length}&vialons=${vialons}&vialats=${vialats}&vianames=${vianames}`
  }

  // Strategy 1: try wx.openUrl (only works for https; app scheme usually fails
  // here, but harmless to attempt — falls through to clipboard otherwise).
  const anyWx = Taro as any
  const openUrlOk = await new Promise<boolean>((resolve) => {
    if (typeof anyWx.openUrl !== 'function') { resolve(false); return }
    anyWx.openUrl({ url: uri, success: () => resolve(true), fail: () => resolve(false) })
  })
  if (openUrlOk) return

  // Strategy 2: copy the scheme to clipboard; pasting it into the phone browser
  // address bar (or Amap's search) launches the Amap app with the full route.
  try {
    await Taro.setClipboardData({ data: uri })
    Taro.showModal({
      title: '高德全程导航',
      content: `已复制导航链接（共 ${waypoints.length + 1} 个配送点）。\n请打开手机浏览器，粘贴到地址栏并访问，将自动唤起高德地图 App 并载入全部途经点。`,
      showCancel: false,
      confirmText: '知道了',
    })
  } catch (_) {
    Taro.showModal({
      title: '无法自动复制',
      content: '请手动复制此链接，在浏览器地址栏打开：\n' + uri.slice(0, 150) + '…',
      showCancel: false,
      confirmText: '关闭',
    })
  }
}

export default function SellerRoutePage() {
  const DEFAULT_COUNT = 3
  const [date, setDate]               = useState('')
  const [driverCount, setDriverCount] = useState(String(DEFAULT_COUNT))
  const [driverNames, setDriverNames] = useState<string[]>(
    Array.from({ length: DEFAULT_COUNT }, () => '')
  )
  const [cluster, setCluster]         = useState<ClusterResult | null>(null)
  const [selected, setSelected]       = useState(0)
  const [originIdx, setOriginIdx]     = useState(-1)
  const [polyPoints, setPolyPoints]   = useState<{ latitude: number; longitude: number }[]>([])
  const [segInfo, setSegInfo]         = useState<{ distance: number; duration: number; count: number } | null>(null)
  const [loading, setLoading]         = useState(false)
  const [delivered, setDelivered]     = useState<Set<number>>(new Set())

  // Resize the names array whenever the driver count input changes.
  // If a route was already generated, debounce an automatic re-cluster.
  const handleCountChange = (val: string) => {
    setDriverCount(val)
    const n = parseInt(val, 10)
    if (!isNaN(n) && n > 0) {
      setDriverNames((prev) => Array.from({ length: n }, (_, i) => prev[i] ?? ''))
      if (cluster) {
        if (reclusterTimer.current) clearTimeout(reclusterTimer.current)
        Taro.showToast({ title: '数量已变更，即将重新生成…', icon: 'none', duration: 1500 })
        reclusterTimer.current = setTimeout(() => runCluster(n), 800)
      }
    }
  }

  // Get display name for driver index i
  const getDriverLabel = (i: number) =>
    driverNames[i]?.trim() || `配送员${i + 1}`

  // Debounce timer: auto re-cluster when driver count changes while a route exists
  const reclusterTimer = useRef<ReturnType<typeof setTimeout> | null>(null)

  // ScrollView ref for scrolling to the current stop card
  const scrollViewId = useRef('scroll-stop-0')

  useLoad(async (opts) => {
    const d = (opts.deliveryDate as string) || ''
    setDate(d)
    Taro.setNavigationBarTitle({ title: '配送路线' })
    try {
      const data = await request<ClusterResult>({ url: `/seller/routes/cluster?deliveryDate=${d}` })
      if (data?.drivers?.length) {
        setCluster(data)
        const n = data.driverCount
        if (n) {
          setDriverCount(String(n))
          // Keep names that the user already typed; fill gaps with empty string
          setDriverNames((prev) => Array.from({ length: n }, (_, i) => prev[i] ?? ''))
        }
      }
    } catch (_) { /* not yet clustered */ }
  })

  const runCluster = async (overrideCount?: number) => {
    const n = overrideCount ?? parseInt(driverCount, 10)
    if (!n || n < 1) { Taro.showToast({ title: '请输入配送员数量', icon: 'none' }); return }
    setLoading(true)
    setDelivered(new Set())
    try {
      const data = await request<ClusterResult>({
        url: '/seller/routes/cluster',
        method: 'POST',
        data: { deliveryDate: date, driverCount: n }
      })
      setCluster(data)
      setSelected(0)
      setOriginIdx(-1)
    } catch (e: any) {
      Taro.showToast({ title: e?.message || '聚类失败', icon: 'none' })
    } finally { setLoading(false) }
  }

  const activeDriver = cluster?.drivers[selected]

  // Index of the first undelivered stop = "current stop" for the rider
  const currentStopIdx = activeDriver
    ? activeDriver.stops.findIndex((s) => !delivered.has(s.orderId))
    : -1
  const currentStop = currentStopIdx >= 0 ? activeDriver?.stops[currentStopIdx] : null

  const loadSegment = useCallback(async () => {
    if (!cluster || !activeDriver || activeDriver.stops.length === 0) {
      setPolyPoints([]); setSegInfo(null); return
    }
    const stops  = activeDriver.stops
    const origin = originIdx < 0
      ? { lat: cluster.sellerLat, lng: cluster.sellerLng }
      : { lat: stops[originIdx].lat, lng: stops[originIdx].lng }
    const startAt  = originIdx < 0 ? 0 : originIdx + 1
    const segment  = stops.slice(startAt, startAt + SLICE)
    if (segment.length === 0) { setPolyPoints([]); setSegInfo({ distance: 0, duration: 0, count: 0 }); return }
    const destination = segment[segment.length - 1]
    const waypoints   = segment.slice(0, segment.length - 1)
    setLoading(true)
    try {
      const dir = await request<Directions>({
        url: '/seller/routes/directions', method: 'POST',
        data: {
          origin,
          destination: { lat: destination.lat, lng: destination.lng },
          waypoints: waypoints.map((w) => ({ lat: w.lat, lng: w.lng }))
        }
      })
      setPolyPoints((dir.points || []).map((p) => ({ latitude: p.lat, longitude: p.lng })))
      setSegInfo({ distance: dir.distance, duration: dir.duration, count: segment.length })
    } catch (_) {
      const pts = [origin, ...segment].map((p) => ({ latitude: p.lat, longitude: p.lng }))
      setPolyPoints(pts)
      setSegInfo(null)
      Taro.showToast({ title: '道路路线获取失败，显示直线', icon: 'none' })
    } finally { setLoading(false) }
  }, [cluster, activeDriver, originIdx])

  useEffect(() => { loadSegment() }, [loadSegment])

  // After marking a stop as delivered, scroll list to the new current stop
  useEffect(() => {
    if (currentStopIdx >= 0) {
      scrollViewId.current = `scroll-stop-${currentStopIdx}`
    }
  }, [currentStopIdx])

  const selectDriver = (idx: number) => { setSelected(idx); setOriginIdx(-1) }

  const onMarkerTap = (e: any) => {
    const id = e?.detail?.markerId
    if (id == null || id === SHOP_MARKER_ID || !activeDriver) return
    const idx = activeDriver.stops.findIndex((s) => s.orderId === id)
    if (idx >= 0) setOriginIdx(idx)
  }

  const toggleDelivered = (orderId: number, e?: any) => {
    e?.stopPropagation?.()
    setDelivered((prev) => {
      const next = new Set(prev)
      if (next.has(orderId)) next.delete(orderId)
      else next.add(orderId)
      return next
    })
  }

  // ── Map markers ───────────────────────────────────────────
  const center = activeDriver?.stops[0] ||
    (cluster ? { lat: cluster.sellerLat, lng: cluster.sellerLng } : { lat: 31.2, lng: 121.5 })

  const markers: any[] = []
  if (cluster?.sellerLat && cluster?.sellerLng) {
    markers.push({
      id: SHOP_MARKER_ID, latitude: cluster.sellerLat, longitude: cluster.sellerLng,
      width: 28, height: 28,
      callout: { content: '店铺', color: '#000', fontSize: 12, borderRadius: 4, padding: 4, display: 'ALWAYS' }
    })
  }
  // Only show stops for the currently selected driver
  if (activeDriver) {
    activeDriver.stops.forEach((s, si) => {
      const done   = delivered.has(s.orderId)
      const isCurr = si === currentStopIdx
      markers.push({
        id: s.orderId, latitude: s.lat, longitude: s.lng,
        alpha:  done ? 0.25 : 1,
        width:  isCurr ? 28 : (done ? 18 : 20),
        height: isCurr ? 28 : (done ? 18 : 20),
        callout: {
          content:  done ? `✓ ${s.contactName}` : (isCurr ? `📍${si + 1}.${s.contactName}` : `${si + 1}.${s.contactName}`),
          color:    done ? '#aaa' : (isCurr ? '#07c160' : activeDriver.color),
          fontSize: isCurr ? 13 : 11,
          borderRadius: 4, padding: 4,
          display: done ? 'BYCLICK' : 'ALWAYS'
        }
      })
    })
  }

  const polyline = polyPoints.length > 1
    ? [{ points: polyPoints, color: activeDriver?.color || '#07c160', width: 5, arrowLine: true }]
    : []

  const hasMore = !!activeDriver && originIdx >= 0 && originIdx + SLICE < activeDriver.stops.length
  const segmentStartLabel = originIdx < 0 ? '店铺' : `第 ${originIdx + 1} 站`

  const deliveredCount = activeDriver?.stops.filter((s) => delivered.has(s.orderId)).length ?? 0
  const totalCount     = activeDriver?.stops.length ?? 0
  const allDone        = totalCount > 0 && deliveredCount === totalCount

  return (
    <View className='route-page'>

      {/* ══ 固定顶部：地图 + 进度 ════════════════════════════════ */}
      <View className='sticky-top'>

        {/* Config card */}
        <View className='config-card'>
          {/* Row 1: count + date + button */}
          <View className='config-row'>
            <Text className='label'>配送员数量</Text>
            <Input
              className='count-input'
              type='number'
              value={driverCount}
              onInput={(e) => handleCountChange(e.detail.value)}
            />
            {date ? <Text className='date-label'>{date}</Text> : null}
            <Button size='mini' className='btn-primary' loading={loading} onClick={runCluster}>
              生成路线
            </Button>
          </View>

          {/* Row 2: driver name inputs (dynamic, one per driver) */}
          {driverNames.length > 0 && (
            <View className='names-grid'>
              {driverNames.map((name, i) => (
                <View key={i} className='name-item'>
                  <View className='name-badge' style={{ background: DRIVER_COLORS[i % DRIVER_COLORS.length] }}>
                    {i + 1}
                  </View>
                  <Input
                    className='name-input'
                    placeholder={`配送员${i + 1}姓名`}
                    value={name}
                    onInput={(e) => {
                      const val = e.detail.value
                      setDriverNames((prev) => {
                        const next = [...prev]
                        next[i] = val
                        return next
                      })
                    }}
                  />
                </View>
              ))}
            </View>
          )}
        </View>

        {/* Driver tabs - show real names */}
        {cluster && cluster.drivers.length > 0 && (
          <View className='driver-tabs'>
            {cluster.drivers.map((dr, i) => (
              <View
                key={i}
                className={`driver-tab ${i === selected ? 'active' : ''}`}
                style={{ borderColor: dr.color, color: i === selected ? '#fff' : dr.color, background: i === selected ? dr.color : '#fff' }}
                onClick={() => selectDriver(i)}
              >
                {getDriverLabel(i)}（{dr.stops.length}单）
              </View>
            ))}
          </View>
        )}

        {/* Map */}
        <Map
          className='map'
          latitude={center.lat}
          longitude={center.lng}
          markers={markers}
          polyline={polyline}
          scale={12}
          showLocation
          onMarkerTap={onMarkerTap}
        />

        {/* Progress + Next-stop nav button */}
        {activeDriver && (
          <View className='progress-section'>
            {allDone ? (
              <View className='all-done-banner'>
                <Text>🎉 全部配送完成！</Text>
              </View>
            ) : (
              <>
                <View className='progress-header'>
                  <Text className='progress-text'>
                    已送达 {deliveredCount} / {totalCount}
                  </Text>
                  {deliveredCount > 0 && (
                    <Text className='progress-reset' onClick={() => setDelivered(new Set())}>重置</Text>
                  )}
                </View>
                <View className='progress-bar-wrap'>
                  <View className='progress-bar-fill' style={{ width: `${(deliveredCount / totalCount) * 100}%` }} />
                </View>

                {/* ── 高德全程导航（所有未送达点） ── */}
                {activeDriver.stops.some((s) => !delivered.has(s.orderId) && s.lat && s.lng) && (
                  <View
                    className='amap-full-btn'
                    onClick={() => void openAmapFullRoute(
                      activeDriver.stops.filter((s) => !delivered.has(s.orderId)),
                      cluster?.sellerLat,
                      cluster?.sellerLng,
                    )}
                  >
                    <Text className='amap-full-icon'>🗺</Text>
                    <View className='amap-full-text'>
                      <Text className='amap-full-title'>高德全程导航</Text>
                      <Text className='amap-full-sub'>
                        将剩余 {activeDriver.stops.filter((s) => !delivered.has(s.orderId)).length} 个配送点一键导入高德
                      </Text>
                    </View>
                    <Text className='amap-full-arrow'>›</Text>
                  </View>
                )}

                {/* ── 骑手核心操作：导航到下一站 ── */}
                {currentStop && (
                  <View className='next-nav-bar' onClick={() => navigateTo(currentStop)}>
                    <View className='next-nav-info'>
                      <Text className='next-nav-label'>下一站</Text>
                      <Text className='next-nav-name'>{currentStopIdx + 1}. {currentStop.contactName}</Text>
                      <Text className='next-nav-addr' numberOfLines={1}>{currentStop.address}</Text>
                    </View>
                    <View className='next-nav-btn'>
                      <Text className='next-nav-icon'>◎</Text>
                      <Text>导航</Text>
                    </View>
                  </View>
                )}
              </>
            )}
          </View>
        )}
      </View>

      {/* ══ 可滚动列表 ═══════════════════════════════════════════ */}
      <ScrollView
        scrollY
        className='scroll-list'
        scrollIntoView={scrollViewId.current}
        scrollWithAnimation
      >
        {/* Inner wrapper carries the horizontal padding — ScrollView right-padding is unreliable in WeChat */}
        <View className='scroll-inner'>

        {/* Segment info */}
        {activeDriver && (
          <View className='seg-card'>
            <View className='seg-row'>
              <Text className='seg-title'>从{segmentStartLabel}起 · {segInfo?.count ?? 0} 个点</Text>
              {segInfo && (
                <Text className='seg-meta'>
                  {(segInfo.distance / 1000).toFixed(1)} km · {Math.round(segInfo.duration / 60)} 分钟
                </Text>
              )}
            </View>
            <View className='seg-actions'>
              <Button size='mini' onClick={() => setOriginIdx(-1)}>重置路线段</Button>
              {hasMore && (
                <Button size='mini' className='btn-primary' onClick={() => setOriginIdx(originIdx + SLICE)}>
                  下一段
                </Button>
              )}
            </View>
          </View>
        )}

        {/* Stop cards */}
        {activeDriver?.stops.map((s, i) => {
          const done    = delivered.has(s.orderId)
          const isCurr  = i === currentStopIdx
          return (
            <View
              id={`scroll-stop-${i}`}
              key={s.orderId}
              className={`stop-card ${done ? 'done' : ''} ${isCurr ? 'current' : ''}`}
              onClick={() => !done && setOriginIdx(i)}
            >
              {/* Sequence badge */}
              <View className='stop-seq' style={{ background: done ? '#ccc' : activeDriver.color }}>
                {done ? '✓' : i + 1}
              </View>

              {/* Info */}
              <View className='stop-info'>
                {isCurr && !done && <Text className='curr-label'>当前站</Text>}
                <Text className={`stop-name ${done ? 'done-text' : ''}`}>{s.contactName}</Text>
                <Text className={`stop-addr ${done ? 'done-text' : ''}`}>{s.address}</Text>
                {s.phone ? <Text className='stop-phone'>{s.phone}</Text> : null}
              </View>

              {/* Action buttons */}
              <View className='stop-actions'>
                {!done && (
                  <View
                    className={`nav-icon-btn ${isCurr ? 'nav-icon-btn-active' : ''}`}
                    onClick={(e) => { e.stopPropagation?.(); navigateTo(s) }}
                  >
                    <Text>导航</Text>
                  </View>
                )}
                <View
                  className={`deliver-btn ${done ? 'deliver-btn-done' : ''}`}
                  onClick={(e) => toggleDelivered(s.orderId, e)}
                >
                  {done ? '已送达' : '送达'}
                </View>
              </View>
            </View>
          )
        })}

        {!cluster && <View className='empty-hint'>输入配送员数量后点击「生成路线」</View>}
        {cluster && cluster.drivers.length === 0 && (
          <View className='empty-hint'>暂无可配送订单（订单需有买家坐标）</View>
        )}
        <View style={{ height: '40px' }} />

        </View>{/* /scroll-inner */}
      </ScrollView>
    </View>
  )
}
