package log

import (
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	originalStdout = os.Stdout
	originalStderr = os.Stderr
	logFile        *os.File
	mu             sync.Mutex
	initialized    bool
)

// StartLoggingToFile 启动透明日志记录
// 所有后续写入 os.Stdout / os.Stderr 的内容都会同时写入 YYYY-MM-DD.log
func StartLoggingToFile(logPath string) error {
	mu.Lock()
	defer mu.Unlock()

	if initialized {
		return nil
	}

	// 创建日志文件
	logFileName := filepath.Join(
		logPath,
		time.Now().Format("2006-01-02")+".log",
	)
	file, err := os.Create(logFileName)
	if err != nil {
		return err
	}
	logFile = file

	// 创建管道
	stdoutReader, stdoutWriter, err := os.Pipe()
	if err != nil {
		file.Close()
		return err
	}

	stderrReader, stderrWriter, err := os.Pipe()
	if err != nil {
		stdoutReader.Close()
		stdoutWriter.Close()
		file.Close()
		return err
	}

	// 替换全局标准流
	os.Stdout = stdoutWriter
	os.Stderr = stderrWriter

	// 启动转发 goroutine：从管道读取并同时写入原始终端 + 日志文件
	go forwardStream(stdoutReader, originalStdout, file)
	go forwardStream(stderrReader, originalStderr, file)

	initialized = true
	return nil
}

// forwardStream 将输入流复制到两个输出：终端和日志文件
func forwardStream(input io.Reader, terminalOut, fileOut io.Writer) {
	buf := make([]byte, 4096)
	for {
		n, err := input.Read(buf)
		if n > 0 {
			// 写入终端和日志文件
			terminalOut.Write(buf[:n])
			fileOut.Write(buf[:n])
		}
		if err != nil {
			break // 通常是 EOF（writer 已关闭）
		}
	}
}

// StopLogging 停止日志记录并恢复标准输出（可选）
func StopLogging() {
	mu.Lock()
	defer mu.Unlock()

	if !initialized {
		return
	}

	// 恢复原始 stdout/stderr
	os.Stdout = originalStdout
	os.Stderr = originalStderr

	if logFile != nil {
		logFile.Close()
		logFile = nil
	}
	initialized = false
}
