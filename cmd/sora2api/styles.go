package main

import "github.com/charmbracelet/lipgloss"

var (
	// 颜色
	colorPrimary = lipgloss.AdaptiveColor{Light: "#5B5BD6", Dark: "#7C7CFF"}
	colorSuccess = lipgloss.AdaptiveColor{Light: "#30A46C", Dark: "#3DD68C"}
	colorError   = lipgloss.AdaptiveColor{Light: "#E5484D", Dark: "#FF6369"}
	colorWarning = lipgloss.AdaptiveColor{Light: "#F5A623", Dark: "#FFB84D"}
	colorMuted   = lipgloss.AdaptiveColor{Light: "#8B8D98", Dark: "#70727F"}
	colorBorder  = lipgloss.AdaptiveColor{Light: "#D3D4DB", Dark: "#3A3B45"}

	// 标题
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			MarginBottom(1)

	// 状态栏
	statusBarStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			PaddingLeft(2)

	// 菜单分组标题
	groupTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			PaddingLeft(2).
			MarginTop(1)

	// 菜单项
	menuItemStyle = lipgloss.NewStyle().
			PaddingLeft(4)

	menuSelectedStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(colorPrimary).
				Bold(true)

	menuDescStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	// 输入框标签
	labelStyle = lipgloss.NewStyle().
			Bold(true).
			MarginTop(1)

	// 成功/错误
	successStyle = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorError).
			Bold(true)

	warnStyle = lipgloss.NewStyle().
			Foreground(colorWarning)

	// 帮助栏
	helpStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			MarginTop(1)

	// 边框容器
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(1, 2)

	// URL 链接样式
	urlStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Underline(true)

	// 账号信息卡片
	accountCardStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorBorder).
				Padding(0, 1).
				MarginLeft(2)
)
