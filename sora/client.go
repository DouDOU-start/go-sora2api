package sora

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"

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

	return &Client{httpClient: c}, nil
}

func (c *Client) doPost(url string, headers map[string]string, body interface{}) (map[string]interface{}, error) {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(bodyBytes)))
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

func (c *Client) doGet(url string, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
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

func (c *Client) doPostMultipart(url string, headers map[string]string, body *bytes.Buffer, contentType string) (map[string]interface{}, error) {
	req, err := http.NewRequest("POST", url, body)
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

// GenerateSentinelToken 获取 sentinel token（含 PoW 计算）
func (c *Client) GenerateSentinelToken(accessToken string) (string, error) {
	reqID := generateUUID()
	userAgent := desktopUserAgents[rand.Intn(len(desktopUserAgents))]
	powToken := getPowToken(userAgent)

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

	resp, err := c.doPost(chatgptBaseURL+"/backend-api/sentinel/req", headers, payload)
	if err != nil {
		return "", fmt.Errorf("sentinel 请求失败: %w", err)
	}

	return buildSentinelToken(sentinelFlow, reqID, powToken, resp, userAgent), nil
}
