package main

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type resultModel struct {
	resultURL string
	err       error
	funcType  funcType
	width     int
}

func newResultModel(resultURL string, err error, ft funcType, width int) resultModel {
	return resultModel{
		resultURL: resultURL,
		err:       err,
		funcType:  ft,
		width:     width,
	}
}

func (m resultModel) Init() tea.Cmd { return nil }

func (m resultModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", "q", "esc":
			return m, func() tea.Msg { return switchPageMsg{pageMenu} }
		}
	}
	return m, nil
}

func (m resultModel) View() string {
	var b strings.Builder

	// box 内容宽度 = 终端宽度 - 边框(2) - padding(2*2)
	contentWidth := m.width - 6
	if contentWidth < 40 {
		contentWidth = 40
	}
	box := boxStyle.Width(contentWidth)

	if m.err != nil {
		b.WriteString(errorStyle.Render("  " + funcName(m.funcType) + " - 失败"))
		b.WriteString("\n\n")

		errContent := errorStyle.Render("✗ " + m.err.Error())
		b.WriteString(box.Render(errContent))
	} else {
		b.WriteString(successStyle.Render("  " + funcName(m.funcType) + " - 完成"))
		b.WriteString("\n\n")

		switch m.funcType {
		case funcEnhancePrompt:
			var content strings.Builder
			content.WriteString(labelStyle.Render("优化后的提示词:"))
			content.WriteString("\n\n")
			content.WriteString(m.resultURL)
			b.WriteString(box.Render(content.String()))
		case funcCreditBalance:
			b.WriteString(box.Render(m.resultURL))
		default:
			var content strings.Builder
			if strings.HasPrefix(m.resultURL, "http") {
				// 远程 URL（下载失败回退），放在 box 外避免截断
				b.WriteString(box.Render(labelStyle.Render("下载链接:")))
				b.WriteString("\n\n")
				b.WriteString("  " + m.resultURL)
			} else {
				// 本地文件路径
				content.WriteString(successStyle.Render("文件已保存:"))
				content.WriteString("\n\n")
				content.WriteString(m.resultURL)
				b.WriteString(box.Render(content.String()))
			}
		}
	}

	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("  Enter 返回菜单  |  Ctrl+C 退出"))

	return b.String()
}
