# RelayDeck - 统一管理平台

## 项目简介

RelayDeck 是一个统一管理多个大模型站点（基于 new-api 或 sub2api）的平台。通过集中管理，智能路由，自动将客户端请求转发到对应的上游站点，降低管理成本和提高服务可用性。

## 核心功能

- 🌐 **站点管理** - 统一管理多个上游站点，实时监控状态
- 🤖 **模型管理** - 配置和管理各种 AI 模型，灵活绑定站点
- 📊 **统计分析** - 详细的数据分析和成本统计
- 🔑 **密钥管理** - API 密钥生成和配额管理
- 📝 **日志管理** - 完整的请求日志记录和查询
- ⚠️ **告警通知** - 实时告警和消息通知
- 👥 **用户管理** - 多用户支持和权限管理
- ⚙️ **系统设置** - 灵活的路由策略和负载均衡配置

## 技术架构

### 前端
- **框架**: React 18 + TypeScript
- **构建工具**: Vite 5
- **UI**: Tailwind CSS + 自研控制台组件
- **图标**: lucide-react
- **路由**: 当前原型使用应用内页面状态，后续接入真实 API 时再引入 URL 路由
- **主题**: 支持亮色/暗色切换

### 后端
- **语言**: Go
- **HTTP**: 标准库 `net/http`
- **数据库**: PostgreSQL 作为主存储；当前已支持管理端用户、上游站点、站点状态、模型同步结果和操作历史持久化
- **缓存**: Redis 用于管理端 session；后续继续承载分布式限流和共享熔断状态
- **API**: OpenAI 兼容 `/v1/*` 网关 API + RelayDeck 管理端 `/api/admin/*`

## 项目结构

```
RelayDeck/
├── Design_image/      # UI 设计稿
│   ├── 1.PNG         # 仪表盘（暗色）
│   ├── 2.PNG         # 登录页
│   ├── 3.PNG         # 站点管理
│   ├── 4.PNG         # 模型管理
│   ├── 5.PNG         # 统计分析（亮色）
│   ├── 6.PNG         # 统计分析（暗色）
│   ├── 7.PNG         # 消息通知
│   └── test.png      # 测试 API 调用弹窗原型
├── src/               # 当前主线前端原型
├── backend/           # Go 后端服务
├── docs/              # 设计和实施计划
└── package.json       # 前端开发脚本
```

## 编译与运行

### 安装前端依赖

```bash
npm install
```

### 启动基础设施

本地 PostgreSQL 和 Redis 使用 Docker Compose 启动：

```bash
docker compose up -d postgres redis
```

`docker-compose.yml` 默认使用镜像源地址，避免本地开发时直接拉取 Docker Hub 超时。

### 启动前端开发服务器

```bash
npm run dev
```

访问 Vite 输出的本地地址，默认通常是 http://localhost:5173。前端开发服务器会把 `/api` 和 `/v1` 请求代理到后端。

### 启动后端开发服务器

后端依赖 PostgreSQL、Redis 和根目录 `.env` 中的配置：

```bash
cd backend
go run ./cmd/relaydeck
```

默认后端监听 `http://localhost:8080`。启动后可检查健康接口：

```bash
curl http://localhost:8080/healthz
```

### 生产构建

前端构建：

```bash
npm run build
```

构建产物输出到 `dist/`。

后端编译：

```bash
cd backend
go build -o relaydeck ./cmd/relaydeck
```

本地编译出的 `backend/relaydeck` 是机器相关二进制文件，已加入 `.gitignore`，不提交到仓库。运行编译产物：

```bash
cd backend
./relaydeck
```

### 环境变量

项目根目录提供 `.env.example` 作为唯一环境变量模板。本地开发或集成测试时复制为 `.env` 并填写真实值：

```bash
cp .env.example .env
```

`.env` 不提交到 Git。后端启动和 Go 集成测试会自动读取 `.env` 中的 `DATABASE_URL`、`REDIS_URL`、`APP_BOOTSTRAP_OWNER_EMAIL` 等变量。

### 获取 New API 账号凭据

RelayDeck 的站点管理中，如果账号凭据类型选择 `New API Access Token`，需要填写上游 New API 账号的 `Access Token` 和 `User ID`。这两个值来自上游 New API 站点，不是在 RelayDeck 里生成。

1. 打开上游 New API 站点并完成登录。
2. 在该上游站点页面打开浏览器开发者工具 Console。
3. 如果浏览器提示禁止粘贴代码，先确认代码内容无误，再按提示输入 `允许粘贴`。
4. 执行下面代码：

```js
const uid = localStorage.getItem('uid') || JSON.parse(localStorage.getItem('user') || '{}').id;
console.log('user_id:', uid);

const tokenRes = await fetch('/api/user/token', {
  headers: { 'New-Api-User': String(uid) },
  credentials: 'include',
}).then((response) => response.json());

console.log('access_token:', tokenRes.data);
```

源码行为说明：

- New API 前端会从 `localStorage.uid` 读取当前用户 ID，并作为 `New-Api-User` 请求头发送。
- `GET /api/user/token` 的返回格式是 `{ "success": true, "data": "<access-token>" }`，所以 token 在 `data` 字段里。
- 这段代码必须在“上游 New API 站点”的 Console 中执行，不要在 RelayDeck 页面执行。

### 站点管理数据说明

站点管理数据使用 PostgreSQL 持久化，不写入 Redis。Redis 当前只保存管理端登录 session。

| 数据 | PostgreSQL 表 | 说明 |
|------|---------------|------|
| 站点配置 | `upstream_accounts` | 名称、代码、平台、Base URL、API Key 密文、账号凭据密文和自动化开关 |
| 状态快照 | `upstream_account_status` | API 状态、账号凭据状态、模型数量、延迟、余额、最近刷新时间和错误信息 |
| 已同步模型 | `upstream_synced_models` | 站点模型列表，来自同步或全量刷新动作 |
| 操作历史 | `upstream_account_events` | 检测、同步、刷新额度、签到和全量刷新等操作记录 |

站点列表中的模型数量、测试调用弹窗中的模型下拉，都读取数据库中的已同步模型。需要更新模型列表时，先对站点执行“全量刷新”或模型同步动作。

New API 额度按上游 `/api/status` 返回的 `quota_per_unit` 换算为 USD；如果上游未返回该值，则使用 New API 默认换算 `500000 quota = 1 USD`。

“测试 API”按钮会打开测试调用弹窗，并通过当前站点 API Key 向上游发送真实请求。RelayDeck 不会把该调用写入自身统计；实际上游是否计费或扣量以上游平台规则为准。

### 功能页面

| 页面 | 功能 |
|------|------|
| 登录 | 管理端登录入口，当前为前端原型状态 |
| 概览 | 数据概览、趋势图表、站点状态 |
| 站点管理 | 管理上游站点，支持分页、批量刷新、删除确认、测试调用、模型同步结果和状态快照 |
| 模型管理 | 管理 AI 模型和站点映射 |
| 智能路由 | 配置候选线路和路由策略 |
| 签到中心 / 额度管理 | 管理站点签到和额度状态 |
| 调用测试 | 验证模型和上游线路可用性 |
| 用户管理 | 管理后台用户、权限和配额 |
| API Keys | 管理下游调用凭证 |
| 任务日志 | 请求日志查询 |
| 系统设置 | 系统配置 |

## 设计特色

- 🎨 **现代化 UI** - 基于 Tailwind CSS 和控制台组件体系
- 🌓 **主题切换** - 支持亮色/暗色主题无缝切换
- 📊 **数据可视化** - 丰富的图表展示（折线图、环形图、进度条等）
- 💫 **流畅体验** - 热更新、平滑过渡动画
- 📱 **响应式设计** - 适配不同屏幕尺寸

## 工程原则

- 默认采用业内主流、可生产演进的实现方式。
- 本地内存 fallback 仅用于开发便利，生产路径应优先使用 PostgreSQL、Redis 等明确的基础设施。
- 临时简化必须在文档或代码边界中标明，后续实施应优先替换为正式实现。

## 开发计划

### 已完成 ✅
- [x] 前端 UI 框架搭建
- [x] 10 个核心页面开发
- [x] 主题切换功能
- [x] 基础路由配置
- [x] 组件和样式设计

### 进行中 🚧
- [ ] 后端 API 开发
- [x] 数据库设计草案
- [x] 管理端登录/session MVP
- [x] PostgreSQL 管理员用户持久化
- [x] Redis session 存储
- [x] 平台账号聚合与原生协议路由设计
- [x] 上游站点 PostgreSQL 持久化
- [x] 站点分页、批量操作和测试调用弹窗
- [ ] 网关配置持久化
- [ ] new-api/sub2api 账号聚合和模型同步完善
- [ ] 原生协议优先路由与协议转换
- [ ] 站点健康检查

### 计划中 📋
- [ ] WebSocket 实时推送
- [ ] 告警规则引擎
- [ ] 数据导出功能
- [ ] Docker 部署
- [ ] 多语言支持

## 应用场景

1. **多站点统一管理** - 管理多个 OpenAI API 中转站点
2. **智能负载均衡** - 根据站点状态智能路由请求
3. **成本优化** - 实时监控成本，优化使用策略
4. **高可用保障** - 自动故障切换，提高服务可用性
5. **集中监控** - 统一的日志和监控面板

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License
