package sora

import (
	"bytes"
	"fmt"
	"math/rand"
	"mime/multipart"
	"net/textproto"
	"path/filepath"
	"strings"
)

// UploadImage 上传图片，返回 mediaID，用于图生图/图生视频
// imageData 为图片二进制数据，filename 为文件名（如 "image.png"）
func (c *Client) UploadImage(accessToken string, imageData []byte, filename string) (string, error) {
	userAgent := mobileUserAgents[rand.Intn(len(mobileUserAgents))]

	headers := map[string]string{
		"Authorization": "Bearer " + accessToken,
		"User-Agent":    userAgent,
		"Origin":        "https://sora.chatgpt.com",
		"Referer":       "https://sora.chatgpt.com/",
	}

	// 检测 MIME 类型
	mimeType := "image/png"
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		mimeType = "image/jpeg"
	case ".webp":
		mimeType = "image/webp"
	}

	// 构造 multipart body
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 添加文件部分
	partHeader := make(textproto.MIMEHeader)
	partHeader.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, filename))
	partHeader.Set("Content-Type", mimeType)
	part, err := writer.CreatePart(partHeader)
	if err != nil {
		return "", err
	}
	if _, err := part.Write(imageData); err != nil {
		return "", err
	}

	// 添加 file_name 字段
	if err := writer.WriteField("file_name", filename); err != nil {
		return "", err
	}
	writer.Close()

	resp, err := c.doPostMultipart(soraBaseURL+"/uploads", headers, &buf, writer.FormDataContentType())
	if err != nil {
		return "", fmt.Errorf("上传图片失败: %w", err)
	}

	mediaID, ok := resp["id"].(string)
	if !ok || mediaID == "" {
		return "", fmt.Errorf("响应中无 media_id: %v", resp)
	}

	return mediaID, nil
}

// CreateVideoTask 创建视频生成任务（文生视频）
// 如需图生视频，先调用 UploadImage 获取 mediaID，再传入 CreateVideoTaskWithImage
func (c *Client) CreateVideoTask(accessToken, sentinelToken, prompt, orientation string, nFrames int, model, size string) (string, error) {
	return c.CreateVideoTaskWithOptions(accessToken, sentinelToken, prompt, orientation, nFrames, model, size, "", "")
}

// CreateVideoTaskWithImage 创建图生视频任务
func (c *Client) CreateVideoTaskWithImage(accessToken, sentinelToken, prompt, orientation string, nFrames int, model, size, mediaID string) (string, error) {
	return c.CreateVideoTaskWithOptions(accessToken, sentinelToken, prompt, orientation, nFrames, model, size, mediaID, "")
}

// CreateVideoTaskWithOptions 创建视频任务的完整方法
// mediaID 为空表示文生视频，非空表示图生视频
// styleID 为空表示无风格，可选值见 ValidStyles
func (c *Client) CreateVideoTaskWithOptions(accessToken, sentinelToken, prompt, orientation string, nFrames int, model, size, mediaID, styleID string) (string, error) {
	userAgent := mobileUserAgents[rand.Intn(len(mobileUserAgents))]

	headers := map[string]string{
		"Authorization":         "Bearer " + accessToken,
		"openai-sentinel-token": sentinelToken,
		"Content-Type":          "application/json",
		"User-Agent":            userAgent,
		"Origin":                "https://sora.chatgpt.com",
		"Referer":               "https://sora.chatgpt.com/",
	}

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
		"kind":          "video",
		"prompt":        prompt,
		"orientation":   orientation,
		"size":          size,
		"n_frames":      nFrames,
		"model":         model,
		"inpaint_items": inpaintItems,
		"style_id":      nilIfEmpty(styleID),
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

// CreateImageTask 创建图片生成任务（文生图）
// 如需图生图，先调用 UploadImage 获取 mediaID，再传入 CreateImageTaskWithImage
func (c *Client) CreateImageTask(accessToken, sentinelToken, prompt string, width, height int) (string, error) {
	return c.CreateImageTaskWithImage(accessToken, sentinelToken, prompt, width, height, "")
}

// CreateImageTaskWithImage 创建图生图任务
// mediaID 为空表示文生图，非空表示图生图
func (c *Client) CreateImageTaskWithImage(accessToken, sentinelToken, prompt string, width, height int, mediaID string) (string, error) {
	userAgent := mobileUserAgents[rand.Intn(len(mobileUserAgents))]

	headers := map[string]string{
		"Authorization":         "Bearer " + accessToken,
		"openai-sentinel-token": sentinelToken,
		"Content-Type":          "application/json",
		"User-Agent":            userAgent,
		"Origin":                "https://sora.chatgpt.com",
		"Referer":               "https://sora.chatgpt.com/",
	}

	operation := "simple_compose"
	inpaintItems := []interface{}{}
	if mediaID != "" {
		operation = "remix"
		inpaintItems = []interface{}{
			map[string]interface{}{
				"type":            "image",
				"frame_index":     0,
				"upload_media_id": mediaID,
			},
		}
	}

	payload := map[string]interface{}{
		"type":          "image_gen",
		"operation":     operation,
		"prompt":        prompt,
		"width":         width,
		"height":        height,
		"n_variants":    1,
		"n_frames":      1,
		"inpaint_items": inpaintItems,
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

// RemixVideo 基于已有视频创建 Remix 任务
// remixTargetID 为 Sora 分享链接中的视频 ID，格式: s_[hex32]
func (c *Client) RemixVideo(accessToken, sentinelToken, remixTargetID, prompt, orientation string, nFrames int, styleID string) (string, error) {
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
		"kind":               "video",
		"prompt":             prompt,
		"inpaint_items":      []interface{}{},
		"remix_target_id":    remixTargetID,
		"cameo_ids":          []interface{}{},
		"cameo_replacements": map[string]interface{}{},
		"model":              "sy_8",
		"orientation":        orientation,
		"n_frames":           nFrames,
		"style_id":           nilIfEmpty(styleID),
	}

	resp, err := c.doPost(soraBaseURL+"/nf/create", headers, payload)
	if err != nil {
		return "", fmt.Errorf("创建 Remix 任务失败: %w", err)
	}

	taskID, ok := resp["id"].(string)
	if !ok || taskID == "" {
		return "", fmt.Errorf("响应中无 task_id: %v", resp)
	}

	return taskID, nil
}

// EnhancePrompt 使用 Sora 的提示词优化 API 增强提示词
// expansionLevel: "medium" 或 "long"
// durationSec: 10、15 或 20
func (c *Client) EnhancePrompt(accessToken, prompt, expansionLevel string, durationSec int) (string, error) {
	userAgent := mobileUserAgents[rand.Intn(len(mobileUserAgents))]

	headers := map[string]string{
		"Authorization": "Bearer " + accessToken,
		"Content-Type":  "application/json",
		"User-Agent":    userAgent,
	}

	payload := map[string]interface{}{
		"prompt":          prompt,
		"expansion_level": expansionLevel,
		"duration_s":      durationSec,
	}

	resp, err := c.doPost(soraBaseURL+"/editor/enhance_prompt", headers, payload)
	if err != nil {
		return "", fmt.Errorf("提示词优化失败: %w", err)
	}

	enhanced, _ := resp["enhanced_prompt"].(string)
	if enhanced == "" {
		return prompt, nil
	}
	return enhanced, nil
}
