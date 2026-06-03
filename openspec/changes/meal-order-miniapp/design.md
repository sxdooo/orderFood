## Context

本项目为全新构建的订饭微信小程序，面向「今日下单、明日中午配送」的轻量餐饮场景。卖家可提前发布多日菜单，买家仅能购买明天的餐食。系统需在截单后汇总订单、规划配送路线，并支持订单内买卖双方沟通。

**技术约束（来自需求）：**
- 前端：Taro + React（微信小程序）
- 后端：Go
- 数据库：MySQL（主库）+ Redis（缓存/会话）
- 部署：阿里云 ECS

**业务关键约束：**
- 买家永远只能购买「明天」（配送日 = 下单日 + 1 天）的菜单
- 订单按配送日期隔离管理
- 卖家为「下单日」设置截单时间，到期自动停止接单
- 时区统一使用 `Asia/Shanghai`

## Goals / Non-Goals

**Goals:**

- 实现买家默认资料、菜单发布、明日下单、微信支付、截单控制、订单管理（含退单/地图）、配送路线规划、订单对话的完整业务闭环
- 前后端分离，API 清晰，支持买家/卖家角色权限隔离
- 可部署至单台阿里云 ECS（初期），架构可水平扩展
- 数据模型支持按配送日期高效查询与统计

**Non-Goals:**

- 多卖家入驻/平台化（首期一个店铺、一个卖家账号，但须完整支持买家与卖家两种角色）
- 复杂库存与促销系统
- 实时 WebSocket IM（首期采用轮询或长轮询）
- 买家修改已截单订单（需通过订单对话与卖家协商）

## Decisions

### 1. 数据库选型：MySQL 为主库

**选择：** MySQL 8.x

**理由：** 订单按日期查询、销量统计、关联查询（用户-订单-菜品）场景多，关系型模型更清晰；团队对 SQL 统计更熟悉。

**备选：** MongoDB 文档模型灵活，但复杂统计需额外聚合，首期不采用。

### 2. 项目结构：Monorepo 三端分离

```
orderFood/
├── miniapp/          # Taro + React 微信小程序
├── server/           # Go API 服务
├── deploy/           # Docker Compose / Nginx 配置
└── openspec/         # 规格与变更管理
```

**理由：** 前后端独立构建与部署，Go 服务通过 REST API 对外，便于 ECS 上用 Nginx 反向代理。

### 3. 认证与角色生命周期

- 小程序 `wx.login` 获取 code，后端调用微信 `jscode2session` 换取 OpenID
- 后端签发 JWT（存 Redis 黑名单支持登出），请求头 `Authorization: Bearer <token>`
- 用户表 `role` 字段：`buyer`（默认）/ `seller`，卖家由管理员后台设置或申请审核

**买家生命周期：**
1. 首次登录 → 检测 `profile_completed`，未完成则引导填写默认配送信息
2. 浏览明日菜单 → 下单（自动带入默认信息，可临时修改，不影响默认资料）
3. 微信支付 → 查看订单状态（待确认/已确认/配送中/已完成/已退单）
4. 与卖家订单内对话

**卖家生命周期：**
1. 发布某日菜单 → 设置截单时间
2. 查看某日订单列表 → 进入详情（地址、距离、地图双点）
3. 确认/配送/完成订单，或退单（填原因）
4. 截单后生成配送路线

### 4. 买家默认配送信息

```
buyer_profiles: user_id, contact_name, contact_phone, address, address_lat, address_lng,
                profile_completed, updated_at
```

- 首次登录后强制引导填写，完成后 `profile_completed = true`
- 「我的」页面可查看/修改，修改后仅影响后续下单的默认值
- 下单时将当前表单值**快照**写入 `orders` 表，历史订单不受默认信息变更影响

### 5. 卖家店铺位置

```
seller_profiles: user_id, shop_name, address, address_lat, address_lng
```

- 初始化卖家账号时配置店铺地址与坐标（用于订单详情距离计算与地图展示）

### 6. 日期与截单逻辑

- 所有业务日期使用 `DATE` 类型，时区 `Asia/Shanghai`
- **下单日（order_date）**：买家今天下单的日期
- **配送日（delivery_date）**：`order_date + 1 天`，订单归属日期
- **明日菜单**：`delivery_date = today + 1` 且 `status = published` 的菜单
- 截单时间存储在 `cutoff_settings` 表，按 `order_date` 配置；定时任务每分钟检查，到期后将当日接单状态写入 Redis `cutoff:{order_date}` = closed

### 7. 菜单与菜品模型

```
menus: id, seller_id, delivery_date, status(draft|published|expired), created_at
menu_items: id, menu_id, name, image_url, price, description, sort_order, is_available
```

- 卖家可提前创建未来多天的菜单
- 菜单过期：配送日过后自动标记 `expired`（定时任务）
- 已有订单的菜品下架：设置 `is_available = false`，不删除记录

### 8. 订单模型

```
orders: id, order_no, buyer_id, menu_id, delivery_date, total_amount, status,
        contact_name, contact_phone, address, address_lat, address_lng,
        delivery_time_pref, remark,
        refund_reason, refund_remark, refunded_at,
        created_at
order_items: id, order_id, menu_item_id, name_snapshot, price_snapshot, quantity
```

**状态流转：**
- `pending_payment` → `pending`（支付成功，待卖家确认）
- `pending` → `confirmed` → `delivering` → `completed`
- `refunded`（卖家退单，已支付则触发微信退款）
- `cancelled`（买家截单前主动取消）

订单联系人与地址字段为下单时**快照**，与 `buyer_profiles` 后续修改无关。

### 9. 卖家订单详情增强

- 展示买家地址文字、联系人、电话
- 根据卖家店铺坐标与订单 `address_lat/lng` 计算直线距离（高德距离 API 或 Haversine，展示如「约 2.3 km」）
- 小程序 `map` 组件展示卖家与买家两个标记点
- **退单**：卖家必填 `refund_reason`，可选 `refund_remark`；状态变为 `refunded`，已支付订单自动发起微信退款；买家在订单列表/详情看到「已退单」状态（后续可扩展订阅消息通知）

### 10. 配送路线规划

- 截单后卖家触发「生成路线」，后端收集当日所有订单地址（lat/lng）
- 调用**高德地图**「路径规划 API」（驾车路线 + 途经点优化）或「配送路线规划」
- 结果存储 `delivery_routes` 表：途经点顺序 JSON、总距离、预估时长
- 前端使用微信小程序 `map` 组件展示路线，支持卖家手动拖拽调整顺序后保存

**备选：** 百度地图 API，功能类似，按账号已有密钥选择。

### 11. 订单对话（轻量 IM）

- 每个订单一个聊天室 `order_messages`：`order_id, sender_id, sender_role, type(text|image), content, created_at`
- 首期采用**短轮询**（每 5s 拉取新消息），降低 WebSocket 运维复杂度
- 图片上传至阿里云 OSS（与 ECS 同账号开通，按量计费），消息存 URL

### 12. 微信支付

- 下单后调用微信支付 JSAPI（小程序支付），订单状态 `pending_payment` → 支付成功 `pending`（待卖家确认）
- 后端使用微信支付 API v3：统一下单、支付回调通知、退款（取消订单时）
- 需申请：**微信支付商户号**（与小程序 AppID 绑定），配置 API 证书与回调 URL
- 支付记录表 `payments`：`order_id, transaction_id, amount, status, paid_at`

### 13. Redis 用途

| Key 模式 | 用途 |
|----------|------|
| `session:{user_id}` | JWT 会话 / 登出黑名单 |
| `cutoff:{order_date}` | 当日截单状态缓存 |
| `menu:tomorrow` | 明日菜单缓存（TTL 5min） |
| `chat:unread:{order_id}:{user_id}` | 未读消息计数 |

### 14. API 风格

- RESTful，`/api/v1/` 前缀
- 统一响应：`{ code, message, data }`
- 卖家接口需 `role=seller` 中间件校验
- 核心端点见下方架构图

### 15. 部署架构（阿里云 ECS）

```
                    ┌─────────────┐
  微信小程序 ──────►│   Nginx     │
                    │  (443/80)   │
                    └──────┬──────┘
                           │
              ┌────────────┼────────────┐
              ▼            ▼            ▼
         Go API       静态资源      (未来扩展)
              │
       ┌──────┴──────┐
       ▼             ▼
    MySQL 8       Redis 7
```

- Docker Compose 编排：nginx + api + mysql + redis
- 环境变量管理敏感配置（微信 AppID/Secret、地图 API Key、OSS 密钥）
- 定时任务：Go 内置 cron（`robfig/cron`）处理截单与菜单过期

## Risks / Trade-offs

| 风险 | 缓解措施 |
|------|----------|
| 地图 API 配额/费用 | 仅截单后调用一次路线规划；缓存结果 |
| 地址无经纬度导致路线失败 | 下单时调用地理编码 API 补全；无坐标订单单独列表提示 |
| 轮询 IM 延迟与流量 | 5s 间隔可接受；后期可升级 WebSocket |
| 单 ECS 性能瓶颈 | 初期用户量小足够；MySQL/Redis 可迁 RDS |
| 时区/夏令时问题 | 固定 Asia/Shanghai，不使用 DST |
| 截单定时任务延迟 | 每分钟检查 + Redis 实时拦截下单请求 |

## Migration Plan

本项目为全新构建，无存量数据迁移。

**部署步骤：**
1. ECS 安装 Docker + Docker Compose
2. 配置环境变量与 SSL 证书（微信小程序要求 HTTPS）
3. `docker compose up -d` 启动 MySQL、Redis、API、Nginx
4. 执行数据库迁移（`golang-migrate` 或 GORM AutoMigrate）
5. 微信公众平台配置服务器域名、上传小程序代码审核发布
6. 初始化首个卖家账号（数据库或管理脚本）

**回滚：** Docker 镜像版本回退 + 数据库备份恢复

## Resolved Decisions（已确认）

| 问题 | 决策 |
|------|------|
| 卖家模式 | **单店铺**：一个卖家经营，系统完整支持**买家**与**卖家**两种角色；用户默认买家，卖家由后台授权。非多卖家入驻平台。 |
| 支付 | **需要微信支付**（小程序 JSAPI），下单后在线支付，支付成功订单才进入待确认流程。 |
| 地图 API | **高德地图**，密钥待申请（见下方「第三方账号申请清单」）。 |
| 图片存储 | **使用阿里云 OSS**。OSS 与 ECS 是独立产品，但可用同一阿里云账号按量开通；小流量场景月费通常几元，新用户常有免费额度。 |

## 第三方账号申请清单

### 高德开放平台（lbs.amap.com）

1. 注册开发者账号 → 进入「应用管理」→ 创建应用（类型选「微信小程序」+「Web 服务」各建一个 Key，或按平台指引配置）
2. 需开通的 **Web 服务 API**（服务端 Key，绑定 ECS 公网 IP 或留空开发期）：
   - **地理编码**：地址 → 经纬度（下单时补全坐标）
   - **路径规划 2.0**（驾车）：多途经点配送路线优化
   - **距离测量**（可选）：订单详情页卖家到买家距离展示
3. 小程序端 `map` 组件展示路线时，在微信公众平台配置「位置接口」权限，并按高德文档配置小程序 Key（若使用高德地图插件）
4. 个人开发者有每日免费调用额度；商用需企业认证并关注配额

### 微信支付

1. 注册 [微信支付商户平台](https://pay.weixin.qq.com/)（需营业执照，个体户/企业均可）
2. 商户号与小程序 AppID 绑定（微信商户平台 → 产品中心 → AppID 账号管理）
3. 开通「JSAPI 支付」产品
4. 下载 API 证书，配置支付回调域名（须 HTTPS，与 ECS 域名一致）
5. 开发阶段可用微信支付沙箱（部分接口）或 0.01 元实付联调

### 阿里云 OSS

1. 登录与 ECS **同一阿里云账号** → 对象存储 OSS → 开通服务（无需单独买服务器）
2. 创建 Bucket（地域建议与 ECS 同区域，如华东），读写权限「私有」，通过后端签名 URL 访问
3. 创建 RAM 子账号 + AccessKey，仅授予该 Bucket 读写权限（勿把 Key 写进前端）
4. **费用**：按存储量 + 流量计费；订饭小程序图片量小，一般 **每月几元～十几元**；新用户常有 6～12 个月免费额度（以官网活动为准）
5. **仅有 ECS 的替代方案**（不推荐生产）：图片存 ECS 本地磁盘 + Nginx 静态目录——占用系统盘、无 CDN、扩容和备份麻烦，仅适合本地开发演示
