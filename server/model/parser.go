package model

import (
	"fmt"
	"strings"
)

// ModelParams 解析后的模型参数
type ModelParams struct {
	Orientation string // landscape / portrait
	NFrames     int    // 300=10s / 450=15s / 750=25s
	Model       string // sy_8 / sy_ore
	Size        string // small / large
	Duration    int    // 10 / 15 / 25（秒）
}

// ParseModelName 解析模型名称为 Sora 原生参数
//
// 支持两种命名格式（K8Ray Creator 使用下划线前缀）：
//   - sora-2-landscape-10s / sora_video2-landscape-10s
//   - sora-2-pro-landscape-hd-10s / sora_video2-pro-landscape-hd-10s
func ParseModelName(name string) (*ModelParams, error) {
	// 统一转换为小写
	n := strings.ToLower(name)

	// 兼容 K8Ray Creator 命名（sora_video2-xxx → sora-2-xxx）
	n = strings.ReplaceAll(n, "sora_video2-", "sora-2-")

	params := &ModelParams{}

	// 解析 pro → sy_ore，否则 sy_8
	if strings.Contains(n, "-pro-") || strings.Contains(n, "-pro_") {
		params.Model = "sy_ore"
	} else {
		params.Model = "sy_8"
	}

	// 解析 hd → large，否则 small
	if strings.Contains(n, "-hd-") || strings.Contains(n, "_hd_") || strings.Contains(n, "-hd_") {
		params.Size = "large"
	} else {
		params.Size = "small"
	}

	// 解析方向
	if strings.Contains(n, "landscape") {
		params.Orientation = "landscape"
	} else if strings.Contains(n, "portrait") {
		params.Orientation = "portrait"
	} else {
		return nil, fmt.Errorf("模型名称中未找到方向（landscape/portrait）: %s", name)
	}

	// 解析时长 → nFrames
	switch {
	case strings.HasSuffix(n, "10s") || strings.Contains(n, "10s-"):
		params.Duration = 10
		params.NFrames = 300
	case strings.HasSuffix(n, "15s") || strings.Contains(n, "15s-"):
		params.Duration = 15
		params.NFrames = 450
	case strings.HasSuffix(n, "25s") || strings.Contains(n, "25s-"):
		params.Duration = 25
		params.NFrames = 750
	default:
		return nil, fmt.Errorf("模型名称中未找到时长（10s/15s/25s）: %s", name)
	}

	return params, nil
}

// SizeToResolution 将 size 转为分辨率字符串
func SizeToResolution(size, orientation string) string {
	if size == "large" {
		if orientation == "landscape" {
			return "1920x1080"
		}
		return "1080x1920"
	}
	if orientation == "landscape" {
		return "1280x720"
	}
	return "720x1280"
}
