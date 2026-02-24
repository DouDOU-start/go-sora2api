package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"go-sora2api/internal/client"
	"go-sora2api/internal/util"
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
	proxyURL := util.ParseProxy(readLine(reader))

	// 选择生成类型
	fmt.Println("\n请选择生成类型:")
	fmt.Println("  1) 图片")
	fmt.Println("  2) 视频")
	fmt.Print("请输入 (1/2) [默认 1]: ")
	genChoice := readLine(reader)
	if genChoice == "" {
		genChoice = "1"
	}

	// 输入提示词
	fmt.Print("\n请输入提示词 [默认: 一只可爱的小猫在草地上奔跑]: ")
	prompt := readLine(reader)
	if prompt == "" {
		prompt = "一只可爱的小猫在草地上奔跑"
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
	fmt.Printf("提示词: %s\n", prompt)

	// 创建客户端
	c, err := client.New(proxyURL)
	if err != nil {
		fmt.Printf("[错误] 创建客户端失败: %v\n", err)
		return
	}

	// 获取 sentinel token
	fmt.Println("\n[步骤 1] 正在获取 sentinel token...")
	sentinelToken, err := c.GenerateSentinelToken(accessToken)
	if err != nil {
		fmt.Printf("[错误] 获取 sentinel token 失败: %v\n", err)
		return
	}
	fmt.Println("[步骤 1] sentinel token 获取成功")

	if genChoice == "1" {
		generateImage(reader, c, accessToken, sentinelToken, prompt)
	} else {
		generateVideo(reader, c, accessToken, sentinelToken, prompt)
	}
}

func generateImage(reader *bufio.Reader, c *client.SoraClient, accessToken, sentinelToken, prompt string) {
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

	fmt.Printf("\n--- 图片参数 ---\n")
	fmt.Printf("尺寸: %dx%d\n\n", width, height)

	// 创建图片任务
	fmt.Println("[步骤 2/3] 正在创建图片任务...")
	taskID, err := c.CreateImageTask(accessToken, sentinelToken, prompt, width, height)
	if err != nil {
		fmt.Printf("[错误] 创建任务失败: %v\n", err)
		return
	}
	fmt.Printf("[步骤 2/3] 任务创建成功! ID: %s\n", taskID)

	// 轮询进度并获取结果
	fmt.Printf("[步骤 3/3] 开始轮询任务进度 (超时: %v)...\n", pollTimeout)
	imageURL, err := c.PollImageTask(accessToken, taskID, pollInterval, pollTimeout)
	if err != nil {
		fmt.Printf("[错误] 轮询失败: %v\n", err)
		return
	}

	fmt.Printf("\n[完成] 图片下载链接:\n  %s\n", imageURL)
}

func generateVideo(reader *bufio.Reader, c *client.SoraClient, accessToken, sentinelToken, prompt string) {
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

	// 选择模型
	fmt.Println("\n请选择模型:")
	fmt.Println("  1) 标准 (sy_8)")
	fmt.Println("  2) Pro  (sy_ore)")
	fmt.Print("请输入 (1/2) [默认 1]: ")
	modelChoice := readLine(reader)

	model := "sy_8"
	size := "small"
	if modelChoice == "2" {
		model = "sy_ore"
		// Pro 模型可选高清
		fmt.Println("\n请选择清晰度:")
		fmt.Println("  1) 标准 (small)")
		fmt.Println("  2) 高清 (large)")
		fmt.Print("请输入 (1/2) [默认 1]: ")
		hdChoice := readLine(reader)
		if hdChoice == "2" {
			size = "large"
		}
	}

	fmt.Printf("\n--- 视频参数 ---\n")
	fmt.Printf("方向: %s  帧数: %d  模型: %s  尺寸: %s\n\n", orientation, nFrames, model, size)

	// 创建视频任务
	fmt.Println("[步骤 2/4] 正在创建视频任务...")
	taskID, err := c.CreateVideoTask(accessToken, sentinelToken, prompt, orientation, nFrames, model, size)
	if err != nil {
		fmt.Printf("[错误] 创建任务失败: %v\n", err)
		return
	}
	fmt.Printf("[步骤 2/4] 任务创建成功! ID: %s\n", taskID)

	// 轮询进度
	fmt.Printf("[步骤 3/4] 开始轮询任务进度 (超时: %v)...\n", pollTimeout)
	err = c.PollVideoTask(accessToken, taskID, pollInterval, pollTimeout)
	if err != nil {
		fmt.Printf("[错误] 轮询失败: %v\n", err)
		return
	}

	// 获取下载链接
	fmt.Println("[步骤 4/4] 正在获取视频下载链接...")
	downloadURL, err := c.GetDownloadURL(accessToken, taskID)
	if err != nil {
		fmt.Printf("[错误] 获取下载链接失败: %v\n", err)
		return
	}

	fmt.Printf("\n[完成] 视频下载链接:\n  %s\n", downloadURL)
}

func readLine(reader *bufio.Reader) string {
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}
