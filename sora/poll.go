package sora

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"
)

// ── 轮询接口的响应结构体（替代 map[string]interface{} 提升反序列化性能） ──

type recentTasksResp struct {
	TaskResponses []taskResponseItem `json:"task_responses"`
}

type taskResponseItem struct {
	ID            string           `json:"id"`
	Status        string           `json:"status"`
	FailureReason string           `json:"failure_reason"`
	ProgressPct   json.Number      `json:"progress_pct"`
	Generations   []generationItem `json:"generations"`
}

type generationItem struct {
	ID  string `json:"id"` // generation ID（如 gen_xxx），用于发布帖子
	URL string `json:"url"`
}

type pendingTaskItem struct {
	ID            string      `json:"id"`
	Status        string      `json:"status"`
	FailureReason string      `json:"failure_reason"`
	ProgressPct   json.Number `json:"progress_pct"`
}

type draftsResp struct {
	Items []draftItem `json:"items"`
}

type draftItem struct {
	TaskID          string `json:"task_id"`
	GenerationID    string `json:"generation_id"` // gen_xxx，用于发布帖子
	Kind            string `json:"kind"`
	ReasonStr       string `json:"reason_str"`
	MarkdownReason  string `json:"markdown_reason_str"`
	DownloadableURL string `json:"downloadable_url"`
	URL             string `json:"url"`
}

// parseProgressFromNumber 从 json.Number 解析进度百分比
func parseProgressFromNumber(n json.Number) int {
	f, err := n.Float64()
	if err != nil {
		return 0
	}
	if f > 0 && f <= 1 {
		return int(f * 100)
	}
	return int(f)
}

// backoff 计算退避间隔：min(base * 2^attempt, maxInterval)
func backoff(base time.Duration, attempt int, maxInterval time.Duration) time.Duration {
	d := time.Duration(float64(base) * math.Pow(1.5, float64(attempt)))
	if d > maxInterval {
		d = maxInterval
	}
	return d
}

// sleepWithContext 可被 ctx 取消的 sleep
func sleepWithContext(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// PollImageTask 轮询图片任务进度，返回图片 URL
// onProgress 可为 nil，非 nil 时在每次轮询后回调进度
func (c *Client) PollImageTask(ctx context.Context, accessToken, taskID string, pollInterval, pollTimeout time.Duration, onProgress ProgressFunc) (string, error) {
	headers := c.baseHeaders(accessToken)

	startTime := time.Now()
	if err := sleepWithContext(ctx, 2*time.Second); err != nil {
		return "", err
	}

	failCount := 0
	for {
		elapsed := time.Since(startTime)
		if elapsed > pollTimeout {
			return "", fmt.Errorf("轮询超时 (%v)", pollTimeout)
		}

		body, err := c.doGet(ctx, soraBaseURL+"/v2/recent_tasks?limit=20", headers)
		if err != nil {
			failCount++
			if err := sleepWithContext(ctx, backoff(pollInterval, failCount, 30*time.Second)); err != nil {
				return "", err
			}
			continue
		}
		failCount = 0

		var result recentTasksResp
		if err := json.Unmarshal(body, &result); err != nil {
			if err := sleepWithContext(ctx, pollInterval); err != nil {
				return "", err
			}
			continue
		}

		for i := range result.TaskResponses {
			task := &result.TaskResponses[i]
			if task.ID != taskID {
				continue
			}

			progressPct := parseProgressFromNumber(task.ProgressPct)

			if onProgress != nil {
				onProgress(Progress{
					Percent: progressPct,
					Status:  task.Status,
					Elapsed: int(elapsed.Seconds()),
				})
			}

			if task.Status == "failed" || task.Status == "error" {
				return "", fmt.Errorf("任务失败: %s", task.FailureReason)
			}

			if task.Status == "succeeded" {
				for j := range task.Generations {
					if task.Generations[j].URL != "" {
						return task.Generations[j].URL, nil
					}
				}
				return "", fmt.Errorf("任务成功但未找到图片 URL")
			}

			break
		}

		if err := sleepWithContext(ctx, pollInterval); err != nil {
			return "", err
		}
	}
}

// PollVideoTask 轮询视频任务进度
// onProgress 可为 nil，非 nil 时在每次轮询后回调进度
func (c *Client) PollVideoTask(ctx context.Context, accessToken, taskID string, pollInterval, pollTimeout time.Duration, onProgress ProgressFunc) error {
	headers := c.baseHeaders(accessToken)

	startTime := time.Now()
	maxProgress := 0
	everFound := false
	notFoundCount := 0

	if err := sleepWithContext(ctx, 2*time.Second); err != nil {
		return err
	}

	failCount := 0
	for {
		elapsed := time.Since(startTime)
		if elapsed > pollTimeout {
			return fmt.Errorf("轮询超时 (%v)", pollTimeout)
		}

		body, err := c.doGet(ctx, soraBaseURL+"/nf/pending/v2", headers)
		if err != nil {
			failCount++
			if err := sleepWithContext(ctx, backoff(pollInterval, failCount, 30*time.Second)); err != nil {
				return err
			}
			continue
		}
		failCount = 0

		var tasks []pendingTaskItem
		if err := json.Unmarshal(body, &tasks); err != nil {
			if err := sleepWithContext(ctx, pollInterval); err != nil {
				return err
			}
			continue
		}

		found := false
		for i := range tasks {
			task := &tasks[i]
			if task.ID != taskID {
				continue
			}

			found = true
			everFound = true
			notFoundCount = 0

			progressPct := parseProgressFromNumber(task.ProgressPct)
			if progressPct > maxProgress {
				maxProgress = progressPct
			}

			if onProgress != nil {
				onProgress(Progress{
					Percent: maxProgress,
					Status:  task.Status,
					Elapsed: int(elapsed.Seconds()),
				})
			}

			if task.Status == "failed" || task.Status == "error" {
				return fmt.Errorf("任务失败: %s", task.FailureReason)
			}
			break
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

		if err := sleepWithContext(ctx, pollInterval); err != nil {
			return err
		}
	}
}

// GetDownloadURL 从 drafts 接口获取下载链接
func (c *Client) GetDownloadURL(ctx context.Context, accessToken, taskID string) (string, error) {
	headers := c.baseHeaders(accessToken)

	for attempt := 0; attempt < 3; attempt++ {
		body, err := c.doGet(ctx, soraBaseURL+"/project_y/profile/drafts?limit=15", headers)
		if err != nil {
			if attempt < 2 {
				if err := sleepWithContext(ctx, backoff(3*time.Second, attempt, 15*time.Second)); err != nil {
					return "", err
				}
			}
			continue
		}

		var result draftsResp
		if err := json.Unmarshal(body, &result); err != nil {
			if attempt < 2 {
				if err := sleepWithContext(ctx, 3*time.Second); err != nil {
					return "", err
				}
			}
			continue
		}

		for i := range result.Items {
			item := &result.Items[i]
			if item.TaskID != taskID {
				continue
			}

			if item.Kind == "sora_content_violation" || item.ReasonStr != "" || item.MarkdownReason != "" {
				reason := item.ReasonStr
				if reason == "" {
					reason = item.MarkdownReason
				}
				if reason == "" {
					reason = "内容违反使用政策"
				}
				return "", fmt.Errorf("内容违规: %s", reason)
			}

			downloadURL := item.DownloadableURL
			if downloadURL == "" {
				downloadURL = item.URL
			}
			if downloadURL != "" {
				return downloadURL, nil
			}
		}

		if attempt < 2 {
			if err := sleepWithContext(ctx, 3*time.Second); err != nil {
				return "", err
			}
		}
	}

	return "", fmt.Errorf("在最近草稿中未找到任务 %s", taskID)
}

// GetGenerationID 从 recent_tasks 或 drafts 接口获取任务的 generation ID
// 用于发布帖子时传入 PublishVideo
func (c *Client) GetGenerationID(ctx context.Context, accessToken, taskID string) (string, error) {
	headers := c.baseHeaders(accessToken)

	// 先从 recent_tasks 获取
	body, err := c.doGet(ctx, soraBaseURL+"/v2/recent_tasks?limit=20", headers)
	if err == nil {
		var result recentTasksResp
		if err := json.Unmarshal(body, &result); err == nil {
			for i := range result.TaskResponses {
				task := &result.TaskResponses[i]
				if task.ID != taskID {
					continue
				}
				for j := range task.Generations {
					if task.Generations[j].ID != "" {
						return task.Generations[j].ID, nil
					}
				}
			}
		}
	}

	// 回退到 drafts 接口
	body, err = c.doGet(ctx, soraBaseURL+"/project_y/profile/drafts?limit=15", headers)
	if err != nil {
		return "", fmt.Errorf("获取 generation ID 失败: %w", err)
	}

	var drafts draftsResp
	if err := json.Unmarshal(body, &drafts); err != nil {
		return "", fmt.Errorf("解析 drafts 响应失败: %w", err)
	}

	for i := range drafts.Items {
		item := &drafts.Items[i]
		if item.TaskID == taskID && item.GenerationID != "" {
			return item.GenerationID, nil
		}
	}

	return "", fmt.Errorf("未找到任务 %s 的 generation ID", taskID)
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
func (c *Client) QueryImageTaskOnce(ctx context.Context, accessToken, taskID string, startTime time.Time) ImageTaskResult {
	headers := c.baseHeaders(accessToken)

	elapsed := time.Since(startTime)

	body, err := c.doGet(ctx, soraBaseURL+"/v2/recent_tasks?limit=20", headers)
	if err != nil {
		return ImageTaskResult{Err: fmt.Errorf("查询失败: %w", err)}
	}

	var result recentTasksResp
	if err := json.Unmarshal(body, &result); err != nil {
		return ImageTaskResult{Err: fmt.Errorf("解析失败: %w", err)}
	}

	for i := range result.TaskResponses {
		task := &result.TaskResponses[i]
		if task.ID != taskID {
			continue
		}

		progressPct := parseProgressFromNumber(task.ProgressPct)
		progress := Progress{Percent: progressPct, Status: task.Status, Elapsed: int(elapsed.Seconds())}

		if task.Status == "failed" || task.Status == "error" {
			return ImageTaskResult{Progress: progress, Done: true, Err: fmt.Errorf("任务失败: %s", task.FailureReason)}
		}

		if task.Status == "succeeded" {
			for j := range task.Generations {
				if task.Generations[j].URL != "" {
					return ImageTaskResult{Progress: progress, Done: true, ImageURL: task.Generations[j].URL}
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
func (c *Client) QueryVideoTaskOnce(ctx context.Context, accessToken, taskID string, startTime time.Time, maxProgress int) VideoTaskResult {
	headers := c.baseHeaders(accessToken)

	elapsed := time.Since(startTime)

	body, err := c.doGet(ctx, soraBaseURL+"/nf/pending/v2", headers)
	if err != nil {
		return VideoTaskResult{Err: fmt.Errorf("查询失败: %w", err)}
	}

	var tasks []pendingTaskItem
	if err := json.Unmarshal(body, &tasks); err != nil {
		return VideoTaskResult{Err: fmt.Errorf("解析失败: %w", err)}
	}

	for i := range tasks {
		task := &tasks[i]
		if task.ID != taskID {
			continue
		}

		progressPct := parseProgressFromNumber(task.ProgressPct)
		if progressPct > maxProgress {
			maxProgress = progressPct
		}

		progress := Progress{Percent: maxProgress, Status: task.Status, Elapsed: int(elapsed.Seconds())}

		if task.Status == "failed" || task.Status == "error" {
			return VideoTaskResult{Progress: progress, Done: true, Err: fmt.Errorf("任务失败: %s", task.FailureReason)}
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
	RemainingCount    int  // 剩余可用次数
	RateLimitReached  bool // 是否触发速率限制
	AccessResetsInSec int  // 访问权限重置时间（秒）
}

// creditBalanceResp 配额 API 响应结构体
type creditBalanceResp struct {
	RateLimitAndCreditBalance struct {
		EstimatedNumVideosRemaining float64 `json:"estimated_num_videos_remaining"`
		RateLimitReached            bool    `json:"rate_limit_reached"`
		AccessResetsInSeconds       float64 `json:"access_resets_in_seconds"`
	} `json:"rate_limit_and_credit_balance"`
}

// GetCreditBalance 获取当前账号的可用次数和配额信息
func (c *Client) GetCreditBalance(ctx context.Context, accessToken string) (CreditBalance, error) {
	headers := c.baseHeaders(accessToken)
	headers["Accept"] = "application/json"

	body, err := c.doGet(ctx, soraBaseURL+"/nf/check", headers)
	if err != nil {
		return CreditBalance{}, fmt.Errorf("获取配额信息失败: %w", err)
	}

	var result creditBalanceResp
	if err := json.Unmarshal(body, &result); err != nil {
		return CreditBalance{}, fmt.Errorf("解析响应失败: %w", err)
	}

	info := result.RateLimitAndCreditBalance
	return CreditBalance{
		RemainingCount:    int(info.EstimatedNumVideosRemaining),
		RateLimitReached:  info.RateLimitReached,
		AccessResetsInSec: int(info.AccessResetsInSeconds),
	}, nil
}

// SubscriptionInfo 订阅信息
type SubscriptionInfo struct {
	PlanID    string // 套餐类型，例如 "chatgptplusplan"
	PlanTitle string // 套餐名称，例如 "ChatGPT Plus"
	EndTs     int64  // 订阅到期时间戳（秒）
}

// flexFloat64 兼容 JSON 中数字或字符串形式的 float64
type flexFloat64 float64

func (f *flexFloat64) UnmarshalJSON(data []byte) error {
	// 先尝试直接解析为数字
	var num float64
	if err := json.Unmarshal(data, &num); err == nil {
		*f = flexFloat64(num)
		return nil
	}
	// 再尝试解析为字符串
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("end_ts 既不是数字也不是字符串: %s", string(data))
	}
	// 尝试解析为数字字符串
	if num, err := strconv.ParseFloat(s, 64); err == nil {
		*f = flexFloat64(num)
		return nil
	}
	// 尝试解析为 ISO 8601 时间字符串
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		*f = flexFloat64(t.Unix())
		return nil
	}
	return fmt.Errorf("end_ts 无法解析: %s", s)
}

// subscriptionResp 订阅 API 响应结构体
type subscriptionResp struct {
	Data []struct {
		Plan struct {
			ID    string `json:"id"`
			Title string `json:"title"`
		} `json:"plan"`
		EndTs flexFloat64 `json:"end_ts"`
	} `json:"data"`
}

// UserInfo 用户信息
type UserInfo struct {
	Email string
	Name  string
}

// GetUserInfo 获取当前用户信息（邮箱、名称等）
func (c *Client) GetUserInfo(ctx context.Context, accessToken string) (UserInfo, error) {
	headers := c.baseHeaders(accessToken)
	headers["Accept"] = "application/json"

	body, err := c.doGet(ctx, soraBaseURL+"/me", headers)
	if err != nil {
		return UserInfo{}, fmt.Errorf("获取用户信息失败: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return UserInfo{}, fmt.Errorf("解析用户信息失败: %w", err)
	}

	info := UserInfo{}
	if email, ok := result["email"].(string); ok {
		info.Email = email
	}
	if name, ok := result["name"].(string); ok {
		info.Name = name
	}
	return info, nil
}

// GetSubscriptionInfo 获取当前账号的订阅信息（套餐类型、到期时间）
func (c *Client) GetSubscriptionInfo(ctx context.Context, accessToken string) (SubscriptionInfo, error) {
	headers := c.baseHeaders(accessToken)
	headers["Accept"] = "application/json"

	body, err := c.doGet(ctx, soraBaseURL+"/billing/subscriptions", headers)
	if err != nil {
		return SubscriptionInfo{}, fmt.Errorf("获取订阅信息失败: %w", err)
	}

	var result subscriptionResp
	if err := json.Unmarshal(body, &result); err != nil {
		return SubscriptionInfo{}, fmt.Errorf("解析响应失败: %w", err)
	}

	if len(result.Data) == 0 {
		return SubscriptionInfo{}, fmt.Errorf("未找到订阅信息")
	}

	sub := result.Data[0]
	return SubscriptionInfo{
		PlanID:    sub.Plan.ID,
		PlanTitle: sub.Plan.Title,
		EndTs:     int64(sub.EndTs),
	}, nil
}
