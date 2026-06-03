## 1. 项目脚手架与基础设施

- [x] 1.1 初始化 Monorepo 目录结构（`miniapp/`、`server/`、`deploy/`）
- [x] 1.2 初始化 Go 后端项目（Gin/Fiber 框架、配置管理、统一响应格式）
- [x] 1.3 初始化 Taro + React 微信小程序项目
- [x] 1.4 配置 Docker Compose（MySQL 8、Redis 7、Nginx、API 服务）
- [x] 1.5 编写数据库迁移脚本（users、buyer_profiles、seller_profiles、menus、menu_items、orders、order_items、payments、cutoff_settings、order_messages、delivery_routes）
- [x] 1.6 配置环境变量模板（微信 AppID/Secret、微信支付商户号/证书、JWT Secret、高德 API Key、OSS AccessKey）

## 2. 用户认证（user-auth）

- [x] 2.1 实现微信 code2session 登录接口 `POST /api/v1/auth/wechat`
- [x] 2.2 实现 JWT 签发与验证中间件
- [x] 2.3 实现 Redis 会话管理与登出黑名单
- [x] 2.4 实现角色中间件（buyer/seller 权限校验）
- [x] 2.5 小程序端：登录页、`wx.login` 调用、Token 存储与请求拦截器
- [x] 2.6 编写卖家账号初始化脚本（含 seller_profiles 店铺地址与坐标）

## 3. 买家资料（buyer-profile）

- [x] 3.1 实现买家资料 API `GET/PUT /api/v1/buyer/profile`
- [x] 3.2 实现首次登录资料完成状态检测（`profile_completed`）
- [x] 3.3 实现地址地理编码（保存默认资料时补全 lat/lng）
- [x] 3.4 买家端：首次引导页（地址、联系人、电话）
- [x] 3.5 买家端：「我的」页面资料查看与编辑
- [x] 3.6 下单页自动带入默认资料，支持单次覆盖（不更新默认资料）

## 4. 菜单管理（menu-management）

- [x] 4.1 实现菜单 CRUD API（按 delivery_date 创建/编辑/查询）
- [x] 4.2 实现菜品 CRUD API（含软下架 `is_available`）
- [x] 4.3 实现菜单发布/状态流转（draft → published → expired）
- [x] 4.4 实现菜单过期定时任务（配送日后自动 expired）
- [x] 4.5 卖家端：菜单列表页（按日期浏览、状态标签）
- [x] 4.6 卖家端：菜单编辑页（菜品增删改、发布按钮）
- [ ] 4.7 实现图片上传接口（阿里云 OSS）— **暂缓**

## 5. 截单控制（cutoff-control）

- [x] 5.1 实现截单时间设置 API `PUT /api/v1/seller/cutoff`
- [x] 5.2 实现截单状态查询 API `GET /api/v1/cutoff/status`
- [x] 5.3 实现截单定时检查任务（每分钟检查 + Redis 缓存截单状态）
- [x] 5.4 实现下单拦截中间件（截单后拒绝新订单）
- [x] 5.5 卖家端：截单时间设置页面
- [x] 5.6 买家端：菜单页展示截单倒计时/已截单状态

## 6. 下单模块（order-placement）

- [x] 6.1 实现明日菜单查询 API `GET /api/v1/menus/tomorrow`（仅 published + 明日日期）
- [x] 6.2 实现下单 API `POST /api/v1/orders`（校验截单、快照菜品与配送信息，跳过支付直接 `pending`）
- [x] 6.3 实现买家取消订单 API（截单前可取消）
- [x] 6.4 实现买家订单列表/详情 API
- [x] 6.5 买家端：明日菜单浏览页（菜品列表、购物车）
- [x] 6.6 买家端：下单确认页（默认资料带入、可临时修改、备注）
- [x] 6.7 买家端：我的订单页（按配送日期分组、含已退单状态）

## 7. 微信支付（wechat-pay）— **暂缓**

- [ ] 7.1 集成微信支付 API v3（统一下单、签名验证）
- [ ] 7.2 实现支付下单接口 `POST /api/v1/orders/:id/pay`（返回小程序支付参数）
- [ ] 7.3 实现支付回调通知接口 `POST /api/v1/pay/notify`（验签、更新订单状态）
- [ ] 7.4 实现退款接口（买家截单前取消、卖家退单）
- [ ] 7.5 实现支付超时自动取消订单定时任务
- [ ] 7.6 买家端：下单后调起 `wx.requestPayment` 支付流程
- [ ] 7.7 买家端：支付结果页（成功/失败/重试）

## 8. 订单管理（order-management）

- [x] 8.1 实现卖家按配送日期查询订单 API
- [x] 8.2 实现卖家订单详情 API（含距离计算、卖家/买家坐标）
- [x] 8.3 实现卖家更新订单状态 API（confirm/delivering/completed）
- [x] 8.4 实现卖家退单 API `POST /api/v1/seller/orders/:id/refund`（必填原因、可选备注）
- [x] 8.5 实现卖家日销售统计 API（订单数、总金额、状态分布）
- [x] 8.6 卖家端：订单列表页（日期选择器、订单卡片）
- [x] 8.7 卖家端：订单详情页（地址、距离、地图双点、退单入口）
- [x] 8.8 卖家端：退单弹窗（原因必填、备注选填）
- [ ] 8.9 卖家端：日统计面板

## 9. 配送路线规划（delivery-routing）

- [x] 9.1 实现配送地址列表 API（按配送日汇总所有订单地址）
- [x] 9.2 集成路线规划（最近邻启发式 + 直线距离）
- [x] 9.3 实现路线生成与保存 API `POST /api/v1/seller/routes`
- [x] 9.4 实现路线查询与手动调整顺序 API
- [x] 9.5 卖家端：配送路线页（地图组件、标记点、路线折线）
- [ ] 9.6 卖家端：手动拖拽调整配送顺序并保存

## 10. 订单对话（order-chat）

- [x] 10.1 实现发送消息 API `POST /api/v1/orders/:id/messages`（文本）
- [x] 10.2 实现消息历史查询 API `GET /api/v1/orders/:id/messages`
- [x] 10.3 实现消息轮询 API（`since` 参数增量拉取）
- [ ] 10.4 实现未读消息计数（Redis + API）
- [x] 10.5 买家端：订单聊天页（消息列表、发送文本、轮询刷新）
- [x] 10.6 卖家端：订单聊天页（复用订单详情对话页）
- [ ] 10.7 订单列表未读消息角标

## 11. 前端公共与角色路由

- [x] 11.1 实现买家/卖家入口（我的页卖家中心，按手机号白名单识别角色）
- [x] 11.2 封装公共组件（加载态、空状态、错误提示、价格格式化）
- [x] 11.3 统一 API 请求层（错误处理、Token 刷新、401 跳转登录）
- [x] 11.4 首次登录资料未完成时全局路由拦截

## 12. 部署与上线

- [ ] 12.1 编写 Nginx 配置（HTTPS 反向代理、静态资源）
- [ ] 12.2 编写生产环境 Docker Compose 与启动脚本
- [ ] 12.3 配置阿里云 ECS 部署（安全组、域名、SSL 证书）
- [ ] 12.4 微信公众平台配置服务器域名白名单 + 开通支付相关权限
- [ ] 12.5 端到端联调测试（资料引导 → 发菜单 → 下单 → 支付 → 退单 → 截单 → 路线 → 对话）
- [ ] 12.6 小程序提交审核与发布
