import Taro from '@tarojs/taro'

export interface CartItem {
  menuItemId: number
  name: string
  price: number
  quantity: number
}

const CART_KEY = 'cart_items'

export function getCart(): CartItem[] {
  return Taro.getStorageSync(CART_KEY) || []
}

export function setCart(items: CartItem[]): void {
  Taro.setStorageSync(CART_KEY, items)
}

export function clearCart(): void {
  Taro.removeStorageSync(CART_KEY)
}

export function cartTotal(items: CartItem[]): number {
  return items.reduce((s, i) => s + i.price * i.quantity, 0)
}
