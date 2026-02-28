package sora

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"math/rand"
	"net/url"
	"path"
	"strings"

	http "github.com/bogdanfinn/fhttp"
)

// randSessionID 生成 8 位随机字母数字字符串，用于代理 session ID
func randSessionID() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

// ParseProxy 解析代理字符串
//
// 支持格式:
//   - ip:port:username:password -> http://username:password@ip:port
//   - http://username:password@ip:port (原样返回)
//   - socks5://username:password@ip:port (原样返回)
//   - 空字符串 -> ""
func ParseProxy(proxy string) string {
	proxy = strings.TrimSpace(proxy)
	if proxy == "" {
		return ""
	}

	if strings.HasPrefix(proxy, "http://") || strings.HasPrefix(proxy, "https://") || strings.HasPrefix(proxy, "socks5://") {
		// 替换 %s 占位符为随机 session ID（兼容代理商的粘性会话格式）
		if strings.Contains(proxy, "%s") {
			proxy = strings.ReplaceAll(proxy, "%s", randSessionID())
		}
		return proxy
	}

	parts := strings.Split(proxy, ":")
	if len(parts) == 4 {
		return fmt.Sprintf("http://%s:%s@%s:%s", parts[2], parts[3], parts[0], parts[1])
	}
	if len(parts) == 2 {
		return fmt.Sprintf("http://%s", proxy)
	}

	return ""
}

// generateUUID 生成随机 UUID v4（使用 Client 实例的随机数生成器）
func (c *Client) generateUUID() string {
	b := make([]byte, 16)
	c.randRead(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// readAll 读取 io.Reader 全部内容
func readAll(r io.Reader) ([]byte, error) {
	return io.ReadAll(r)
}

// DownloadFile 下载 URL 内容并返回字节数据
func (c *Client) DownloadFile(ctx context.Context, fileURL string) ([]byte, error) {
	return c.doGet(ctx, fileURL, nil)
}

// TestConnectivity 测试代理连通性，向目标 URL 发送 GET 请求，只要收到响应即视为成功
func (c *Client) TestConnectivity(ctx context.Context, targetURL string) (statusCode int, err error) {
	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", desktopUserAgents[c.randIntn(len(desktopUserAgents))])

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}

// IsDataURI 判断字符串是否为 data URI（data:...;base64,...）
func IsDataURI(s string) bool {
	return strings.HasPrefix(s, "data:")
}

// ParseDataURI 解析 data URI，返回二进制数据和对应的文件扩展名
// 支持格式: data:image/png;base64,iVBOR... 或 data:video/mp4;base64,AAAA...
func ParseDataURI(dataURI string) (data []byte, ext string, err error) {
	// 格式: data:[<mediatype>][;base64],<data>
	if !strings.HasPrefix(dataURI, "data:") {
		return nil, "", fmt.Errorf("不是有效的 data URI")
	}

	commaIdx := strings.Index(dataURI, ",")
	if commaIdx < 0 {
		return nil, "", fmt.Errorf("data URI 格式错误: 缺少逗号分隔符")
	}

	meta := dataURI[5:commaIdx] // 跳过 "data:"
	payload := dataURI[commaIdx+1:]

	if !strings.Contains(meta, "base64") {
		return nil, "", fmt.Errorf("仅支持 base64 编码的 data URI")
	}

	data, err = base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return nil, "", fmt.Errorf("base64 解码失败: %w", err)
	}

	// 从 MIME 类型推断扩展名
	mimeType := strings.Split(meta, ";")[0] // e.g. "image/png"
	switch mimeType {
	case "image/png":
		ext = ".png"
	case "image/jpeg", "image/jpg":
		ext = ".jpg"
	case "image/webp":
		ext = ".webp"
	case "video/mp4":
		ext = ".mp4"
	case "video/webm":
		ext = ".webm"
	default:
		// 尝试从 MIME 中提取子类型作为扩展名
		parts := strings.SplitN(mimeType, "/", 2)
		if len(parts) == 2 && parts[1] != "" {
			ext = "." + parts[1]
		} else {
			ext = ".bin"
		}
	}

	return data, ext, nil
}

// ExtFromURL 从 URL 中提取文件扩展名，默认返回 fallback
func ExtFromURL(rawURL, fallback string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fallback
	}
	ext := path.Ext(u.Path)
	if ext == "" {
		return fallback
	}
	return ext
}

// extractAPIError 从 API 响应中提取可读的错误信息
// 支持 {"error": {"message": "xxx", "code": "yyy"}} 格式
func extractAPIError(result map[string]interface{}) string {
	if result == nil {
		return "未知错误"
	}
	if errObj, ok := result["error"]; ok {
		if errMap, ok := errObj.(map[string]interface{}); ok {
			msg, _ := errMap["message"].(string)
			code, _ := errMap["code"].(string)
			if msg != "" {
				if code != "" {
					return code + ": " + msg
				}
				return msg
			}
		}
	}
	return fmt.Sprintf("%v", result)
}

// truncate 截断字符串
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
