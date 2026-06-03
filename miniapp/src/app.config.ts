export default defineAppConfig({
  pages: [
    'pages/index/index',
    'pages/login/index',
    'pages/profile-setup/index',
    'pages/menu/index',
    'pages/checkout/index',
    'pages/orders/index',
    'pages/order-detail/index',
    'pages/mine/index',
    'pages/seller/menus/index',
    'pages/seller/orders/index',
    'pages/seller/order-detail/index',
    'pages/seller/route/index'
  ],
  window: {
    backgroundTextStyle: 'light',
    navigationBarBackgroundColor: '#fff',
    navigationBarTitleText: '订饭',
    navigationBarTextStyle: 'black'
  },
  tabBar: {
    color: '#666',
    selectedColor: '#07c160',
    list: [
      { pagePath: 'pages/menu/index', text: '订餐' },
      { pagePath: 'pages/orders/index', text: '订单' },
      { pagePath: 'pages/mine/index', text: '我的' }
    ]
  },
  permission: {
    'scope.userLocation': {
      desc: '用于展示配送地图'
    }
  }
})

function defineAppConfig(config: Taro.Config) {
  return config
}
