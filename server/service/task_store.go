package service

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/DouDOU-start/go-sora2api/server/model"
	"github.com/DouDOU-start/go-sora2api/sora"
	"gorm.io/gorm"
)

// TaskStore 任务存储与后台轮询
type TaskStore struct {
	db        *gorm.DB
	scheduler *Scheduler
	polls     sync.Map // taskID → cancel func
}

// NewTaskStore 创建任务存储
func NewTaskStore(db *gorm.DB, scheduler *Scheduler) *TaskStore {
	return &TaskStore{db: db, scheduler: scheduler}
}

// Create 创建任务记录
func (ts *TaskStore) Create(task *model.SoraTask) error {
	return ts.db.Create(task).Error
}

// Get 获取任务
func (ts *TaskStore) Get(taskID string) (*model.SoraTask, error) {
	var task model.SoraTask
	if err := ts.db.Where("id = ?", taskID).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// ListTasks 分页查询任务
func (ts *TaskStore) ListTasks(status string, page, pageSize int) ([]model.SoraTask, int64, error) {
	var tasks []model.SoraTask
	var total int64

	query := ts.db.Model(&model.SoraTask{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// StartPolling 启动后台轮询任务状态
func (ts *TaskStore) StartPolling(task *model.SoraTask, account *model.SoraAccount) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	ts.polls.Store(task.ID, cancel)

	go func() {
		defer cancel()
		defer ts.polls.Delete(task.ID)

		client, err := sora.New(ts.scheduler.GetProxyURL())
		if err != nil {
			ts.failTask(task.ID, fmt.Sprintf("创建 Sora 客户端失败: %v", err))
			return
		}

		// 更新为进行中
		ts.db.Model(&model.SoraTask{}).Where("id = ?", task.ID).Updates(map[string]interface{}{
			"status":   model.TaskStatusInProgress,
			"progress": 5,
		})

		startTime := time.Now()
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		maxProgress := 0

		for {
			select {
			case <-ctx.Done():
				ts.failTask(task.ID, "轮询超时")
				return
			case <-ticker.C:
				switch task.Type {
				case "image":
					ts.pollImageTask(ctx, client, account.AccessToken, task, startTime)
				default:
					ts.pollVideoTask(ctx, client, account.AccessToken, task, startTime, &maxProgress)
				}

				// 检查任务是否已完成
				var current model.SoraTask
				if err := ts.db.Where("id = ?", task.ID).First(&current).Error; err == nil {
					if current.Status == model.TaskStatusCompleted || current.Status == model.TaskStatusFailed {
						return
					}
				}
			}
		}
	}()
}

// pollVideoTask 轮询视频任务
func (ts *TaskStore) pollVideoTask(ctx context.Context, client *sora.Client, at string, task *model.SoraTask, startTime time.Time, maxProgress *int) {
	result := client.QueryVideoTaskOnce(ctx, at, task.SoraTaskID, startTime, *maxProgress)
	if result.Err != nil {
		log.Printf("[poll] 视频任务 %s 查询失败: %v", task.ID, result.Err)
		return
	}

	// 更新进度
	if result.Progress.Percent > *maxProgress {
		*maxProgress = result.Progress.Percent
		ts.db.Model(&model.SoraTask{}).Where("id = ?", task.ID).Update("progress", result.Progress.Percent)
	}

	if result.Done {
		// 获取下载链接
		downloadURL, err := client.GetDownloadURL(ctx, at, task.SoraTaskID)
		if err != nil {
			log.Printf("[poll] 视频任务 %s 获取下载链接失败: %v", task.ID, err)
			ts.failTask(task.ID, fmt.Sprintf("获取下载链接失败: %v", err))
			return
		}
		ts.completeTask(task.ID, downloadURL, "")

		// 异步更新账号配额
		go ts.syncAccountCredit(ctx, client, at, task.AccountID)
	}
}

// pollImageTask 轮询图片任务
func (ts *TaskStore) pollImageTask(ctx context.Context, client *sora.Client, at string, task *model.SoraTask, startTime time.Time) {
	result := client.QueryImageTaskOnce(ctx, at, task.SoraTaskID, startTime)
	if result.Err != nil {
		log.Printf("[poll] 图片任务 %s 查询失败: %v", task.ID, result.Err)
		return
	}

	// 更新进度
	ts.db.Model(&model.SoraTask{}).Where("id = ?", task.ID).Update("progress", result.Progress.Percent)

	if result.Done {
		ts.completeTask(task.ID, "", result.ImageURL)

		// 异步更新账号配额
		go ts.syncAccountCredit(ctx, client, at, task.AccountID)
	}
}

// completeTask 标记任务完成
func (ts *TaskStore) completeTask(taskID, downloadURL, imageURL string) {
	now := time.Now()
	updates := map[string]interface{}{
		"status":       model.TaskStatusCompleted,
		"progress":     100,
		"completed_at": &now,
	}
	if downloadURL != "" {
		updates["download_url"] = downloadURL
	}
	if imageURL != "" {
		updates["image_url"] = imageURL
	}
	ts.db.Model(&model.SoraTask{}).Where("id = ?", taskID).Updates(updates)
	log.Printf("[poll] 任务 %s 已完成", taskID)
}

// failTask 标记任务失败
func (ts *TaskStore) failTask(taskID, errMsg string) {
	now := time.Now()
	ts.db.Model(&model.SoraTask{}).Where("id = ?", taskID).Updates(map[string]interface{}{
		"status":        model.TaskStatusFailed,
		"error_message": errMsg,
		"completed_at":  &now,
	})
	log.Printf("[poll] 任务 %s 失败: %s", taskID, errMsg)
}

// syncAccountCredit 同步账号配额
func (ts *TaskStore) syncAccountCredit(ctx context.Context, client *sora.Client, at string, accountID int64) {
	balance, err := client.GetCreditBalance(ctx, at)
	if err != nil {
		log.Printf("[poll] 同步账号 %d 配额失败: %v", accountID, err)
		return
	}

	updates := map[string]interface{}{
		"remaining_count":    balance.RemainingCount,
		"rate_limit_reached": balance.RateLimitReached,
		"last_sync_at":       time.Now(),
	}

	if balance.RateLimitReached && balance.AccessResetsInSec > 0 {
		resetsAt := time.Now().Add(time.Duration(balance.AccessResetsInSec) * time.Second)
		updates["rate_limit_resets_at"] = resetsAt
	}

	// 额度耗尽 → 标记状态
	if balance.RemainingCount == 0 {
		updates["status"] = model.AccountStatusQuotaExhausted
	}

	ts.db.Model(&model.SoraAccount{}).Where("id = ?", accountID).Updates(updates)
}

// RecoverInProgressTasks 服务重启后恢复进行中的任务轮询
func (ts *TaskStore) RecoverInProgressTasks() {
	var tasks []model.SoraTask
	if err := ts.db.Where("status IN ?", []string{model.TaskStatusQueued, model.TaskStatusInProgress}).Find(&tasks).Error; err != nil {
		log.Printf("[task_store] 查询进行中任务失败: %v", err)
		return
	}

	for i := range tasks {
		task := &tasks[i]
		var account model.SoraAccount
		if err := ts.db.Where("id = ?", task.AccountID).First(&account).Error; err != nil {
			log.Printf("[task_store] 恢复任务 %s 失败：找不到账号 %d", task.ID, task.AccountID)
			ts.failTask(task.ID, "服务重启后找不到关联账号")
			continue
		}
		log.Printf("[task_store] 恢复轮询任务 %s（Sora: %s）", task.ID, task.SoraTaskID)
		ts.StartPolling(task, &account)
	}

	if len(tasks) > 0 {
		log.Printf("[task_store] 已恢复 %d 个进行中的任务", len(tasks))
	}
}

// DownloadVideo 下载视频并流式转发
func (ts *TaskStore) DownloadVideo(ctx context.Context, task *model.SoraTask) (io.ReadCloser, int64, string, error) {
	downloadURL := task.DownloadURL

	// 如果没有缓存的下载链接，尝试实时获取
	if downloadURL == "" {
		var account model.SoraAccount
		if err := ts.db.Where("id = ?", task.AccountID).First(&account).Error; err != nil {
			return nil, 0, "", fmt.Errorf("找不到关联账号: %w", err)
		}
		client, err := sora.New(ts.scheduler.GetProxyURL())
		if err != nil {
			return nil, 0, "", fmt.Errorf("创建 Sora 客户端失败: %w", err)
		}
		url, err := client.GetDownloadURL(ctx, account.AccessToken, task.SoraTaskID)
		if err != nil {
			return nil, 0, "", fmt.Errorf("获取下载链接失败: %w", err)
		}
		downloadURL = url
		// 缓存下载链接
		ts.db.Model(&model.SoraTask{}).Where("id = ?", task.ID).Update("download_url", downloadURL)
	}

	// HTTP GET 下载
	resp, err := http.Get(downloadURL) //nolint:gosec // 来自 Sora 官方的下载链接
	if err != nil {
		return nil, 0, "", fmt.Errorf("下载视频失败: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, 0, "", fmt.Errorf("下载视频返回 %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "video/mp4"
	}

	return resp.Body, resp.ContentLength, contentType, nil
}
