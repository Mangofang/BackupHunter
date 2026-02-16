package handlers

import (
	"BackupHunter/config"
	"BackupHunter/types"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 修改扫描器配置
func ChangeScannerConfig(c *gin.Context) {
	scan_proxy := c.PostForm("scan_proxy")
	scan_rate_limit, err := strconv.Atoi(c.PostForm("scan_rate_limit"))
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(200, gin.H{
			"state": "error",
			"msg":   "扫描速率限制必须为数字",
		})
		return
	}
	scan_timeout, err := strconv.Atoi(c.PostForm("scan_timeout"))
	if err != nil {
		c.JSON(200, gin.H{
			"state": "error",
			"msg":   "扫描timeout时间必须为数字",
		})
		return
	}
	suffixes := c.PostFormArray("suffixes")
	staticWords := c.PostFormArray("staticwords")
	fmt.Println(suffixes)
	fmt.Println(staticWords)
	err = config.ChangeConfig(scan_proxy, scan_rate_limit, scan_timeout, suffixes, staticWords)
	if err != nil {
		c.JSON(200, gin.H{
			"state": "error",
			"msg":   "配置修改失败",
		})
		return
	}
	c.JSON(200, gin.H{
		"state": "success",
		"msg":   "配置修改成功",
	})
}

// 获取扫描器配置
func GetScannerConfig(c *gin.Context) {
	scan_proxy := types.GlobalConfig.Scanner.ScanProxy
	scan_rate_limit := types.GlobalConfig.Scanner.ScanRateLimit
	scan_timeout := types.GlobalConfig.Scanner.ScanTimeout
	suffixes := types.GlobalConfig.Scanner.Suffixes
	staticWords := types.GlobalConfig.Scanner.StaticWords
	c.JSON(200, gin.H{
		"state":           "success",
		"scan_proxy":      scan_proxy,
		"scan_rate_limit": scan_rate_limit,
		"scan_timeout":    scan_timeout,
		"suffixes":        suffixes,
		"staticWords":     staticWords,
	})
}
