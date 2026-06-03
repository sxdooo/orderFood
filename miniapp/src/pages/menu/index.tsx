import { View, Text, Button } from '@tarojs/components'
import Taro, { useDidShow } from '@tarojs/taro'
import { useState } from 'react'
import { request } from '../../utils/request'
import { CartItem, getCart, setCart, cartTotal } from '../../utils/cart'
import { getUser } from '../../utils/auth'
import { formatPrice } from '../../utils/format'
import './index.scss'

interface MenuItem {
  id: number
  name: string
  price: number
  description: string
}

interface Menu {
  id: number
  deliveryDate: string
  items: MenuItem[]
}

interface CutoffStatus {
  isOpen: boolean
  cutoffTime: string
  secondsLeft: number
}

// ─── Seller Hub ───────────────────────────────────────────────
function SellerHub() {
  const today = new Date().toLocaleDateString('zh-CN', { month: 'long', day: 'numeric', weekday: 'short' })

  return (
    <View className='container seller-hub'>
      <View className='hub-header card'>
        <Text className='hub-title'>卖家工作台</Text>
        <Text className='hub-date'>{today}</Text>
      </View>
      <View className='hub-actions'>
        <View className='hub-action-card card' onClick={() => Taro.navigateTo({ url: '/pages/seller/menus/index' })}>
          <View className='hub-action-icon menu-icon' />
          <View className='hub-action-body'>
            <Text className='hub-action-title'>发布菜单</Text>
            <Text className='hub-action-desc'>管理明日及以后的菜单</Text>
          </View>
          <Text className='hub-chevron'>›</Text>
        </View>
        <View className='hub-action-card card' onClick={() => Taro.navigateTo({ url: '/pages/seller/orders/index' })}>
          <View className='hub-action-icon order-icon' />
          <View className='hub-action-body'>
            <Text className='hub-action-title'>今日订单</Text>
            <Text className='hub-action-desc'>查看订单、设置截单、生成路线</Text>
          </View>
          <Text className='hub-chevron'>›</Text>
        </View>
      </View>
    </View>
  )
}

// ─── Buyer Menu ───────────────────────────────────────────────
export default function MenuPage() {
  const [menu, setMenu] = useState<Menu | null>(null)
  const [cart, setCartState] = useState<CartItem[]>([])
  const [cutoff, setCutoff] = useState<CutoffStatus | null>(null)
  const [isSeller, setIsSeller] = useState(false)
  const [cartOpen, setCartOpen] = useState(false)

  const load = async () => {
    const user = getUser()
    if (user?.role === 'seller') {
      setIsSeller(true)
      Taro.setTabBarItem({ index: 0, text: '工作台' })
      Taro.setTabBarItem({ index: 1, text: '订单管理' })
      return
    }
    setIsSeller(false)
    Taro.setTabBarItem({ index: 0, text: '订餐' })
    Taro.setTabBarItem({ index: 1, text: '我的订单' })
    try {
      const [m, c] = await Promise.all([
        request<Menu | null>({ url: '/menus/tomorrow' }),
        request<CutoffStatus>({ url: '/cutoff/status' })
      ])
      setMenu(m)
      setCutoff(c)
      setCartState(getCart())
    } catch { /* ignore */ }
  }

  useDidShow(() => {
    Taro.setNavigationBarTitle({ title: isSeller ? '工作台' : '明日菜单' })
    load()
  })

  if (isSeller) return <SellerHub />

  const addItem = (item: MenuItem) => {
    if (!cutoff?.isOpen) {
      Taro.showToast({ title: '今日已截单', icon: 'none' })
      return
    }
    const next = [...cart]
    const idx = next.findIndex((i) => i.menuItemId === item.id)
    if (idx >= 0) {
      next[idx].quantity += 1
    } else {
      next.push({ menuItemId: item.id, name: item.name, price: item.price, quantity: 1 })
    }
    setCart(next)
    setCartState(next)
  }

  const updateQty = (menuItemId: number, delta: number) => {
    const next = cart
      .map((i) => i.menuItemId === menuItemId ? { ...i, quantity: i.quantity + delta } : i)
      .filter((i) => i.quantity > 0)
    setCart(next)
    setCartState(next)
    if (next.length === 0) setCartOpen(false)
  }

  const goCheckout = () => {
    setCartOpen(false)
    Taro.navigateTo({ url: '/pages/checkout/index' })
  }

  const formatCountdown = (s: number) => {
    const h = Math.floor(s / 3600)
    const m = Math.floor((s % 3600) / 60)
    if (h > 0) return `${h}小时${m}分后截单`
    return `${m}分钟后截单`
  }

  const totalCount = cart.reduce((s, i) => s + i.quantity, 0)

  return (
    <View className='container menu-page'>
      {/* Cutoff banner */}
      {cutoff && (
        <View className={`cutoff-bar ${cutoff.isOpen ? '' : 'closed'}`}>
          {cutoff.isOpen ? (
            <Text>截单时间 {cutoff.cutoffTime} · {formatCountdown(cutoff.secondsLeft)}</Text>
          ) : (
            <Text>今日已截单，明天再来看看</Text>
          )}
        </View>
      )}

      {/* Menu items */}
      {!menu ? (
        <View className='empty-tip'>卖家还没发布明日菜单，请稍后再来</View>
      ) : (
        menu.items?.map((item) => {
          const qty = cart.find((c) => c.menuItemId === item.id)?.quantity ?? 0
          return (
            <View key={item.id} className='card dish-card'>
              <View className='dish-info'>
                <Text className='dish-name'>{item.name}</Text>
                {item.description ? <Text className='dish-desc'>{item.description}</Text> : null}
                <Text className='dish-price'>¥{formatPrice(item.price)}</Text>
              </View>
              <View className='dish-ctrl'>
                {qty > 0 ? (
                  <View className='qty-row'>
                    <View className='qty-btn minus' onClick={() => updateQty(item.id, -1)}>－</View>
                    <Text className='qty-num'>{qty}</Text>
                    <View className='qty-btn plus' onClick={() => addItem(item)}>＋</View>
                  </View>
                ) : (
                  <View
                    className={`add-btn ${!cutoff?.isOpen ? 'disabled' : ''}`}
                    onClick={() => addItem(item)}
                  >
                    ＋
                  </View>
                )}
              </View>
            </View>
          )
        })
      )}

      {/* Cart overlay */}
      {cartOpen && <View className='cart-overlay' onClick={() => setCartOpen(false)} />}

      {/* Cart panel */}
      {cartOpen && (
        <View className='cart-panel'>
          <View className='cart-panel-header'>
            <Text className='cart-panel-title'>已选菜品</Text>
            <Text className='cart-panel-close' onClick={() => setCartOpen(false)}>✕</Text>
          </View>
          {cart.map((item) => (
            <View key={item.menuItemId} className='cart-row'>
              <Text className='cart-item-name'>{item.name}</Text>
              <View className='qty-row'>
                <View className='qty-btn minus' onClick={() => updateQty(item.menuItemId, -1)}>－</View>
                <Text className='qty-num'>{item.quantity}</Text>
                <View className='qty-btn plus' onClick={() => updateQty(item.menuItemId, 1)}>＋</View>
              </View>
              <Text className='cart-item-subtotal'>¥{formatPrice(item.price * item.quantity)}</Text>
            </View>
          ))}
          <View className='cart-panel-footer'>
            <Text className='cart-total-text'>共 {totalCount} 件 · <Text className='cart-total-price'>¥{formatPrice(cartTotal(cart))}</Text></Text>
            <Button className='btn-primary checkout-btn' onClick={goCheckout}>去结算</Button>
          </View>
        </View>
      )}

      {/* Cart bar */}
      {cart.length > 0 && !cartOpen && (
        <View className='cart-bar' onClick={() => setCartOpen(true)}>
          <View className='cart-bar-left'>
            <View className='cart-badge'>{totalCount}</View>
            <Text className='cart-bar-total'>¥{formatPrice(cartTotal(cart))}</Text>
          </View>
          <View className='cart-bar-right'>
            <Text>查看购物车</Text>
            <Text className='cart-chevron'>›</Text>
          </View>
        </View>
      )}
    </View>
  )
}
