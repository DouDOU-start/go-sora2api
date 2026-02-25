package main

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/DouDOU-start/go-sora2api/sora"
)

type pageType int

const (
	pageSetup pageType = iota
	pageMenu
	pageParam
	pageTask
	pageResult
)

type funcType int

const (
	funcTextToImage funcType = iota
	funcImageToImage
	funcTextToVideo
	funcImageToVideo
	funcRemixVideo
	funcEnhancePrompt
	funcWatermarkFree
	funcCreditBalance
	funcCreateCharacter
	funcDeleteCharacter
	funcStoryboard
	funcPublishVideo
	funcDeletePost
)

// appModel 顶层模型
type appModel struct {
	currentPage pageType
	setup       setupModel
	menu        menuModel
	param       paramModel
	task        taskModel
	result      resultModel

	// 共享状态
	client      *sora.Client
	accessToken string
	proxyURL    string
	width       int
	height      int
}

func newAppModel() appModel {
	return appModel{
		currentPage: pageSetup,
		setup:       newSetupModel(),
		menu:        newMenuModel(),
	}
}

func (m appModel) Init() tea.Cmd {
	return m.setup.Init()
}

func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.menu.width = msg.Width
		m.task.width = msg.Width
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case switchPageMsg:
		m.currentPage = msg.page
		switch msg.page {
		case pageSetup:
			m.setup = newSetupModel()
			m.client = nil
			return m, m.setup.Init()
		case pageMenu:
			m.menu = newMenuModel()
			m.menu.width = m.width
			return m, tea.Batch(m.menu.Init(), m.fetchAccountInfo())
		}
		return m, nil

	case funcSelectedMsg:
		// 无需参数的功能直接进入任务页
		if msg.funcType == funcCreditBalance {
			m.currentPage = pageTask
			m.task = newTaskModel(m.client, m.accessToken, msg.funcType, map[string]string{})
			m.task.width = m.width
			return m, m.task.Init()
		}
		m.currentPage = pageParam
		m.param = newParamModel(msg.funcType)
		return m, m.param.Init()
	}

	// 委托给当前页面处理
	var cmd tea.Cmd
	switch m.currentPage {
	case pageSetup:
		var updated tea.Model
		updated, cmd = m.setup.Update(msg)
		m.setup = updated.(setupModel)

		// 检查是否配置完成
		if m.setup.done {
			var err error
			m.accessToken = m.setup.accessToken
			m.proxyURL = m.setup.proxyURL
			m.client, err = sora.New(m.proxyURL)
			if err != nil {
				m.setup.err = fmt.Sprintf("创建客户端失败: %v", err)
				m.setup.done = false
				return m, nil
			}
			m.currentPage = pageMenu
			m.menu = newMenuModel()
			m.menu.width = m.width
			return m, tea.Batch(m.menu.Init(), m.fetchAccountInfo())
		}

	case pageMenu:
		var updated tea.Model
		updated, cmd = m.menu.Update(msg)
		m.menu = updated.(menuModel)

	case pageParam:
		var updated tea.Model
		updated, cmd = m.param.Update(msg)
		m.param = updated.(paramModel)

		if m.param.submitted {
			m.currentPage = pageTask
			m.task = newTaskModel(m.client, m.accessToken, m.param.funcType, m.param.values)
			m.task.width = m.width
			return m, m.task.Init()
		}
		if m.param.cancelled {
			m.currentPage = pageMenu
			m.menu = newMenuModel()
			m.menu.width = m.width
			return m, tea.Batch(m.menu.Init(), m.fetchAccountInfo())
		}

	case pageTask:
		var updated tea.Model
		updated, cmd = m.task.Update(msg)
		m.task = updated.(taskModel)

		if m.task.done {
			m.currentPage = pageResult
			m.result = newResultModel(m.task.resultURL, m.task.taskErr, m.task.funcType, m.width)
			return m, nil
		}

	case pageResult:
		var updated tea.Model
		updated, cmd = m.result.Update(msg)
		m.result = updated.(resultModel)
	}

	return m, cmd
}

func (m appModel) View() string {
	var content string

	switch m.currentPage {
	case pageSetup:
		content = m.renderSetupPage()
	case pageMenu:
		content = m.renderMenuPage()
	case pageParam:
		content = m.renderParamPage()
	case pageTask:
		content = m.renderTaskPage()
	case pageResult:
		content = m.renderResultPage()
	}

	return content
}

// fetchAccountInfo 异步获取账号配额和订阅信息
func (m appModel) fetchAccountInfo() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		var balance *sora.CreditBalance
		var sub *sora.SubscriptionInfo

		// 获取配额
		b, err := m.client.GetCreditBalance(ctx, m.accessToken)
		if err == nil {
			balance = &b
		}

		// 获取订阅信息
		s, err2 := m.client.GetSubscriptionInfo(ctx, m.accessToken)
		if err2 == nil {
			sub = &s
		}

		// 只要有一个成功就不算错误
		var retErr error
		if err != nil && err2 != nil {
			retErr = err
		}

		return accountInfoMsg{balance: balance, sub: sub, err: retErr}
	}
}

// ─── 页面渲染 ───

func (m appModel) renderSetupPage() string {
	return m.wrapPage("", m.setup.View())
}

func (m appModel) renderMenuPage() string {
	var b strings.Builder

	// 标题
	b.WriteString(titleStyle.Render("  Sora 视频/图片生成工具"))
	b.WriteString("\n\n")

	// 账号信息卡片
	accountContent := m.menu.renderAccountCard()
	if accountContent != "" {
		card := accountCardStyle.Render(accountContent)
		b.WriteString(card)
		b.WriteString("\n\n")
	}

	// 连接信息
	tokenDisplay := m.accessToken
	if len(tokenDisplay) > 30 {
		tokenDisplay = tokenDisplay[:15] + "..." + tokenDisplay[len(tokenDisplay)-8:]
	}
	proxyDisplay := "无"
	if m.proxyURL != "" {
		proxyDisplay = m.proxyURL
	}
	connInfo := menuDescStyle.Render(fmt.Sprintf("  Token: %s  |  代理: %s", tokenDisplay, proxyDisplay))
	b.WriteString(connInfo)
	b.WriteString("\n")

	// 菜单
	b.WriteString(m.menu.View())

	return b.String()
}

func (m appModel) renderParamPage() string {
	return m.wrapPage("", m.param.View())
}

func (m appModel) renderTaskPage() string {
	return m.wrapPage("", m.task.View())
}

func (m appModel) renderResultPage() string {
	return m.wrapPage("", m.result.View())
}

func (m appModel) wrapPage(_ string, content string) string {
	// 计算可用宽度
	maxWidth := m.width
	if maxWidth <= 0 {
		maxWidth = 80
	}
	if maxWidth > 100 {
		maxWidth = 100
	}

	return lipgloss.NewStyle().
		MaxWidth(maxWidth).
		Render(content)
}
