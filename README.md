# go-sora2api

Sora 视频生成 Go 实现，通过 TLS 指纹模拟绕过 Cloudflare 验证。

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

运行后按提示输入 `access_token` 和代理地址，程序将自动完成：

1. 获取 sentinel token（含 PoW 计算）
2. 创建视频生成任务
3. 轮询任务进度
4. 获取视频下载链接

## 代理格式

支持以下格式：

| 格式 | 示例 |
|------|------|
| ip:port:user:pass | `45.56.182.13:7902:user:pass` |
| ip:port | `45.56.182.13:7902` |
| 标准 URL | `http://user:pass@45.56.182.13:7902` |
| SOCKS5 | `socks5://user:pass@45.56.182.13:1080` |

## 视频参数

在 `cmd/main.go` 中修改：

| 参数 | 说明 | 可选值 |
|------|------|--------|
| orientation | 方向 | `landscape` / `portrait` |
| nFrames | 时长 | `150`(5s) / `300`(10s) / `450`(15s) / `750`(25s) |
| model | 模型 | `sy_8`(标准) / `sy_ore`(Pro) |
| size | 尺寸 | `small`(标准) / `large`(高清, 仅Pro) |
