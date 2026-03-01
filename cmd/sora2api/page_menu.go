package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/DouDOU-start/go-sora2api/sora"
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

	// 账号信息
	accountLoading bool
	accountSpinner spinner.Model
	balance        *sora.CreditBalance
	sub            *sora.SubscriptionInfo
	accountErr     error
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
			title: "角色管理",
			items: []menuItem{
				{name: "创建角色", desc: "上传视频自动创建角色（全流程）", funcType: funcCreateCharacter},
				{name: "删除角色", desc: "输入 character_id 删除角色", funcType: funcDeleteCharacter},
			},
		},
		{
			title: "视频发布",
			items: []menuItem{
				{name: "分镜任务", desc: "使用分镜格式生成视频", funcType: funcStoryboard},
				{name: "发布去水印", desc: "发布视频获取去水印链接", funcType: funcPublishVideo},
				{name: "删除帖子", desc: "删除已发布的帖子", funcType: funcDeletePost},
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

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle()

	return menuModel{
		groups:         groups,
		allItems:       all,
		accountLoading: true,
		accountSpinner: s,
	}
}

func (m menuModel) Init() tea.Cmd { return m.accountSpinner.Tick }

func (m menuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		if m.accountLoading {
			var cmd tea.Cmd
			m.accountSpinner, cmd = m.accountSpinner.Update(msg)
			return m, cmd
		}

	case accountInfoMsg:
		m.accountLoading = false
		m.accountErr = msg.err
		m.balance = msg.balance
		m.sub = msg.sub
		return m, nil

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

	// 菜单列表
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
			fmt.Fprintf(&b, "%s %s\n", name, desc)
			globalIdx++
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  ↑/↓ 移动  |  Enter 选择  |  q 退出"))

	return b.String()
}

// renderAccountCard 渲染账号信息卡片
func (m menuModel) renderAccountCard() string {
	var b strings.Builder

	if m.accountLoading {
		fmt.Fprintf(&b, "  %s 正在获取账号信息...", m.accountSpinner.View())
		return b.String()
	}

	if m.accountErr != nil {
		b.WriteString(warnStyle.Render("  ⚠ 获取账号信息失败"))
		return b.String()
	}

	// 订阅信息
	if m.sub != nil && m.sub.PlanTitle != "" {
		b.WriteString(labelStyle.Render("  套餐"))
		b.WriteString("  ")
		b.WriteString(successStyle.Render(m.sub.PlanTitle))
		if m.sub.EndTs > 0 {
			expireTime := time.Unix(m.sub.EndTs, 0)
			remaining := time.Until(expireTime)
			if remaining > 0 {
				days := int(remaining.Hours() / 24)
				b.WriteString(menuDescStyle.Render(fmt.Sprintf("  到期 %s（%d天后）", expireTime.Format("2006-01-02"), days)))
			} else {
				b.WriteString(errorStyle.Render("  已过期"))
			}
		}
		b.WriteString("\n")
	}

	// 配额信息
	if m.balance != nil {
		b.WriteString(labelStyle.Render("  配额"))
		b.WriteString("  ")
		countStr := fmt.Sprintf("剩余 %d 次", m.balance.RemainingCount)
		if m.balance.RemainingCount > 0 {
			b.WriteString(successStyle.Render(countStr))
		} else {
			b.WriteString(errorStyle.Render(countStr))
		}
		if m.balance.RateLimitReached {
			b.WriteString(warnStyle.Render("  (已限速)"))
		}
		if m.balance.AccessResetsInSec > 0 {
			resetMin := m.balance.AccessResetsInSec / 60
			resetHour := resetMin / 60
			resetMin = resetMin % 60
			b.WriteString(menuDescStyle.Render(fmt.Sprintf("  %d时%d分后重置", resetHour, resetMin)))
		}
	}

	return b.String()
}
