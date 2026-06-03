# OrderFood - 订饭小程序

Monorepo for meal ordering WeChat mini-program.

```
orderFood/
├── miniapp/    # Taro + React WeChat mini-program
├── server/     # Go API service
├── deploy/     # Docker Compose & Nginx
└── openspec/   # Specifications & change management
```

## Quick Start

### Backend (local)

```bash
cd server
cp .env.example .env
go run ./cmd/api
```

### Miniapp (local)

```bash
cd miniapp
npm install
npm run dev:weapp
```

### 初始化卖家手机号

```bash
cd server
go run ./scripts/init_seller_phone.go 13800138000 "我的小店"
```

买家资料中填写该手机号并保存后，自动获得卖家角色，「我的」页出现卖家中心。

### 重启 API

修改后端后需重启 `go run ./cmd/api`（若 8080 被占用先 `taskkill /IM api.exe /F`）。

```bash
cd deploy
# 仅启动 MySQL + Redis（本地 go run API 时使用）
docker compose up -d mysql redis
# 或启动全部服务
docker compose up -d
```

**Docker Hub 拉取超时（TLS handshake timeout）**

国内网络常见。任选其一：

1. **使用项目内置镜像加速**（已配置 `deploy/.env` 使用 DaoCloud 镜像）：
   ```bash
   cd deploy
   docker compose up -d mysql redis
   ```

2. **配置 Docker Desktop 镜像加速**：设置 → Docker Engine，在 JSON 中加入：
   ```json
   "registry-mirrors": [
     "https://docker.m.daocloud.io",
     "https://docker.1panel.live"
   ]
   ```
   保存并重启 Docker，再执行 `docker compose up -d`。

3. **不用 Docker**：本机安装 MySQL 8 + Redis，保持 `server/.env` 中 `localhost` 即可。
