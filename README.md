# go-sora2api

Sora 视频/图片生成 Go SDK，通过 TLS 指纹模拟绕过 Cloudflare 验证。

## 功能

- 文生图 / 图生图
- 文生视频 / 图生视频
- 视频 Remix（基于已有视频再创作）
- 10 种视频风格（anime、retro、comic 等）
- 提示词优化（AI 自动扩展提示词）
- 进度回调
- 代理支持

## 安装

```bash
go get github.com/DouDOU-start/go-sora2api/sora
```

## 快速开始

### 文生图

```go
c, _ := sora.New("")
token, _ := c.GenerateSentinelToken(accessToken)
taskID, _ := c.CreateImageTask(accessToken, token, "a cute cat", 360, 360)
imageURL, _ := c.PollImageTask(accessToken, taskID, 3*time.Second, 600*time.Second, nil)
```

### 图生图

```go
// 先上传图片
mediaID, _ := c.UploadImage(accessToken, imageData, "input.png")

// 基于图片生成新图片
token, _ := c.GenerateSentinelToken(accessToken)
taskID, _ := c.CreateImageTaskWithImage(accessToken, token, "make it more colorful", 360, 360, mediaID)
imageURL, _ := c.PollImageTask(accessToken, taskID, 3*time.Second, 600*time.Second, nil)
```

### 文生视频

```go
token, _ := c.GenerateSentinelToken(accessToken)
taskID, _ := c.CreateVideoTask(accessToken, token, "a cat running", "landscape", 150, "sy_8", "small")
_ = c.PollVideoTask(accessToken, taskID, 3*time.Second, 600*time.Second, nil)
url, _ := c.GetDownloadURL(accessToken, taskID)
```

### 图生视频

```go
mediaID, _ := c.UploadImage(accessToken, imageData, "input.png")

token, _ := c.GenerateSentinelToken(accessToken)
taskID, _ := c.CreateVideoTaskWithImage(accessToken, token, "animate this scene", "landscape", 150, "sy_8", "small", mediaID)
_ = c.PollVideoTask(accessToken, taskID, 3*time.Second, 600*time.Second, nil)
url, _ := c.GetDownloadURL(accessToken, taskID)
```

### 带风格的视频

```go
// 方式一：通过参数指定风格
taskID, _ := c.CreateVideoTaskWithOptions(accessToken, token, "a cat running", "landscape", 150, "sy_8", "small", "", "anime")

// 方式二：从提示词中自动提取 {style}
prompt, styleID := sora.ExtractStyle("a cat running {anime}")
// prompt = "a cat running", styleID = "anime"
```

可选风格：`festive`, `kakalaka`, `news`, `selfie`, `handheld`, `golden`, `anime`, `retro`, `nostalgic`, `comic`

### Remix 视频

```go
// 从分享链接提取视频 ID
remixID := sora.ExtractRemixID("https://sora.chatgpt.com/p/s_690d100857248191b679e6de12db840e")

token, _ := c.GenerateSentinelToken(accessToken)
taskID, _ := c.RemixVideo(accessToken, token, remixID, "make it snowy", "landscape", 150, "")
_ = c.PollVideoTask(accessToken, taskID, 3*time.Second, 600*time.Second, nil)
url, _ := c.GetDownloadURL(accessToken, taskID)
```

### 提示词优化

```go
enhanced, _ := c.EnhancePrompt(accessToken, "a cat", "medium", 10)
// enhanced = "A playful orange tabby cat sits gracefully..."
```

### 进度回调

```go
c.PollImageTask(accessToken, taskID, 3*time.Second, 600*time.Second, func(p sora.Progress) {
    fmt.Printf("\r进度: %d%% 状态: %s 耗时: %ds", p.Percent, p.Status, p.Elapsed)
})
```

### 代理支持

```go
c, _ := sora.New("http://user:pass@ip:port")
c, _ := sora.New("socks5://user:pass@ip:port")

// ParseProxy 解析简写格式
proxy := sora.ParseProxy("ip:port:user:pass")
c, _ := sora.New(proxy)
```

## API 参考

| 方法 | 说明 |
|------|------|
| `New(proxyURL)` | 创建客户端 |
| `GenerateSentinelToken(accessToken)` | 获取 sentinel token（含 PoW），创建任务前必须调用 |
| `UploadImage(accessToken, imageData, filename)` | 上传图片，返回 mediaID |
| `CreateImageTask(accessToken, sentinelToken, prompt, w, h)` | 文生图 |
| `CreateImageTaskWithImage(..., mediaID)` | 图生图 |
| `CreateVideoTask(accessToken, sentinelToken, prompt, orientation, nFrames, model, size)` | 文生视频 |
| `CreateVideoTaskWithImage(..., mediaID)` | 图生视频 |
| `CreateVideoTaskWithOptions(..., mediaID, styleID)` | 完整视频创建（含风格） |
| `RemixVideo(accessToken, sentinelToken, remixTargetID, prompt, orientation, nFrames, styleID)` | Remix 视频 |
| `EnhancePrompt(accessToken, prompt, expansionLevel, durationSec)` | 提示词优化 |
| `PollImageTask(accessToken, taskID, interval, timeout, onProgress)` | 轮询图片任务 |
| `PollVideoTask(accessToken, taskID, interval, timeout, onProgress)` | 轮询视频任务 |
| `GetDownloadURL(accessToken, taskID)` | 获取视频下载链接 |
| `ExtractStyle(prompt)` | 从提示词提取 `{style}` 风格 |
| `ExtractRemixID(text)` | 从 URL 提取 Remix 视频 ID |
| `ParseProxy(proxy)` | 解析代理字符串 |

### 视频参数

| 参数 | 可选值 |
|------|--------|
| orientation | `landscape` / `portrait` |
| nFrames | `150`(5s) / `300`(10s) / `450`(15s) / `750`(25s) |
| model | `sy_8`(标准) / `sy_ore`(Pro) |
| size | `small`(标准) / `large`(高清, 仅Pro) |

### 图片参数

| 尺寸 | width | height |
|------|-------|--------|
| 正方形 | 360 | 360 |
| 横向 | 540 | 360 |
| 纵向 | 360 | 540 |

## CLI 工具

提供交互式命令行工具，支持全部功能：

```bash
go install github.com/DouDOU-start/go-sora2api/cmd/sora2api@latest
sora2api
```

或编译后运行：

```bash
go build -o sora2api ./cmd/sora2api/
./sora2api
```

## 项目结构

```
go-sora2api/
├── sora/                    # 公开 SDK 包
│   ├── client.go            # 客户端基础 + Sentinel Token
│   ├── task.go              # 任务创建（Upload/Image/Video/Remix/Enhance）
│   ├── poll.go              # 任务轮询 + 下载链接
│   ├── style.go             # 风格提取 + Remix ID 解析
│   ├── pow.go               # PoW (SHA3-512) 算法
│   └── util.go              # 代理解析等工具
├── cmd/sora2api/
│   └── main.go              # 交互式 CLI 工具
├── go.mod
└── README.md
```

## 免责声明

本项目仅供学习和研究使用，不得用于任何商业或非法用途。使用者应自行承担使用本项目所产生的一切风险和责任，项目作者不对因使用本项目而导致的任何直接或间接损失承担责任。

使用本项目即表示您已阅读并同意以上声明。

## 许可证

[MIT License](LICENSE)
