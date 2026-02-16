package handlers

import (
	"time"

	"github.com/gin-gonic/gin"
)

func Ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
		"ip":      c.ClientIP(),                             // 返回客户端IP
		"time":    time.Now().Format("2006-01-02 15:04:05"), // 返回当前时间
	})
}
