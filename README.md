# go-sora2api

Sora 视频/图片生成 Go 实现，通过 TLS 指纹模拟绕过 Cloudflare 验证。

## 项目结构

```
go-sora2api/
├── cmd/main.go                # 入口
├── internal/
│   ├── client/client.go       # TLS 客户端 + Sora API 调用
│   ├── pow/pow.go             # PoW (SHA3-512) 算法
│   └── util/util.go           # 代理解析、UUID 等工具
├── go.mod
└── README.md
```

## 构建

```bash
go build -o sora2api ./cmd/
```

## 使用

```bash
./sora2api
```

运行后按提示输入 `access_token` 和代理地址，选择生成类型（图片/视频），程序将自动完成：

1. 获取 sentinel token（含 PoW 计算）
2. 创建图片/视频生成任务
3. 轮询任务进度
4. 获取下载链接

## 代理格式

支持以下格式：

| 格式 | 示例 |
|------|------|
| ip:port:user:pass | `45.56.182.13:7902:user:pass` |
| ip:port | `45.56.182.13:7902` |
| 标准 URL | `http://user:pass@45.56.182.13:7902` |
| SOCKS5 | `socks5://user:pass@45.56.182.13:1080` |

## 运行示例

### 图片生成

```
$ go run ./cmd/main.go
请输入 access_token: eyJhbGciOiJSUzI1NiIs...
请输入代理 (留空不使用代理):
请选择生成类型: 1) 图片  2) 视频 → 1
请输入提示词: 一只可爱的小猫在草地上奔跑
请选择图片尺寸: 1) 正方形  2) 横向  3) 纵向 → 1

[步骤 1] sentinel token 获取成功
[步骤 2/3] 任务创建成功! ID: task_01kj7e8ssze66rw1e5kzv2jtfk
[步骤 3/3] 进度: 75%  状态: succeeded  耗时: 15s

[完成] 图片下载链接:
  https://videos.openai.com/...
```

### 视频生成

```
$ go run ./cmd/main.go
请输入 access_token: eyJhbGciOiJSUzI1NiIs...
请输入代理 (留空不使用代理): 45.56.182.13:7902:user:pass
请选择生成类型: 1) 图片  2) 视频 → 2
请输入提示词: 一只可爱的小猫在草地上奔跑
请选择视频方向: 1) 横向  2) 纵向 → 1
请选择视频时长: 1) 5s  2) 10s  3) 15s  4) 25s → 1
请选择模型: 1) 标准  2) Pro → 1

[步骤 1] sentinel token 获取成功
[步骤 2/4] 任务创建成功! ID: task_01kj7djj0cfn69p1fmtw6sawfd
[步骤 3/4] 进度: 100%  耗时: 100s
[步骤 4/4] 获取视频下载链接...

[完成] 视频下载链接:
  https://videos.openai.com/...
```

## 参数说明

### 图片参数

| 尺寸选项 | 分辨率 |
|----------|--------|
| 正方形 | 360x360 |
| 横向 | 540x360 |
| 纵向 | 360x540 |

### 视频参数

| 参数 | 说明 | 可选值 |
|------|------|--------|
| 方向 | 横/纵 | `landscape` / `portrait` |
| 时长 | 帧数 | `150`(5s) / `300`(10s) / `450`(15s) / `750`(25s) |
| 模型 | 标准/Pro | `sy_8`(标准) / `sy_ore`(Pro) |
| 尺寸 | 清晰度 | `small`(标准) / `large`(高清, 仅Pro) |

## 免责声明

本项目仅供学习和研究使用，不得用于任何商业或非法用途。使用者应自行承担使用本项目所产生的一切风险和责任，项目作者不对因使用本项目而导致的任何直接或间接损失承担责任。

使用本项目即表示您已阅读并同意以上声明。

## 许可证

[MIT License](LICENSE)
