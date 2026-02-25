package sora

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/textproto"
	"time"
)

// CameoStatus 角色处理状态
type CameoStatus struct {
	ID              string // 角色 cameo ID
	Status          string // 处理状态
	DisplayNameHint string // 推荐的显示名称
	UsernameHint    string // 推荐的用户名
	ProfileAssetURL string // 角色头像 URL
}

// UploadCharacterVideo 上传角色视频，返回 cameoID
// videoData 为视频二进制数据（mp4 格式），timestamps 默认 "0,3"
func (c *Client) UploadCharacterVideo(ctx context.Context, accessToken string, videoData []byte) (string, error) {
	headers := c.baseHeaders(accessToken)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 添加视频文件
	partHeader := make(textproto.MIMEHeader)
	partHeader.Set("Content-Disposition", `form-data; name="file"; filename="video.mp4"`)
	partHeader.Set("Content-Type", "video/mp4")
	part, err := writer.CreatePart(partHeader)
	if err != nil {
		return "", err
	}
	if _, err := part.Write(videoData); err != nil {
		return "", err
	}

	// 添加 timestamps 字段
	if err := writer.WriteField("timestamps", "0,3"); err != nil {
		return "", err
	}
	writer.Close()

	resp, err := c.doPostMultipart(ctx, soraBaseURL+"/characters/upload", headers, &buf, writer.FormDataContentType())
	if err != nil {
		return "", fmt.Errorf("上传角色视频失败: %w", err)
	}

	cameoID, ok := resp["id"].(string)
	if !ok || cameoID == "" {
		return "", fmt.Errorf("响应中无 cameo_id: %v", resp)
	}

	return cameoID, nil
}

// GetCameoStatus 获取角色处理状态
func (c *Client) GetCameoStatus(ctx context.Context, accessToken, cameoID string) (CameoStatus, error) {
	headers := c.baseHeaders(accessToken)

	body, err := c.doGet(ctx, soraBaseURL+"/project_y/cameos/in_progress/"+cameoID, headers)
	if err != nil {
		return CameoStatus{}, fmt.Errorf("获取角色状态失败: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return CameoStatus{}, fmt.Errorf("解析响应失败: %w", err)
	}

	status := CameoStatus{
		ID: cameoID,
	}
	status.Status, _ = result["status"].(string)
	status.DisplayNameHint, _ = result["display_name_hint"].(string)
	status.UsernameHint, _ = result["username_hint"].(string)
	status.ProfileAssetURL, _ = result["profile_asset_url"].(string)

	return status, nil
}

// PollCameoStatus 轮询角色处理状态直到完成
func (c *Client) PollCameoStatus(ctx context.Context, accessToken, cameoID string, pollInterval, pollTimeout time.Duration, onProgress ProgressFunc) (CameoStatus, error) {
	startTime := time.Now()
	if err := sleepWithContext(ctx, 2*time.Second); err != nil {
		return CameoStatus{}, err
	}

	failCount := 0
	for {
		elapsed := time.Since(startTime)
		if elapsed > pollTimeout {
			return CameoStatus{}, fmt.Errorf("轮询超时 (%v)", pollTimeout)
		}

		status, err := c.GetCameoStatus(ctx, accessToken, cameoID)
		if err != nil {
			failCount++
			if err := sleepWithContext(ctx, backoff(pollInterval, failCount, 30*time.Second)); err != nil {
				return CameoStatus{}, err
			}
			continue
		}
		failCount = 0

		if onProgress != nil {
			onProgress(Progress{
				Status:  status.Status,
				Elapsed: int(elapsed.Seconds()),
			})
		}

		if status.Status == "failed" || status.Status == "error" {
			return status, fmt.Errorf("角色处理失败: %s", status.Status)
		}

		// 当 profile_asset_url 出现时表示处理完成
		if status.ProfileAssetURL != "" {
			return status, nil
		}

		if err := sleepWithContext(ctx, pollInterval); err != nil {
			return CameoStatus{}, err
		}
	}
}

// DownloadCharacterImage 下载角色头像图片
func (c *Client) DownloadCharacterImage(ctx context.Context, imageURL string) ([]byte, error) {
	userAgent := desktopUserAgents[c.randIntn(len(desktopUserAgents))]
	headers := map[string]string{
		"User-Agent": userAgent,
	}

	body, err := c.doGet(ctx, imageURL, headers)
	if err != nil {
		return nil, fmt.Errorf("下载角色图片失败: %w", err)
	}

	return body, nil
}

// UploadCharacterImage 上传角色头像图片，返回 assetPointer
// imageData 为图片二进制数据（webp 格式）
func (c *Client) UploadCharacterImage(ctx context.Context, accessToken string, imageData []byte) (string, error) {
	headers := c.baseHeaders(accessToken)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 添加图片文件
	partHeader := make(textproto.MIMEHeader)
	partHeader.Set("Content-Disposition", `form-data; name="file"; filename="profile.webp"`)
	partHeader.Set("Content-Type", "image/webp")
	part, err := writer.CreatePart(partHeader)
	if err != nil {
		return "", err
	}
	if _, err := part.Write(imageData); err != nil {
		return "", err
	}

	// 添加 use_case 字段
	if err := writer.WriteField("use_case", "profile"); err != nil {
		return "", err
	}
	writer.Close()

	resp, err := c.doPostMultipart(ctx, soraBaseURL+"/project_y/file/upload", headers, &buf, writer.FormDataContentType())
	if err != nil {
		return "", fmt.Errorf("上传角色头像失败: %w", err)
	}

	assetPointer, ok := resp["asset_pointer"].(string)
	if !ok || assetPointer == "" {
		return "", fmt.Errorf("响应中无 asset_pointer: %v", resp)
	}

	return assetPointer, nil
}

// FinalizeCharacter 定稿角色，返回 characterID
func (c *Client) FinalizeCharacter(ctx context.Context, accessToken, cameoID, username, displayName, profileAssetPointer string) (string, error) {
	headers := c.jsonHeaders(accessToken)

	payload := map[string]interface{}{
		"cameo_id":               cameoID,
		"username":               username,
		"display_name":           displayName,
		"profile_asset_pointer":  profileAssetPointer,
		"instruction_set":        nil,
		"safety_instruction_set": nil,
	}

	resp, err := c.doPost(ctx, soraBaseURL+"/characters/finalize", headers, payload)
	if err != nil {
		return "", fmt.Errorf("定稿角色失败: %w", err)
	}

	character, _ := resp["character"].(map[string]interface{})
	if character == nil {
		return "", fmt.Errorf("响应中无 character: %v", resp)
	}

	characterID, _ := character["character_id"].(string)
	if characterID == "" {
		return "", fmt.Errorf("响应中无 character_id: %v", resp)
	}

	return characterID, nil
}

// SetCharacterPublic 设置角色为公开
func (c *Client) SetCharacterPublic(ctx context.Context, accessToken, cameoID string) error {
	headers := c.jsonHeaders(accessToken)

	payload := map[string]interface{}{
		"visibility": "public",
	}

	_, err := c.doPost(ctx, soraBaseURL+"/project_y/cameos/by_id/"+cameoID+"/update_v2", headers, payload)
	if err != nil {
		return fmt.Errorf("设置角色公开失败: %w", err)
	}

	return nil
}

// DeleteCharacter 删除角色
func (c *Client) DeleteCharacter(ctx context.Context, accessToken, characterID string) error {
	return c.doDelete(ctx, soraBaseURL+"/project_y/characters/"+characterID, c.baseHeaders(accessToken))
}
