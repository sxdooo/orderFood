import { View, Button, Text } from '@tarojs/components'
import Taro, { useLoad } from '@tarojs/taro'
import { useState } from 'react'
import { request } from '../../utils/request'
import { setToken, setUser } from '../../utils/auth'
import './index.scss'

export default function LoginPage() {
  const [loading, setLoading] = useState(false)

  useLoad(() => {
    Taro.setNavigationBarTitle({ title: '登录' })
  })

  // Shared post-login navigation logic.
  const afterLogin = async (token: string, user: { id: number; role: 'buyer' | 'seller' }) => {
    setToken(token)
    setUser(user)
    if (user.role === 'seller') {
      await Taro.switchTab({ url: '/pages/menu/index' })
      return
    }
    const profile = await request<{ profileCompleted: boolean }>({ url: '/buyer/profile' })
    if (!profile.profileCompleted) {
      await Taro.reLaunch({ url: '/pages/profile-setup/index' })
      return
    }
    await Taro.switchTab({ url: '/pages/menu/index' })
  }

  // Called by the getPhoneNumber button; phoneCode may be empty if user denied.
  const handleGetPhoneNumber = async (e: { detail: { code?: string; errMsg: string } }) => {
    setLoading(true)
    try {
      const loginRes = await Taro.login()
      const phoneCode = e.detail.code ?? ''
      const data = await request<{ token: string; phone: string; user: { id: number; role: 'buyer' | 'seller' } }>({
        url: '/auth/wechat',
        method: 'POST',
        data: { code: loginRes.code || 'dev_code', phoneCode }
      })
      // Persist resolved phone so profile-setup can pre-fill it.
      if (data.phone) {
        Taro.setStorageSync('wechat_phone', data.phone)
      }
      await afterLogin(data.token, data.user)
    } catch (err) {
      Taro.showToast({ title: err instanceof Error ? err.message : '登录失败', icon: 'none' })
    } finally {
      setLoading(false)
    }
  }

  return (
    <View className='login-page'>
      <View className='logo'>
        <Text className='logo-text'>饭</Text>
      </View>
      <View className='title'>订饭小程序</View>
      <View className='subtitle'>一键下单，明日准时送达{'\n'}授权手机号，卖家自动获得管理权限</View>
      <Button
        className='btn-primary'
        loading={loading}
        openType='getPhoneNumber'
        onGetPhoneNumber={handleGetPhoneNumber}
      >
        微信一键登录
      </Button>
      <Text className='login-hint'>点击即授权微信手机号，用于识别卖家身份</Text>
    </View>
  )
}
