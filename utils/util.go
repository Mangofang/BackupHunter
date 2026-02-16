package utils

import (
	"BackupHunter/types"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"
)

// 生成不小于16位的随机密码
func RandomPassword(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	if length < 16 {
		return "", fmt.Errorf("password length must be greater than 0")
	}
	result := make([]byte, length)
	randomBytes := make([]byte, length+(length/2))
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	for i := 0; i < length; i++ {
		result[i] = charset[randomBytes[i]%byte(len(charset))]
	}
	return string(result), nil
}

// 生成唯一 Session ID
func GenerateSessionID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// URL去重
func DeduplicateURLs(urls []string) []string {
	hostMap := make(map[string]string)

	for _, raw := range urls {
		uStr := strings.TrimSpace(raw)
		if uStr == "" {
			continue
		}

		u, err := url.Parse(uStr)
		if err != nil {
			continue
		}

		host := u.Host
		scheme := strings.ToLower(u.Scheme)
		if scheme != "http" && scheme != "https" {
			continue
		}
		existing, exists := hostMap[host]
		if !exists {
			hostMap[host] = uStr
		} else {
			existingURL, _ := url.Parse(existing)
			if existingURL.Scheme != "https" && scheme == "https" {
				hostMap[host] = uStr
			}
		}
	}
	result := make([]string, 0, len(hostMap))
	for _, v := range hostMap {
		result = append(result, v)
	}

	sort.Strings(result)

	return result
}

// 根据目标域名生成备份文件字典
func GenerateDict(domain string) []string {
	cleanHost := domain
	if strings.HasPrefix(cleanHost, "http://") {
		cleanHost = cleanHost[7:]
	} else if strings.HasPrefix(cleanHost, "https://") {
		cleanHost = cleanHost[8:]
	}
	if i := strings.Index(cleanHost, "/"); i != -1 {
		cleanHost = cleanHost[:i]
	}
	if i := strings.Index(cleanHost, ":"); i != -1 {
		cleanHost = cleanHost[:i]
	}

	parts := strings.Split(cleanHost, ".")
	var wwwhost string
	for i := 1; i < len(parts); i++ {
		wwwhost += parts[i]
	}

	domainDic := []string{
		cleanHost,
		strings.ReplaceAll(cleanHost, ".", ""),
		strings.ReplaceAll(cleanHost, ".", "_"),
		wwwhost,
		strings.Join(parts[1:], "."),
		strings.ReplaceAll(strings.Join(parts[1:], "."), ".", "_"),
	}
	if len(parts) > 0 {
		domainDic = append(domainDic, parts[0])
	}
	if len(parts) > 1 {
		domainDic = append(domainDic, parts[1])
	}
	seen := make(map[string]bool)
	uniqueDomains := []string{}
	for _, d := range domainDic {
		if !seen[d] && d != "" {
			seen[d] = true
			uniqueDomains = append(uniqueDomains, d)
		}
	}
	var suffixes []string
	var staticWords []string
	if len(types.GlobalConfig.Scanner.Suffixes) != 0 {
		suffixes = types.GlobalConfig.Scanner.Suffixes
	} else {
		suffixes = []string{
			".zip", ".rar", ".tar.gz", ".tgz", ".tar.bz2", ".tar", ".jar", ".war",
			".7z", ".bak", ".sql", ".gz", ".sql.gz", ".tar.tgz", ".backup", "",
		}
	}
	if len(types.GlobalConfig.Scanner.StaticWords) != 0 {
		staticWords = types.GlobalConfig.Scanner.StaticWords
	} else {
		staticWords = []string{
			"0", "00", "000", "012", "1", "111", "123", "127.0.0.1", "2", "2010", "2011", "2012", "2013", "2014", "2015",
			"2016", "2017", "2018", "2019", "2020", "2021", "2022", "2023", "2024", "2025", "234", "3", "333", "4", "444",
			"5", "555", "6", "666", "7", "777", "8", "888", "9", "999", "a", "about", "admin", "app", "application",
			"archive", "asp", "aspx", "auth", "b", "back", "backup", "backups", "bak", "bbs", "beifen", "bin", "cache",
			"clients", "code", "com", "config", "core", "customers", "dat", "data", "database", "db", "download", "dump",
			"engine", "error_log", "extend", "files", "forum", "ftp", "home", "html", "img", "include", "index", "install",
			"joomla", "js", "jsp", "local", "login", "localhost", "master", "media", "members", "my", "mysql", "new", "old",
			"orders", "output", "package", "php", "public", "root", "runtime", "sales", "server", "shujuku", "site", "sjk",
			"sql", "store", "tar", "template", "test", "upload", "user", "users", "vb", "vendor", "wangzhan", "web", "website",
			"wordpress", "wp", "www", "wwwroot", "wz", "log", "数据库", "数据库备份", "网站", "网站备份",
		}
	}
	dictMap := make(map[string]bool)
	for _, word := range staticWords {
		for _, suf := range suffixes {
			dictMap[word+suf] = true
		}
	}
	for _, domain := range uniqueDomains {
		for _, suf := range suffixes {
			dictMap[domain+suf] = true
		}
	}
	result := make([]string, 0, len(dictMap))
	for k := range dictMap {
		result = append(result, k)
	}
	return result
}
func DownloadFileWithURLFilename(id, url, dir string) error {
	fmt.Println("[Message] Find bakfile, try download it")

	parsedURL := strings.Split(url, "?")[0]
	_, filename := filepath.Split(parsedURL)

	if filename == "" || filename == "/" {
		filename = "downloaded_file"
	}
	filename = SanitizeFilename(filename)
	if filename == "" {
		filename = "unnamed_file"
	}
	filename = id + "_" + filename
	savePath := filepath.Join(dir, filename)

	// 创建目录
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("[ERROR] failed to create directory %s: %w", dir, err)
	}

	// 使用带智能超时的 client（不限制整个传输时间）
	client := &http.Client{
		Transport: &http.Transport{
			// 连接建立超时
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			// 响应头超时（防止服务器挂起）
			ResponseHeaderTimeout: 15 * time.Second,
			// 空闲连接超时
			IdleConnTimeout: 90 * time.Second,
			// 最大空闲连接数
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
		},
	}

	// 重试（最多3次）
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			fmt.Printf("[INFO] Attempt %d after error: %v\n", attempt+1, lastErr)
			time.Sleep(time.Duration(2<<attempt) * time.Second) // 指数退避
		}

		// 下载并验证完整性
		err := downloadWithValidation(client, url, savePath)
		if err == nil {
			fmt.Printf("[Success] Downloaded to %s\n", savePath)
			return nil
		}

		// 判断是否可重试
		if !isRetryableError(err) {
			return fmt.Errorf("[ERROR] non-retryable error: %w", err)
		}
		lastErr = err
	}

	return fmt.Errorf("[ERROR] download failed after 3 attempts: %w", lastErr)
}

// 执行单次下载并验证完整性
func downloadWithValidation(client *http.Client, url, savePath string) error {
	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("[ERROR] server returned %d for %s", resp.StatusCode, url)
	}

	// 获取预期大小（用于完整性校验）
	var expectedSize int64 = -1
	if resp.ContentLength > 0 {
		expectedSize = resp.ContentLength
		fmt.Printf("[Info] Expected file size: %d bytes\n", expectedSize)
	}

	// 创建临时文件（避免写入不完整文件）
	tmpPath := savePath + ".tmp"
	out, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	defer func() {
		out.Close()
		if err != nil {
			os.Remove(tmpPath)
		}
	}()

	// 分块复制（便于监控和中断恢复）
	buf := make([]byte, 1024*1024) // 1MB buffer
	written, err := io.CopyBuffer(out, resp.Body, buf)
	if err != nil {
		return fmt.Errorf("[ERROR] io.Copy failed: %w", err)
	}

	// 强制刷盘
	if err := out.Sync(); err != nil {
		return fmt.Errorf("[ERROR] fsync failed: %w", err)
	}
	if err := out.Close(); err != nil {
		return fmt.Errorf("[ERROR] file close failed: %w", err)
	}

	// 完整性校验
	if expectedSize > 0 && written != expectedSize {
		return fmt.Errorf("[ERROR] incomplete download: got %d bytes, expected %d", written, expectedSize)
	}

	// 重命名为最终文件
	if err := os.Rename(tmpPath, savePath); err != nil {
		return fmt.Errorf("[ERROR]rename failed: %w", err)
	}

	return nil
}

// 判断错误是否可重试
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// 网络相关错误
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return true
	}

	// unexpected EOF 属于可重试的临时错误
	if strings.Contains(err.Error(), "unexpected EOF") ||
		strings.Contains(err.Error(), "connection reset") ||
		strings.Contains(err.Error(), "broken pipe") ||
		strings.Contains(err.Error(), "connection closed") {
		return true
	}

	// syscall 错误
	if syscallErr, ok := err.(*os.SyscallError); ok {
		return syscallErr.Err == syscall.ECONNRESET ||
			syscallErr.Err == syscall.ECONNREFUSED ||
			syscallErr.Err == syscall.ETIMEDOUT
	}

	return false
}

// 安全化文件名
func SanitizeFilename(name string) string {
	// 移除路径分隔符、控制字符等
	re := regexp.MustCompile(`[\\/:*?"<>|\x00-\x1f]`)
	name = re.ReplaceAllString(name, "_")

	// 去除首尾空格和点（避免 . 或 ..）
	name = strings.Trim(name, " .")

	// 限制长度（可选）
	if len(name) > 255 {
		name = name[:255]
	}

	return name
}

// 检查url有效性
func CheckURLsConcurrent(urls []string) []string {
	var validURLs []string
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 创建带缓冲的channel来控制并发数
	sem := make(chan struct{}, 10) // 最多10个并发

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 处理每个URL
	for _, url := range urls {
		wg.Add(1)
		sem <- struct{}{} // 获取信号量

		go func(u string) {
			defer wg.Done()
			defer func() { <-sem }() // 释放信号量

			if isValidURL(client, u) {
				mu.Lock()
				validURLs = append(validURLs, u)
				mu.Unlock()
			}
		}(url)
	}

	wg.Wait()
	return validURLs
}
func isValidURL(client *http.Client, url string) bool {
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return true
}

// 获取本机内网IP地址
func GetInternalIP() (string, error) {
	// 获取所有网络接口
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	// 遍历网络接口，找到第一个内网IPv4地址
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ip, _, err := net.ParseCIDR(addr.String())
			if err == nil && ip.To4() != nil && isPrivateIP(ip) {
				return ip.String(), nil
			}
		}
	}

	return "", fmt.Errorf("No internal IPv4 address found")
}

// 判断IP地址是否为内网地址
func isPrivateIP(ip net.IP) bool {
	privateIPBlocks := []*net.IPNet{
		{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(8, 32)},
		{IP: net.ParseIP("172.16.0.0"), Mask: net.CIDRMask(12, 32)},
		{IP: net.ParseIP("192.168.0.0"), Mask: net.CIDRMask(16, 32)},
	}

	for _, block := range privateIPBlocks {
		if block.Contains(ip) {
			return true
		}
	}

	return false
}

// 获取外网ip
func GetExternalIP() (string, error) {
	var ipServices = []string{
		"https://api.ipify.org",
		"https://icanhazip.com",
		"https://ident.me",
		"https://ipecho.net/plain",
	}
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
		},
	}

	for _, url := range ipServices {
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36 Edg/144.0.0.0")

		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			ip := strings.TrimSpace(string(body))
			if net.ParseIP(ip) != nil {
				return ip, nil
			}
		}
	}
	return "", fmt.Errorf("all IP services failed")
}
