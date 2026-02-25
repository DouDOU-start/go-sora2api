package sora

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

const (
	soraBaseURL    = "https://sora.chatgpt.com/backend"
	chatgptBaseURL = "https://chatgpt.com"
	sentinelFlow   = "sora_2_create_task"
)

var desktopUserAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
}

var mobileUserAgents = []string{
	"Sora/1.2026.007 (Android 15; 24122RKC7C; build 2600700)",
	"Sora/1.2026.007 (Android 14; SM-G998B; build 2600700)",
	"Sora/1.2026.007 (Android 15; Pixel 8 Pro; build 2600700)",
	"Sora/1.2026.007 (Android 14; Pixel 7; build 2600700)",
	"Sora/1.2026.007 (Android 15; OnePlus 12; build 2600700)",
}

// Progress 任务进度信息
type Progress struct {
	Percent int    // 进度百分比 0-100
	Status  string // 任务状态
	Elapsed int    // 已耗时（秒）
}

// ProgressFunc 进度回调函数类型
type ProgressFunc func(Progress)

// Client Sora API 客户端
type Client struct {
	httpClient tls_client.HttpClient
	rng        *rand.Rand
	rngMu      sync.Mutex
}

// New 创建客户端，proxyURL 为空则不使用代理
func New(proxyURL string) (*Client, error) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_131),
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithNotFollowRedirects(),
	}

	if proxyURL != "" {
		options = append(options, tls_client.WithProxyUrl(proxyURL))
	}

	c, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		return nil, fmt.Errorf("创建 TLS 客户端失败: %w", err)
	}

	return &Client{
		httpClient: c,
		rng:        rand.New(rand.NewSource(rand.Int63())),
	}, nil
}

// randIntn 使用实例级别的随机数生成器，避免全局锁竞争
func (c *Client) randIntn(n int) int {
	c.rngMu.Lock()
	v := c.rng.Intn(n)
	c.rngMu.Unlock()
	return v
}

// randFloat64 使用实例级别的随机数生成器
func (c *Client) randFloat64() float64 {
	c.rngMu.Lock()
	v := c.rng.Float64()
	c.rngMu.Unlock()
	return v
}

// randRead 使用实例级别的随机数生成器填充字节切片
func (c *Client) randRead(b []byte) {
	c.rngMu.Lock()
	c.rng.Read(b)
	c.rngMu.Unlock()
}

func (c *Client) doPost(ctx context.Context, url string, headers map[string]string, body interface{}) (map[string]interface{}, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败 (HTTP %d): %w", resp.StatusCode, err)
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return result, fmt.Errorf("HTTP %d: %v", resp.StatusCode, result)
	}

	return result, nil
}

func (c *Client) doGet(ctx context.Context, url string, headers map[string]string) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	buf, err := readAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return buf, fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncate(string(buf), 200))
	}

	return buf, nil
}

func (c *Client) doPostMultipart(ctx context.Context, url string, headers map[string]string, body *bytes.Buffer, contentType string) (map[string]interface{}, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("Content-Type", contentType)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败 (HTTP %d): %w", resp.StatusCode, err)
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return result, fmt.Errorf("HTTP %d: %v", resp.StatusCode, result)
	}

	return result, nil
}

func (c *Client) doDelete(ctx context.Context, url string, headers map[string]string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		buf, _ := readAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncate(string(buf), 200))
	}

	return nil
}

// baseHeaders 返回基础请求头（Authorization + User-Agent + Origin + Referer）
func (c *Client) baseHeaders(accessToken string) map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + accessToken,
		"User-Agent":    mobileUserAgents[c.randIntn(len(mobileUserAgents))],
		"Origin":        "https://sora.chatgpt.com",
		"Referer":       "https://sora.chatgpt.com/",
	}
}

// jsonHeaders 返回 JSON POST 请求头
func (c *Client) jsonHeaders(accessToken string) map[string]string {
	h := c.baseHeaders(accessToken)
	h["Content-Type"] = "application/json"
	return h
}

// sentinelHeaders 返回带 sentinel token 的请求头
func (c *Client) sentinelHeaders(accessToken, sentinelToken string) map[string]string {
	h := c.jsonHeaders(accessToken)
	h["openai-sentinel-token"] = sentinelToken
	return h
}

// GenerateSentinelToken 获取 sentinel token（含 PoW 计算）
func (c *Client) GenerateSentinelToken(ctx context.Context, accessToken string) (string, error) {
	reqID := c.generateUUID()
	userAgent := desktopUserAgents[c.randIntn(len(desktopUserAgents))]
	powToken := c.getPowToken(userAgent)

	headers := map[string]string{
		"Accept":        "application/json, text/plain, */*",
		"Content-Type":  "application/json",
		"Origin":        "https://sora.chatgpt.com",
		"Referer":       "https://sora.chatgpt.com/",
		"User-Agent":    userAgent,
		"Authorization": "Bearer " + accessToken,
	}

	payload := map[string]string{
		"p":    powToken,
		"flow": sentinelFlow,
		"id":   reqID,
	}

	resp, err := c.doPost(ctx, chatgptBaseURL+"/backend-api/sentinel/req", headers, payload)
	if err != nil {
		return "", fmt.Errorf("sentinel 请求失败: %w", err)
	}

	return c.buildSentinelToken(sentinelFlow, reqID, powToken, resp, userAgent), nil
}
