package handlers

import (
	"BackupHunter/config"
	"BackupHunter/db"
	"BackupHunter/types"
	"BackupHunter/utils"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	attempts = make(map[string]types.LoginAttempt) // Key: "ip"，记录对应ip的登录尝试次数
	sessions = make(map[string]types.SessionData)  // 登录session
	mutex    = &sync.RWMutex{}
)

const (
	maxAttempts    = 5  // 最大登录失败次数
	lockoutMinutes = 10 // 登录失败锁定时间（分钟）
	sessionId      = "session_id"
	sessionTimeout = 30 * time.Minute // session过期时间
)

// 登录
func Login(c *gin.Context) {
	if isLoginAllowed(c.RemoteIP()) {
		inputUserName := c.PostForm("username")
		inputPassWord := c.PostForm("password")
		hostname := c.PostForm("hostname")
		if inputUserName == types.GlobalConfig.Auth.UserName && inputPassWord == types.GlobalConfig.Auth.Password {
			if hostname == "" {
				c.JSON(200, gin.H{
					"state": "error",
					"msg":   "缺少参数",
				})
				return
			}
			session, _ := createSession(inputUserName)
			c.SetCookie(sessionId, session, int(sessionTimeout), "/", hostname, false, true)
			c.JSON(200, gin.H{
				"state": "success",
				"token": session, // 返回token
				"msg":   "登录成功",
			})
		} else {
			recordFailedAttempt(c.RemoteIP())
			c.JSON(200, gin.H{
				"state": "error",
				"msg":   "用户名或密码错误，剩余" + strconv.Itoa((maxAttempts - attempts[c.RemoteIP()].Count)) + "次机会",
			})
			return
		}
	} else {
		c.JSON(200, gin.H{
			"state": "error",
			"msg":   "登录失败次数过多，请稍后再试",
		})
		return
	}
}

// 登出
func Logout(c *gin.Context) {
	id, _ := c.Cookie(sessionId)
	hostname := c.Query("hostname")
	delete(sessions, id)
	c.SetCookie(sessionId, "", -1, "/", hostname, false, true)
	c.JSON(200, gin.H{
		"state": "success",
		"msg":   "登出成功",
	})
}

// 修改密码
func ChangePassword(c *gin.Context) {
	oldPassword := c.PostForm("oldpassword")
	newPassword := c.PostForm("newpassword")
	var isSuccess bool
	isSuccess, err := config.ChangePassword(oldPassword, newPassword)
	if err != nil {
		fmt.Println("[ERROR] Change password error: " + err.Error())
		c.JSON(200, gin.H{
			"state": "error",
			"msg":   "Server error",
		})
		return
	}
	if isSuccess {
		id, _ := c.Cookie(sessionId)
		delete(sessions, id)
		c.JSON(200, gin.H{
			"state": "success",
			"msg":   "修改密码成功",
		})
		return
	}
	c.JSON(200, gin.H{
		"state": "error",
		"msg":   "修改失败，原密码错误",
	})
}

// 返回统计信息：总任务数量、已完成任务次数、扫描到的URL、下载的文件数
func GetStatisticsInfo(c *gin.Context) {
	tasks, _ := db.GetAllTasks()
	results, _ := db.GetResults()
	Files, _ := GetBakFileList_()

	totalExec := 0
	totalUrls := len(results)
	totalTasks := len(tasks)
	totalFiles := len(Files)
	for _, task := range tasks {
		totalExec += task.Exec_count
	}

	monthlyChart := db.GetMonthlyScanStats(6)
	c.JSON(200, gin.H{
		"state": "success",
		"msg":   "获取统计信息成功",
		"data": gin.H{
			"total_tasks":      totalTasks,
			"total_executions": totalExec,
			"total_urls":       totalUrls,
			"total_files":      totalFiles,
			"chart":            monthlyChart, // 近6个月的任务数量
		},
	})
}

// 返回当前用户信息
func GetUserInfo(c *gin.Context) {
	cookie, _ := c.Cookie("session_id")
	sessionData := types.SessionData{
		Username: sessions[cookie].Username,
		Created:  sessions[cookie].Created,
		Expires:  sessions[cookie].Expires,
	}
	for id := range sessions {
		if cookie == id {
			c.JSON(200, gin.H{
				"state": "success",
				"msg":   "获取用户信息成功",
				"data":  sessionData,
			})
			return
		}
	}
}

// 检查是否登录
func IsLogined(c *gin.Context) {
	session, err := c.Cookie(sessionId)
	if err != nil {
		c.JSON(200, gin.H{
			"state": "error",
		})
		return
	}
	for id := range sessions {
		if session == id {
			c.JSON(200, gin.H{
				"state": "success",
			})
			return
		}
	}
	c.JSON(200, gin.H{
		"state": "error",
	})
}

// 鉴权
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now()
		session, err := c.Cookie(sessionId)
		if err != nil {
			c.AbortWithStatusJSON(200, gin.H{
				"state": "error",
				"msg":   "未登录",
			})
			return
		}

		mutex.Lock()
		defer mutex.Unlock()

		for id := range sessions {
			if now.After(sessions[id].Expires) {
				delete(sessions, id)
				continue
			}
			if session == id {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(200, gin.H{
			"state": "error",
			"msg":   "未登录",
		})
	}
}

// 创建新 Session
func createSession(username string) (string, error) {
	id, err := utils.GenerateSessionID()
	if err != nil {
		return "", err
	}

	now := time.Now()
	session := types.SessionData{
		Username: username,
		Created:  now,
		Expires:  now.Add(sessionTimeout),
	}

	mutex.Lock()
	sessions[id] = session
	mutex.Unlock()

	return id, nil
}

// 是否允许登录
func isLoginAllowed(ip string) bool {
	key := ip

	mutex.RLock()
	attempt, exists := attempts[key]
	mutex.RUnlock()

	if !exists { // 首次登录
		return true
	}
	if time.Since(attempt.LastFailed) > lockoutMinutes*time.Minute {
		mutex.Lock()
		delete(attempts, key)
		mutex.Unlock()
		return true
	}
	return attempt.Count < maxAttempts
}

// 登录失败记录
func recordFailedAttempt(ip string) {
	key := ip

	mutex.Lock()
	defer mutex.Unlock()

	if attempt, exists := attempts[key]; exists {
		attempt.Count++
		attempt.LastFailed = time.Now()
		attempts[key] = attempt
	} else {
		attempts[key] = types.LoginAttempt{
			Count:      1,
			LastFailed: time.Now(),
		}
	}
}
