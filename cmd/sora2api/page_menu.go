package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type menuItem struct {
	name     string
	desc     string
	funcType funcType
	isAction bool // true 表示非功能项（如重新配置）
	action   string
}

type menuGroup struct {
	title string
	items []menuItem
}

type menuModel struct {
	groups   []menuGroup
	cursor   int // 全局游标
	width    int
	allItems []menuItem // 扁平化列表
}

func newMenuModel() menuModel {
	groups := []menuGroup{
		{
			title: "图片生成",
			items: []menuItem{
				{name: "文生图", desc: "从文字描述生成图片", funcType: funcTextToImage},
				{name: "图生图", desc: "基于参考图生成新图片", funcType: funcImageToImage},
			},
		},
		{
			title: "视频生成",
			items: []menuItem{
				{name: "文生视频", desc: "从文字描述生成视频", funcType: funcTextToVideo},
				{name: "图生视频", desc: "从图片生成视频动画", funcType: funcImageToVideo},
				{name: "Remix 视频", desc: "基于已有视频重新创作", funcType: funcRemixVideo},
			},
		},
		{
			title: "工具",
			items: []menuItem{
				{name: "提示词优化", desc: "AI 增强你的提示词", funcType: funcEnhancePrompt},
				{name: "去水印链接", desc: "获取无水印下载链接", funcType: funcWatermarkFree},
			},
		},
		{
			title: "设置",
			items: []menuItem{
				{name: "重新配置", desc: "更换 Token 或代理", isAction: true, action: "reset"},
			},
		},
	}

	// 扁平化
	var all []menuItem
	for _, g := range groups {
		all = append(all, g.items...)
	}

	return menuModel{
		groups:   groups,
		allItems: all,
	}
}

func (m menuModel) Init() tea.Cmd { return nil }

func (m menuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.allItems)-1 {
				m.cursor++
			}
		case "enter":
			item := m.allItems[m.cursor]
			if item.isAction {
				switch item.action {
				case "reset":
					return m, func() tea.Msg { return switchPageMsg{pageSetup} }
				}
			}
			return m, func() tea.Msg { return funcSelectedMsg{item.funcType} }
		case "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m menuModel) View() string {
	var b strings.Builder

	globalIdx := 0
	for _, g := range m.groups {
		b.WriteString(groupTitleStyle.Render(g.title))
		b.WriteString("\n")

		for _, item := range g.items {
			cursor := "  "
			style := menuItemStyle
			if globalIdx == m.cursor {
				cursor = "▸ "
				style = menuSelectedStyle
			}

			name := style.Render(fmt.Sprintf("%s%-10s", cursor, item.name))
			desc := menuDescStyle.Render(item.desc)
			b.WriteString(fmt.Sprintf("%s %s\n", name, desc))
			globalIdx++
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  ↑/↓ 移动  |  Enter 选择  |  q 退出"))

	return b.String()
}
