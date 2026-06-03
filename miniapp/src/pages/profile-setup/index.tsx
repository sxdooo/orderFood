import { View, Input, Button, Text } from '@tarojs/components'
import Taro, { useLoad } from '@tarojs/taro'
import { useState } from 'react'
import { request } from '../../utils/request'
import './index.scss'

export default function ProfileSetupPage() {
  const [contactName, setContactName] = useState('')
  const [contactPhone, setContactPhone] = useState('')
  const [address, setAddress] = useState('')
  const [loading, setLoading] = useState(false)

  useLoad(() => {
    Taro.setNavigationBarTitle({ title: '完善资料' })
    // Pre-fill phone number obtained from WeChat authorization during login.
    const savedPhone = Taro.getStorageSync('wechat_phone')
    if (savedPhone) {
      setContactPhone(savedPhone)
    }
  })

  const handleSubmit = async () => {
    if (!contactName || !contactPhone || !address) {
      Taro.showToast({ title: '请填写完整信息', icon: 'none' })
      return
    }
    setLoading(true)
    try {
      await request({
        url: '/buyer/profile',
        method: 'PUT',
        data: { contactName, contactPhone, address }
      })
      Taro.showToast({ title: '保存成功', icon: 'success' })
      Taro.switchTab({ url: '/pages/menu/index' })
    } catch (err) {
      Taro.showToast({ title: err instanceof Error ? err.message : '保存失败', icon: 'none' })
    } finally {
      setLoading(false)
    }
  }

  return (
    <View className='profile-setup-page'>
      <View className='setup-header'>
        <Text className='setup-title'>完善资料</Text>
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
      <Button className='btn-primary submit-btn' loading={loading} onClick={handleSubmit}>
        保存并开始订餐
      </Button>
    </View>
  )
}
