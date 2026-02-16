package config

import (
	"BackupHunter/types"
	"BackupHunter/utils"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// 载入配置文件
func LoadConfig() (*types.Config, error) {
	defaultPwd, err := utils.RandomPassword(16)
	if err != nil {
		return nil, err
	}

	// 默认扫描器后缀
	def_suffixes := []string{
		".zip", ".rar", ".tar.gz", ".tgz", ".tar.bz2", ".tar", ".jar", ".war",
		".7z", ".bak", ".sql", ".gz", ".sql.gz", ".tar.tgz", ".backup",
	}
	// 默认扫描器字典
	def_staticWords := []string{
		"0", "00", "000", "012", "1", "111", "123", "127.0.0.1", "2", "2010", "2011", "2012", "2013", "2014", "2015",
		"2016", "2017", "2018", "2019", "2020", "2021", "2022", "2023", "2024", "2025", "2026", "234", "3", "333", "4", "444",
		"5", "555", "6", "666", "7", "777", "8", "888", "9", "999", "a", "about", "admin", "app", "application",
		"archive", "asp", "aspx", "auth", "b", "back", "backup", "backups", "bak", "blog", "bbs", "beifen", "bin", "cache",
		"clients", "code", "com", "config", "core", "customers", "dat", "data", "database", "db", "download", "dump",
		"engine", "error_log", "extend", "files", "forum", "ftp", "home", "html", "img", "include", "index", "install",
		"joomla", "js", "jsp", "local", "login", "localhost", "master", "media", "members", "my", "mysql", "new", "old",
		"orders", "output", "package", "php", "public", "root", "runtime", "sales", "server", "shujuku", "site", "sjk",
		"sql", "store", "tar", "template", "test", "upload", "user", "users", "vb", "vendor", "wangzhan", "web", "website",
		"wordpress", "wp", "www", "wwwroot", "wz", "log", "数据库", "数据库备份", "网站", "网站备份",
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	// 先从环境变量中拿数据
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.BindEnv("server.db_host", "DB_HOST")
	viper.BindEnv("server.db_port", "DB_PORT")
	viper.BindEnv("server.db_username", "DB_USER")
	viper.BindEnv("server.db_password", "DB_PASS")
	viper.BindEnv("server.db_name", "DB_NAME")

	// 配置默认值
	viper.SetDefault("server.port", 5555)
	viper.SetDefault("server.db_username", "root")
	viper.SetDefault("server.db_password", "root")
	viper.SetDefault("server.db_host", "localhost")
	viper.SetDefault("server.db_port", 3306)
	viper.SetDefault("server.db_name", "backup_hunter")
	viper.SetDefault("scanner.scan_timeout", 10)
	viper.SetDefault("scanner.scan_rate_limit", 5)
	viper.SetDefault("scanner.scan_proxy", "")
	viper.SetDefault("scanner.suffixes", def_suffixes)
	viper.SetDefault("scanner.static_words", def_staticWords)
	viper.SetDefault("storage.bak_file_save_path", "./data/files/")
	viper.SetDefault("storage.log_save_path", "./data/logs/")
	viper.SetDefault("auth.username", "admin")
	viper.SetDefault("auth.password", defaultPwd)

	// 检查文件是否存在 自动生成配置文件
	if _, err := os.Stat("config.yaml"); os.IsNotExist(err) {
		if err := viper.SafeWriteConfig(); err != nil {
			return nil, err
		}
		intIp, err := utils.GetInternalIP()
		if err != nil {
			intIp = "127.0.0.1"
		}
		extIp, err := utils.GetExternalIP()
		if err != nil {
			extIp = "127.0.0.1"
		}
		fmt.Println("--- [Success] Init config success ---")
		fmt.Println("[+] Username: admin")
		fmt.Println("[+] Password: " + defaultPwd)
		fmt.Println("[+] Intranet Web: http://" + intIp + ":5555")
		fmt.Println("[+] External Web: http://" + extIp + ":5555")
		fmt.Println("[+] Now please modify the database configuration in config.ymal")
		os.Exit(0)
	}
	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	// 解析配置文件
	var cfg types.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	types.GlobalConfig = &cfg
	return &cfg, nil
}

// 修改当前密码并写入配置文件
func ChangePassword(oldPassword string, newPassword string) (bool, error) {
	if oldPassword != types.GlobalConfig.Auth.Password {
		fmt.Println(oldPassword + " | " + types.GlobalConfig.Auth.Password)
		return false, nil
	} else {
		types.GlobalConfig.Auth.Password = newPassword
		return true, SaveConfig()
	}
}

// 修改扫描器配置
func ChangeConfig(scan_proxy string, scan_rate_limit int, scan_timeout int, suffixes []string, staticWords []string) error {
	types.GlobalConfig.Scanner.ScanProxy = scan_proxy
	types.GlobalConfig.Scanner.ScanRateLimit = scan_rate_limit
	types.GlobalConfig.Scanner.ScanTimeout = scan_timeout
	types.GlobalConfig.Scanner.Suffixes = suffixes
	types.GlobalConfig.Scanner.StaticWords = staticWords
	return SaveConfig()
}

// 写入配置文件
func SaveConfig() error {
	cfg := types.GlobalConfig
	viper.Set("server.port", cfg.Server.Port)
	viper.Set("server.db_username", cfg.Server.DbUserName)
	viper.Set("server.db_password", cfg.Server.DbPassword)
	viper.Set("server.db_host", cfg.Server.DbHost)
	viper.Set("server.db_port", cfg.Server.DbPort)
	viper.Set("server.db_name", cfg.Server.DbName)

	viper.Set("scanner.scan_timeout", cfg.Scanner.ScanTimeout)
	viper.Set("scanner.scan_rate_limit", cfg.Scanner.ScanRateLimit)
	viper.Set("scanner.scan_proxy", cfg.Scanner.ScanProxy)
	viper.Set("scanner.suffixes", cfg.Scanner.Suffixes)
	viper.Set("scanner.static_words", cfg.Scanner.StaticWords)

	viper.Set("storage.bak_file_save_path", cfg.Storage.BakFileSavePath)
	viper.Set("storage.log_save_path", cfg.Storage.LogSavePath)

	viper.Set("auth.username", cfg.Auth.UserName)
	viper.Set("auth.password", cfg.Auth.Password)

	return viper.WriteConfig()
}
