import { View } from '@tarojs/components'
import Taro, { useLoad } from '@tarojs/taro'
import { getToken } from '../../utils/auth'
import { request } from '../../utils/request'

// Single entry: avoid concurrent reLaunch + switchTab (webviewId route errors)
export default function IndexPage() {
  useLoad(async () => {
    const token = getToken()
    if (!token) {
      await Taro.reLaunch({ url: '/pages/login/index' })
      return
    }
    try {
      const profile = await request<{ profileCompleted: boolean }>({ url: '/buyer/profile' })
      if (!profile.profileCompleted) {
        await Taro.reLaunch({ url: '/pages/profile-setup/index' })
        return
      }
      await Taro.switchTab({ url: '/pages/menu/index' })
    } catch {
      await Taro.reLaunch({ url: '/pages/login/index' })
    }
  })

  return <View className='container'>加载中...</View>
}
