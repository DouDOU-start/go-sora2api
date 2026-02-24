package client

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"go-sora2api/internal/pow"
	"go-sora2api/internal/util"

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

// SoraClient Sora API 客户端
type SoraClient struct {
	httpClient tls_client.HttpClient
}

// New 创建客户端
func New(proxyURL string) (*SoraClient, error) {
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

	return &SoraClient{httpClient: c}, nil
}

func (c *SoraClient) doPost(url string, headers map[string]string, body interface{}) (map[string]interface{}, error) {
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

func (c *SoraClient) doGet(url string, headers map[string]string) ([]byte, error) {
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

	buf, err := util.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return buf, fmt.Errorf("HTTP %d: %s", resp.StatusCode, util.Truncate(string(buf), 200))
	}

	return buf, nil
}

// GenerateSentinelToken 获取 sentinel token（含 PoW 计算）
func (c *SoraClient) GenerateSentinelToken(accessToken string) (string, error) {
	reqID := util.GenerateUUID()
	userAgent := desktopUserAgents[rand.Intn(len(desktopUserAgents))]
	powToken := pow.GetToken(userAgent)

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

	return pow.BuildSentinelToken(sentinelFlow, reqID, powToken, resp, userAgent), nil
}

// CreateVideoTask 创建视频生成任务
func (c *SoraClient) CreateVideoTask(accessToken, sentinelToken, prompt, orientation string, nFrames int, model, size string) (string, error) {
	userAgent := mobileUserAgents[rand.Intn(len(mobileUserAgents))]

	headers := map[string]string{
		"Authorization":         "Bearer " + accessToken,
		"openai-sentinel-token": sentinelToken,
		"Content-Type":          "application/json",
		"User-Agent":            userAgent,
		"Origin":                "https://sora.chatgpt.com",
		"Referer":               "https://sora.chatgpt.com/",
	}

	payload := map[string]interface{}{
		"kind":          "video",
		"prompt":        prompt,
		"orientation":   orientation,
		"size":          size,
		"n_frames":      nFrames,
		"model":         model,
		"inpaint_items": []interface{}{},
		"style_id":      nil,
	}

	resp, err := c.doPost(soraBaseURL+"/nf/create", headers, payload)
	if err != nil {
		return "", fmt.Errorf("创建任务失败: %w", err)
	}

	taskID, ok := resp["id"].(string)
	if !ok || taskID == "" {
		return "", fmt.Errorf("响应中无 task_id: %v", resp)
	}

	return taskID, nil
}

// CreateImageTask 创建图片生成任务
func (c *SoraClient) CreateImageTask(accessToken, sentinelToken, prompt string, width, height int) (string, error) {
	userAgent := mobileUserAgents[rand.Intn(len(mobileUserAgents))]

	headers := map[string]string{
		"Authorization":         "Bearer " + accessToken,
		"openai-sentinel-token": sentinelToken,
		"Content-Type":          "application/json",
		"User-Agent":            userAgent,
		"Origin":                "https://sora.chatgpt.com",
		"Referer":               "https://sora.chatgpt.com/",
	}

	payload := map[string]interface{}{
		"type":          "image_gen",
		"operation":     "simple_compose",
		"prompt":        prompt,
		"width":         width,
		"height":        height,
		"n_variants":    1,
		"n_frames":      1,
		"inpaint_items": []interface{}{},
	}

	resp, err := c.doPost(soraBaseURL+"/video_gen", headers, payload)
	if err != nil {
		return "", fmt.Errorf("创建图片任务失败: %w", err)
	}

	taskID, ok := resp["id"].(string)
	if !ok || taskID == "" {
		return "", fmt.Errorf("响应中无 task_id: %v", resp)
	}

	return taskID, nil
}

// PollImageTask 轮询图片任务进度，返回图片 URL
func (c *SoraClient) PollImageTask(accessToken, taskID string, pollInterval, pollTimeout time.Duration) (string, error) {
	userAgent := mobileUserAgents[rand.Intn(len(mobileUserAgents))]
	headers := map[string]string{
		"Authorization": "Bearer " + accessToken,
		"User-Agent":    userAgent,
	}

	startTime := time.Now()
	time.Sleep(2 * time.Second)

	for {
		elapsed := time.Since(startTime)
		if elapsed > pollTimeout {
			return "", fmt.Errorf("轮询超时 (%v)", pollTimeout)
		}

		body, err := c.doGet(soraBaseURL+"/v2/recent_tasks?limit=20", headers)
		if err != nil {
			fmt.Printf("  [轮询错误] %v\n", err)
			time.Sleep(pollInterval)
			continue
		}

		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			fmt.Printf("  [解析错误] %v\n", err)
			time.Sleep(pollInterval)
			continue
		}

		taskResponses, _ := result["task_responses"].([]interface{})
		for _, taskRaw := range taskResponses {
			task, ok := taskRaw.(map[string]interface{})
			if !ok {
				continue
			}

			id, _ := task["id"].(string)
			if id != taskID {
				continue
			}

			status, _ := task["status"].(string)
			progressPct := 0
			if p, ok := task["progress_pct"].(float64); ok {
				if p > 0 && p <= 1 {
					progressPct = int(p * 100)
				} else {
					progressPct = int(p)
				}
			}

			fmt.Printf("\r  进度: %d%%  状态: %s  耗时: %ds    ", progressPct, status, int(elapsed.Seconds()))

			if status == "failed" || status == "error" {
				fmt.Println()
				reason, _ := task["failure_reason"].(string)
				return "", fmt.Errorf("任务失败: %s", reason)
			}

			if status == "succeeded" {
				fmt.Println()
				// 从 generations 中提取图片 URL
				generations, _ := task["generations"].([]interface{})
				for _, genRaw := range generations {
					gen, ok := genRaw.(map[string]interface{})
					if !ok {
						continue
					}
					url, _ := gen["url"].(string)
					if url != "" {
						return url, nil
					}
				}
				return "", fmt.Errorf("任务成功但未找到图片 URL")
			}

			break
		}

		time.Sleep(pollInterval)
	}
}

// PollVideoTask 轮询视频任务进度
func (c *SoraClient) PollVideoTask(accessToken, taskID string, pollInterval, pollTimeout time.Duration) error {
	userAgent := mobileUserAgents[rand.Intn(len(mobileUserAgents))]
	headers := map[string]string{
		"Authorization": "Bearer " + accessToken,
		"User-Agent":    userAgent,
	}

	startTime := time.Now()
	maxProgress := 0
	everFound := false
	notFoundCount := 0

	time.Sleep(2 * time.Second)

	for {
		elapsed := time.Since(startTime)
		if elapsed > pollTimeout {
			return fmt.Errorf("轮询超时 (%v)", pollTimeout)
		}

		body, err := c.doGet(soraBaseURL+"/nf/pending/v2", headers)
		if err != nil {
			fmt.Printf("  [轮询错误] %v\n", err)
			time.Sleep(pollInterval)
			continue
		}

		var tasks []map[string]interface{}
		if err := json.Unmarshal(body, &tasks); err != nil {
			fmt.Printf("  [解析错误] %v\n", err)
			time.Sleep(pollInterval)
			continue
		}

		found := false
		for _, task := range tasks {
			id, _ := task["id"].(string)
			if id == taskID {
				found = true
				everFound = true
				notFoundCount = 0

				progressPct := 0
				if p, ok := task["progress_pct"].(float64); ok {
					if p > 0 && p <= 1 {
						progressPct = int(p * 100)
					} else {
						progressPct = int(p)
					}
				}

				status, _ := task["status"].(string)

				if progressPct > maxProgress {
					maxProgress = progressPct
				}

				fmt.Printf("\r  进度: %d%%  状态: %s  耗时: %ds    ", maxProgress, status, int(elapsed.Seconds()))

				if status == "failed" || status == "error" {
					fmt.Println()
					reason, _ := task["failure_reason"].(string)
					return fmt.Errorf("任务失败: %s", reason)
				}
				break
			}
		}

		if !found {
			notFoundCount++
			if everFound && notFoundCount >= 2 {
				fmt.Printf("\n[信息] 任务已从 pending 列表移除，视频生成完成! 耗时: %ds\n", int(elapsed.Seconds()))
				return nil
			}
			if !everFound && elapsed.Seconds() > 30 {
				fmt.Println("\n[信息] 任务未在 pending 列表中出现，可能已快速完成")
				return nil
			}
		}

		time.Sleep(pollInterval)
	}
}

// GetDownloadURL 从 drafts 接口获取下载链接
func (c *SoraClient) GetDownloadURL(accessToken, taskID string) (string, error) {
	userAgent := mobileUserAgents[rand.Intn(len(mobileUserAgents))]
	headers := map[string]string{
		"Authorization": "Bearer " + accessToken,
		"User-Agent":    userAgent,
	}

	for attempt := 0; attempt < 3; attempt++ {
		body, err := c.doGet(soraBaseURL+"/project_y/profile/drafts?limit=15", headers)
		if err != nil {
			fmt.Printf("  [获取结果错误] %v\n", err)
			if attempt < 2 {
				time.Sleep(3 * time.Second)
			}
			continue
		}

		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			fmt.Printf("  [解析错误] %v\n", err)
			if attempt < 2 {
				time.Sleep(3 * time.Second)
			}
			continue
		}

		items, _ := result["items"].([]interface{})
		for _, itemRaw := range items {
			item, ok := itemRaw.(map[string]interface{})
			if !ok {
				continue
			}

			tid, _ := item["task_id"].(string)
			if tid != taskID {
				continue
			}

			kind, _ := item["kind"].(string)
			reasonStr, _ := item["reason_str"].(string)
			markdownReason, _ := item["markdown_reason_str"].(string)

			if kind == "sora_content_violation" || reasonStr != "" || markdownReason != "" {
				reason := reasonStr
				if reason == "" {
					reason = markdownReason
				}
				if reason == "" {
					reason = "内容违反使用政策"
				}
				return "", fmt.Errorf("内容违规: %s", reason)
			}

			downloadURL, _ := item["downloadable_url"].(string)
			if downloadURL == "" {
				downloadURL, _ = item["url"].(string)
			}
			if downloadURL != "" {
				return downloadURL, nil
			}

			fmt.Println("  [警告] 找到任务但无下载链接，等待重试...")
		}

		if attempt < 2 {
			time.Sleep(3 * time.Second)
		}
	}

	return "", fmt.Errorf("在最近草稿中未找到任务 %s", taskID)
}
