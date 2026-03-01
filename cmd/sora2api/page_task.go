package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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
	spinner     spinner.Model
	progress    progress.Model
	progressPct float64
	statusText  string

	startTime   time.Time
	maxProgress int

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

	case charCreateStepMsg:
		if msg.err != nil || msg.step == -1 {
			// 错误：标记当前步骤失败并结束
			err := msg.err
			if err == nil {
				err = fmt.Errorf("未知错误")
			}
			if len(m.steps) > 0 {
				m.steps[len(m.steps)-1].done = true
				m.steps[len(m.steps)-1].err = err
			}
			m.done = true
			m.taskErr = err
			return m, nil
		}
		// 触发下一步
		return m, m.charCreateStep(msg)

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
		fmt.Fprintf(&card, "%s 准备中...", m.spinner.View())
	}
	for _, step := range m.steps {
		if step.err != nil {
			card.WriteString(errorStyle.Render("✗ " + step.name + ": " + step.err.Error()))
		} else if step.done {
			card.WriteString(successStyle.Render("✓ " + step.name))
		} else {
			fmt.Fprintf(&card, "%s %s", m.spinner.View(), step.name)
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
	case funcCreateCharacter:
		return m.executeCreateCharacter()
	case funcDeleteCharacter:
		return m.executeDeleteCharacter()
	case funcStoryboard:
		return m.executeStoryboard()
	case funcPublishVideo:
		return m.executePublishVideo()
	case funcDeletePost:
		return m.executeDeletePost()
	}
	return nil
}

func (m taskModel) executeImageTask() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
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

			id, err := c.UploadImage(ctx, at, imageData, filename)
			if err != nil {
				return taskCompleteMsg{err: fmt.Errorf("上传图片失败: %w", err)}
			}
			mediaID = id
		}

		// 获取 sentinel token
		sentinelToken, err := c.GenerateSentinelToken(ctx, at)
		if err != nil {
			return taskCompleteMsg{err: fmt.Errorf("获取 sentinel token 失败: %w", err)}
		}

		// 创建任务
		taskID, err := c.CreateImageTaskWithImage(ctx, at, sentinelToken, prompt, width, height, mediaID)
		if err != nil {
			return taskCompleteMsg{err: err}
		}

		// 轮询
		imageURL, err := c.PollImageTask(ctx, at, taskID, 3*time.Second, 600*time.Second, nil)
		if err != nil {
			return taskCompleteMsg{err: err}
		}

		// 下载到本地
		localPath, err := downloadToLocal(ctx, c, imageURL, taskID, ".png")
		if err != nil {
			return taskCompleteMsg{resultURL: imageURL}
		}
		return taskCompleteMsg{resultURL: localPath}
	}
}

func (m taskModel) executeVideoTask() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
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
			nFrames = 300
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

			id, err := c.UploadImage(ctx, at, imageData, filename)
			if err != nil {
				return taskCompleteMsg{err: fmt.Errorf("上传图片失败: %w", err)}
			}
			mediaID = id
		}

		// 获取 sentinel token
		sentinelToken, err := c.GenerateSentinelToken(ctx, at)
		if err != nil {
			return taskCompleteMsg{err: fmt.Errorf("获取 sentinel token 失败: %w", err)}
		}

		// 创建任务
		taskID, err := c.CreateVideoTaskWithOptions(ctx, at, sentinelToken, prompt, orientation, nFrames, model, size, mediaID, styleID)
		if err != nil {
			return taskCompleteMsg{err: err}
		}

		// 轮询
		err = c.PollVideoTask(ctx, at, taskID, 3*time.Second, 600*time.Second, nil)
		if err != nil {
			return taskCompleteMsg{err: err}
		}

		// 获取下载链接
		dlURL, err := c.GetDownloadURL(ctx, at, taskID)
		if err != nil {
			return taskCompleteMsg{err: err}
		}

		// 下载到本地
		localPath, err := downloadToLocal(ctx, c, dlURL, taskID, ".mp4")
		if err != nil {
			return taskCompleteMsg{resultURL: dlURL}
		}
		return taskCompleteMsg{resultURL: localPath}
	}
}

func (m taskModel) executeRemixTask() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
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
			nFrames = 300
		}
		styleID := m.params["style"]

		prompt, extractedStyle := sora.ExtractStyle(prompt)
		if extractedStyle != "" {
			styleID = extractedStyle
		}

		sentinelToken, err := c.GenerateSentinelToken(ctx, at)
		if err != nil {
			return taskCompleteMsg{err: fmt.Errorf("获取 sentinel token 失败: %w", err)}
		}

		taskID, err := c.RemixVideo(ctx, at, sentinelToken, remixID, prompt, orientation, nFrames, styleID)
		if err != nil {
			return taskCompleteMsg{err: err}
		}

		err = c.PollVideoTask(ctx, at, taskID, 3*time.Second, 600*time.Second, nil)
		if err != nil {
			return taskCompleteMsg{err: err}
		}

		dlURL, err := c.GetDownloadURL(ctx, at, taskID)
		if err != nil {
			return taskCompleteMsg{err: err}
		}

		localPath, err := downloadToLocal(ctx, c, dlURL, taskID, ".mp4")
		if err != nil {
			return taskCompleteMsg{resultURL: dlURL}
		}
		return taskCompleteMsg{resultURL: localPath}
	}
}

func (m taskModel) executeEnhancePrompt() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
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

		enhanced, err := m.client.EnhancePrompt(ctx, m.accessToken, prompt, expansion, duration)
		if err != nil {
			return taskCompleteMsg{err: err}
		}

		return taskCompleteMsg{resultURL: enhanced}
	}
}

func (m taskModel) executeWatermarkFree() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		refreshToken := m.params["refresh_token"]
		if refreshToken == "" {
			return taskCompleteMsg{err: fmt.Errorf("refresh_token 不能为空")}
		}
		clientID := m.params["client_id"]
		videoID := m.params["video_id"]
		if videoID == "" {
			return taskCompleteMsg{err: fmt.Errorf("视频 ID 不能为空")}
		}

		soraToken, _, err := m.client.RefreshAccessToken(ctx, refreshToken, clientID)
		if err != nil {
			return taskCompleteMsg{err: fmt.Errorf("刷新 token 失败: %w", err)}
		}

		url, err := m.client.GetWatermarkFreeURL(ctx, soraToken, videoID)
		if err != nil {
			return taskCompleteMsg{err: err}
		}

		return taskCompleteMsg{resultURL: url}
	}
}

func (m taskModel) executeCreditBalance() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		balance, err := m.client.GetCreditBalance(ctx, m.accessToken)
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
		sub, err := m.client.GetSubscriptionInfo(ctx, m.accessToken)
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

// executeCreateCharacter 启动角色创建的第一步
func (m taskModel) executeCreateCharacter() tea.Cmd {
	return m.charCreateStep(charCreateStepMsg{step: 0})
}

// charCreateStep 角色创建状态机，每步发送进度并触发下一步
func (m taskModel) charCreateStep(msg charCreateStepMsg) tea.Cmd {
	c := m.client
	at := m.accessToken

	switch msg.step {
	case 0: // 读取并上传视频
		return tea.Sequence(
			func() tea.Msg { return taskStepMsg{step: "读取并上传角色视频"} },
			func() tea.Msg {
				ctx := context.Background()
				videoPath := m.params["video_path"]
				if videoPath == "" {
					return charCreateStepMsg{step: -1, err: fmt.Errorf("视频路径不能为空")}
				}
				videoData, err := os.ReadFile(videoPath)
				if err != nil {
					return charCreateStepMsg{step: -1, err: fmt.Errorf("读取视频失败: %w", err)}
				}
				cameoID, err := c.UploadCharacterVideo(ctx, at, videoData)
				if err != nil {
					return charCreateStepMsg{step: -1, err: fmt.Errorf("上传角色视频失败: %w", err)}
				}
				return charCreateStepMsg{step: 1, cameoID: cameoID}
			},
		)

	case 1: // 轮询角色处理状态
		return tea.Sequence(
			func() tea.Msg { return taskStepMsg{step: "读取并上传角色视频", done: true} },
			func() tea.Msg { return taskStepMsg{step: "等待角色处理完成"} },
			func() tea.Msg {
				ctx := context.Background()
				status, err := c.PollCameoStatus(ctx, at, msg.cameoID, 3*time.Second, 300*time.Second, nil)
				if err != nil {
					return charCreateStepMsg{step: -1, err: err}
				}
				if status.ProfileAssetURL == "" {
					return charCreateStepMsg{step: -1, err: fmt.Errorf("角色处理完成但无头像 URL")}
				}
				return charCreateStepMsg{step: 2, cameoID: msg.cameoID, profileAssetURL: status.ProfileAssetURL}
			},
		)

	case 2: // 下载角色头像（直接使用上一步传入的 URL，避免重复请求被 403）
		return tea.Sequence(
			func() tea.Msg { return taskStepMsg{step: "等待角色处理完成", done: true} },
			func() tea.Msg { return taskStepMsg{step: "下载角色头像"} },
			func() tea.Msg {
				ctx := context.Background()
				imageData, err := c.DownloadCharacterImage(ctx, msg.profileAssetURL)
				if err != nil {
					return charCreateStepMsg{step: -1, err: fmt.Errorf("下载角色图片失败: %w", err)}
				}
				return charCreateStepMsg{step: 3, cameoID: msg.cameoID, imageData: imageData}
			},
		)

	case 3: // 上传角色头像
		return tea.Sequence(
			func() tea.Msg { return taskStepMsg{step: "下载角色头像", done: true} },
			func() tea.Msg { return taskStepMsg{step: "上传角色头像"} },
			func() tea.Msg {
				ctx := context.Background()
				assetPointer, err := c.UploadCharacterImage(ctx, at, msg.imageData)
				if err != nil {
					return charCreateStepMsg{step: -1, err: fmt.Errorf("上传角色头像失败: %w", err)}
				}
				return charCreateStepMsg{step: 4, cameoID: msg.cameoID, assetPointer: assetPointer}
			},
		)

	case 4: // 定稿角色
		return tea.Sequence(
			func() tea.Msg { return taskStepMsg{step: "上传角色头像", done: true} },
			func() tea.Msg { return taskStepMsg{step: "定稿角色"} },
			func() tea.Msg {
				ctx := context.Background()
				displayName := m.params["display_name"]
				if displayName == "" {
					displayName = "My Character"
				}
				username := m.params["username"]
				if username == "" {
					username = "my_character"
				}
				characterID, err := c.FinalizeCharacter(ctx, at, msg.cameoID, username, displayName, msg.assetPointer)
				if err != nil {
					return charCreateStepMsg{step: -1, err: fmt.Errorf("定稿角色失败: %w", err)}
				}
				return charCreateStepMsg{step: 5, cameoID: msg.cameoID, characterID: characterID}
			},
		)

	case 5: // 设置公开
		return tea.Sequence(
			func() tea.Msg { return taskStepMsg{step: "定稿角色", done: true} },
			func() tea.Msg { return taskStepMsg{step: "设置角色公开"} },
			func() tea.Msg {
				ctx := context.Background()
				_ = c.SetCharacterPublic(ctx, at, msg.cameoID)
				return charCreateStepMsg{step: 6, cameoID: msg.cameoID, characterID: msg.characterID}
			},
		)

	case 6: // 完成
		displayName := m.params["display_name"]
		if displayName == "" {
			displayName = "My Character"
		}
		return tea.Sequence(
			func() tea.Msg { return taskStepMsg{step: "设置角色公开", done: true} },
			func() tea.Msg {
				return taskCompleteMsg{resultURL: fmt.Sprintf("角色创建成功\n  Character ID: %s\n  Cameo ID: %s\n  名称: %s", msg.characterID, msg.cameoID, displayName)}
			},
		)
	}
	return nil
}

func (m taskModel) executeDeleteCharacter() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		characterID := m.params["character_id"]
		if characterID == "" {
			return taskCompleteMsg{err: fmt.Errorf("character_id 不能为空")}
		}

		err := m.client.DeleteCharacter(ctx, m.accessToken, characterID)
		if err != nil {
			return taskCompleteMsg{err: fmt.Errorf("删除角色失败: %w", err)}
		}

		return taskCompleteMsg{resultURL: fmt.Sprintf("角色 %s 已删除", characterID)}
	}
}

func (m taskModel) executeStoryboard() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		c := m.client
		at := m.accessToken

		prompt := m.params["prompt"]
		if prompt == "" {
			return taskCompleteMsg{err: fmt.Errorf("分镜提示词不能为空")}
		}
		orientation := m.params["orientation"]
		if orientation == "" {
			orientation = "landscape"
		}
		nFrames, _ := strconv.Atoi(m.params["n_frames"])
		if nFrames == 0 {
			nFrames = 450
		}

		// 获取 sentinel token
		sentinelToken, err := c.GenerateSentinelToken(ctx, at)
		if err != nil {
			return taskCompleteMsg{err: fmt.Errorf("获取 sentinel token 失败: %w", err)}
		}

		// 创建分镜任务
		taskID, err := c.CreateStoryboardTask(ctx, at, sentinelToken, prompt, orientation, nFrames, "", "")
		if err != nil {
			return taskCompleteMsg{err: err}
		}

		// 轮询
		err = c.PollVideoTask(ctx, at, taskID, 3*time.Second, 600*time.Second, nil)
		if err != nil {
			return taskCompleteMsg{err: err}
		}

		// 获取下载链接
		dlURL, err := c.GetDownloadURL(ctx, at, taskID)
		if err != nil {
			return taskCompleteMsg{err: err}
		}

		localPath, err := downloadToLocal(ctx, c, dlURL, taskID, ".mp4")
		if err != nil {
			return taskCompleteMsg{resultURL: dlURL}
		}
		return taskCompleteMsg{resultURL: localPath}
	}
}

func (m taskModel) executePublishVideo() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		c := m.client
		at := m.accessToken

		generationID := m.params["generation_id"]
		if generationID == "" {
			return taskCompleteMsg{err: fmt.Errorf("generation_id 不能为空")}
		}

		sentinelToken, err := c.GenerateSentinelToken(ctx, at)
		if err != nil {
			return taskCompleteMsg{err: fmt.Errorf("获取 sentinel token 失败: %w", err)}
		}

		postID, err := c.PublishVideo(ctx, at, sentinelToken, generationID)
		if err != nil {
			return taskCompleteMsg{err: fmt.Errorf("发布视频失败: %w", err)}
		}

		return taskCompleteMsg{resultURL: fmt.Sprintf("发布成功\n  Post ID: %s\n  去水印链接: https://sora.chatgpt.com/p/%s", postID, postID)}
	}
}

func (m taskModel) executeDeletePost() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		postID := m.params["post_id"]
		if postID == "" {
			return taskCompleteMsg{err: fmt.Errorf("post_id 不能为空")}
		}

		err := m.client.DeletePost(ctx, m.accessToken, postID)
		if err != nil {
			return taskCompleteMsg{err: fmt.Errorf("删除帖子失败: %w", err)}
		}

		return taskCompleteMsg{resultURL: fmt.Sprintf("帖子 %s 已删除", postID)}
	}
}

func (m taskModel) pollOnce() tea.Cmd {
	// 此方法预留给未来的非阻塞轮询模式使用
	return nil
}

// downloadToLocal 下载文件到当前目录，返回本地文件路径
func downloadToLocal(ctx context.Context, c *sora.Client, fileURL, taskID, defaultExt string) (string, error) {
	data, err := c.DownloadFile(ctx, fileURL)
	if err != nil {
		return "", err
	}

	ext := sora.ExtFromURL(fileURL, defaultExt)
	filename := taskID + ext
	absPath, _ := filepath.Abs(filename)

	if err := os.WriteFile(absPath, data, 0644); err != nil {
		return "", err
	}
	return absPath, nil
}
