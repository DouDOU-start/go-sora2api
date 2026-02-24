package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/DouDOU-start/go-sora2api/sora"
)

type setupModel struct {
	tokenInput textinput.Model
	proxyInput textinput.Model
	focusIndex int // 0: token, 1: proxy
	done       bool
	err        string

	accessToken string
	proxyURL    string
}

func newSetupModel() setupModel {
	ti := textinput.New()
	ti.Placeholder = "粘贴你的 access_token..."
	ti.Focus()
	ti.CharLimit = 4096
	ti.Width = 60
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '*'

	pi := textinput.New()
	pi.Placeholder = "留空不使用代理 (支持 ip:port:user:pass 格式)"
	pi.CharLimit = 256
	pi.Width = 60

	return setupModel{
		tokenInput: ti,
		proxyInput: pi,
	}
}

func (m setupModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m setupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab":
			if m.focusIndex == 0 {
				m.focusIndex = 1
				m.tokenInput.Blur()
				m.proxyInput.Focus()
			} else {
				m.focusIndex = 0
				m.proxyInput.Blur()
				m.tokenInput.Focus()
			}
			return m, nil

		case "enter":
			if m.focusIndex == 0 && m.tokenInput.Value() != "" {
				// Token 已输入，跳到代理输入
				m.focusIndex = 1
				m.tokenInput.Blur()
				m.proxyInput.Focus()
				return m, nil
			}
			if m.tokenInput.Value() == "" {
				m.err = "access_token 不能为空"
				return m, nil
			}
			// 提交
			m.accessToken = m.tokenInput.Value()
			m.proxyURL = sora.ParseProxy(m.proxyInput.Value())
			m.done = true
			return m, nil
		}
	}

	m.err = ""
	var cmd tea.Cmd
	if m.focusIndex == 0 {
		m.tokenInput, cmd = m.tokenInput.Update(msg)
	} else {
		m.proxyInput, cmd = m.proxyInput.Update(msg)
	}
	return m, cmd
}

func (m setupModel) View() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(titleStyle.Render("  Sora 视频/图片生成工具"))
	b.WriteString("\n\n")

	// 表单内容
	var form strings.Builder

	// Token 输入
	if m.focusIndex == 0 {
		form.WriteString(labelStyle.Foreground(colorPrimary).Render("▸ Access Token"))
	} else {
		form.WriteString(labelStyle.Render("  Access Token"))
	}
	form.WriteString("\n")
	form.WriteString(fmt.Sprintf("  %s\n", m.tokenInput.View()))

	// 代理输入
	if m.focusIndex == 1 {
		form.WriteString(labelStyle.Foreground(colorPrimary).Render("\n▸ 代理地址 (可选)"))
	} else {
		form.WriteString(labelStyle.Render("\n  代理地址 (可选)"))
	}
	form.WriteString("\n")
	form.WriteString(fmt.Sprintf("  %s", m.proxyInput.View()))

	b.WriteString(boxStyle.Render(form.String()))
	b.WriteString("\n")

	// 错误提示
	if m.err != "" {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render("  ✗ " + m.err))
	}

	// 帮助
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("  Tab 切换  |  Enter 确认  |  Ctrl+C 退出"))

	return b.String()
}
