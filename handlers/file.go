package handlers

import (
	"BackupHunter/db"
	"BackupHunter/types"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// 获取备份文件列表
func GetDownloadedBakFileList(c *gin.Context) {
	fileList, err := GetBakFileList_()
	if err != nil {
		c.JSON(500, gin.H{
			"state": "error",
			"msg":   "Server error",
		})
		return
	}
	c.JSON(200, gin.H{
		"state":    "success",
		"filelist": fileList,
	})
}

// 根据ID获取扫描结果
func GetResultsById(c *gin.Context) {
	id := c.Query("id")
	result, err := db.GetUrlsById(id)
	if err != nil {
		c.JSON(500, gin.H{
			"state": "error",
			"msg":   err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"state":  "success",
		"result": result,
	})
}

// 获取所有扫描结果
func GetAllResults(c *gin.Context) {
	result, err := db.GetResults()
	if err != nil {
		c.JSON(500, gin.H{
			"state": "error",
			"msg":   "Server error",
		})
		return
	}
	c.JSON(200, gin.H{
		"state":  "success",
		"result": result,
	})
}

// 获取备份文件列表
func GetBakFileList_() ([]string, error) {
	filePath := types.GlobalConfig.Storage.BakFileSavePath
	entries, err := os.ReadDir(filePath)
	if err != nil {
		fmt.Println("[Error] Failed to read backup file directory: " + err.Error())
		return nil, err
	}
	var fileList []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fileList = append(fileList, entry.Name())
	}
	return fileList, nil
}

// 下载文件
func DownloadFile(c *gin.Context) {
	id := c.Query("id")
	result, _ := db.GetUrlsById(id)
	fileName := id + "_" + result.FileName
	downloadFilePath := filepath.Join(types.GlobalConfig.Storage.BakFileSavePath, fileName)
	c.FileAttachment(downloadFilePath, result.FileName)
}

// 删除结果和转储文件
func DeleteResultAndFile(c *gin.Context) {
	id := c.Query("id")
	result, _ := db.GetUrlsById(id)
	fileName := id + "_" + result.FileName
	err := db.DeleteResultById(id)
	if err != nil {
		fmt.Println("[ERROR] Delete result error: ", err.Error())
		c.JSON(200, gin.H{
			"state": "error",
		})
		return
	}
	filePath := filepath.Join(types.GlobalConfig.Storage.BakFileSavePath, fileName)
	err = os.Remove(filePath)
	if err != nil {
		fmt.Println("[ERROR] Delete file error: ", err.Error())
		c.JSON(200, gin.H{
			"state": "success",
		})
		return
	}
	c.JSON(200, gin.H{
		"state": "success",
	})
}
