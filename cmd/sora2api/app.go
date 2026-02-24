package main

import (
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
		// 传播到子模型
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
			return m, nil
		}
		return m, nil

	case funcSelectedMsg:
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
			return m, nil
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
			return m, nil
		}

	case pageTask:
		var updated tea.Model
		updated, cmd = m.task.Update(msg)
		m.task = updated.(taskModel)

		if m.task.done {
			m.currentPage = pageResult
			m.result = newResultModel(m.task.resultURL, m.task.taskErr, m.task.funcType)
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
		content = m.setup.View()
	case pageMenu:
		content = m.menuViewWithStatus()
	case pageParam:
		content = m.param.View()
	case pageTask:
		content = m.task.View()
	case pageResult:
		content = m.result.View()
	}

	return content
}

func (m appModel) menuViewWithStatus() string {
	var b strings.Builder

	header := titleStyle.Render("Sora 视频/图片生成工具")
	b.WriteString(header)
	b.WriteString("\n")

	// 状态栏
	tokenDisplay := m.accessToken
	if len(tokenDisplay) > 30 {
		tokenDisplay = tokenDisplay[:15] + "..." + tokenDisplay[len(tokenDisplay)-8:]
	}
	proxyDisplay := "无"
	if m.proxyURL != "" {
		proxyDisplay = m.proxyURL
	}
	status := lipgloss.NewStyle().Foreground(colorMuted).Render(
		fmt.Sprintf("Token: %s  |  代理: %s", tokenDisplay, proxyDisplay))
	b.WriteString(status)
	b.WriteString("\n")

	b.WriteString(m.menu.View())

	return b.String()
}
