package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/DouDOU-start/go-sora2api/sora"
)

var (
	pollInterval = 3 * time.Second
	pollTimeout  = 600 * time.Second
)

func main() {
	fmt.Println("============================================================")
	fmt.Println("  Sora 视频/图片生成工具 (go-sora2api)")
	fmt.Println("============================================================")

	reader := bufio.NewReader(os.Stdin)

	// 输入 access_token
	fmt.Print("\n请输入 access_token: ")
	accessToken := readLine(reader)
	if accessToken == "" {
		fmt.Println("[错误] access_token 不能为空!")
		return
	}

	// 输入代理
	fmt.Print("请输入代理 (留空不使用代理): ")
	proxyURL := sora.ParseProxy(readLine(reader))

	// 选择生成类型
	fmt.Println("\n请选择功能:")
	fmt.Println("  1) 文生图")
	fmt.Println("  2) 图生图")
	fmt.Println("  3) 文生视频")
	fmt.Println("  4) 图生视频")
	fmt.Println("  5) Remix 视频")
	fmt.Println("  6) 提示词优化")
	fmt.Println("  7) 获取去水印链接")
	fmt.Print("请输入 (1-7) [默认 1]: ")
	genChoice := readLine(reader)
	if genChoice == "" {
		genChoice = "1"
	}

	// 创建客户端
	c, err := sora.New(proxyURL)
	if err != nil {
		fmt.Printf("[错误] 创建客户端失败: %v\n", err)
		return
	}

	// 显示配置
	tokenDisplay := accessToken
	if len(accessToken) > 30 {
		tokenDisplay = accessToken[:20] + "..." + accessToken[len(accessToken)-10:]
	}
	fmt.Printf("\n--- 配置信息 ---\n")
	fmt.Printf("Token: %s\n", tokenDisplay)
	if proxyURL != "" {
		fmt.Printf("代理: %s\n", proxyURL)
	} else {
		fmt.Printf("代理: 无\n")
	}

	switch genChoice {
	case "1":
		generateImage(reader, c, accessToken, "")
	case "2":
		generateImageFromImage(reader, c, accessToken)
	case "3":
		generateVideo(reader, c, accessToken, "")
	case "4":
		generateVideoFromImage(reader, c, accessToken)
	case "5":
		remixVideo(reader, c, accessToken)
	case "6":
		enhancePrompt(reader, c, accessToken)
	case "7":
		getWatermarkFreeURL(reader, c, accessToken)
	default:
		fmt.Println("[错误] 无效的选择")
	}
}

// 进度回调：在终端打印进度
func printProgress(p sora.Progress) {
	fmt.Printf("\r  进度: %d%%  状态: %s  耗时: %ds    ", p.Percent, p.Status, p.Elapsed)
}

func inputPrompt(reader *bufio.Reader) string {
	fmt.Print("\n请输入提示词 [默认: 一只可爱的小猫在草地上奔跑]: ")
	prompt := readLine(reader)
	if prompt == "" {
		prompt = "一只可爱的小猫在草地上奔跑"
	}
	return prompt
}

func uploadImageFromPath(reader *bufio.Reader, c *sora.Client, accessToken string) string {
	fmt.Print("\n请输入图片路径: ")
	imagePath := readLine(reader)
	if imagePath == "" {
		fmt.Println("[错误] 图片路径不能为空!")
		return ""
	}

	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		fmt.Printf("[错误] 读取图片失败: %v\n", err)
		return ""
	}

	// 从路径中提取文件名
	parts := strings.Split(strings.ReplaceAll(imagePath, "\\", "/"), "/")
	filename := parts[len(parts)-1]

	fmt.Println("[上传] 正在上传图片...")
	mediaID, err := c.UploadImage(accessToken, imageData, filename)
	if err != nil {
		fmt.Printf("[错误] 上传图片失败: %v\n", err)
		return ""
	}
	fmt.Printf("[上传] 图片上传成功! MediaID: %s\n", mediaID)
	return mediaID
}

func getSentinelToken(c *sora.Client, accessToken string) (string, bool) {
	fmt.Println("\n[步骤] 正在获取 sentinel token...")
	sentinelToken, err := c.GenerateSentinelToken(accessToken)
	if err != nil {
		fmt.Printf("[错误] 获取 sentinel token 失败: %v\n", err)
		return "", false
	}
	fmt.Println("[步骤] sentinel token 获取成功")
	return sentinelToken, true
}

func generateImage(reader *bufio.Reader, c *sora.Client, accessToken, mediaID string) {
	prompt := inputPrompt(reader)

	// 选择图片尺寸
	fmt.Println("\n请选择图片尺寸:")
	fmt.Println("  1) 正方形 (360x360)")
	fmt.Println("  2) 横向   (540x360)")
	fmt.Println("  3) 纵向   (360x540)")
	fmt.Print("请输入 (1/2/3) [默认 1]: ")
	sizeChoice := readLine(reader)

	width, height := 360, 360
	switch sizeChoice {
	case "2":
		width, height = 540, 360
	case "3":
		width, height = 360, 540
	}

	fmt.Printf("提示词: %s\n", prompt)
	fmt.Printf("尺寸: %dx%d\n\n", width, height)

	sentinelToken, ok := getSentinelToken(c, accessToken)
	if !ok {
		return
	}

	// 创建图片任务
	fmt.Println("[步骤] 正在创建图片任务...")
	taskID, err := c.CreateImageTaskWithImage(accessToken, sentinelToken, prompt, width, height, mediaID)
	if err != nil {
		fmt.Printf("[错误] 创建任务失败: %v\n", err)
		return
	}
	fmt.Printf("[步骤] 任务创建成功! ID: %s\n", taskID)

	// 轮询进度
	fmt.Printf("[步骤] 开始轮询任务进度 (超时: %v)...\n", pollTimeout)
	imageURL, err := c.PollImageTask(accessToken, taskID, pollInterval, pollTimeout, printProgress)
	if err != nil {
		fmt.Printf("\n[错误] 轮询失败: %v\n", err)
		return
	}

	fmt.Printf("\n[完成] 图片下载链接:\n  %s\n", imageURL)
}

func generateImageFromImage(reader *bufio.Reader, c *sora.Client, accessToken string) {
	mediaID := uploadImageFromPath(reader, c, accessToken)
	if mediaID == "" {
		return
	}
	generateImage(reader, c, accessToken, mediaID)
}

func inputVideoParams(reader *bufio.Reader) (orientation string, nFrames int, model, size string) {
	// 选择方向
	fmt.Println("\n请选择视频方向:")
	fmt.Println("  1) 横向 (landscape)")
	fmt.Println("  2) 纵向 (portrait)")
	fmt.Print("请输入 (1/2) [默认 1]: ")
	orientChoice := readLine(reader)

	orientation = "landscape"
	if orientChoice == "2" {
		orientation = "portrait"
	}

	// 选择时长
	fmt.Println("\n请选择视频时长:")
	fmt.Println("  1) 5 秒  (150 帧)")
	fmt.Println("  2) 10 秒 (300 帧)")
	fmt.Println("  3) 15 秒 (450 帧)")
	fmt.Println("  4) 25 秒 (750 帧)")
	fmt.Print("请输入 (1/2/3/4) [默认 1]: ")
	durChoice := readLine(reader)

	nFrames = 150
	switch durChoice {
	case "2":
		nFrames = 300
	case "3":
		nFrames = 450
	case "4":
		nFrames = 750
	}

	// 选择模型
	fmt.Println("\n请选择模型:")
	fmt.Println("  1) 标准 (sy_8)")
	fmt.Println("  2) Pro  (sy_ore)")
	fmt.Print("请输入 (1/2) [默认 1]: ")
	modelChoice := readLine(reader)

	model = "sy_8"
	size = "small"
	if modelChoice == "2" {
		model = "sy_ore"
		fmt.Println("\n请选择清晰度:")
		fmt.Println("  1) 标准 (small)")
		fmt.Println("  2) 高清 (large)")
		fmt.Print("请输入 (1/2) [默认 1]: ")
		hdChoice := readLine(reader)
		if hdChoice == "2" {
			size = "large"
		}
	}

	return
}

func inputStyleID(reader *bufio.Reader, prompt string) (string, string) {
	// 自动从提示词提取风格
	cleanedPrompt, styleID := sora.ExtractStyle(prompt)
	if styleID != "" {
		fmt.Printf("[风格] 从提示词中提取到风格: %s\n", styleID)
		return cleanedPrompt, styleID
	}

	// 手动选择风格
	fmt.Println("\n请选择视频风格 (留空不使用):")
	fmt.Printf("  可选: %s\n", strings.Join(sora.ValidStyles, ", "))
	fmt.Print("请输入风格名称: ")
	style := readLine(reader)
	return prompt, style
}

func generateVideo(reader *bufio.Reader, c *sora.Client, accessToken, mediaID string) {
	prompt := inputPrompt(reader)
	orientation, nFrames, model, size := inputVideoParams(reader)
	prompt, styleID := inputStyleID(reader, prompt)

	fmt.Printf("\n--- 视频参数 ---\n")
	fmt.Printf("提示词: %s\n", prompt)
	fmt.Printf("方向: %s  帧数: %d  模型: %s  尺寸: %s\n", orientation, nFrames, model, size)
	if styleID != "" {
		fmt.Printf("风格: %s\n", styleID)
	}
	if mediaID != "" {
		fmt.Printf("输入图片: %s\n", mediaID)
	}
	fmt.Println()

	sentinelToken, ok := getSentinelToken(c, accessToken)
	if !ok {
		return
	}

	// 创建视频任务
	fmt.Println("[步骤] 正在创建视频任务...")
	taskID, err := c.CreateVideoTaskWithOptions(accessToken, sentinelToken, prompt, orientation, nFrames, model, size, mediaID, styleID)
	if err != nil {
		fmt.Printf("[错误] 创建任务失败: %v\n", err)
		return
	}
	fmt.Printf("[步骤] 任务创建成功! ID: %s\n", taskID)

	// 轮询进度
	fmt.Printf("[步骤] 开始轮询任务进度 (超时: %v)...\n", pollTimeout)
	err = c.PollVideoTask(accessToken, taskID, pollInterval, pollTimeout, printProgress)
	if err != nil {
		fmt.Printf("\n[错误] 轮询失败: %v\n", err)
		return
	}

	// 获取下载链接
	fmt.Println("\n[步骤] 正在获取视频下载链接...")
	downloadURL, err := c.GetDownloadURL(accessToken, taskID)
	if err != nil {
		fmt.Printf("[错误] 获取下载链接失败: %v\n", err)
		return
	}

	fmt.Printf("\n[完成] 视频下载链接:\n  %s\n", downloadURL)
}

func generateVideoFromImage(reader *bufio.Reader, c *sora.Client, accessToken string) {
	mediaID := uploadImageFromPath(reader, c, accessToken)
	if mediaID == "" {
		return
	}
	generateVideo(reader, c, accessToken, mediaID)
}

func remixVideo(reader *bufio.Reader, c *sora.Client, accessToken string) {
	fmt.Print("\n请输入 Remix 视频 ID 或分享链接: ")
	input := readLine(reader)
	remixID := sora.ExtractRemixID(input)
	if remixID == "" {
		fmt.Println("[错误] 无法解析 Remix 视频 ID，格式应为 s_[hex32]")
		return
	}
	fmt.Printf("[Remix] 视频 ID: %s\n", remixID)

	prompt := inputPrompt(reader)

	// 选择方向
	fmt.Println("\n请选择视频方向:")
	fmt.Println("  1) 横向 (landscape)")
	fmt.Println("  2) 纵向 (portrait)")
	fmt.Print("请输入 (1/2) [默认 1]: ")
	orientChoice := readLine(reader)
	orientation := "landscape"
	if orientChoice == "2" {
		orientation = "portrait"
	}

	// 选择时长
	fmt.Println("\n请选择视频时长:")
	fmt.Println("  1) 5 秒  (150 帧)")
	fmt.Println("  2) 10 秒 (300 帧)")
	fmt.Println("  3) 15 秒 (450 帧)")
	fmt.Println("  4) 25 秒 (750 帧)")
	fmt.Print("请输入 (1/2/3/4) [默认 1]: ")
	durChoice := readLine(reader)
	nFrames := 150
	switch durChoice {
	case "2":
		nFrames = 300
	case "3":
		nFrames = 450
	case "4":
		nFrames = 750
	}

	prompt, styleID := inputStyleID(reader, prompt)

	fmt.Printf("\n--- Remix 参数 ---\n")
	fmt.Printf("原视频: %s\n", remixID)
	fmt.Printf("提示词: %s\n", prompt)
	fmt.Printf("方向: %s  帧数: %d\n", orientation, nFrames)
	if styleID != "" {
		fmt.Printf("风格: %s\n", styleID)
	}
	fmt.Println()

	sentinelToken, ok := getSentinelToken(c, accessToken)
	if !ok {
		return
	}

	fmt.Println("[步骤] 正在创建 Remix 任务...")
	taskID, err := c.RemixVideo(accessToken, sentinelToken, remixID, prompt, orientation, nFrames, styleID)
	if err != nil {
		fmt.Printf("[错误] 创建 Remix 任务失败: %v\n", err)
		return
	}
	fmt.Printf("[步骤] 任务创建成功! ID: %s\n", taskID)

	fmt.Printf("[步骤] 开始轮询任务进度 (超时: %v)...\n", pollTimeout)
	err = c.PollVideoTask(accessToken, taskID, pollInterval, pollTimeout, printProgress)
	if err != nil {
		fmt.Printf("\n[错误] 轮询失败: %v\n", err)
		return
	}

	fmt.Println("\n[步骤] 正在获取视频下载链接...")
	downloadURL, err := c.GetDownloadURL(accessToken, taskID)
	if err != nil {
		fmt.Printf("[错误] 获取下载链接失败: %v\n", err)
		return
	}

	fmt.Printf("\n[完成] 视频下载链接:\n  %s\n", downloadURL)
}

func enhancePrompt(reader *bufio.Reader, c *sora.Client, accessToken string) {
	fmt.Print("\n请输入要优化的提示词: ")
	prompt := readLine(reader)
	if prompt == "" {
		fmt.Println("[错误] 提示词不能为空!")
		return
	}

	fmt.Println("\n请选择扩展程度:")
	fmt.Println("  1) 中等 (medium)")
	fmt.Println("  2) 详细 (long)")
	fmt.Print("请输入 (1/2) [默认 1]: ")
	levelChoice := readLine(reader)
	expansionLevel := "medium"
	if levelChoice == "2" {
		expansionLevel = "long"
	}

	fmt.Println("\n请选择目标时长:")
	fmt.Println("  1) 10 秒")
	fmt.Println("  2) 15 秒")
	fmt.Println("  3) 20 秒")
	fmt.Print("请输入 (1/2/3) [默认 1]: ")
	durChoice := readLine(reader)
	durationSec := 10
	switch durChoice {
	case "2":
		durationSec = 15
	case "3":
		durationSec = 20
	}

	fmt.Println("\n[步骤] 正在优化提示词...")
	enhanced, err := c.EnhancePrompt(accessToken, prompt, expansionLevel, durationSec)
	if err != nil {
		fmt.Printf("[错误] 提示词优化失败: %v\n", err)
		return
	}

	fmt.Printf("\n[完成] 优化后的提示词:\n  %s\n", enhanced)
}

func getWatermarkFreeURL(reader *bufio.Reader, c *sora.Client, _ string) {
	fmt.Println("\n--- 去水印功能需要 refresh_token ---")
	fmt.Print("请输入 refresh_token: ")
	refreshToken := readLine(reader)
	if refreshToken == "" {
		fmt.Println("[错误] refresh_token 不能为空!")
		return
	}

	fmt.Print("请输入 client_id (留空使用默认值): ")
	clientID := readLine(reader)

	// 刷新获取专用 access_token
	fmt.Println("\n[步骤] 正在刷新 access_token...")
	soraToken, newRT, err := c.RefreshAccessToken(refreshToken, clientID)
	if err != nil {
		fmt.Printf("[错误] 刷新 token 失败: %v\n", err)
		return
	}
	fmt.Println("[步骤] access_token 获取成功")
	if newRT != refreshToken {
		fmt.Printf("[提示] refresh_token 已更新，请保存新值:\n  %s\n", newRT)
	}

	fmt.Print("\n请输入 Sora 视频分享链接或视频 ID: ")
	input := readLine(reader)
	if input == "" {
		fmt.Println("[错误] 输入不能为空!")
		return
	}

	fmt.Println("[步骤] 正在获取去水印链接...")
	url, err := c.GetWatermarkFreeURL(soraToken, input)
	if err != nil {
		fmt.Printf("[错误] 获取失败: %v\n", err)
		return
	}

	fmt.Printf("\n[完成] 去水印下载链接:\n  %s\n", url)
}

func readLine(reader *bufio.Reader) string {
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}
