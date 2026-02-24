package sora

import (
	"regexp"
	"strings"
)

// ValidStyles 有效的视频风格列表
var ValidStyles = []string{
	"festive", "kakalaka", "news", "selfie", "handheld",
	"golden", "anime", "retro", "nostalgic", "comic",
}

// ExtractStyle 从提示词中提取 {style} 风格标记
// 返回清理后的提示词和风格 ID（无风格时为空字符串）
// 示例: "一只猫 {anime}" -> ("一只猫", "anime")
func ExtractStyle(prompt string) (string, string) {
	re := regexp.MustCompile(`\{([^}]+)\}`)
	match := re.FindStringSubmatch(prompt)
	if match == nil {
		return prompt, ""
	}

	candidate := strings.TrimSpace(match[1])
	if strings.Contains(candidate, " ") {
		return prompt, ""
	}

	lower := strings.ToLower(candidate)
	for _, s := range ValidStyles {
		if s == lower {
			cleaned := re.ReplaceAllString(prompt, "")
			cleaned = strings.Join(strings.Fields(cleaned), " ")
			return cleaned, lower
		}
	}

	return prompt, ""
}

// ExtractRemixID 从文本或 URL 中提取 Remix 视频 ID
// 支持: https://sora.chatgpt.com/p/s_[hex32] 或直接 s_[hex32]
func ExtractRemixID(text string) string {
	re := regexp.MustCompile(`s_[a-f0-9]{32}`)
	match := re.FindString(text)
	return match
}

// ExtractVideoID 从 Sora 分享链接或文本中提取视频 ID
// 支持: https://sora.chatgpt.com/p/s_xxx 或直接 s_xxx 等格式
func ExtractVideoID(text string) string {
	re := regexp.MustCompile(`sora\.chatgpt\.com/p/([a-zA-Z0-9_]+)`)
	match := re.FindStringSubmatch(text)
	if match != nil {
		return match[1]
	}
	return ""
}

func nilIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
