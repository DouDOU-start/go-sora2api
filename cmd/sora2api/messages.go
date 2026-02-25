package main

import "github.com/DouDOU-start/go-sora2api/sora"

// 页面切换消息
type switchPageMsg struct{ page pageType }

// 功能选择消息
type funcSelectedMsg struct{ funcType funcType }

// 任务阶段消息
type taskStepMsg struct {
	step string
	done bool
	err  error
}

// 任务进度消息
type taskProgressMsg struct {
	progress sora.Progress
}

// 任务完成消息
type taskCompleteMsg struct {
	resultURL string
	err       error
}

// 轮询计时器到期消息
type tickPollMsg struct{}

// 角色创建步骤消息（多步骤状态机）
type charCreateStepMsg struct {
	step            int
	cameoID         string
	profileAssetURL string
	imageData       []byte
	assetPointer    string
	characterID     string
	err             error
}

// 账号信息加载完成消息
type accountInfoMsg struct {
	balance *sora.CreditBalance
	sub     *sora.SubscriptionInfo
	err     error
}
