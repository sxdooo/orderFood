import { View, Input, Button, Text } from '@tarojs/components'
import Taro, { useLoad } from '@tarojs/taro'
import { useState } from 'react'
import { request } from '../../utils/request'
import { getUser } from '../../utils/auth'
import './index.scss'

export default function ProfileSetupPage() {
  const [role, setRole] = useState<'buyer' | 'seller' | null>(null)
  const [loading, setLoading] = useState(false)

  // Buyer fields
  const [contactName, setContactName] = useState('')
  const [contactPhone, setContactPhone] = useState('')
  const [address, setAddress] = useState('')

  // Seller fields
  const [shopName, setShopName] = useState('')
  const [shopAddress, setShopAddress] = useState('')

  useLoad(() => {
    Taro.setNavigationBarTitle({ title: '完善资料' })
    // Role is determined by the backend; pick it from local storage so we can
    // skip the selection step and show the right form directly.
    const user = getUser()
    if (user?.role) {
      setRole(user.role)
    }
    // Pre-fill phone number obtained from WeChat authorization during login.
    const savedPhone = Taro.getStorageSync('wechat_phone')
    if (savedPhone) {
      setContactPhone(savedPhone)
    }
  })

  const submitBuyer = async () => {
    if (!contactName || !contactPhone || !address) {
      Taro.showToast({ title: '请填写完整信息', icon: 'none' })
      return
    }
    setLoading(true)
    try {
      await request({ url: '/buyer/profile', method: 'PUT', data: { contactName, contactPhone, address } })
      Taro.showToast({ title: '保存成功', icon: 'success' })
      Taro.switchTab({ url: '/pages/menu/index' })
    } catch (err) {
      Taro.showToast({ title: err instanceof Error ? err.message : '保存失败', icon: 'none' })
    } finally {
      setLoading(false)
    }
  }

  const submitSeller = async () => {
    if (!shopName || !shopAddress) {
      Taro.showToast({ title: '请填写店铺信息', icon: 'none' })
      return
    }
    setLoading(true)
    try {
      await request({ url: '/seller/profile', method: 'PUT', data: { shopName, address: shopAddress } })
      Taro.showToast({ title: '保存成功', icon: 'success' })
      Taro.switchTab({ url: '/pages/menu/index' })
    } catch (err) {
      Taro.showToast({ title: err instanceof Error ? err.message : '保存失败', icon: 'none' })
    } finally {
      setLoading(false)
    }
  }

  // ── Role not determined yet (shouldn't normally happen, but fallback) ──
  if (role === null) {
    return (
      <View className='profile-setup-page'>
        <View className='setup-header'>
          <Text className='setup-title'>你的身份</Text>
          <Text className='setup-hint'>请选择你的使用角色</Text>
        </View>
        <View className='role-cards'>
          <View className='role-card' onClick={() => setRole('buyer')}>
            <Text className='role-icon'>🛍️</Text>
            <Text className='role-name'>买家</Text>
            <Text className='role-desc'>浏览菜单、下单订餐</Text>
          </View>
          <View className='role-card' onClick={() => setRole('seller')}>
            <Text className='role-icon'>🍱</Text>
            <Text className='role-name'>卖家</Text>
            <Text className='role-desc'>发布菜单、管理订单</Text>
          </View>
        </View>
        <Text className='role-tip'>* 卖家身份需经平台授权，未授权账号仍以买家身份使用</Text>
      </View>
    )
  }

  // ── Buyer setup form ──────────────────────────────────────────────────
  if (role === 'buyer') {
    return (
      <View className='profile-setup-page'>
        <View className='setup-header'>
          <Text className='setup-title'>买家资料</Text>
          <Text className='setup-hint'>填写默认配送信息，下单时自动带入，随时可修改</Text>
        </View>
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
          <Input value={address} onInput={(e) => setAddress(e.detail.value)} placeholder='请输入详细地址（街道门牌号）' />
        </View>
        <Button className='btn-primary submit-btn' loading={loading} onClick={submitBuyer}>
          保存并开始订餐
        </Button>
      </View>
    )
  }

  // ── Seller setup form ─────────────────────────────────────────────────
  return (
    <View className='profile-setup-page'>
      <View className='setup-header'>
        <Text className='setup-title'>店铺资料</Text>
        <Text className='setup-hint'>填写店铺信息，用于买家查看和路线规划定位</Text>
      </View>
      <View className='field'>
        <Text className='field-label'>店铺名称</Text>
        <Input value={shopName} onInput={(e) => setShopName(e.detail.value)} placeholder='请输入店铺名称' />
      </View>
      <View className='field'>
        <Text className='field-label'>店铺地址</Text>
        <Input value={shopAddress} onInput={(e) => setShopAddress(e.detail.value)} placeholder='请输入详细地址（用于距离计算）' />
      </View>
      <Text className='field-tip'>填写准确地址后系统会自动解析坐标，用于配送路线规划</Text>
      <Button className='btn-primary submit-btn' loading={loading} onClick={submitSeller}>
        保存并开始使用
      </Button>
    </View>
  )
}
