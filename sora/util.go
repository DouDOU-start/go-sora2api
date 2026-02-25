package sora

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path"
	"strings"
)

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
