export function formatPrice(cents: number): string {
  return (cents / 100).toFixed(2)
}

export const ORDER_STATUS_LABEL: Record<string, string> = {
  pending: '待确认',
  confirmed: '已确认',
  delivering: '配送中',
  completed: '已完成',
  refunded: '已退单',
  cancelled: '已取消'
}
