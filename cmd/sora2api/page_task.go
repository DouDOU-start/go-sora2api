package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/DouDOU-start/go-sora2api/sora"
)

type taskStep struct {
	name string
	done bool
	err  error
}

type taskModel struct {
	client      *sora.Client
	accessToken string
	funcType    funcType
	params      map[string]string

	steps       []taskStep
	currentStep int
	spinner     spinner.Model
	progress    progress.Model
	progressPct float64
	statusText  string

	taskID       string
	startTime    time.Time
	maxProgress  int
	everFound    bool
	notFoundCnt  int

	done      bool
	resultURL string
	taskErr   error
	width     int
}

func newTaskModel(client *sora.Client, accessToken string, ft funcType, params map[string]string) taskModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle()

	p := progress.New(progress.WithDefaultGradient())
	p.Width = 40

	return taskModel{
		client:      client,
		accessToken: accessToken,
		funcType:    ft,
		params:      params,
		spinner:     s,
		progress:    p,
		startTime:   time.Now(),
	}
}

func spinnerStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(colorPrimary)
}

func (m taskModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.startExecution())
}

func (m taskModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case progress.FrameMsg:
		model, cmd := m.progress.Update(msg)
		m.progress = model.(progress.Model)
		return m, cmd

	case taskStepMsg:
		if msg.err != nil {
			m.steps = append(m.steps, taskStep{name: msg.step, done: true, err: msg.err})
			m.done = true
			m.taskErr = msg.err
			return m, nil
		}
		if msg.done {
			// 标记步骤完成
			if len(m.steps) > 0 {
				m.steps[len(m.steps)-1].done = true
			}
		} else {
			m.steps = append(m.steps, taskStep{name: msg.step})
		}
		return m, nil

	case taskProgressMsg:
		m.progressPct = float64(msg.progress.Percent) / 100.0
		m.statusText = fmt.Sprintf("%s | 已耗时: %ds", msg.progress.Status, msg.progress.Elapsed)
		m.maxProgress = msg.progress.Percent
		return m, m.progress.SetPercent(m.progressPct)

	case taskCompleteMsg:
		m.done = true
		m.resultURL = msg.resultURL
		m.taskErr = msg.err
		if len(m.steps) > 0 {
			last := &m.steps[len(m.steps)-1]
			last.done = true
			if msg.err != nil {
				last.err = msg.err
			}
		}
		return m, nil

	case tickPollMsg:
		return m, m.pollOnce()
	}

	return m, nil
}

func (m taskModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  " + funcName(m.funcType) + " - 执行中"))
	b.WriteString("\n\n")

	// 卡片内容
	var card strings.Builder

	// 参数摘要
	if prompt, ok := m.params["prompt"]; ok && prompt != "" {
		display := prompt
		if len(display) > 55 {
			display = display[:52] + "..."
		}
		card.WriteString(menuDescStyle.Render("提示词: " + display))
		card.WriteString("\n\n")
	}

	// 步骤列表
	if len(m.steps) == 0 {
		card.WriteString(fmt.Sprintf("%s 准备中...", m.spinner.View()))
	}
	for _, step := range m.steps {
		if step.err != nil {
			card.WriteString(errorStyle.Render("✗ " + step.name + ": " + step.err.Error()))
		} else if step.done {
			card.WriteString(successStyle.Render("✓ " + step.name))
		} else {
			card.WriteString(fmt.Sprintf("%s %s", m.spinner.View(), step.name))
		}
		card.WriteString("\n")
	}

	// 进度条
	if m.progressPct > 0 {
		card.WriteString("\n")
		card.WriteString(m.progress.View())
		if m.statusText != "" {
			card.WriteString("\n")
			card.WriteString(menuDescStyle.Render(m.statusText))
		}
	}

	b.WriteString(boxStyle.Render(card.String()))

	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("  Ctrl+C 退出"))

	return b.String()
}

// startExecution 根据功能类型启动对应的任务链
func (m taskModel) startExecution() tea.Cmd {
	switch m.funcType {
	case funcTextToImage, funcImageToImage:
		return m.executeImageTask()
	case funcTextToVideo, funcImageToVideo:
		return m.executeVideoTask()
	case funcRemixVideo:
		return m.executeRemixTask()
	case funcEnhancePrompt:
		return m.executeEnhancePrompt()
	case funcWatermarkFree:
		return m.executeWatermarkFree()
	case funcCreditBalance:
		return m.executeCreditBalance()
	}
	return nil
}

func (m taskModel) executeImageTask() tea.Cmd {
	return func() tea.Msg {
		c := m.client
		at := m.accessToken
		prompt := m.params["prompt"]
		if prompt == "" {
			prompt = "一只可爱的小猫在草地上奔跑"
		}

		// 解析尺寸
		width, height := 360, 360
		if size := m.params["size"]; size != "" {
			parts := strings.Split(size, "x")
			if len(parts) == 2 {
				width, _ = strconv.Atoi(parts[0])
				height, _ = strconv.Atoi(parts[1])
			}
		}

		// 上传图片（图生图）
		var mediaID string
		if m.funcType == funcImageToImage {
			p := tea.NewProgram(nil) // 占位，实际不使用
			_ = p
			imagePath := m.params["image_path"]
			if imagePath == "" {
				return taskCompleteMsg{err: fmt.Errorf("图片路径不能为空")}
			}
			imageData, err := os.ReadFile(imagePath)
			if err != nil {
				return taskCompleteMsg{err: fmt.Errorf("读取图片失败: %w", err)}
			}
			parts := strings.Split(strings.ReplaceAll(imagePath, "\\", "/"), "/")
			filename := parts[len(parts)-1]

			id, err := c.UploadImage(at, imageData, filename)
			if err != nil {
				return taskCompleteMsg{err: fmt.Errorf("上传图片失败: %w", err)}
			}
			mediaID = id
		}

		// 获取 sentinel token
		sentinelToken, err := c.GenerateSentinelToken(at)
		if err != nil {
			return taskCompleteMsg{err: fmt.Errorf("获取 sentinel token 失败: %w", err)}
		}

		// 创建任务
		taskID, err := c.CreateImageTaskWithImage(at, sentinelToken, prompt, width, height, mediaID)
		if err != nil {
			return taskCompleteMsg{err: fmt.Errorf("创建任务失败: %w", err)}
		}

		// 轮询
		imageURL, err := c.PollImageTask(at, taskID, 3*time.Second, 600*time.Second, nil)
		if err != nil {
			return taskCompleteMsg{err: err}
		}

		return taskCompleteMsg{resultURL: imageURL}
	}
}

func (m taskModel) executeVideoTask() tea.Cmd {
	return func() tea.Msg {
		c := m.client
		at := m.accessToken
		prompt := m.params["prompt"]
		if prompt == "" {
			prompt = "一只可爱的小猫在草地上奔跑"
		}
		orientation := m.params["orientation"]
		if orientation == "" {
			orientation = "landscape"
		}
		nFrames, _ := strconv.Atoi(m.params["n_frames"])
		if nFrames == 0 {
			nFrames = 150
		}
		model := m.params["model"]
		if model == "" {
			model = "sy_8"
		}
		size := m.params["size"]
		if size == "" {
			size = "small"
		}
		styleID := m.params["style"]

		// 提取风格
		prompt, extractedStyle := sora.ExtractStyle(prompt)
		if extractedStyle != "" {
			styleID = extractedStyle
		}

		// 上传图片（图生视频）
		var mediaID string
		if m.funcType == funcImageToVideo {
			imagePath := m.params["image_path"]
			if imagePath == "" {
				return taskCompleteMsg{err: fmt.Errorf("图片路径不能为空")}
			}
			imageData, err := os.ReadFile(imagePath)
			if err != nil {
				return taskCompleteMsg{err: fmt.Errorf("读取图片失败: %w", err)}
			}
			parts := strings.Split(strings.ReplaceAll(imagePath, "\\", "/"), "/")
			filename := parts[len(parts)-1]

			id, err := c.UploadImage(at, imageData, filename)
			if err != nil {
				return taskCompleteMsg{err: fmt.Errorf("上传图片失败: %w", err)}
			}
			mediaID = id
		}

		// 获取 sentinel token
		sentinelToken, err := c.GenerateSentinelToken(at)
		if err != nil {
			return taskCompleteMsg{err: fmt.Errorf("获取 sentinel token 失败: %w", err)}
		}

		// 创建任务
		taskID, err := c.CreateVideoTaskWithOptions(at, sentinelToken, prompt, orientation, nFrames, model, size, mediaID, styleID)
		if err != nil {
			return taskCompleteMsg{err: fmt.Errorf("创建任务失败: %w", err)}
		}

		// 轮询
		err = c.PollVideoTask(at, taskID, 3*time.Second, 600*time.Second, nil)
		if err != nil {
			return taskCompleteMsg{err: err}
		}

		// 获取下载链接
		url, err := c.GetDownloadURL(at, taskID)
		if err != nil {
			return taskCompleteMsg{err: err}
		}

		return taskCompleteMsg{resultURL: url}
	}
}

func (m taskModel) executeRemixTask() tea.Cmd {
	return func() tea.Msg {
		c := m.client
		at := m.accessToken
		remixInput := m.params["remix_id"]
		remixID := sora.ExtractRemixID(remixInput)
		if remixID == "" {
			return taskCompleteMsg{err: fmt.Errorf("无法解析 Remix 视频 ID")}
		}

		prompt := m.params["prompt"]
		if prompt == "" {
			prompt = "make it different"
		}
		orientation := m.params["orientation"]
		if orientation == "" {
			orientation = "landscape"
		}
		nFrames, _ := strconv.Atoi(m.params["n_frames"])
		if nFrames == 0 {
			nFrames = 150
		}
		styleID := m.params["style"]

		prompt, extractedStyle := sora.ExtractStyle(prompt)
		if extractedStyle != "" {
			styleID = extractedStyle
		}

		sentinelToken, err := c.GenerateSentinelToken(at)
		if err != nil {
			return taskCompleteMsg{err: fmt.Errorf("获取 sentinel token 失败: %w", err)}
		}

		taskID, err := c.RemixVideo(at, sentinelToken, remixID, prompt, orientation, nFrames, styleID)
		if err != nil {
			return taskCompleteMsg{err: fmt.Errorf("创建 Remix 任务失败: %w", err)}
		}

		err = c.PollVideoTask(at, taskID, 3*time.Second, 600*time.Second, nil)
		if err != nil {
			return taskCompleteMsg{err: err}
		}

		url, err := c.GetDownloadURL(at, taskID)
		if err != nil {
			return taskCompleteMsg{err: err}
		}

		return taskCompleteMsg{resultURL: url}
	}
}

func (m taskModel) executeEnhancePrompt() tea.Cmd {
	return func() tea.Msg {
		prompt := m.params["prompt"]
		if prompt == "" {
			return taskCompleteMsg{err: fmt.Errorf("提示词不能为空")}
		}

		expansion := m.params["expansion"]
		if expansion == "" {
			expansion = "medium"
		}
		duration, _ := strconv.Atoi(m.params["duration"])
		if duration == 0 {
			duration = 10
		}

		enhanced, err := m.client.EnhancePrompt(m.accessToken, prompt, expansion, duration)
		if err != nil {
			return taskCompleteMsg{err: err}
		}

		return taskCompleteMsg{resultURL: enhanced}
	}
}

func (m taskModel) executeWatermarkFree() tea.Cmd {
	return func() tea.Msg {
		refreshToken := m.params["refresh_token"]
		if refreshToken == "" {
			return taskCompleteMsg{err: fmt.Errorf("refresh_token 不能为空")}
		}
		clientID := m.params["client_id"]
		videoID := m.params["video_id"]
		if videoID == "" {
			return taskCompleteMsg{err: fmt.Errorf("视频 ID 不能为空")}
		}

		soraToken, _, err := m.client.RefreshAccessToken(refreshToken, clientID)
		if err != nil {
			return taskCompleteMsg{err: fmt.Errorf("刷新 token 失败: %w", err)}
		}

		url, err := m.client.GetWatermarkFreeURL(soraToken, videoID)
		if err != nil {
			return taskCompleteMsg{err: err}
		}

		return taskCompleteMsg{resultURL: url}
	}
}

func (m taskModel) executeCreditBalance() tea.Cmd {
	return func() tea.Msg {
		balance, err := m.client.GetCreditBalance(m.accessToken)
		if err != nil {
			return taskCompleteMsg{err: err}
		}

		resetMin := balance.AccessResetsInSec / 60
		resetHour := resetMin / 60
		resetMin = resetMin % 60

		result := fmt.Sprintf("剩余可用次数: %d\n", balance.RemainingCount)
		if balance.RateLimitReached {
			result += "  速率限制: 已触发\n"
		}
		if balance.AccessResetsInSec > 0 {
			result += fmt.Sprintf("  配额重置: %d小时%d分钟后\n", resetHour, resetMin)
		}

		// 查询订阅信息
		sub, err := m.client.GetSubscriptionInfo(m.accessToken)
		if err == nil {
			result += "\n"
			if sub.PlanTitle != "" {
				result += fmt.Sprintf("  账号类型: %s\n", sub.PlanTitle)
			}
			if sub.EndTs > 0 {
				expireTime := time.Unix(sub.EndTs, 0)
				remaining := time.Until(expireTime)
				if remaining > 0 {
					days := int(remaining.Hours() / 24)
					result += fmt.Sprintf("  到期时间: %s（剩余 %d 天）", expireTime.Format("2006-01-02"), days)
				} else {
					result += fmt.Sprintf("  到期时间: %s（已过期）", expireTime.Format("2006-01-02"))
				}
			}
		}

		return taskCompleteMsg{resultURL: result}
	}
}

func (m taskModel) pollOnce() tea.Cmd {
	// 此方法预留给未来的非阻塞轮询模式使用
	return nil
}

