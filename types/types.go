package types

import (
	"database/sql"
	"time"
)

var GlobalConfig *Config // 全局配置
var DB *sql.DB           // 全局数据库

// 配置文件结构体
type Config struct {
	Server struct {
		Port       int    `mapstructure:"port"`        // 服务器端口
		DbUserName string `mapstructure:"db_username"` // 数据库用户名
		DbPassword string `mapstructure:"db_password"` // 数据库密码
		DbHost     string `mapstructure:"db_host"`     // 数据库地址
		DbPort     int    `mapstructure:"db_port"`     // 数据库端口
		DbName     string `mapstructure:"db_name"`     // 数据库名称
	}
	Scanner struct {
		ScanTimeout   int      `mapstructure:"scan_timeout"`    // 扫描器超时时间
		ScanRateLimit int      `mapstructure:"scan_rate_limit"` // 扫描器并发限制
		ScanProxy     string   `mapstructure:"scan_proxy"`      // 扫描器代理
		Suffixes      []string `mapstructure:"suffixes"`        // 扫描器域名后缀
		StaticWords   []string `mapstructure:"static_words"`    // 扫描器字典
	}
	Storage struct {
		BakFileSavePath string `mapstructure:"bak_file_save_path"` // 扫描到的备份文件保存路径
		LogSavePath     string `mapstructure:"log_save_path"`      // 日志保存路径
	}
	Auth struct {
		UserName string `mapstructure:"username"` // 系统用户名
		Password string `mapstructure:"password"` // 系统密码
	}
}

// 计划任务结构体
type TableTasks struct {
	Id             string       // 任务ID
	Name           string       // 任务名
	Cron_expr      string       // cron表达式
	Domain         string       // 目标域名
	Active         bool         // 是否启用
	Exec_count     int          // 已执行次数
	State          int          // 任务状态 0:未扫描子域名 1:等待执行 2:正在执行 3:执行成功 4:执行失败 5:子域名发现
	Last_exec_time sql.NullTime // 最后执行时间
	Created_time   sql.NullTime // 创建时间
	Updated_time   sql.NullTime // 更新时间
}

// 登录尝试记录
type LoginAttempt struct {
	Count      int
	LastFailed time.Time
}

// Session
type SessionData struct {
	Username string
	Created  time.Time
	Expires  time.Time
}

// 扫描结果结构体
type Result struct {
	Id       string
	URL      string
	FileName string
	Size     string
}

// 统计信息结构体，主要用于前端页面的统计图信息显示
type MonthlyScanStat struct {
	YearMonth string `json:"year_month"` // e.g. "2026-01"
	Count     int    `json:"count"`
}

// 系统信息结构体，主要用于前端页面的系统资源使用情况显示
type SystemMetrics struct {
	CPU    int
	Memory int
	Disk   int
	Time   string
}
