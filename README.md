# go-sora2api

Sora 视频/图片生成 Go SDK，通过 TLS 指纹模拟绕过 Cloudflare 验证。

## 安装

```bash
go get github.com/DouDOU-start/go-sora2api/sora
```

## 作为库使用

```go
package main

import (
	"fmt"
	"time"

	"github.com/DouDOU-start/go-sora2api/sora"
)

func main() {
	// 创建客户端（传入代理地址，空字符串表示不使用代理）
	c, err := sora.New("")
	if err != nil {
		panic(err)
	}

	accessToken := "your_access_token"

	// 1. 获取 sentinel token
	sentinelToken, err := c.GenerateSentinelToken(accessToken)
	if err != nil {
		panic(err)
	}

	// 2. 创建图片任务
	taskID, err := c.CreateImageTask(accessToken, sentinelToken, "a cute cat running on grass", 360, 360)
	if err != nil {
		panic(err)
	}

	// 3. 轮询进度（回调可传 nil 忽略进度）
	imageURL, err := c.PollImageTask(accessToken, taskID, 3*time.Second, 600*time.Second, func(p sora.Progress) {
		fmt.Printf("\r进度: %d%% 状态: %s 耗时: %ds", p.Percent, p.Status, p.Elapsed)
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("\n图片链接:", imageURL)
}
```

### 视频生成

```go
// 创建视频任务
taskID, err := c.CreateVideoTask(accessToken, sentinelToken,
	"a cute cat running on grass", // prompt
	"landscape",                    // orientation: landscape / portrait
	150,                            // nFrames: 150(5s) / 300(10s) / 450(15s) / 750(25s)
	"sy_8",                         // model: sy_8(标准) / sy_ore(Pro)
	"small",                        // size: small(标准) / large(高清, 仅Pro)
)

// 轮询视频进度
err = c.PollVideoTask(accessToken, taskID, 3*time.Second, 600*time.Second, nil)

// 获取下载链接
downloadURL, err := c.GetDownloadURL(accessToken, taskID)
```

### 代理支持

```go
// 支持多种代理格式
c, _ := sora.New("http://user:pass@ip:port")
c, _ := sora.New("socks5://user:pass@ip:port")

// 也可以用 ParseProxy 解析简写格式
proxy := sora.ParseProxy("ip:port:user:pass")
c, _ := sora.New(proxy)
```

## API 参考

### `sora.New(proxyURL string) (*Client, error)`

创建客户端。`proxyURL` 为空则不使用代理。

### `Client.GenerateSentinelToken(accessToken string) (string, error)`

获取 sentinel token（含 PoW 计算），创建任务前必须调用。

### `Client.CreateImageTask(accessToken, sentinelToken, prompt string, width, height int) (string, error)`

创建图片生成任务，返回 taskID。

| 尺寸选项 | width | height |
|----------|-------|--------|
| 正方形 | 360 | 360 |
| 横向 | 540 | 360 |
| 纵向 | 360 | 540 |

### `Client.CreateVideoTask(accessToken, sentinelToken, prompt, orientation string, nFrames int, model, size string) (string, error)`

创建视频生成任务，返回 taskID。

| 参数 | 可选值 |
|------|--------|
| orientation | `landscape` / `portrait` |
| nFrames | `150`(5s) / `300`(10s) / `450`(15s) / `750`(25s) |
| model | `sy_8`(标准) / `sy_ore`(Pro) |
| size | `small`(标准) / `large`(高清, 仅Pro) |

### `Client.PollImageTask(accessToken, taskID string, pollInterval, pollTimeout time.Duration, onProgress ProgressFunc) (string, error)`

轮询图片任务，完成后返回图片 URL。`onProgress` 可为 `nil`。

### `Client.PollVideoTask(accessToken, taskID string, pollInterval, pollTimeout time.Duration, onProgress ProgressFunc) error`

轮询视频任务进度。`onProgress` 可为 `nil`。

### `Client.GetDownloadURL(accessToken, taskID string) (string, error)`

获取视频下载链接（视频轮询完成后调用）。

### `sora.ParseProxy(proxy string) string`

解析代理字符串，支持 `ip:port:user:pass`、`ip:port`、标准 URL 等格式。

## CLI 工具

也提供交互式命令行工具：

```bash
go install github.com/DouDOU-start/go-sora2api/cmd@latest
```

或编译后运行：

```bash
go build -o sora2api ./cmd/
./sora2api
```

## 项目结构

```
go-sora2api/
├── sora/              # 公开 SDK 包
│   ├── client.go      # 客户端 + API 方法
│   ├── pow.go         # PoW (SHA3-512) 算法
│   └── util.go        # 代理解析等工具
├── cmd/
│   └── main.go        # CLI 工具入口
├── go.mod
└── README.md
```

## 免责声明

本项目仅供学习和研究使用，不得用于任何商业或非法用途。使用者应自行承担使用本项目所产生的一切风险和责任，项目作者不对因使用本项目而导致的任何直接或间接损失承担责任。

使用本项目即表示您已阅读并同意以上声明。

## 许可证

[MIT License](LICENSE)
