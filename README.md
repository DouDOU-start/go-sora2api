# go-sora2api

Sora 视频/图片生成 Go SDK，通过 TLS 指纹模拟绕过 Cloudflare 验证。

## 功能

- 文生图 / 图生图
- 文生视频 / 图生视频
- 视频 Remix（基于已有视频再创作）
- 分镜视频（多场景拼接生成）
- 角色管理（创建 / 删除角色）
- 视频发布（发布去水印 / 删除帖子）
- 获取去水印下载链接
- 10 种视频风格（anime、retro、comic 等）
- 提示词优化（AI 自动扩展提示词）
- 账号信息查询（配额 / 订阅）
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
taskID, _ := c.CreateVideoTask(accessToken, token, "a cat running", "landscape", 300, "sy_8", "small")
_ = c.PollVideoTask(accessToken, taskID, 3*time.Second, 600*time.Second, nil)
url, _ := c.GetDownloadURL(accessToken, taskID)
```

### 图生视频

```go
mediaID, _ := c.UploadImage(accessToken, imageData, "input.png")

token, _ := c.GenerateSentinelToken(accessToken)
taskID, _ := c.CreateVideoTaskWithImage(accessToken, token, "animate this scene", "landscape", 300, "sy_8", "small", mediaID)
_ = c.PollVideoTask(accessToken, taskID, 3*time.Second, 600*time.Second, nil)
url, _ := c.GetDownloadURL(accessToken, taskID)
```

### 带风格的视频

```go
// 方式一：通过参数指定风格
taskID, _ := c.CreateVideoTaskWithOptions(accessToken, token, "a cat running", "landscape", 300, "sy_8", "small", "", "anime")

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
taskID, _ := c.RemixVideo(accessToken, token, remixID, "make it snowy", "landscape", 300, "")
_ = c.PollVideoTask(accessToken, taskID, 3*time.Second, 600*time.Second, nil)
url, _ := c.GetDownloadURL(accessToken, taskID)
```

### 获取去水印下载链接

> 注意：此功能需要 `refresh_token`，不支持普通的 ChatGPT `access_token`。

```go
c, _ := sora.New("")

// 1. 使用 refresh_token 刷新获取专用 access_token
soraToken, newRefreshToken, _ := c.RefreshAccessToken(refreshToken, "")
// newRefreshToken 已更新，需保存供下次使用

// 2. 获取去水印链接（支持传入完整链接或视频 ID）
url, _ := c.GetWatermarkFreeURL(soraToken, "https://sora.chatgpt.com/p/s_xxx")
// 或
url, _ := c.GetWatermarkFreeURL(soraToken, "s_xxx")
```

### 分镜视频

```go
// 分镜格式：[时长]场景描述
prompt := "[5.0s]一只猫在草地上奔跑 [5.0s]猫跳上了树"

token, _ := c.GenerateSentinelToken(accessToken)
taskID, _ := c.CreateStoryboardTask(accessToken, token, prompt, "landscape", 450, "", "")
_ = c.PollVideoTask(accessToken, taskID, 3*time.Second, 600*time.Second, nil)
url, _ := c.GetDownloadURL(accessToken, taskID)
```

### 角色管理

```go
// 创建角色（全流程：上传视频 → 轮询处理 → 下载头像 → 上传头像 → 定稿 → 设置公开）
cameoID, _ := c.UploadCharacterVideo(accessToken, videoData)
status, _ := c.PollCameoStatus(accessToken, cameoID, 3*time.Second, 300*time.Second, nil)
imageData, _ := c.DownloadCharacterImage(status.ProfileAssetURL)
assetPointer, _ := c.UploadCharacterImage(accessToken, imageData)
characterID, _ := c.FinalizeCharacter(accessToken, cameoID, "username", "显示名称", assetPointer)
_ = c.SetCharacterPublic(accessToken, cameoID)

// 删除角色
_ = c.DeleteCharacter(accessToken, characterID)
```

### 视频发布

```go
// 发布视频获取去水印链接
token, _ := c.GenerateSentinelToken(accessToken)
postID, _ := c.PublishVideo(accessToken, token, "gen_xxx")
// 去水印链接: https://sora.chatgpt.com/p/{postID}

// 删除帖子
_ = c.DeletePost(accessToken, postID)
```

### 账号信息查询

```go
// 查询配额
balance, _ := c.GetCreditBalance(accessToken)
fmt.Printf("剩余次数: %d\n", balance.RemainingCount)

// 查询订阅
sub, _ := c.GetSubscriptionInfo(accessToken)
fmt.Printf("套餐: %s, 到期: %s\n", sub.PlanTitle, time.Unix(sub.EndTs, 0).Format("2006-01-02"))
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
| `RefreshAccessToken(refreshToken, clientID)` | 刷新获取专用 access_token（去水印用） |
| `GetWatermarkFreeURL(accessToken, videoID)` | 获取无水印下载链接 |
| `GetCreditBalance(accessToken)` | 查询配额（剩余次数、速率限制） |
| `GetSubscriptionInfo(accessToken)` | 查询订阅信息（套餐类型、到期时间） |
| `CreateStoryboardTask(accessToken, sentinelToken, prompt, orientation, nFrames, mediaID, styleID)` | 创建分镜视频任务 |
| `UploadCharacterVideo(accessToken, videoData)` | 上传角色视频，返回 cameoID |
| `GetCameoStatus(accessToken, cameoID)` | 获取角色处理状态 |
| `PollCameoStatus(accessToken, cameoID, interval, timeout, onProgress)` | 轮询角色处理状态 |
| `DownloadCharacterImage(imageURL)` | 下载角色头像图片 |
| `UploadCharacterImage(accessToken, imageData)` | 上传角色头像，返回 assetPointer |
| `FinalizeCharacter(accessToken, cameoID, username, displayName, assetPointer)` | 定稿角色 |
| `SetCharacterPublic(accessToken, cameoID)` | 设置角色为公开 |
| `DeleteCharacter(accessToken, characterID)` | 删除角色 |
| `PublishVideo(accessToken, sentinelToken, generationID)` | 发布视频帖子 |
| `DeletePost(accessToken, postID)` | 删除已发布帖子 |
| `QueryImageTaskOnce(accessToken, taskID, startTime)` | 单次查询图片任务状态（非阻塞） |
| `QueryVideoTaskOnce(accessToken, taskID, startTime, maxProgress)` | 单次查询视频任务状态（非阻塞） |
| `ExtractStyle(prompt)` | 从提示词提取 `{style}` 风格 |
| `IsStoryboardPrompt(prompt)` | 检测是否为分镜格式 |
| `FormatStoryboardPrompt(prompt)` | 转换分镜格式为 API 格式 |
| `ExtractRemixID(text)` | 从 URL 提取 Remix 视频 ID |
| `ExtractVideoID(text)` | 从分享链接提取视频 ID |
| `ParseProxy(proxy)` | 解析代理字符串 |

### 视频参数

| 参数 | 可选值 |
|------|--------|
| orientation | `landscape` / `portrait` |
| nFrames | `300`(10s) / `450`(15s) / `750`(25s) |
| model | `sy_8`(标准) / `sy_ore`(Pro) |
| size | `small`(标准) / `large`(高清, 仅Pro) |

### 图片参数

| 尺寸 | width | height |
|------|-------|--------|
| 正方形 | 360 | 360 |
| 横向 | 540 | 360 |
| 纵向 | 360 | 540 |

## CLI 工具

提供基于 [Bubble Tea](https://github.com/charmbracelet/bubbletea) 的交互式 TUI 工具，支持全部功能：

- 键盘导航（↑/↓ 菜单选择，Tab 切换字段，←/→ 选择选项）
- 自动展示账号信息（配额、订阅类型、到期时间）
- 分组功能菜单（图片生成 / 视频生成 / 角色管理 / 视频发布 / 工具 / 设置）
- 动态参数表单（不同功能自动展示对应参数）
- 任务进度条和状态显示

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
│   ├── client.go            # 客户端基础 + 请求头封装 + Sentinel Token
│   ├── task.go              # 任务创建（Upload/Image/Video/Remix/Enhance/Watermark）
│   ├── poll.go              # 任务轮询 + 单次查询 + 下载链接 + 配额/订阅查询
│   ├── character.go         # 角色管理（上传/轮询/定稿/公开/删除）
│   ├── storyboard.go        # 分镜视频（格式检测/转换/创建任务）
│   ├── post.go              # 视频发布（发布/删除帖子）
│   ├── style.go             # 风格提取 + ID 解析工具
│   ├── pow.go               # PoW (SHA3-512) 算法
│   └── util.go              # 代理解析等工具
├── cmd/sora2api/            # TUI 交互式工具
│   ├── main.go              # 入口
│   ├── app.go               # 顶层模型 + 页面切换
│   ├── messages.go          # 消息类型定义
│   ├── styles.go            # 样式常量
│   ├── page_setup.go        # Token/代理配置页
│   ├── page_menu.go         # 功能菜单页
│   ├── page_param.go        # 动态参数表单页
│   ├── page_task.go         # 任务执行页
│   └── page_result.go       # 结果展示页
├── go.mod
└── README.md
```

## 免责声明

本项目仅供学习和研究使用，不得用于任何商业或非法用途。使用者应自行承担使用本项目所产生的一切风险和责任，项目作者不对因使用本项目而导致的任何直接或间接损失承担责任。

使用本项目即表示您已阅读并同意以上声明。

## 许可证

[MIT License](LICENSE)
