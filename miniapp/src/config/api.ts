// API base URL for WeChat mini program requests.
//
// Production (real users): MUST be an HTTPS domain that is added to the
// mini-program's request 合法域名 whitelist in the WeChat MP console, e.g.
//   https://api.your-domain.com/api/v1
//
// Development:
// - DevTools simulator: 127.0.0.1 works because it runs on the same machine.
// - Real device / preview: use the LAN IP of the dev machine and enable
//   详情 → 本地设置 → 不校验合法域名 + 不校验HTTPS证书.
//
// Switch by build mode: `taro build --type weapp` (prod) vs `--watch` (dev).
const PROD_API_BASE_URL = 'https://your-domain.com/api/v1'
const DEV_API_BASE_URL = 'http://192.168.1.9:8080/api/v1'

export const API_BASE_URL =
  process.env.NODE_ENV === 'production' ? PROD_API_BASE_URL : DEV_API_BASE_URL
