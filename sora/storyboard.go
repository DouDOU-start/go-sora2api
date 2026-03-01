package sora

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// IsStoryboardPrompt 检测提示词是否为分镜模式格式
// 格式: [5.0s]场景描述 [5.0s]场景描述
func IsStoryboardPrompt(prompt string) bool {
	re := regexp.MustCompile(`\[\d+(?:\.\d+)?s\]`)
	matches := re.FindAllString(prompt, -1)
	return len(matches) >= 1
}

// FormatStoryboardPrompt 将分镜格式的提示词转换为 API 所需格式
// 输入: "总体描述\n[5.0s]场景1 [5.0s]场景2"
// 输出: "current timeline:\nShot 1:\nduration: 5.0sec\nScene: 场景1\n..."
func FormatStoryboardPrompt(prompt string) string {
	re := regexp.MustCompile(`\[(\d+(?:\.\d+)?)s\]([^[\]]+)`)
	matches := re.FindAllStringSubmatch(prompt, -1)
	if len(matches) == 0 {
		return prompt
	}

	// 提取非分镜部分作为 instructions
	instructions := re.ReplaceAllString(prompt, "")
	instructions = strings.TrimSpace(instructions)

	var b strings.Builder
	b.WriteString("current timeline:\n")

	for i, match := range matches {
		duration := match[1]
		scene := strings.TrimSpace(match[2])
		fmt.Fprintf(&b, "Shot %d:\n", i+1)
		fmt.Fprintf(&b, "duration: %ssec\n", duration)
		fmt.Fprintf(&b, "Scene: %s\n\n", scene)
	}

	if instructions != "" {
		b.WriteString("instructions:\n")
		b.WriteString(instructions)
	}

	return b.String()
}

// CreateStoryboardTask 创建分镜视频任务
// prompt 应为分镜格式（会自动调用 FormatStoryboardPrompt 转换）
func (c *Client) CreateStoryboardTask(ctx context.Context, accessToken, sentinelToken, prompt, orientation string, nFrames int, mediaID, styleID string) (string, error) {
	headers := c.sentinelHeaders(accessToken, sentinelToken)

	// 格式化分镜提示词
	formattedPrompt := FormatStoryboardPrompt(prompt)

	inpaintItems := []interface{}{}
	if mediaID != "" {
		inpaintItems = []interface{}{
			map[string]interface{}{
				"kind":      "upload",
				"upload_id": mediaID,
			},
		}
	}

	payload := map[string]interface{}{
		"kind":               "video",
		"prompt":             formattedPrompt,
		"title":              "Draft your video",
		"orientation":        orientation,
		"size":               "small",
		"n_frames":           nFrames,
		"storyboard_id":      nil,
		"inpaint_items":      inpaintItems,
		"remix_target_id":    nil,
		"model":              "sy_8",
		"metadata":           nil,
		"style_id":           nilIfEmpty(styleID),
		"cameo_ids":          nil,
		"cameo_replacements": nil,
		"audio_caption":      nil,
		"audio_transcript":   nil,
		"video_caption":      nil,
	}

	resp, err := c.doPost(ctx, soraBaseURL+"/nf/create/storyboard", headers, payload)
	if err != nil {
		return "", fmt.Errorf("创建分镜任务失败: %w", err)
	}

	taskID, ok := resp["id"].(string)
	if !ok || taskID == "" {
		return "", fmt.Errorf("响应中无 task_id: %v", resp)
	}

	return taskID, nil
}
