import { View, Input, Button, Textarea, Text } from '@tarojs/components'
import Taro, { useLoad } from '@tarojs/taro'
import { useState } from 'react'
import { request } from '../../utils/request'
import { getCart, clearCart, cartTotal } from '../../utils/cart'
import { formatPrice } from '../../utils/format'
import './index.scss'

interface BuyerProfile {
  contactName: string
  contactPhone: string
  address: string
}

export default function CheckoutPage() {
  const [cart] = useState(getCart())
  const [contactName, setContactName] = useState('')
  const [contactPhone, setContactPhone] = useState('')
  const [address, setAddress] = useState('')
  const [remark, setRemark] = useState('')
  const [loading, setLoading] = useState(false)

  useLoad(async () => {
    Taro.setNavigationBarTitle({ title: '确认订单' })
    if (cart.length === 0) {
      Taro.showToast({ title: '购物车为空', icon: 'none' })
      setTimeout(() => Taro.navigateBack(), 500)
      return
    }
    try {
      const profile = await request<BuyerProfile>({ url: '/buyer/profile' })
      setContactName(profile.contactName || '')
      setContactPhone(profile.contactPhone || '')
      setAddress(profile.address || '')
    } catch { /* ignore */ }
  })

  const submit = async () => {
    if (!contactName || !contactPhone || !address) {
      Taro.showToast({ title: '请填写完整配送信息', icon: 'none' })
      return
    }
    setLoading(true)
    try {
      await request({
        url: '/orders',
        method: 'POST',
        data: {
          items: cart.map((c) => ({ menuItemId: c.menuItemId, quantity: c.quantity })),
          contactName,
          contactPhone,
          address,
          remark
        }
      })
      clearCart()
      Taro.showToast({ title: '下单成功！', icon: 'success' })
      setTimeout(() => Taro.switchTab({ url: '/pages/orders/index' }), 800)
    } catch (err) {
      Taro.showToast({ title: err instanceof Error ? err.message : '下单失败', icon: 'none' })
    } finally {
      setLoading(false)
    }
  }

  return (
    <View className='container checkout-page'>
      {/* Order items */}
      <View className='card order-items-card'>
        <Text className='section-title'>订单菜品</Text>
        {cart.map((c) => (
          <View key={c.menuItemId} className='item-row'>
            <Text className='item-name'>{c.name}</Text>
            <Text className='item-qty'>×{c.quantity}</Text>
            <Text className='item-price'>¥{formatPrice(c.price * c.quantity)}</Text>
          </View>
        ))}
        <View className='total-row'>
          <Text className='total-label'>合计</Text>
          <Text className='total-amount'>¥{formatPrice(cartTotal(cart))}</Text>
        </View>
      </View>

      {/* Delivery info */}
      <View className='card delivery-card'>
        <Text className='delivery-title'>配送信息</Text>
        <Text className='delivery-hint'>已自动带入默认信息，可临时修改（不影响默认配置）</Text>
        <View className='field'>
          <Text className='field-label'>联系人</Text>
          <Input value={contactName} onInput={(e) => setContactName(e.detail.value)} placeholder='请输入姓名' />
        </View>
        <View className='field'>
          <Text className='field-label'>手机号</Text>
          <Input type='number' value={contactPhone} onInput={(e) => setContactPhone(e.detail.value)} placeholder='请输入手机号' />
        </View>
        <View className='field'>
          <Text className='field-label'>收货地址</Text>
          <Input value={address} onInput={(e) => setAddress(e.detail.value)} placeholder='请输入详细地址' />
        </View>
        <View className='field'>
          <Text className='field-label'>备注（选填）</Text>
          <Textarea value={remark} onInput={(e) => setRemark(e.detail.value)} placeholder='如：不要辣、放门口等' />
        </View>
      </View>

      <Button className='btn-primary submit-btn' loading={loading} onClick={submit}>
        提交订单
      </Button>
    </View>
  )
}
