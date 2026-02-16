package handlers

import (
	"BackupHunter/db"
	"BackupHunter/types"
	"BackupHunter/utils"
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var findTaskRuning = false // 用于判断当前任务执行之前是否还有任务在执行中

// 调用Oneforall寻找子域名
func FindSubdomainTask(task types.TableTasks) {
	targets := db.GetSubDomain(task.Domain)
	if len(targets) != 0 {
		fmt.Println("[INFO] Subdomain already exists, skip find subdomain task")
		return // 如果已经存在子域名，则不进行子域名发现
	}
	findTaskRuning = true
	db.SetTaskState(task.Id, 5)
	fmt.Printf("[INFO] Create new find subdomain task [%s]\n", time.Now().Format("2006-01-02 15:04:05"))
	cmd := exec.Command("python", "oneforall.py", "--target", task.Domain, "--dns=114.114.114.114", "run")
	cmd.Dir = "./OneForAll"
	//cmd.Stdout = os.Stdout
	//cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("[INFO] Oneforall run error wite command python: %v\n", err)
		fmt.Println("[MESSAGE] Try to run oneforall again with command python3")
		cmd := exec.Command("python3", "oneforall.py", "--target", task.Domain, "run")
		cmd.Dir = "./oneforall"
		err = cmd.Run()
		if err != nil {
			fmt.Printf("[ERROR] Oneforall run error: %v\n", err)
		}
	}
	onFindComplete(task) // 回调函数
	fmt.Printf("[INFO] Find subdomain task completed [%s]\n", time.Now().Format("2006-01-02 15:04:05"))
}

// 手动触发子域名发现任务
func FindSubdomain(c *gin.Context) {
	if findTaskRuning {
		c.JSON(200, gin.H{
			"state":   "error",
			"message": "已有一个任务正在进行中...",
		})
		return
	}
	taskid := c.PostForm("taskid")
	task, err := db.GetTaskById(taskid)
	if err != nil {
		fmt.Printf("[ERROR] Get task by id error: %v\n", err)
		c.JSON(200, gin.H{
			"state":   "error",
			"message": "Server error",
		})
		return
	}
	go FindSubdomainTask(task) // 调用Oneforall寻找子域名
	c.JSON(200, gin.H{
		"state": "success",
	})
}

// 手动添加新子域名
func AddSubDomain(c *gin.Context) {
	domain := c.PostForm("domain")
	subdomain := c.PostForm("subdomain")
	err := db.AddSubDomain(subdomain, domain)
	if err != nil {
		c.JSON(200, gin.H{
			"state": "error",
			"msg":   "Server error",
		})
		return
	}
	c.JSON(200, gin.H{
		"state": "success",
	})
}

// 手动删除单个子域名
func DeleteSingleSubDomain(c *gin.Context) {
	domain := c.PostForm("domain")
	subdomain := c.PostForm("subdomain")
	err := db.DeleteSingleSubDomain(subdomain, domain)
	if err != nil {
		c.JSON(200, gin.H{
			"state": "error",
			"msg":   "Server error",
		})
		return
	}
	c.JSON(200, gin.H{
		"state": "success",
	})
}

// 获取子域名
func GetSubDomain(c *gin.Context) {
	domain := c.Query("domain")
	subdomains := db.GetSubDomain(domain)
	c.JSON(200, gin.H{
		"state": "success",
		"data":  subdomains,
	})
}

// 获取所有子域名
func GetAllSubDomain(c *gin.Context) {
	subdomains := db.GetAllSubDomain()
	if subdomains == nil {
		c.JSON(200, gin.H{
			"state": "error",
			"msg":   "Server error",
		})
		return
	}
	c.JSON(200, gin.H{
		"state": "success",
		"data":  subdomains,
	})
}

// 从文件获取对应域名的Oneforall结果文件
func getSubDomainfromCsv(domain string) []string {
	if domain == "" {
		return nil
	}
	csvFile := "./oneforall/results/" + domain + ".csv"
	file, err := os.Open(csvFile)
	if err != nil {
		fmt.Printf("[ERROR] Open oneforall result csvfile error: %v\n", err)
		return nil
	}
	defer file.Close()
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Printf("[ERROR] Read oneforall result csvfile error: %v\n", err)
		return nil
	}
	var subdomains []string
	for i, record := range records {
		if i == 0 {
			continue
		}
		if len(record) > 3 {
			url := strings.TrimSpace(record[4])
			if url != "" {
				subdomains = append(subdomains, url)
			}
		}
	}
	subdomains = utils.DeduplicateURLs(subdomains)
	return subdomains
}

// Oneforall寻找完成后的回调函数
func onFindComplete(task types.TableTasks) {
	findTaskRuning = false
	subdomains := getSubDomainfromCsv(task.Domain)
	fmt.Println("[Message] Subdomain find success,check url valid")
	subdomains = utils.CheckURLsConcurrent(subdomains)
	fmt.Printf("[Message] Valid success valid subdomain count: %d\n", len(subdomains))
	err := db.BatchInsertSubDomain(subdomains, task.Domain)
	if err != nil {
		fmt.Printf("[ERROR] Batch insert subdomain error: %v\n", err)
		return
	}
	err = db.SetTaskState(task.Id, 1)
	if err != nil {
		fmt.Printf("[ERROR] Set task state error: %v\n", err)
		return
	}
	fmt.Printf("[Success] Batch insert subdomain success\n")
}
