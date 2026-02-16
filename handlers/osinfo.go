package handlers

import (
	"BackupHunter/types"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

var (
	metrics     types.SystemMetrics
	metricsMu   sync.RWMutex
	collectDone = make(chan struct{})
)

// 后台采集协程
func StartMetricsCollector() {
	ticker := time.NewTicker(5 * time.Second) // 5s采集一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// CPU
			cpuPercents, _ := cpu.Percent(0, false)
			cpuUsage := float64(0)
			if len(cpuPercents) > 0 {
				cpuUsage = cpuPercents[0]
			}

			// 内存
			vmStat, _ := mem.VirtualMemory()
			memUsage := vmStat.UsedPercent

			// 磁盘
			diskPath := "/"
			if runtime.GOOS == "windows" {
				diskPath = "C:\\"
			}
			diskStat, _ := disk.Usage(diskPath)
			diskUsage := diskStat.UsedPercent

			// 更新缓存
			metricsMu.Lock()
			metrics = types.SystemMetrics{
				CPU:    int(cpuUsage),
				Memory: int(memUsage),
				Disk:   int(diskUsage),
				Time:   time.Now().Format("15:04:05"),
			}
			metricsMu.Unlock()

		case <-collectDone:
			return
		}
	}
}

// 获取系统信息接口
func GetSystemInfo(c *gin.Context) {
	metricsMu.RLock()
	info := metrics
	metricsMu.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"cpu":    info.CPU,
		"memory": info.Memory,
		"disk":   info.Disk,
		"time":   info.Time,
		"msg":    "success",
	})
}
