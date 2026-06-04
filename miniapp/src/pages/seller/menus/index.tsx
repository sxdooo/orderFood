import { View, Input, Button, Picker, Text, ScrollView } from '@tarojs/components'
import Taro, { useDidShow } from '@tarojs/taro'
import { useState } from 'react'
import { request } from '../../../utils/request'
import { formatPrice } from '../../../utils/format'
import './index.scss'

interface MenuItem {
  name: string
  price: number
}

interface Menu {
  id: number
  deliveryDate: string
  status: string
  items: MenuItem[]
}

const getTomorrow = () => {
  const d = new Date()
  d.setDate(d.getDate() + 1)
  return d.toISOString().slice(0, 10)
}

const STATUS_FILTERS: { key: string; label: string }[] = [
  { key: 'all', label: '全部' },
  { key: 'published', label: '已发布' },
  { key: 'draft', label: '草稿' },
  { key: 'expired', label: '已过期' },
]

export default function SellerMenusPage() {
  const [menus, setMenus] = useState<Menu[]>([])
  const [deliveryDate, setDeliveryDate] = useState(getTomorrow())
  const [dishName, setDishName] = useState('')
  const [dishPrice, setDishPrice] = useState('')
  const [dishDesc, setDishDesc] = useState('')
  const [items, setItems] = useState<MenuItem[]>([])
  const [creating, setCreating] = useState(false)
  const [filterStatus, setFilterStatus] = useState('all')

  const load = async () => {
    try {
      const data = await request<Menu[]>({ url: '/seller/menus' })
      setMenus(data || [])
    } catch { /* ignore */ }
  }

  useDidShow(() => {
    Taro.setNavigationBarTitle({ title: '菜单管理' })
    load()
  })

  const addDish = () => {
    const name = dishName.trim()
    const price = parseFloat(dishPrice)
    if (!name) {
      Taro.showToast({ title: '请输入菜名', icon: 'none' })
      return
    }
    if (isNaN(price) || price <= 0) {
      Taro.showToast({ title: '请输入有效价格', icon: 'none' })
      return
    }
    setItems([...items, { name, price: Math.round(price * 100) }])
    setDishName('')
    setDishPrice('')
    setDishDesc('')
  }

  const removeDish = (idx: number) => {
    setItems(items.filter((_, i) => i !== idx))
  }

  const createMenu = async () => {
    if (items.length === 0) {
      Taro.showToast({ title: '请至少添加一道菜', icon: 'none' })
      return
    }
    setCreating(true)
    try {
      const menu = await request<Menu>({
        url: '/seller/menus',
        method: 'POST',
        data: {
          deliveryDate,
          items: items.map((it, i) => ({ name: it.name, price: it.price, sortOrder: i }))
        }
      })
      await request({ url: `/seller/menus/${menu.id}/publish`, method: 'POST' })
      Taro.showToast({ title: '菜单已发布', icon: 'success' })
      setItems([])
      setDeliveryDate(getTomorrow())
      load()
    } catch (err) {
      Taro.showToast({ title: err instanceof Error ? err.message : '发布失败', icon: 'none' })
    } finally {
      setCreating(false)
    }
  }

  const STATUS_LABEL: Record<string, string> = {
    draft: '草稿', published: '已发布', expired: '已过期'
  }

  const filteredMenus = filterStatus === 'all'
    ? menus
    : menus.filter((m) => m.status === filterStatus)

  return (
    <View className='container menus-page'>
      {/* Create form */}
      <View className='card'>
        <View className='section-title'>发布新菜单</View>

        {/* Date picker */}
        <View className='field'>
          <View className='field-label'>配送日期</View>
          <Picker
            mode='date'
            value={deliveryDate}
            start={getTomorrow()}
            onChange={(e) => setDeliveryDate(String(e.detail.value))}
          >
            <View className='picker-val'>
              <Text>{deliveryDate}</Text>
              <Text className='picker-arrow'>›</Text>
            </View>
          </Picker>
        </View>

        <View className='divider' />

        {/* Dish form */}
        <View className='dish-form-title'>添加菜品</View>
        <View className='dish-inputs'>
          <Input
            className='dish-name-input'
            placeholder='菜品名称'
            value={dishName}
            onInput={(e) => setDishName(e.detail.value)}
          />
          <View className='price-input-wrap'>
            <Text className='price-prefix'>¥</Text>
            <Input
              type='digit'
              placeholder='价格'
              value={dishPrice}
              onInput={(e) => setDishPrice(e.detail.value)}
            />
          </View>
        </View>
        <Button className='btn-add-dish' onClick={addDish}>＋ 加入菜单</Button>

        {/* Dish list */}
        {items.length > 0 && (
          <View className='dish-list'>
            {items.map((it, i) => (
              <View key={i} className='dish-item'>
                <View className='dish-item-info'>
                  <Text className='dish-item-name'>{it.name}</Text>
                  <Text className='dish-item-price'>¥{formatPrice(it.price)}</Text>
                </View>
                <Button className='btn-remove' size='mini' onClick={() => removeDish(i)}>删除</Button>
              </View>
            ))}
          </View>
        )}

        <Button
          className='btn-primary publish-btn'
          loading={creating}
          onClick={createMenu}
          disabled={items.length === 0}
        >
          创建并发布 ({items.length} 道菜)
        </Button>
      </View>

      {/* Existing menus */}
      {menus.length > 0 && (
        <>
          <View className='section-title'>已有菜单</View>

          {/* 状态筛选 Tab（通用小程序筛选交互） */}
          <ScrollView scrollX className='filter-tabs' enableFlex>
            {STATUS_FILTERS.map((f) => {
              const cnt = f.key === 'all'
                ? menus.length
                : menus.filter((m) => m.status === f.key).length
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

          {filteredMenus.map((m) => (
            <View key={m.id} className='card menu-history-card'>
              <View className='row'>
                <Text className='menu-date'>{String(m.deliveryDate).slice(0, 10)}</Text>
                <Text className={`status-badge ${m.status === 'published' ? 's-confirmed' : 's-cancelled'}`}>
                  {STATUS_LABEL[m.status] || m.status}
                </Text>
              </View>
              <View className='menu-items-list'>
                {m.items?.map((it, i) => (
                  <Text key={i} className='menu-item-chip'>
                    {it.name} ¥{formatPrice(it.price)}
                  </Text>
                ))}
              </View>
            </View>
          ))}

          {filteredMenus.length === 0 && (
            <View className='empty-filter-hint'>该状态下暂无菜单</View>
          )}
        </>
      )}
    </View>
  )
}
