package sora

import (
	"encoding/json"
	"fmt"
	"time"
)

// PollImageTask 轮询图片任务进度，返回图片 URL
// onProgress 可为 nil，非 nil 时在每次轮询后回调进度
func (c *Client) PollImageTask(accessToken, taskID string, pollInterval, pollTimeout time.Duration, onProgress ProgressFunc) (string, error) {
	headers := baseHeaders(accessToken)

	startTime := time.Now()
	time.Sleep(2 * time.Second)

	for {
		elapsed := time.Since(startTime)
		if elapsed > pollTimeout {
			return "", fmt.Errorf("轮询超时 (%v)", pollTimeout)
		}

		body, err := c.doGet(soraBaseURL+"/v2/recent_tasks?limit=20", headers)
		if err != nil {
			time.Sleep(pollInterval)
			continue
		}

		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
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
			progressPct := parseProgressPct(task)

			if onProgress != nil {
				onProgress(Progress{
					Percent: progressPct,
					Status:  status,
					Elapsed: int(elapsed.Seconds()),
				})
			}

			if status == "failed" || status == "error" {
				reason, _ := task["failure_reason"].(string)
				return "", fmt.Errorf("任务失败: %s", reason)
			}

			if status == "succeeded" {
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
// onProgress 可为 nil，非 nil 时在每次轮询后回调进度
func (c *Client) PollVideoTask(accessToken, taskID string, pollInterval, pollTimeout time.Duration, onProgress ProgressFunc) error {
	headers := baseHeaders(accessToken)

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
			time.Sleep(pollInterval)
			continue
		}

		var tasks []map[string]interface{}
		if err := json.Unmarshal(body, &tasks); err != nil {
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

				progressPct := parseProgressPct(task)
				status, _ := task["status"].(string)

				if progressPct > maxProgress {
					maxProgress = progressPct
				}

				if onProgress != nil {
					onProgress(Progress{
						Percent: maxProgress,
						Status:  status,
						Elapsed: int(elapsed.Seconds()),
					})
				}

				if status == "failed" || status == "error" {
					reason, _ := task["failure_reason"].(string)
					return fmt.Errorf("任务失败: %s", reason)
				}
				break
			}
		}

		if !found {
			notFoundCount++
			if everFound && notFoundCount >= 2 {
				return nil
			}
			if !everFound && elapsed.Seconds() > 30 {
				return nil
			}
		}

		time.Sleep(pollInterval)
	}
}

// GetDownloadURL 从 drafts 接口获取下载链接
func (c *Client) GetDownloadURL(accessToken, taskID string) (string, error) {
	headers := baseHeaders(accessToken)

	for attempt := 0; attempt < 3; attempt++ {
		body, err := c.doGet(soraBaseURL+"/project_y/profile/drafts?limit=15", headers)
		if err != nil {
			if attempt < 2 {
				time.Sleep(3 * time.Second)
			}
			continue
		}

		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
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
		}

		if attempt < 2 {
			time.Sleep(3 * time.Second)
		}
	}

	return "", fmt.Errorf("在最近草稿中未找到任务 %s", taskID)
}

// ImageTaskResult 图片任务单次查询结果
type ImageTaskResult struct {
	Progress Progress
	Done     bool
	ImageURL string
	Err      error
}

// VideoTaskResult 视频任务单次查询结果
type VideoTaskResult struct {
	Progress Progress
	Done     bool
	Err      error
}

// QueryImageTaskOnce 单次查询图片任务状态（非阻塞，供 TUI 使用）
func (c *Client) QueryImageTaskOnce(accessToken, taskID string, startTime time.Time) ImageTaskResult {
	headers := baseHeaders(accessToken)

	elapsed := time.Since(startTime)

	body, err := c.doGet(soraBaseURL+"/v2/recent_tasks?limit=20", headers)
	if err != nil {
		return ImageTaskResult{Err: fmt.Errorf("查询失败: %w", err)}
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return ImageTaskResult{Err: fmt.Errorf("解析失败: %w", err)}
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
		progressPct := parseProgressPct(task)
		progress := Progress{Percent: progressPct, Status: status, Elapsed: int(elapsed.Seconds())}

		if status == "failed" || status == "error" {
			reason, _ := task["failure_reason"].(string)
			return ImageTaskResult{Progress: progress, Done: true, Err: fmt.Errorf("任务失败: %s", reason)}
		}

		if status == "succeeded" {
			generations, _ := task["generations"].([]interface{})
			for _, genRaw := range generations {
				gen, ok := genRaw.(map[string]interface{})
				if !ok {
					continue
				}
				url, _ := gen["url"].(string)
				if url != "" {
					return ImageTaskResult{Progress: progress, Done: true, ImageURL: url}
				}
			}
			return ImageTaskResult{Progress: progress, Done: true, Err: fmt.Errorf("任务成功但未找到图片 URL")}
		}

		return ImageTaskResult{Progress: progress}
	}

	return ImageTaskResult{Progress: Progress{Status: "waiting", Elapsed: int(elapsed.Seconds())}}
}

// QueryVideoTaskOnce 单次查询视频任务状态（非阻塞，供 TUI 使用）
// maxProgress 应传入之前的最大进度值，返回的结果中会包含更新后的进度
func (c *Client) QueryVideoTaskOnce(accessToken, taskID string, startTime time.Time, maxProgress int) VideoTaskResult {
	headers := baseHeaders(accessToken)

	elapsed := time.Since(startTime)

	body, err := c.doGet(soraBaseURL+"/nf/pending/v2", headers)
	if err != nil {
		return VideoTaskResult{Err: fmt.Errorf("查询失败: %w", err)}
	}

	var tasks []map[string]interface{}
	if err := json.Unmarshal(body, &tasks); err != nil {
		return VideoTaskResult{Err: fmt.Errorf("解析失败: %w", err)}
	}

	for _, task := range tasks {
		id, _ := task["id"].(string)
		if id != taskID {
			continue
		}

		progressPct := parseProgressPct(task)
		status, _ := task["status"].(string)

		if progressPct > maxProgress {
			maxProgress = progressPct
		}

		progress := Progress{Percent: maxProgress, Status: status, Elapsed: int(elapsed.Seconds())}

		if status == "failed" || status == "error" {
			reason, _ := task["failure_reason"].(string)
			return VideoTaskResult{Progress: progress, Done: true, Err: fmt.Errorf("任务失败: %s", reason)}
		}

		return VideoTaskResult{Progress: progress}
	}

	// 任务不在列表中，可能已完成
	return VideoTaskResult{
		Progress: Progress{Percent: maxProgress, Status: "not_found", Elapsed: int(elapsed.Seconds())},
		Done:     true,
	}
}

// CreditBalance 配额信息
type CreditBalance struct {
	RemainingCount      int  // 剩余可用次数
	RateLimitReached    bool // 是否触发速率限制
	AccessResetsInSec   int  // 访问权限重置时间（秒）
}

// GetCreditBalance 获取当前账号的可用次数和配额信息
func (c *Client) GetCreditBalance(accessToken string) (CreditBalance, error) {
	headers := baseHeaders(accessToken)
	headers["Accept"] = "application/json"

	body, err := c.doGet(soraBaseURL+"/nf/check", headers)
	if err != nil {
		return CreditBalance{}, fmt.Errorf("获取配额信息失败: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return CreditBalance{}, fmt.Errorf("解析响应失败: %w", err)
	}

	rateLimitInfo, _ := result["rate_limit_and_credit_balance"].(map[string]interface{})
	if rateLimitInfo == nil {
		return CreditBalance{}, fmt.Errorf("响应中无配额信息: %v", result)
	}

	remaining, _ := rateLimitInfo["estimated_num_videos_remaining"].(float64)
	rateLimitReached, _ := rateLimitInfo["rate_limit_reached"].(bool)
	accessResets, _ := rateLimitInfo["access_resets_in_seconds"].(float64)

	return CreditBalance{
		RemainingCount:    int(remaining),
		RateLimitReached:  rateLimitReached,
		AccessResetsInSec: int(accessResets),
	}, nil
}

// SubscriptionInfo 订阅信息
type SubscriptionInfo struct {
	PlanID    string // 套餐类型，例如 "chatgptplusplan"
	PlanTitle string // 套餐名称，例如 "ChatGPT Plus"
	EndTs     int64  // 订阅到期时间戳（秒）
}

// GetSubscriptionInfo 获取当前账号的订阅信息（套餐类型、到期时间）
func (c *Client) GetSubscriptionInfo(accessToken string) (SubscriptionInfo, error) {
	headers := baseHeaders(accessToken)
	headers["Accept"] = "application/json"

	body, err := c.doGet(soraBaseURL+"/billing/subscriptions", headers)
	if err != nil {
		return SubscriptionInfo{}, fmt.Errorf("获取订阅信息失败: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return SubscriptionInfo{}, fmt.Errorf("解析响应失败: %w", err)
	}

	data, _ := result["data"].([]interface{})
	if len(data) == 0 {
		return SubscriptionInfo{}, fmt.Errorf("未找到订阅信息")
	}

	sub, _ := data[0].(map[string]interface{})
	if sub == nil {
		return SubscriptionInfo{}, fmt.Errorf("订阅数据格式异常")
	}

	plan, _ := sub["plan"].(map[string]interface{})
	planID, _ := plan["id"].(string)
	planTitle, _ := plan["title"].(string)
	endTs, _ := sub["end_ts"].(float64)

	return SubscriptionInfo{
		PlanID:    planID,
		PlanTitle: planTitle,
		EndTs:     int64(endTs),
	}, nil
}

// parseProgressPct 从任务响应中解析进度百分比
func parseProgressPct(task map[string]interface{}) int {
	if p, ok := task["progress_pct"].(float64); ok {
		if p > 0 && p <= 1 {
			return int(p * 100)
		}
		return int(p)
	}
	return 0
}
