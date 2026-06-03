import { View, Input, Button, Text } from '@tarojs/components'
import Taro, { useDidShow } from '@tarojs/taro'
import { useState } from 'react'
import { request } from '../../utils/request'
import { clearAuth, setUser } from '../../utils/auth'
import './index.scss'

interface BuyerProfile {
  contactName: string
  contactPhone: string
  address: string
  profileCompleted: boolean
}

interface SellerProfile {
  shopName: string
  address: string
  addressLat?: number
  addressLng?: number
}

export default function MinePage() {
  const [buyerProfile, setBuyerProfile] = useState<BuyerProfile | null>(null)
  const [sellerProfile, setSellerProfile] = useState<SellerProfile | null>(null)
  const [role, setRole] = useState<'buyer' | 'seller'>('buyer')
  const [editing, setEditing] = useState(false)
  const [sellerEditing, setSellerEditing] = useState(false)
  const [editSeller, setEditSeller] = useState<SellerProfile>({ shopName: '', address: '' })
  const [loading, setLoading] = useState(false)
  const [pageReady, setPageReady] = useState(false)

  const load = async () => {
    try {
      const me = await request<{ id: number; role: 'buyer' | 'seller' }>({ url: '/auth/me' })
      setRole(me.role)
      setUser({ id: me.id, role: me.role })

      if (me.role === 'seller') {
        try {
          const sp = await request<SellerProfile>({ url: '/seller/profile' })
          setSellerProfile(sp)
          setEditSeller({ shopName: sp.shopName || '', address: sp.address || '' })
        } catch { /* profile not set up yet */ }
      } else {
        const bp = await request<BuyerProfile>({ url: '/buyer/profile' })
        setBuyerProfile(bp)
      }
    } catch { /* ignore */ }
    setPageReady(true)
  }

  useDidShow(() => {
    Taro.setNavigationBarTitle({ title: '我的' })
    load()
  })

  const handleSave = async () => {
    if (!buyerProfile) return
    setLoading(true)
    try {
      await request({
        url: '/buyer/profile',
        method: 'PUT',
        data: {
          contactName: buyerProfile.contactName,
          contactPhone: buyerProfile.contactPhone,
          address: buyerProfile.address
        }
      })
      Taro.showToast({ title: '已保存', icon: 'success' })
      setEditing(false)
    } catch (err) {
      Taro.showToast({ title: err instanceof Error ? err.message : '保存失败', icon: 'none' })
    } finally {
      setLoading(false)
    }
  }

  const handleSaveSeller = async () => {
    if (!editSeller.shopName.trim() || !editSeller.address.trim()) {
      Taro.showToast({ title: '请填写店铺名称和地址', icon: 'none' })
      return
    }
    setLoading(true)
    try {
      const sp = await request<SellerProfile>({
        url: '/seller/profile',
        method: 'PUT',
        data: { shopName: editSeller.shopName, address: editSeller.address, lat: 0, lng: 0 }
      })
      setSellerProfile(sp)
      setSellerEditing(false)
      Taro.showToast({ title: '已保存', icon: 'success' })
    } catch (err) {
      Taro.showToast({ title: err instanceof Error ? err.message : '保存失败', icon: 'none' })
    } finally {
      setLoading(false)
    }
  }

  const handleLogout = async () => {
    try { await request({ url: '/auth/logout', method: 'POST' }) } catch { /* ignore */ }
    clearAuth()
    Taro.reLaunch({ url: '/pages/login/index' })
  }

  if (!pageReady) {
    return <View className='container'><Text className='loading-text'>加载中...</Text></View>
  }

  return (
    <View className='container mine-page'>
      {/* ── Seller section ─────────────────────────────── */}
      {role === 'seller' && (
        <>
          <View className='avatar-card card'>
            <View className='avatar-circle seller-avatar'>卖</View>
            <View className='avatar-info'>
              <Text className='avatar-name'>{sellerProfile?.shopName || '我的店铺'}</Text>
              <Text className='avatar-role-tag seller-tag'>卖家</Text>
            </View>
          </View>

          {/* Seller profile card */}
          <View className='card'>
            <View className='row'>
              <Text className='section-title' style='margin-bottom:0'>店铺信息</Text>
              {!sellerEditing && (
                <Text className='edit-link' onClick={() => {
                  setEditSeller({ shopName: sellerProfile?.shopName || '', address: sellerProfile?.address || '' })
                  setSellerEditing(true)
                }}>编辑</Text>
              )}
            </View>
            <View className='divider' />
            {sellerEditing ? (
              <View className='edit-form'>
                <View className='field'>
                  <Text className='field-label'>店铺名称</Text>
                  <Input
                    className='field-input'
                    value={editSeller.shopName}
                    onInput={(e) => setEditSeller({ ...editSeller, shopName: e.detail.value })}
                    placeholder='请输入店铺名称'
                  />
                </View>
                <View className='field'>
                  <Text className='field-label'>店铺地址</Text>
                  <Input
                    className='field-input'
                    value={editSeller.address}
                    onInput={(e) => setEditSeller({ ...editSeller, address: e.detail.value })}
                    placeholder='请输入详细地址（用于路线规划）'
                  />
                </View>
                <View className='edit-btns'>
                  <Button className='btn-outline' onClick={() => setSellerEditing(false)}>取消</Button>
                  <Button className='btn-primary' loading={loading} onClick={handleSaveSeller}>保存</Button>
                </View>
              </View>
            ) : (
              <View className='profile-view'>
                <View className='profile-row'>
                  <Text className='profile-label'>店铺名</Text>
                  <Text className='profile-val'>{sellerProfile?.shopName || '未设置'}</Text>
                </View>
                <View className='profile-row'>
                  <Text className='profile-label'>地址</Text>
                  <Text className='profile-val'>{sellerProfile?.address || '未设置'}</Text>
                </View>
                {sellerProfile?.addressLat ? (
                  <View className='profile-row'>
                    <Text className='profile-label'>坐标</Text>
                    <Text className='profile-val geo-ok'>已获取 ✓</Text>
                  </View>
                ) : (
                  <View className='profile-row'>
                    <Text className='profile-label'>坐标</Text>
                    <Text className='profile-val geo-missing'>未获取（需配置高德API）</Text>
                  </View>
                )}
              </View>
            )}
          </View>

          <View className='card'>
            <Text className='section-title'>卖家中心</Text>
            <View className='action-list'>
              <View
                className='action-item'
                onClick={() => Taro.navigateTo({ url: '/pages/seller/menus/index' })}
              >
                <Text className='action-title'>菜单管理</Text>
                <Text className='action-arrow'>›</Text>
              </View>
              <View className='action-divider' />
              <View
                className='action-item'
                onClick={() => Taro.navigateTo({ url: '/pages/seller/orders/index' })}
              >
                <Text className='action-title'>订单管理</Text>
                <Text className='action-arrow'>›</Text>
              </View>
            </View>
          </View>
        </>
      )}

      {/* ── Buyer section ──────────────────────────────── */}
      {role === 'buyer' && buyerProfile && (
        <>
          <View className='avatar-card card'>
            <View className='avatar-circle buyer-avatar'>订</View>
            <View className='avatar-info'>
              <Text className='avatar-name'>{buyerProfile.contactName || '未设置姓名'}</Text>
              <Text className='avatar-role-tag buyer-tag'>买家</Text>
            </View>
          </View>

          <View className='card'>
            <View className='row'>
              <Text className='section-title' style='margin-bottom:0'>默认配送信息</Text>
              {!editing && (
                <Text className='edit-link' onClick={() => setEditing(true)}>编辑</Text>
              )}
            </View>
            <View className='divider' />
            {editing ? (
              <View className='edit-form'>
                <View className='field'>
                  <Text className='field-label'>联系人</Text>
                  <Input
                    className='field-input'
                    value={buyerProfile.contactName}
                    onInput={(e) => setBuyerProfile({ ...buyerProfile, contactName: e.detail.value })}
                    placeholder='请输入姓名'
                  />
                </View>
                <View className='field'>
                  <Text className='field-label'>手机号</Text>
                  <Input
                    className='field-input'
                    type='number'
                    value={buyerProfile.contactPhone}
                    onInput={(e) => setBuyerProfile({ ...buyerProfile, contactPhone: e.detail.value })}
                    placeholder='请输入手机号'
                  />
                </View>
                <View className='field'>
                  <Text className='field-label'>收货地址</Text>
                  <Input
                    className='field-input'
                    value={buyerProfile.address}
                    onInput={(e) => setBuyerProfile({ ...buyerProfile, address: e.detail.value })}
                    placeholder='请输入详细地址'
                  />
                </View>
                <View className='edit-btns'>
                  <Button className='btn-outline' onClick={() => { setEditing(false); load() }}>取消</Button>
                  <Button className='btn-primary' loading={loading} onClick={handleSave}>保存</Button>
                </View>
              </View>
            ) : (
              <View className='profile-view'>
                <View className='profile-row'>
                  <Text className='profile-label'>联系人</Text>
                  <Text className='profile-val'>{buyerProfile.contactName || '未填写'}</Text>
                </View>
                <View className='profile-row'>
                  <Text className='profile-label'>手机号</Text>
                  <Text className='profile-val'>{buyerProfile.contactPhone || '未填写'}</Text>
                </View>
                <View className='profile-row'>
                  <Text className='profile-label'>地址</Text>
                  <Text className='profile-val'>{buyerProfile.address || '未填写'}</Text>
                </View>
              </View>
            )}
          </View>
        </>
      )}

      {/* ── Logout ─────────────────────────────────────── */}
      <Button className='logout-btn' onClick={handleLogout}>退出登录</Button>
    </View>
  )
}
