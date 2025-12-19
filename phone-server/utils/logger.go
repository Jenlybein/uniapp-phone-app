package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	logFile  *os.File
	logger   *log.Logger
	lastDate string // 用于跟踪当前日志文件的日期
)

// InitLogger 初始化日志系统
func InitLogger() {
	// 创建logs目录（如果不存在）
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Fatalf("创建日志目录失败: %v", err)
	}

	// 初始化日志文件
	if err := updateLogFile(); err != nil {
		log.Fatalf("初始化日志文件失败: %v", err)
	}

	// 创建日志记录器
	logger = log.New(logFile, "", log.Ldate|log.Ltime|log.Lshortfile)

	// 启动定时检查，每天更新日志文件
	go func() {
		ticker := time.NewTicker(1 * time.Hour) // 每小时检查一次
		defer ticker.Stop()

		for {
			<-ticker.C
			if err := updateLogFile(); err != nil {
				log.Printf("更新日志文件失败: %v", err)
			}
		}
	}()
}

// updateLogFile 更新日志文件（如果日期已变化）
func updateLogFile() error {
	// 获取当前日期
	currentDate := time.Now().Format("2006-01-02")

	// 如果日期没有变化，不需要更新
	if currentDate == lastDate && logFile != nil {
		return nil
	}

	// 关闭旧的日志文件
	if logFile != nil {
		if err := logFile.Close(); err != nil {
			return fmt.Errorf("关闭旧日志文件失败: %v", err)
		}
	}

	// 创建新的日志文件名
	logFileName := fmt.Sprintf("%s.log", currentDate)
	logFilePath := filepath.Join("logs", logFileName)

	// 打开新的日志文件（追加模式）
	newFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("打开新日志文件失败: %v", err)
	}

	// 更新全局变量
	logFile = newFile
	lastDate = currentDate

	// 如果logger已经初始化，更新其输出目标
	if logger != nil {
		logger.SetOutput(logFile)
	}

	return nil
}

// Infof 记录信息级别的日志
func Infof(format string, v ...interface{}) {
	logger.Printf("INFO: "+format, v...)
}

// Errorf 记录错误级别的日志
func Errorf(format string, v ...interface{}) {
	logger.Printf("ERROR: "+format, v...)
}

// Fatalf 记录致命级别的日志并退出程序
func Fatalf(format string, v ...interface{}) {
	logger.Fatalf("FATAL: "+format, v...)
}

// CloseLogger 关闭日志系统
func CloseLogger() {
	if logFile != nil {
		logFile.Close()
	}
}
