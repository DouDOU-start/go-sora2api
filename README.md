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
- 全面 `context.Context` 支持（超时控制、取消）

## 安装

```bash
go get github.com/DouDOU-start/go-sora2api/sora
```

## 快速开始

### 文生图

```go
ctx := context.Background()
c, _ := sora.New("")
token, _ := c.GenerateSentinelToken(ctx, accessToken)
taskID, _ := c.CreateImageTask(ctx, accessToken, token, "a cute cat", 360, 360)
imageURL, _ := c.PollImageTask(ctx, accessToken, taskID, 3*time.Second, 600*time.Second, nil)
```

### 图生图

```go
ctx := context.Background()

// 先上传图片
mediaID, _ := c.UploadImage(ctx, accessToken, imageData, "input.png")

// 基于图片生成新图片
token, _ := c.GenerateSentinelToken(ctx, accessToken)
taskID, _ := c.CreateImageTaskWithImage(ctx, accessToken, token, "make it more colorful", 360, 360, mediaID)
imageURL, _ := c.PollImageTask(ctx, accessToken, taskID, 3*time.Second, 600*time.Second, nil)
```

### 文生视频

```go
ctx := context.Background()
token, _ := c.GenerateSentinelToken(ctx, accessToken)
taskID, _ := c.CreateVideoTask(ctx, accessToken, token, "a cat running", "landscape", 300, "sy_8", "small")
_ = c.PollVideoTask(ctx, accessToken, taskID, 3*time.Second, 600*time.Second, nil)
url, _ := c.GetDownloadURL(ctx, accessToken, taskID)
```

### 图生视频

```go
ctx := context.Background()
mediaID, _ := c.UploadImage(ctx, accessToken, imageData, "input.png")

token, _ := c.GenerateSentinelToken(ctx, accessToken)
taskID, _ := c.CreateVideoTaskWithImage(ctx, accessToken, token, "animate this scene", "landscape", 300, "sy_8", "small", mediaID)
_ = c.PollVideoTask(ctx, accessToken, taskID, 3*time.Second, 600*time.Second, nil)
url, _ := c.GetDownloadURL(ctx, accessToken, taskID)
```

### 带风格的视频

```go
// 方式一：通过参数指定风格
taskID, _ := c.CreateVideoTaskWithOptions(ctx, accessToken, token, "a cat running", "landscape", 300, "sy_8", "small", "", "anime")

// 方式二：从提示词中自动提取 {style}
prompt, styleID := sora.ExtractStyle("a cat running {anime}")
// prompt = "a cat running", styleID = "anime"
```

可选风格：`festive`, `kakalaka`, `news`, `selfie`, `handheld`, `golden`, `anime`, `retro`, `nostalgic`, `comic`

### Remix 视频

```go
ctx := context.Background()

// 从分享链接提取视频 ID
remixID := sora.ExtractRemixID("https://sora.chatgpt.com/p/s_690d100857248191b679e6de12db840e")

token, _ := c.GenerateSentinelToken(ctx, accessToken)
taskID, _ := c.RemixVideo(ctx, accessToken, token, remixID, "make it snowy", "landscape", 300, "")
_ = c.PollVideoTask(ctx, accessToken, taskID, 3*time.Second, 600*time.Second, nil)
url, _ := c.GetDownloadURL(ctx, accessToken, taskID)
```

### 获取去水印下载链接

> 注意：此功能需要 `refresh_token`，不支持普通的 ChatGPT `access_token`。

```go
ctx := context.Background()
c, _ := sora.New("")

// 1. 使用 refresh_token 刷新获取专用 access_token
soraToken, newRefreshToken, _ := c.RefreshAccessToken(ctx, refreshToken, "")
// newRefreshToken 已更新，需保存供下次使用

// 2. 获取去水印链接（支持传入完整链接或视频 ID）
url, _ := c.GetWatermarkFreeURL(ctx, soraToken, "https://sora.chatgpt.com/p/s_xxx")
// 或
url, _ := c.GetWatermarkFreeURL(ctx, soraToken, "s_xxx")
```

### 分镜视频

```go
ctx := context.Background()

// 分镜格式：[时长]场景描述
prompt := "[5.0s]一只猫在草地上奔跑 [5.0s]猫跳上了树"

token, _ := c.GenerateSentinelToken(ctx, accessToken)
taskID, _ := c.CreateStoryboardTask(ctx, accessToken, token, prompt, "landscape", 450, "", "")
_ = c.PollVideoTask(ctx, accessToken, taskID, 3*time.Second, 600*time.Second, nil)
url, _ := c.GetDownloadURL(ctx, accessToken, taskID)
```

### 角色管理

```go
ctx := context.Background()

// 创建角色（全流程：上传视频 → 轮询处理 → 下载头像 → 上传头像 → 定稿 → 设置公开）
cameoID, _ := c.UploadCharacterVideo(ctx, accessToken, videoData)
status, _ := c.PollCameoStatus(ctx, accessToken, cameoID, 3*time.Second, 300*time.Second, nil)
imageData, _ := c.DownloadCharacterImage(ctx, status.ProfileAssetURL)
assetPointer, _ := c.UploadCharacterImage(ctx, accessToken, imageData)
characterID, _ := c.FinalizeCharacter(ctx, accessToken, cameoID, "username", "显示名称", assetPointer)
_ = c.SetCharacterPublic(ctx, accessToken, cameoID)

// 删除角色
_ = c.DeleteCharacter(ctx, accessToken, characterID)
```

### 视频发布

```go
ctx := context.Background()

// 发布视频获取去水印链接
token, _ := c.GenerateSentinelToken(ctx, accessToken)
postID, _ := c.PublishVideo(ctx, accessToken, token, "gen_xxx")
// 去水印链接: https://sora.chatgpt.com/p/{postID}

// 删除帖子
_ = c.DeletePost(ctx, accessToken, postID)
```

### 账号信息查询

```go
ctx := context.Background()

// 查询配额
balance, _ := c.GetCreditBalance(ctx, accessToken)
fmt.Printf("剩余次数: %d\n", balance.RemainingCount)

// 查询订阅
sub, _ := c.GetSubscriptionInfo(ctx, accessToken)
fmt.Printf("套餐: %s, 到期: %s\n", sub.PlanTitle, time.Unix(sub.EndTs, 0).Format("2006-01-02"))
```

### 提示词优化

```go
ctx := context.Background()
enhanced, _ := c.EnhancePrompt(ctx, accessToken, "a cat", "medium", 10)
// enhanced = "A playful orange tabby cat sits gracefully..."
```

### 超时控制

```go
// 设置 30 秒超时
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

token, err := c.GenerateSentinelToken(ctx, accessToken)
if err != nil {
    // 可能是超时或主动取消
    log.Fatal(err)
}
```

### 进度回调

```go
c.PollImageTask(ctx, accessToken, taskID, 3*time.Second, 600*time.Second, func(p sora.Progress) {
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

> 所有方法的第一个参数均为 `context.Context`，用于超时控制和取消。

| 方法 | 说明 |
|------|------|
| `New(proxyURL)` | 创建客户端 |
| `GenerateSentinelToken(ctx, accessToken)` | 获取 sentinel token（含 PoW），创建任务前必须调用 |
| `UploadImage(ctx, accessToken, imageData, filename)` | 上传图片，返回 mediaID |
| `CreateImageTask(ctx, accessToken, sentinelToken, prompt, w, h)` | 文生图 |
| `CreateImageTaskWithImage(ctx, ..., mediaID)` | 图生图 |
| `CreateVideoTask(ctx, accessToken, sentinelToken, prompt, orientation, nFrames, model, size)` | 文生视频 |
| `CreateVideoTaskWithImage(ctx, ..., mediaID)` | 图生视频 |
| `CreateVideoTaskWithOptions(ctx, ..., mediaID, styleID)` | 完整视频创建（含风格） |
| `RemixVideo(ctx, accessToken, sentinelToken, remixTargetID, prompt, orientation, nFrames, styleID)` | Remix 视频 |
| `EnhancePrompt(ctx, accessToken, prompt, expansionLevel, durationSec)` | 提示词优化 |
| `PollImageTask(ctx, accessToken, taskID, interval, timeout, onProgress)` | 轮询图片任务 |
| `PollVideoTask(ctx, accessToken, taskID, interval, timeout, onProgress)` | 轮询视频任务 |
| `GetDownloadURL(ctx, accessToken, taskID)` | 获取视频下载链接 |
| `RefreshAccessToken(ctx, refreshToken, clientID)` | 刷新获取专用 access_token（去水印用） |
| `GetWatermarkFreeURL(ctx, accessToken, videoID)` | 获取无水印下载链接 |
| `GetCreditBalance(ctx, accessToken)` | 查询配额（剩余次数、速率限制） |
| `GetSubscriptionInfo(ctx, accessToken)` | 查询订阅信息（套餐类型、到期时间） |
| `CreateStoryboardTask(ctx, accessToken, sentinelToken, prompt, orientation, nFrames, mediaID, styleID)` | 创建分镜视频任务 |
| `UploadCharacterVideo(ctx, accessToken, videoData)` | 上传角色视频，返回 cameoID |
| `GetCameoStatus(ctx, accessToken, cameoID)` | 获取角色处理状态 |
| `PollCameoStatus(ctx, accessToken, cameoID, interval, timeout, onProgress)` | 轮询角色处理状态 |
| `DownloadCharacterImage(ctx, imageURL)` | 下载角色头像图片 |
| `UploadCharacterImage(ctx, accessToken, imageData)` | 上传角色头像，返回 assetPointer |
| `FinalizeCharacter(ctx, accessToken, cameoID, username, displayName, assetPointer)` | 定稿角色 |
| `SetCharacterPublic(ctx, accessToken, cameoID)` | 设置角色为公开 |
| `DeleteCharacter(ctx, accessToken, characterID)` | 删除角色 |
| `PublishVideo(ctx, accessToken, sentinelToken, generationID)` | 发布视频帖子 |
| `DeletePost(ctx, accessToken, postID)` | 删除已发布帖子 |
| `QueryImageTaskOnce(ctx, accessToken, taskID, startTime)` | 单次查询图片任务状态（非阻塞） |
| `QueryVideoTaskOnce(ctx, accessToken, taskID, startTime, maxProgress)` | 单次查询视频任务状态（非阻塞） |
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
