# Sora2API

OpenAI Sora 视频/图片生成 API 代理平台，提供 Go SDK、Web 管理后台和兼容 API 接口。

## 功能概览

**API 代理**
- 视频生成（文生视频 / 图生视频 / Remix / 分镜视频）
- 图片生成（文生图 / 图生图）
- 角色一致性（创建 / 管理角色）
- 提示词增强、视频发布、去水印下载
- 10 种视频风格（anime、retro、comic 等）
- API Key 鉴权，多账号分组轮询

**Web 管理后台**
- 仪表板（账号/任务/角色状态统计）
- 账号管理（分组、Token 刷新、配额同步）
- API Key 管理
- 任务列表与详情查看
- 角色管理
- 系统设置（代理、同步间隔等）
- 内置 API 文档页

**Go SDK**
- 完整的 Sora API 封装
- TLS 指纹模拟绕过 Cloudflare
- 代理支持、`context.Context` 超时控制、进度回调

## 快速部署

### 一键安装（推荐）

```bash
curl -sSL https://raw.githubusercontent.com/DouDOU-start/go-sora2api/master/deploy/install.sh | sudo bash
```

安装过程中会交互式配置：
- 服务监听地址和端口（默认 `0.0.0.0:8686`）
- PostgreSQL 数据库连接
- 管理员账号密码

安装完成后通过 `sora2api` 命令管理服务：

```bash
sora2api status          # 查看服务状态
sora2api start           # 启动服务
sora2api stop            # 停止服务
sora2api restart         # 重启服务
sora2api logs            # 查看最近日志
sora2api logs -f         # 实时跟踪日志
sora2api config          # 编辑配置文件
sora2api version         # 查看当前版本
sora2api upgrade         # 升级到最新版本
sora2api upgrade -v 1.3  # 升级到指定版本
sora2api list-versions   # 查看可用版本
sora2api uninstall       # 卸载
```

### 手动编译

```bash
# 安装依赖
make install

# 构建（含前端）
make build

# 启动
./bin/sora2api-server
```

### 开发模式

```bash
make dev  # 同时启动后端 (8686) + 前端 (5173)
```

## 配置

配置文件 `/etc/sora2api/config.yaml`（一键安装）或 `server/config.yaml`（手动编译）：

```yaml
server:
  host: "0.0.0.0"
  port: 8686
  admin_user: "admin"
  admin_password: "admin123"
  # jwt_secret: ""  # 留空则自动生成

database:
  url: "postgres://postgres:postgres@localhost:5432/sora2api?sslmode=disable"
  log_level: "warn"
  auto_migrate: true
```

支持环境变量覆盖：
- `CONFIG_PATH` — 配置文件路径
- `DATABASE_URL` — 数据库连接串

## Go SDK

### 安装

```bash
go get github.com/DouDOU-start/go-sora2api/sora
```

### 示例

```go
ctx := context.Background()
c, _ := sora.New("") // 可选代理

// 文生视频
token, _ := c.GenerateSentinelToken(ctx, accessToken)
taskID, _ := c.CreateVideoTask(ctx, accessToken, token, "a cat running", "landscape", 300, "sy_8", "small")
_ = c.PollVideoTask(ctx, accessToken, taskID, 3*time.Second, 600*time.Second, nil)
url, _ := c.GetDownloadURL(ctx, accessToken, taskID)

// 文生图
taskID, _ = c.CreateImageTask(ctx, accessToken, token, "a cute cat", 360, 360)
imageURL, _ := c.PollImageTask(ctx, accessToken, taskID, 3*time.Second, 600*time.Second, nil)
```

<details>
<summary>更多 SDK 用法</summary>

#### 图生视频

```go
mediaID, _ := c.UploadImage(ctx, accessToken, imageData, "input.png")
token, _ := c.GenerateSentinelToken(ctx, accessToken)
taskID, _ := c.CreateVideoTaskWithImage(ctx, accessToken, token, "animate this", "landscape", 300, "sy_8", "small", mediaID)
```

#### 带风格的视频

```go
taskID, _ := c.CreateVideoTaskWithOptions(ctx, accessToken, token, "a cat", "landscape", 300, "sy_8", "small", "", "anime")

// 或从提示词中提取 {style}
prompt, styleID := sora.ExtractStyle("a cat {anime}")
```

可选风格：`festive`, `kakalaka`, `news`, `selfie`, `handheld`, `golden`, `anime`, `retro`, `nostalgic`, `comic`

#### Remix 视频

```go
remixID := sora.ExtractRemixID("https://sora.chatgpt.com/p/s_xxx")
taskID, _ := c.RemixVideo(ctx, accessToken, token, remixID, "make it snowy", "landscape", 300, "")
```

#### 分镜视频

```go
prompt := "[5.0s]一只猫在草地上奔跑 [5.0s]猫跳上了树"
taskID, _ := c.CreateStoryboardTask(ctx, accessToken, token, prompt, "landscape", 450, "", "")
```

#### 角色管理

```go
cameoID, _ := c.UploadCharacterVideo(ctx, accessToken, videoData)
status, _ := c.PollCameoStatus(ctx, accessToken, cameoID, 3*time.Second, 300*time.Second, nil)
imageData, _ := c.DownloadCharacterImage(ctx, status.ProfileAssetURL)
assetPointer, _ := c.UploadCharacterImage(ctx, accessToken, imageData)
characterID, _ := c.FinalizeCharacter(ctx, accessToken, cameoID, "name", "显示名", assetPointer)
_ = c.SetCharacterPublic(ctx, accessToken, cameoID)
```

#### 去水印下载

```go
soraToken, newRefreshToken, _ := c.RefreshAccessToken(ctx, refreshToken, "")
url, _ := c.GetWatermarkFreeURL(ctx, soraToken, "https://sora.chatgpt.com/p/s_xxx")
```

#### 提示词增强

```go
enhanced, _ := c.EnhancePrompt(ctx, accessToken, "a cat", "medium", 10)
```

#### 代理支持

```go
c, _ := sora.New("http://user:pass@ip:port")
c, _ := sora.New("socks5://user:pass@ip:port")
proxy := sora.ParseProxy("ip:port:user:pass")
```

</details>

### SDK 方法速查

| 方法 | 说明 |
|------|------|
| `New(proxyURL)` | 创建客户端 |
| `GenerateSentinelToken` | 获取 sentinel token（含 PoW） |
| `UploadImage` | 上传图片 |
| `CreateImageTask` / `CreateImageTaskWithImage` | 文生图 / 图生图 |
| `CreateVideoTask` / `CreateVideoTaskWithImage` | 文生视频 / 图生视频 |
| `CreateVideoTaskWithOptions` | 完整视频创建（含风格） |
| `RemixVideo` | Remix 视频 |
| `CreateStoryboardTask` | 分镜视频 |
| `EnhancePrompt` | 提示词增强 |
| `PollImageTask` / `PollVideoTask` | 轮询任务 |
| `GetDownloadURL` | 获取下载链接 |
| `RefreshAccessToken` | 刷新 Token |
| `GetWatermarkFreeURL` | 去水印链接 |
| `GetCreditBalance` / `GetSubscriptionInfo` | 配额/订阅查询 |
| `UploadCharacterVideo` / `FinalizeCharacter` | 角色创建 |
| `PublishVideo` / `DeletePost` | 发布/删除帖子 |

### 视频参数

| 参数 | 可选值 |
|------|--------|
| orientation | `landscape` / `portrait` |
| nFrames | `300`(10s) / `450`(15s) / `750`(25s) |
| model | `sy_8`(标准) / `sy_ore`(Pro) |
| size | `small`(标准) / `large`(高清, 仅Pro) |

### 图片参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| width | 图片宽度（像素） | 1792 |
| height | 图片高度（像素） | 1024 |
| input_reference | 参考图片，URL 或 base64 data URI（图生图时传入） | — |

## CLI 工具

基于 [Bubble Tea](https://github.com/charmbracelet/bubbletea) 的交互式 TUI 工具：

```bash
go install github.com/DouDOU-start/go-sora2api/cmd/sora2api@latest
sora2api
```

## 项目结构

```
go-sora2api/
├── sora/                    # Go SDK
├── server/                  # Web 后端（Gin）
│   ├── handler/             #   API 路由处理
│   ├── service/             #   业务逻辑（调度/账号管理/任务）
│   ├── model/               #   数据模型
│   ├── config/              #   配置加载
│   └── dist/                #   前端打包产物（编译时嵌入）
├── web/                     # React 前端
│   └── src/
│       ├── pages/           #   页面组件
│       ├── api/             #   API 调用
│       └── components/      #   通用组件
├── cmd/sora2api/            # CLI 工具
├── deploy/                  # 部署脚本
│   ├── install.sh           #   一键安装/管理脚本
│   ├── sora2api.service     #   Systemd 服务文件
│   └── config.example.yaml  #   配置模板
└── .github/workflows/       # CI/CD
    └── release.yml          #   自动构建发布
```

## 免责声明

本项目仅供学习和研究使用，不得用于任何商业或非法用途。使用者应自行承担使用本项目所产生的一切风险和责任，项目作者不对因使用本项目而导致的任何直接或间接损失承担责任。

## 许可证

[MIT License](LICENSE)
