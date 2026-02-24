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
	accessToken = ""
	proxyStr    = ""
	prompt      = "一只可爱的小猫在草地上奔跑"

	orientation  = "landscape" // landscape / portrait
	nFrames      = 150         // 150(5s) / 300(10s) / 450(15s) / 750(25s)
	model        = "sy_8"      // sy_8(标准) / sy_ore(Pro)
	size         = "small"     // small(标准) / large(高清,仅Pro)
	pollInterval = 3 * time.Second
	pollTimeout  = 600 * time.Second
)

func main() {
	fmt.Println("============================================================")
	fmt.Println("  Sora 视频生成工具 (go-sora2api)")
	fmt.Println("============================================================")

	reader := bufio.NewReader(os.Stdin)

	if accessToken == "" {
		fmt.Print("\n请输入 access_token: ")
		line, _ := reader.ReadString('\n')
		accessToken = strings.TrimSpace(line)
		if accessToken == "" {
			fmt.Println("[错误] access_token 不能为空!")
			return
		}
	}

	if proxyStr == "" {
		fmt.Print("请输入代理 (留空不使用代理): ")
		line, _ := reader.ReadString('\n')
		proxyStr = strings.TrimSpace(line)
	}

	proxyURL := util.ParseProxy(proxyStr)

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
	fmt.Printf("提示词: %s\n\n", prompt)

	// 创建客户端
	c, err := client.New(proxyURL)
	if err != nil {
		fmt.Printf("[错误] 创建客户端失败: %v\n", err)
		return
	}

	// 1. 获取 sentinel token
	fmt.Println("[步骤 1/4] 正在获取 sentinel token...")
	sentinelToken, err := c.GenerateSentinelToken(accessToken)
	if err != nil {
		fmt.Printf("[错误] 获取 sentinel token 失败: %v\n", err)
		return
	}
	fmt.Println("[步骤 1/4] sentinel token 获取成功")

	// 2. 创建视频任务
	fmt.Println("[步骤 2/4] 正在创建视频任务...")
	fmt.Printf("  参数: 方向=%s 帧数=%d 模型=%s 尺寸=%s\n", orientation, nFrames, model, size)

	taskID, err := c.CreateVideoTask(accessToken, sentinelToken, prompt, orientation, nFrames, model, size)
	if err != nil {
		fmt.Printf("[错误] 创建任务失败: %v\n", err)
		return
	}
	fmt.Printf("[步骤 2/4] 任务创建成功! ID: %s\n", taskID)

	// 3. 轮询进度
	fmt.Printf("[步骤 3/4] 开始轮询任务进度 (超时: %v)...\n", pollTimeout)
	err = c.PollVideoTask(accessToken, taskID, pollInterval, pollTimeout)
	if err != nil {
		fmt.Printf("[错误] 轮询失败: %v\n", err)
		return
	}

	// 4. 获取下载链接
	fmt.Println("[步骤 4/4] 正在获取视频下载链接...")
	downloadURL, err := c.GetDownloadURL(accessToken, taskID)
	if err != nil {
		fmt.Printf("[错误] 获取下载链接失败: %v\n", err)
		return
	}

	fmt.Printf("\n[完成] 视频下载链接:\n  %s\n", downloadURL)
}
