package main

import (
	"BackupHunter/config"
	"BackupHunter/cron"
	"BackupHunter/db"
	"BackupHunter/handlers"
	"BackupHunter/log"
	"BackupHunter/types"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}
	fmt.Println("[Success] Config load successfully!")

	// 启动日志
	if err := log.StartLoggingToFile(types.GlobalConfig.Storage.LogSavePath); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start logging: %v\n", err)
		return
	}
	defer log.StopLogging()
	gin.DefaultWriter = os.Stdout
	gin.DefaultErrorWriter = os.Stderr
	fmt.Println("[Success] Logging started!")

	gin.SetMode(gin.ReleaseMode) // 生产模式
	r := gin.Default()

	// 添加CORS中间件
	r.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 加载路由
	testGroup := r.Group("/test")
	{
		testGroup.GET("/ping", handlers.Ping) // 测试接口
	}
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/login", handlers.Login)                                                 // 登录
		authGroup.GET("/logout", handlers.Logout)                                                // 登出
		authGroup.GET("/checklogin", handlers.IsLogined)                                         // 检查登录
		authGroup.GET("/getuserinfo", handlers.AuthRequired(), handlers.GetUserInfo)             // 获取当前登录用户信息
		authGroup.POST("/changepassword", handlers.AuthRequired(), handlers.ChangePassword)      // 修改密码
		authGroup.GET("/getstatisticsinfo", handlers.AuthRequired(), handlers.GetStatisticsInfo) // 获取统计信息
	}
	subDomainGroup := r.Group("/subdomain", handlers.AuthRequired())
	{
		subDomainGroup.GET("/getsubdomain", handlers.GetSubDomain)                    // 获取子域名
		subDomainGroup.GET("/getallsubdomain", handlers.GetAllSubDomain)              // 获取所有子域名
		subDomainGroup.POST("/addsubdomain", handlers.AddSubDomain)                   // 手动添加单个子域名
		subDomainGroup.POST("/deletesinglesubdomain", handlers.DeleteSingleSubDomain) // 手动删除单个子域名
		subDomainGroup.POST("/findsubdomain", handlers.FindSubdomain)                 // 手动触发子域名发现任务
	}
	scanGroup := r.Group("/scan", handlers.AuthRequired())
	{
		scanGroup.POST("/scanbackupfiles", cron.CreateScanBackupFilesTask)          // 创建扫描任务
		scanGroup.POST("/removetask", cron.RemoveTask)                              // 删除扫描任务
		scanGroup.POST("/stoptask", cron.StopTask)                                  // 暂停任务
		scanGroup.POST("/starttask", cron.StartTask)                                // 启动任务
		scanGroup.GET("/getscanbackupfilesstate", handlers.GetScanBackupFilesState) // 获取扫描任务状态
		scanGroup.GET("/getalltasks", handlers.GetAllTasks)                         // 获取所有扫描任务
	}
	fileGroup := r.Group("/file", handlers.AuthRequired())
	{
		fileGroup.GET("/getbakfilelist", handlers.GetDownloadedBakFileList) // 获取已转存的备份文件列表
		fileGroup.GET("/downloadfile", handlers.DownloadFile)               // 根据文件ID下载文件
		fileGroup.GET("/getresultsbyid", handlers.GetResultsById)           // 根据ID获取扫描结果
		fileGroup.GET("/getallresults", handlers.GetAllResults)             // 获取所有扫描结果
		fileGroup.GET("/deleteresult", handlers.DeleteResultAndFile)        // 删除result及本地转储文件
	}
	sysGroup := r.Group("/sys", handlers.AuthRequired())
	{
		sysGroup.GET("/osinfo", handlers.GetSystemInfo) // 获取CPU、内存、磁盘使用情况
	}
	configGroup := r.Group("/config", handlers.AuthRequired())
	{
		configGroup.POST("/changescannerconfig", handlers.ChangeScannerConfig) // 修改扫描器配置
		configGroup.GET("/getscannerconfig", handlers.GetScannerConfig)        // 获取扫描器配置
	}
	routes := r.Routes()
	fmt.Println("Registered routes:")
	for _, route := range routes {
		fmt.Printf("%-6s\t%-20s->\t%s\n", route.Method, route.Path, route.Handler)
	}
	// 初始化数据库
	db.Init()
	defer db.Close()
	fmt.Println("[Success] Database connected successfully!")

	// 初始化定时任务调度器
	if err := cron.InitScheduler(); err != nil {
		fmt.Printf("[ERROR] Failed to initialize cron scheduler: %v\n", err)
	}
	defer cron.Stop()
	fmt.Println("[Success] Cron scheduler started!")

	go handlers.StartMetricsCollector() // 启动系统信息采集协程
	fmt.Println("[Success] Metrics collector started!")

	fmt.Println("[Success] Server is runing")
	// 启动服务
	r.Run(fmt.Sprintf(":%d", cfg.Server.Port))
}
