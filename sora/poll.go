package sora

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

// PollImageTask 轮询图片任务进度，返回图片 URL
// onProgress 可为 nil，非 nil 时在每次轮询后回调进度
func (c *Client) PollImageTask(accessToken, taskID string, pollInterval, pollTimeout time.Duration, onProgress ProgressFunc) (string, error) {
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
	userAgent := mobileUserAgents[rand.Intn(len(mobileUserAgents))]
	headers := map[string]string{
		"Authorization": "Bearer " + accessToken,
		"User-Agent":    userAgent,
	}

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
