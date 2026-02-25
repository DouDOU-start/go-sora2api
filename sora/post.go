package sora

import "fmt"

// PublishVideo 发布视频帖子，返回 postID
// generationID 为视频的生成 ID（格式如 gen_xxx）
func (c *Client) PublishVideo(accessToken, sentinelToken, generationID string) (string, error) {
	headers := sentinelHeaders(accessToken, sentinelToken)

	payload := map[string]interface{}{
		"attachments_to_create": []map[string]interface{}{
			{
				"generation_id": generationID,
				"kind":          "sora",
			},
		},
		"post_text": "",
	}

	resp, err := c.doPost(soraBaseURL+"/project_y/post", headers, payload)
	if err != nil {
		return "", fmt.Errorf("发布视频失败: %w", err)
	}

	post, _ := resp["post"].(map[string]interface{})
	if post == nil {
		return "", fmt.Errorf("响应中无 post: %v", resp)
	}

	postID, _ := post["id"].(string)
	if postID == "" {
		return "", fmt.Errorf("响应中无 post_id: %v", resp)
	}

	return postID, nil
}

// DeletePost 删除已发布的帖子
func (c *Client) DeletePost(accessToken, postID string) error {
	return c.doDelete(soraBaseURL+"/project_y/post/"+postID, baseHeaders(accessToken))
}
