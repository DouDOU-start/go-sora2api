package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/DouDOU-start/go-sora2api/sora"
)

// paramField 表单字段
type paramField struct {
	label    string
	key      string
	kind     string // "text", "select"
	options  []string
	optionLabels []string // 显示文字
	selected int
	input    textinput.Model
}

type paramModel struct {
	funcType  funcType
	fields    []paramField
	focus     int
	submitted bool
	cancelled bool
	values    map[string]string
}

func newParamModel(ft funcType) paramModel {
	fields := buildFields(ft)
	if len(fields) > 0 && fields[0].kind == "text" {
		fields[0].input.Focus()
	}
	return paramModel{
		funcType: ft,
		fields:   fields,
		values:   make(map[string]string),
	}
}

func buildFields(ft funcType) []paramField {
	switch ft {
	case funcTextToImage:
		return []paramField{
			textField("prompt", "提示词", "一只可爱的小猫在草地上奔跑"),
			selectField("size", "图片尺寸", []string{"360x360", "540x360", "360x540"}, []string{"正方形 (360x360)", "横向 (540x360)", "纵向 (360x540)"}),
		}
	case funcImageToImage:
		return []paramField{
			textField("image_path", "图片路径", "/path/to/image.png"),
			textField("prompt", "提示词", "make it more colorful"),
			selectField("size", "图片尺寸", []string{"360x360", "540x360", "360x540"}, []string{"正方形 (360x360)", "横向 (540x360)", "纵向 (360x540)"}),
		}
	case funcTextToVideo:
		return videoFields(false)
	case funcImageToVideo:
		fields := []paramField{textField("image_path", "图片路径", "/path/to/image.png")}
		return append(fields, videoFields(false)...)
	case funcRemixVideo:
		return append(
			[]paramField{textField("remix_id", "视频 ID 或分享链接", "https://sora.chatgpt.com/p/s_xxx")},
			videoFields(true)...,
		)
	case funcEnhancePrompt:
		return []paramField{
			textField("prompt", "提示词", "a cute cat"),
			selectField("expansion", "扩展程度", []string{"medium", "long"}, []string{"中等 (medium)", "详细 (long)"}),
			selectField("duration", "目标时长", []string{"5", "10", "15", "25"}, []string{"5 秒", "10 秒", "15 秒", "25 秒"}),
		}
	case funcWatermarkFree:
		rt := textField("refresh_token", "Refresh Token", "")
		rt.input.EchoMode = textinput.EchoPassword
		rt.input.EchoCharacter = '*'
		return []paramField{
			rt,
			textField("client_id", "Client ID (留空使用默认)", ""),
			textField("video_id", "视频 ID 或分享链接", "https://sora.chatgpt.com/p/s_xxx"),
		}
	case funcCreateCharacter:
		return []paramField{
			textField("video_path", "角色视频路径", "/path/to/video.mp4"),
			textField("display_name", "角色名称", "我的角色"),
			textField("username", "角色用户名", "my_character"),
		}
	case funcDeleteCharacter:
		return []paramField{
			textField("character_id", "Character ID", ""),
		}
	case funcStoryboard:
		return []paramField{
			textField("prompt", "分镜提示词", "[5.0s]场景1 [5.0s]场景2"),
			selectField("orientation", "视频方向", []string{"landscape", "portrait"}, []string{"横向 (landscape)", "纵向 (portrait)"}),
			selectField("n_frames", "视频时长", []string{"150", "300", "450", "750"}, []string{"5 秒", "10 秒", "15 秒", "25 秒"}),
		}
	case funcPublishVideo:
		return []paramField{
			textField("generation_id", "Generation ID", "gen_xxx"),
		}
	case funcDeletePost:
		return []paramField{
			textField("post_id", "帖子 ID", "s_xxx"),
		}
	}
	return nil
}

func videoFields(isRemix bool) []paramField {
	fields := []paramField{
		textField("prompt", "提示词", "一只可爱的小猫在草地上奔跑"),
		selectField("orientation", "视频方向", []string{"landscape", "portrait"}, []string{"横向 (landscape)", "纵向 (portrait)"}),
		selectField("n_frames", "视频时长", []string{"150", "300", "450", "750"}, []string{"5 秒", "10 秒", "15 秒", "25 秒"}),
	}
	if !isRemix {
		fields = append(fields,
			selectField("model", "模型", []string{"sy_8", "sy_ore"}, []string{"标准 (sy_8)", "Pro (sy_ore)"}),
			selectField("size", "清晰度", []string{"small", "large"}, []string{"标准 (small)", "高清 (large, 仅Pro)"}),
		)
	}
	// 风格选项
	styleOpts := append([]string{""}, sora.ValidStyles...)
	styleLabels := append([]string{"无"}, sora.ValidStyles...)
	fields = append(fields, selectField("style", "风格", styleOpts, styleLabels))
	return fields
}

func textField(key, label, placeholder string) paramField {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 4096
	ti.Width = 60
	return paramField{
		label: label,
		key:   key,
		kind:  "text",
		input: ti,
	}
}

func selectField(key, label string, options, labels []string) paramField {
	return paramField{
		label:        label,
		key:          key,
		kind:         "select",
		options:      options,
		optionLabels: labels,
	}
}

func (m paramModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m paramModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.cancelled = true
			return m, nil
		case "tab", "down":
			return m.nextField()
		case "shift+tab", "up":
			return m.prevField()
		case "left":
			if m.fields[m.focus].kind == "select" {
				f := &m.fields[m.focus]
				if f.selected > 0 {
					f.selected--
				}
				return m, nil
			}
		case "right":
			if m.fields[m.focus].kind == "select" {
				f := &m.fields[m.focus]
				if f.selected < len(f.options)-1 {
					f.selected++
				}
				return m, nil
			}
		case "enter":
			// 如果在最后一个字段，提交
			if m.focus == len(m.fields)-1 {
				m.collectValues()
				m.submitted = true
				return m, nil
			}
			return m.nextField()
		case "ctrl+s":
			m.collectValues()
			m.submitted = true
			return m, nil
		}
	}

	// 更新当前文本输入
	if m.fields[m.focus].kind == "text" {
		var cmd tea.Cmd
		m.fields[m.focus].input, cmd = m.fields[m.focus].input.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m paramModel) nextField() (tea.Model, tea.Cmd) {
	if m.focus < len(m.fields)-1 {
		if m.fields[m.focus].kind == "text" {
			m.fields[m.focus].input.Blur()
		}
		m.focus++
		if m.fields[m.focus].kind == "text" {
			m.fields[m.focus].input.Focus()
			return m, textinput.Blink
		}
	}
	return m, nil
}

func (m paramModel) prevField() (tea.Model, tea.Cmd) {
	if m.focus > 0 {
		if m.fields[m.focus].kind == "text" {
			m.fields[m.focus].input.Blur()
		}
		m.focus--
		if m.fields[m.focus].kind == "text" {
			m.fields[m.focus].input.Focus()
			return m, textinput.Blink
		}
	}
	return m, nil
}

func (m *paramModel) collectValues() {
	for _, f := range m.fields {
		switch f.kind {
		case "text":
			m.values[f.key] = f.input.Value()
		case "select":
			m.values[f.key] = f.options[f.selected]
		}
	}
}

func (m paramModel) View() string {
	var b strings.Builder

	title := funcName(m.funcType)
	b.WriteString(titleStyle.Render("  " + title + " - 参数设置"))
	b.WriteString("\n\n")

	// 表单内容放入卡片
	var form strings.Builder
	for i, f := range m.fields {
		isFocused := i == m.focus
		indicator := "  "
		if isFocused {
			indicator = "▸ "
		}

		lbl := f.label
		if isFocused {
			lbl = labelStyle.Foreground(colorPrimary).Render(indicator + lbl)
		} else {
			lbl = labelStyle.Render(indicator + lbl)
		}
		form.WriteString(lbl)
		form.WriteString("\n")

		switch f.kind {
		case "text":
			form.WriteString(fmt.Sprintf("    %s\n", f.input.View()))
		case "select":
			form.WriteString("    ")
			for j, label := range f.optionLabels {
				if j == f.selected {
					if isFocused {
						form.WriteString(menuSelectedStyle.Render("● " + label))
					} else {
						form.WriteString(successStyle.Render("● " + label))
					}
				} else {
					form.WriteString(menuDescStyle.Render("○ " + label))
				}
				if j < len(f.optionLabels)-1 {
					form.WriteString("  ")
				}
			}
			form.WriteString("\n")
		}
	}

	b.WriteString(boxStyle.Render(form.String()))

	b.WriteString("\n\n")
	help := "  ↑/↓ 切换字段  |  ←/→ 选择选项  |  Enter 下一项/提交  |  Ctrl+S 直接提交  |  Esc 返回"
	b.WriteString(helpStyle.Render(help))

	return b.String()
}

func funcName(ft funcType) string {
	switch ft {
	case funcTextToImage:
		return "文生图"
	case funcImageToImage:
		return "图生图"
	case funcTextToVideo:
		return "文生视频"
	case funcImageToVideo:
		return "图生视频"
	case funcRemixVideo:
		return "Remix 视频"
	case funcEnhancePrompt:
		return "提示词优化"
	case funcWatermarkFree:
		return "去水印链接"
	case funcCreditBalance:
		return "查询可用次数"
	case funcCreateCharacter:
		return "创建角色"
	case funcDeleteCharacter:
		return "删除角色"
	case funcStoryboard:
		return "分镜任务"
	case funcPublishVideo:
		return "发布去水印"
	case funcDeletePost:
		return "删除帖子"
	}
	return ""
}
