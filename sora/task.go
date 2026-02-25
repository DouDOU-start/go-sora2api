package sora

import (
	"bytes"
	"encoding/json"
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
	headers := baseHeaders(accessToken)

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
	headers := sentinelHeaders(accessToken, sentinelToken)

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
	headers := sentinelHeaders(accessToken, sentinelToken)

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
	headers := sentinelHeaders(accessToken, sentinelToken)

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
// durationSec: 5、10、15 或 25
func (c *Client) EnhancePrompt(accessToken, prompt, expansionLevel string, durationSec int) (string, error) {
	headers := jsonHeaders(accessToken)

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

// DefaultSoraClientID 默认的 Sora 客户端 ID
const DefaultSoraClientID = "app_1LOVEceTvrP2tHFDNnrPLQkJ"

// RefreshAccessToken 使用 refresh_token 获取新的 access_token
// 返回新的 accessToken 和 refreshToken（OpenAI 每次刷新都会返回新的 refresh_token）
func (c *Client) RefreshAccessToken(refreshToken, clientID string) (newAccessToken, newRefreshToken string, err error) {
	if clientID == "" {
		clientID = DefaultSoraClientID
	}

	headers := map[string]string{
		"Content-Type": "application/json",
		"User-Agent":   mobileUserAgents[rand.Intn(len(mobileUserAgents))],
	}

	payload := map[string]string{
		"client_id":    clientID,
		"grant_type":   "refresh_token",
		"redirect_uri": "com.openai.sora://auth.openai.com/android/com.openai.sora/callback",
		"refresh_token": refreshToken,
	}

	resp, err := c.doPost("https://auth.openai.com/oauth/token", headers, payload)
	if err != nil {
		return "", "", fmt.Errorf("刷新 token 失败: %w", err)
	}

	newAccessToken, _ = resp["access_token"].(string)
	newRefreshToken, _ = resp["refresh_token"].(string)
	if newAccessToken == "" {
		return "", "", fmt.Errorf("响应中无 access_token: %v", resp)
	}

	return newAccessToken, newRefreshToken, nil
}

// GetWatermarkFreeURL 获取 Sora 视频的无水印下载链接
// 需要使用 RefreshAccessToken 获取的 token，普通 ChatGPT access_token 不支持
// videoID 为 Sora 分享链接中的视频 ID，也可以传入完整链接（自动提取 ID）
func (c *Client) GetWatermarkFreeURL(accessToken, videoID string) (string, error) {
	// 如果传入的是完整链接，自动提取 ID
	if extracted := ExtractVideoID(videoID); extracted != "" {
		videoID = extracted
	}

	userAgent := mobileUserAgents[rand.Intn(len(mobileUserAgents))]

	headers := map[string]string{
		"Authorization":    "Bearer " + accessToken,
		"User-Agent":       userAgent,
		"Accept":           "application/json",
		"oai-package-name": "com.openai.sora",
	}

	body, err := c.doGet(soraBaseURL+"/project_y/post/"+videoID, headers)
	if err != nil {
		return "", fmt.Errorf("获取视频信息失败: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	// 提取 post.attachments[0].encodings.source.path
	post, _ := result["post"].(map[string]interface{})
	if post == nil {
		return "", fmt.Errorf("响应中无 post 字段: %v", result)
	}

	attachments, _ := post["attachments"].([]interface{})
	if len(attachments) == 0 {
		return "", fmt.Errorf("响应中无 attachments")
	}

	attachment, _ := attachments[0].(map[string]interface{})
	if attachment == nil {
		return "", fmt.Errorf("无法解析 attachment")
	}

	encodings, _ := attachment["encodings"].(map[string]interface{})
	if encodings == nil {
		return "", fmt.Errorf("响应中无 encodings")
	}

	source, _ := encodings["source"].(map[string]interface{})
	if source == nil {
		return "", fmt.Errorf("响应中无 source encoding")
	}

	path, _ := source["path"].(string)
	if path == "" {
		return "", fmt.Errorf("响应中无下载链接")
	}

	return path, nil
}
