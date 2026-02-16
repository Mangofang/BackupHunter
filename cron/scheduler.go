package cron

import (
	"BackupHunter/db"
	"BackupHunter/handlers"
	"BackupHunter/types"
	"fmt"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
)

var (
	cronManager *cron.Cron
	jobIDs      = make(map[string]cron.EntryID)
	mutex       sync.Mutex
)

// 初始化cron调度器
func InitScheduler() error {
	cronManager = cron.New(cron.WithSeconds())

	// 从数据库加载所有启用的任务
	tasks, err := db.GetAllActiveTasks()
	if err != nil {
		return err
	}

	for _, task := range tasks {
		if task.Active {
			if err := addTask(task); err != nil {
				fmt.Printf("[ERROR] Failed to add task %s: %v\n", task.Name, err)
			}
		}
	}

	cronManager.Start()
	return nil
}

// 创建扫描任务
func CreateScanBackupFilesTask(c *gin.Context) {
	domain := c.PostForm("domain")
	taskName := c.PostForm("taskname")
	cronExpr := c.PostForm("cron")

	if domain == "" || taskName == "" || cronExpr == "" {
		c.JSON(400, gin.H{
			"state":   "error",
			"message": "Missing required parameters",
		})
		return
	}

	task := types.TableTasks{
		Id:        uuid.New().String(),
		Name:      taskName,
		Domain:    domain,
		Cron_expr: cronExpr,
		Active:    true,
		State:     0,
	}

	// 保存到数据库
	if err := db.AddCronTask(task); err != nil {
		fmt.Println("[ERROR] Create new scan backupfile task error:", err)
		c.JSON(500, gin.H{
			"state":   "error",
			"message": "Server error",
		})
		return
	}

	// 注册到调度器
	if err := addTask(task); err != nil {
		fmt.Println("[ERROR] Failed to register task to scheduler:", err)
		c.JSON(500, gin.H{
			"state":   "error",
			"message": "Server error",
		})
		return
	}

	c.JSON(200, gin.H{
		"state": "success",
		"data": gin.H{
			"taskId": task.Id,
		},
	})
}

// 删除任务
func RemoveTask(c *gin.Context) {
	taskId := c.PostForm("taskid")
	mutex.Lock()
	defer mutex.Unlock()

	if id, exists := jobIDs[taskId]; exists {
		task, err := db.GetTaskById(taskId)
		cronManager.Remove(id)
		delete(jobIDs, taskId)
		db.DeleteTask(taskId)
		if err != nil {
			fmt.Println("[ERROR] In remove task gettaskbyid error: ", err.Error())
			c.JSON(200, gin.H{
				"state": "error",
				"msg":   "Server error",
			})
			return
		}
		err = db.DeleteSubDomain(task.Domain)
		if err != nil {
			fmt.Println("[ERROR] In remove task deletesubdomain error: ", err.Error())
			c.JSON(200, gin.H{
				"state": "error",
				"msg":   "Server error",
			})
			return
		}
	}
	c.JSON(200, gin.H{
		"state": "success",
	})
}

// 暂停任务
func StopTask(c *gin.Context) {
	taskId := c.PostForm("taskid")
	mutex.Lock()
	defer mutex.Unlock()
	if id, exists := jobIDs[taskId]; exists {
		cronManager.Remove(id)
		delete(jobIDs, taskId)
	}
	db.SetTaskActive(taskId, false)
	taskActive, _ := db.GetTaskById(taskId)
	c.JSON(200, gin.H{
		"state": "success",
		"task": gin.H{
			"taskId": taskId,
			"active": taskActive.Active,
		}})
}

// 开始任务
func StartTask(c *gin.Context) {
	taskId := c.PostForm("taskid")
	task, err := db.GetTaskById(taskId)
	if err != nil {
		fmt.Println("[ERROR] In start task gettaskbyid error: ", err.Error())
		c.JSON(500, gin.H{
			"state":   "error",
			"message": "Server error",
		})
		return
	}
	if err := addTask(task); err != nil {
		fmt.Println("[ERROR] In start task addtask error: ", err.Error())
		c.JSON(500, gin.H{
			"state":   "error",
			"message": "Server error",
		})
		return
	}
	db.SetTaskActive(taskId, true)
	taskActive, _ := db.GetTaskById(taskId)
	c.JSON(200, gin.H{
		"state":  "success",
		"active": taskActive.Active,
	})
}

// 添加任务到调度器
func addTask(task types.TableTasks) error {
	mutex.Lock()
	defer mutex.Unlock()

	// 如果任务已存在，先移除
	if id, exists := jobIDs[task.Id]; exists {
		cronManager.Remove(id)
	}

	id, err := cronManager.AddFunc(task.Cron_expr, func() {
		fmt.Printf("[INFO] Executing task: %s\n", task.Name)
		if err := handlers.ScanBackupFiles(task); err != nil {
			fmt.Printf("[ERROR] Task %s execution failed: %v\n", task.Name, err)
			db.SetTaskState(task.Id, 4)
			return
		}
	})

	if err != nil {
		return err
	}

	jobIDs[task.Id] = id
	return nil
}

// 停止调度器
func Stop() {
	if cronManager != nil {
		cronManager.Stop()
	}
}
