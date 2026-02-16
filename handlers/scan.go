package handlers

import (
	"BackupHunter/db"
	"BackupHunter/types"
	"BackupHunter/utils"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// 由cron调用的扫描任务
func ScanBackupFiles(task types.TableTasks) error {
	targets := db.GetSubDomain(task.Domain)
	if len(targets) == 0 {
		FindSubdomainTask(task)
		targets = db.GetSubDomain(task.Domain)
	}
	db.UpdateTaskExecInfo(task.Id)
	// 为每个子域名生成字典并启动扫描
	for _, target := range targets {
		dict := utils.GenerateDict(target)
		scanBackupFiles(task, target, dict)
	}
	onScanComplete(task.Id)
	return nil
}

// 获取扫描任务状态
func GetScanBackupFilesState(c *gin.Context) {
	taskId := c.Query("taskid")
	state, _ := db.GetTaskState(taskId)
	c.JSON(200, gin.H{
		"state":     "success",
		"taskstate": state,
	})
}

// 获取所有扫描任务
func GetAllTasks(c *gin.Context) {
	tasks, err := db.GetAllTasks()
	if err != nil {
		fmt.Println("[ERROR] In get all tasks error: ", err.Error())
		c.JSON(500, gin.H{
			"state":   "error",
			"message": "Server error",
		})
		return
	}
	c.JSON(200, gin.H{
		"state": "success",
		"data":  tasks,
	})
}

// 扫描目标subdomain完成回调函数
func onScanComplete(taskId string) {
	fmt.Println("[Message] Task " + taskId + "is complete")
	db.SetTaskState(taskId, 1)
}

// 扫描目标subdomain
func scanBackupFiles(task types.TableTasks, target string, dict []string) {
	fmt.Println("[Message] Task " + task.Id + " is scanning " + target)
	db.SetTaskState(task.Id, 2)
	maxWorkers := types.GlobalConfig.Scanner.ScanRateLimit
	timeout := types.GlobalConfig.Scanner.ScanTimeout
	proxy := types.GlobalConfig.Scanner.ScanProxy

	// 规范化 URL
	if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") {
		target = "http://" + target
	}
	if !strings.HasSuffix(target, "/") {
		target += "/"
	}

	// 创建 HTTP 客户端
	tr := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives: true,
	}
	if proxy != "" {
		if p, err := url.Parse(proxy); err == nil {
			tr.Proxy = http.ProxyURL(p)
		}
	}
	client := &http.Client{
		Timeout:   time.Duration(timeout) * time.Second,
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	var wg sync.WaitGroup
	tasks := make(chan string, len(dict))
	results := make(chan types.Result, len(dict))

	// 启动 worker
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range tasks {
				fullURL := target + path
				req, err := http.NewRequestWithContext(context.Background(), "GET", fullURL, nil)
				if err != nil {
					continue
				}
				req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

				resp, err := client.Do(req)
				if err != nil {
					continue
				}
				io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))
				contentType := resp.Header.Get("Content-Type")
				contentLength := resp.Header.Get("Content-Length")
				statusOK := resp.StatusCode == 200
				isBinary := isBinaryContentType(contentType)
				hasSize := false
				if cl, e := strconv.ParseInt(contentLength, 10, 64); e == nil && cl > 0 {
					hasSize = true
				}
				resp.Body.Close()

				if statusOK && isBinary && hasSize {
					results <- types.Result{Id: uuid.New().String(), URL: fullURL, FileName: path, Size: contentLength}
				}
			}
		}()
	}

	// 发送任务
	for _, path := range dict {
		tasks <- path
	}
	close(tasks)
	wg.Wait()
	close(results)

	// 写入结果（加锁保证多 goroutine 安全）
	mu := sync.Mutex{}
	mu.Lock()
	defer mu.Unlock()
	var allResults []types.Result
	for res := range results {
		allResults = append(allResults, res)
	}
	for _, res := range allResults {
		err := utils.DownloadFileWithURLFilename(res.Id, res.URL, types.GlobalConfig.Storage.BakFileSavePath)
		if err != nil {
			fmt.Println("[ERROR] Download bakfile error: ", err.Error())
		}
		fmt.Printf("[ success ] %s  size:%s\n", res.URL, res.Size)
		if err := db.InsResult(res); err != nil {
			fmt.Println("[Error] Failed to insert result:", err)
			return
		}
	}
}
func isBinaryContentType(ct string) bool {
	ct = strings.ToLower(ct)
	textTypes := []string{"text/", "html", "xml", "json", "javascript", "application/json", "application/xml"}
	for _, t := range textTypes {
		if strings.Contains(ct, t) {
			return false
		}
	}
	return true
}
