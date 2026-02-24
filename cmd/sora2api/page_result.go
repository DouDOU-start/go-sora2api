package main

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type resultModel struct {
	resultURL string
	err       error
	funcType  funcType
}

func newResultModel(resultURL string, err error, ft funcType) resultModel {
	return resultModel{
		resultURL: resultURL,
		err:       err,
		funcType:  ft,
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

	if m.err != nil {
		b.WriteString(errorStyle.Render("  " + funcName(m.funcType) + " - 失败"))
		b.WriteString("\n\n")

		errContent := errorStyle.Render("✗ " + m.err.Error())
		b.WriteString(boxStyle.Render(errContent))
	} else {
		b.WriteString(successStyle.Render("  " + funcName(m.funcType) + " - 完成"))
		b.WriteString("\n\n")

		var content strings.Builder
		switch m.funcType {
		case funcEnhancePrompt:
			content.WriteString(labelStyle.Render("优化后的提示词:"))
			content.WriteString("\n\n")
			content.WriteString(m.resultURL)
		case funcCreditBalance:
			content.WriteString(m.resultURL)
		default:
			content.WriteString(labelStyle.Render("下载链接:"))
			content.WriteString("\n\n")
			content.WriteString(urlStyle.Render(m.resultURL))
		}
		b.WriteString(boxStyle.Render(content.String()))
	}

	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("  Enter 返回菜单  |  Ctrl+C 退出"))

	return b.String()
}
