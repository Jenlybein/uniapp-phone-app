package utils

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// 日志级别定义
const (
	DEBUG = iota
	INFO
	WARN
	ERROR
	FATAL
)

// 日志级别名称映射
var levelNames = map[int]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

// 日志类型常量
const (
	LoggerTypeServer   = "server"   // 服务器日志
	LoggerTypeDatabase = "database" // 数据库日志
)

// 日志配置结构体
type LoggerConfig struct {
	Level    int    // 日志级别
	FilePath string // 日志文件路径
	MaxSize  int    // 日志文件最大大小（MB）
	MaxDays  int    // 日志文件保留天数
	Compress bool   // 是否压缩日志文件
	Console  bool   // 是否输出到控制台
	Type     string // 日志类型: server/database
}

// 日志条目结构体
type logEntry struct {
	level     int
	time      time.Time
	file      string
	line      int
	coroutine uint64
	requestID string
	message   string
}

// Logger 日志器结构体
type Logger struct {
	config       LoggerConfig
	console      *log.Logger
	file         *log.Logger
	logFile      *os.File
	lastDate     string
	logChan      chan logEntry
	wg           sync.WaitGroup
	ctx          context.Context
	cancel       context.CancelFunc
	requestIDKey string
}

var (
	defaultLogger  *Logger
	databaseLogger *Logger
	mu             sync.Mutex
)

// init 初始化默认日志器
func init() {
	// 创建默认配置
	defaultConfig := LoggerConfig{
		Level:    INFO,
		FilePath: "logs/",
		MaxSize:  100,
		MaxDays:  7,
		Compress: false,
		Console:  true,
		Type:     LoggerTypeServer,
	}

	// 初始化默认日志器
	logger, err := NewLogger(defaultConfig)
	if err != nil {
		log.Fatalf("初始化默认日志器失败: %v", err)
	}
	defaultLogger = logger
}

// NewLogger 创建新的日志器
func NewLogger(config LoggerConfig) (*Logger, error) {
	// 确保日志目录存在
	if err := os.MkdirAll(config.FilePath, 0755); err != nil {
		return nil, fmt.Errorf("创建日志目录失败: %v", err)
	}

	// 创建上下文用于控制协程
	ctx, cancel := context.WithCancel(context.Background())

	// 创建日志器实例
	logger := &Logger{
		config:       config,
		logChan:      make(chan logEntry, 1000), // 缓冲区大小为1000
		ctx:          ctx,
		cancel:       cancel,
		requestIDKey: "request_id",
	}

	// 初始化控制台日志器
	if config.Console {
		logger.console = log.New(os.Stdout, "", 0)
	}

	// 初始化文件日志器
	if err := logger.updateLogFile(); err != nil {
		return nil, fmt.Errorf("初始化日志文件失败: %v", err)
	}

	// 启动异步写入协程
	logger.wg.Add(1)
	go logger.writeLoop()

	// 启动日志文件检查协程
	logger.wg.Add(1)
	go logger.checkLogFileLoop()

	return logger, nil
}

// updateLogFile 更新日志文件（如果日期已变化或文件大小超过限制）
func (l *Logger) updateLogFile() error {
	currentDate := time.Now().Format("2006-01-02")

	// 检查文件大小
	needUpdate := false
	if l.logFile != nil {
		if currentDate != l.lastDate {
			needUpdate = true
		} else {
			// 检查文件大小
			info, err := l.logFile.Stat()
			if err != nil {
				return fmt.Errorf("获取日志文件信息失败: %v", err)
			}
			if info.Size() > int64(l.config.MaxSize*1024*1024) {
				needUpdate = true
			}
		}
	} else {
		needUpdate = true
	}

	if !needUpdate {
		return nil
	}

	// 关闭旧的日志文件
	if l.logFile != nil {
		if err := l.logFile.Close(); err != nil {
			return fmt.Errorf("关闭旧日志文件失败: %v", err)
		}
	}

	// 创建新的日志文件名
	logFileName := fmt.Sprintf("%s-%s.log", currentDate, l.config.Type)
	logFilePath := filepath.Join(l.config.FilePath, logFileName)

	// 打开新的日志文件（追加模式）
	newFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("打开新日志文件失败: %v", err)
	}

	// 更新日志器和全局变量
	l.logFile = newFile
	l.lastDate = currentDate
	l.file = log.New(newFile, "", 0)

	return nil
}

// writeLoop 异步写入日志循环
func (l *Logger) writeLoop() {
	defer l.wg.Done()

	for {
		select {
		case entry, ok := <-l.logChan:
			if !ok {
				return
			}

			// 格式化日志
			logStr := l.formatLog(entry)

			// 写入控制台
			if l.config.Console {
				l.console.Print(logStr)
			}

			// 写入文件
			if l.file != nil {
				l.file.Print(logStr)
			}

			// 如果是FATAL级别，记录后退出程序
			if entry.level == FATAL {
				os.Exit(1)
			}

		case <-l.ctx.Done():
			return
		}
	}
}

// checkLogFileLoop 定期检查日志文件
func (l *Logger) checkLogFileLoop() {
	defer l.wg.Done()

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := l.updateLogFile(); err != nil {
				// 使用标准日志记录错误，避免死锁
				log.Printf("更新日志文件失败: %v", err)
			}

			// 清理旧日志文件
			l.cleanOldLogFiles()

		case <-l.ctx.Done():
			return
		}
	}
}

// cleanOldLogFiles 清理旧日志文件
func (l *Logger) cleanOldLogFiles() {
	// 获取当前时间
	now := time.Now()

	// 打开日志目录
	dir, err := os.Open(l.config.FilePath)
	if err != nil {
		return
	}
	defer dir.Close()

	// 读取目录内容
	files, err := dir.Readdir(-1)
	if err != nil {
		return
	}

	// 遍历文件
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// 检查文件是否为日志文件
		if !strings.HasSuffix(file.Name(), ".log") {
			continue
		}

		// 检查文件修改时间
		if now.Sub(file.ModTime()) > time.Duration(l.config.MaxDays)*24*time.Hour {
			// 删除旧日志文件
			os.Remove(filepath.Join(l.config.FilePath, file.Name()))
		}
	}
}

// formatLog 格式化日志条目
func (l *Logger) formatLog(entry logEntry) string {
	// 格式化时间戳（精确到毫秒）
	timestamp := entry.time.Format("2006-01-02 15:04:05.000")

	// 获取日志级别名称
	levelName := levelNames[entry.level]

	// 构建日志字符串
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("[%s] [%s] [%s:%d] [goroutine:%d]",
		timestamp, levelName, entry.file, entry.line, entry.coroutine))

	// 如果有请求ID，添加到日志中
	if entry.requestID != "" {
		buf.WriteString(fmt.Sprintf(" [request_id:%s]", entry.requestID))
	}

	// 添加日志消息
	buf.WriteString(fmt.Sprintf(" %s\n", entry.message))

	return buf.String()
}

// getCallerInfo 获取调用者信息
func getCallerInfo() (file string, line int) {
	// 获取调用栈
	pc, file, line, ok := runtime.Caller(3) // 跳过当前函数和日志函数
	if !ok {
		return "unknown", 0
	}

	// 获取函数名
	funcName := runtime.FuncForPC(pc).Name()

	// 提取文件名（仅保留最后一部分）
	file = filepath.Base(file)

	// 提取函数名（仅保留最后一部分）
	funcNameParts := strings.Split(funcName, ".")
	if len(funcNameParts) > 0 {
		funcName = funcNameParts[len(funcNameParts)-1]
	}

	return fmt.Sprintf("%s.%s", file, funcName), line
}

// getCoroutineID 获取当前协程ID
func getCoroutineID() uint64 {
	b := make([]byte, 64)
	runtime.Stack(b, false)
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	id, _ := strconv.ParseUint(string(b), 10, 64)
	return id
}

// getRequestID 从上下文获取请求ID
func getRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	requestID, ok := ctx.Value("request_id").(string)
	if !ok {
		return ""
	}

	return requestID
}

// log 记录日志的通用方法
func (l *Logger) log(ctx context.Context, level int, format string, v ...interface{}) {
	// 检查日志级别
	if level < l.config.Level {
		return
	}

	// 如果上下文为nil，使用context.TODO()
	if ctx == nil {
		ctx = context.TODO()
	}

	// 获取调用者信息
	file, line := getCallerInfo()

	// 创建日志条目
	entry := logEntry{
		level:     level,
		time:      time.Now(),
		file:      file,
		line:      line,
		coroutine: getCoroutineID(),
		requestID: getRequestID(ctx),
		message:   fmt.Sprintf(format, v...),
	}

	// 发送到日志通道
	select {
	case l.logChan <- entry:
		// 成功发送到通道
	default:
		// 通道已满，使用标准日志记录，避免阻塞
		logStr := l.formatLog(entry)
		os.Stderr.WriteString(logStr)
	}
}

// 设置全局日志器配置
func SetLoggerConfig(config LoggerConfig) error {
	mu.Lock()
	defer mu.Unlock()

	// 根据日志类型初始化或更新相应的日志器
	if config.Type == LoggerTypeDatabase {
		// 关闭旧的数据库日志器
		if databaseLogger != nil {
			databaseLogger.cancel()
			databaseLogger.wg.Wait()
			if databaseLogger.logFile != nil {
				databaseLogger.logFile.Close()
			}
		}

		// 创建新的数据库日志器
		logger, err := NewLogger(config)
		if err != nil {
			return err
		}

		databaseLogger = logger
	} else {
		// 关闭旧的服务器日志器
		if defaultLogger != nil {
			defaultLogger.cancel()
			defaultLogger.wg.Wait()
			if defaultLogger.logFile != nil {
				defaultLogger.logFile.Close()
			}
		}

		// 创建新的服务器日志器
		if config.Type == "" {
			config.Type = LoggerTypeServer
		}
		logger, err := NewLogger(config)
		if err != nil {
			return err
		}

		defaultLogger = logger
	}

	return nil
}

// InitDatabaseLogger 初始化数据库日志器
func InitDatabaseLogger(config LoggerConfig) error {
	config.Type = LoggerTypeDatabase
	return SetLoggerConfig(config)
}

// GetLevelFromString 从字符串获取日志级别（公开函数）
func GetLevelFromString(levelStr string) int {
	levelStr = strings.ToUpper(levelStr)
	switch levelStr {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	case "FATAL":
		return FATAL
	default:
		return INFO // 默认INFO级别
	}
}

// InitLogger 初始化日志系统（从配置文件）
func InitLogger() {
	// 这个函数保持兼容，实际初始化在init函数中完成
}

// InitLoggerWithConfig 带配置初始化日志系统
func InitLoggerWithConfig(level string, filePath string, maxSize int, maxDays int, compress bool, console bool) {
	config := LoggerConfig{
		Level:    GetLevelFromString(level),
		FilePath: filePath,
		MaxSize:  maxSize,
		MaxDays:  maxDays,
		Compress: compress,
		Console:  console,
		Type:     LoggerTypeServer,
	}
	if err := SetLoggerConfig(config); err != nil {
		log.Fatalf("初始化日志系统失败: %v", err)
	}
}

// CloseLogger 关闭日志系统
func CloseLogger() {
	mu.Lock()
	defer mu.Unlock()

	// 关闭服务器日志器
	if defaultLogger != nil {
		defaultLogger.cancel()
		defaultLogger.wg.Wait()
		if defaultLogger.logFile != nil {
			defaultLogger.logFile.Close()
		}
	}

	// 关闭数据库日志器
	if databaseLogger != nil {
		databaseLogger.cancel()
		databaseLogger.wg.Wait()
		if databaseLogger.logFile != nil {
			databaseLogger.logFile.Close()
		}
	}
}

// Debugf 记录DEBUG级别的日志
func Debugf(format string, v ...interface{}) {
	defaultLogger.log(context.TODO(), DEBUG, format, v...)
}

// Infof 记录INFO级别的日志
func Infof(format string, v ...interface{}) {
	defaultLogger.log(context.TODO(), INFO, format, v...)
}

// Warnf 记录WARN级别的日志
func Warnf(format string, v ...interface{}) {
	defaultLogger.log(context.TODO(), WARN, format, v...)
}

// Errorf 记录ERROR级别的日志
func Errorf(format string, v ...interface{}) {
	defaultLogger.log(context.TODO(), ERROR, format, v...)
}

// Fatalf 记录FATAL级别的日志并退出程序
func Fatalf(format string, v ...interface{}) {
	defaultLogger.log(context.TODO(), FATAL, format, v...)
}

// Debugfc 带上下文记录DEBUG级别的日志
func Debugfc(ctx context.Context, format string, v ...interface{}) {
	defaultLogger.log(ctx, DEBUG, format, v...)
}

// Infofc 带上下文记录INFO级别的日志
func Infofc(ctx context.Context, format string, v ...interface{}) {
	defaultLogger.log(ctx, INFO, format, v...)
}

// Warnfc 带上下文记录WARN级别的日志
func Warnfc(ctx context.Context, format string, v ...interface{}) {
	defaultLogger.log(ctx, WARN, format, v...)
}

// Errorfc 带上下文记录ERROR级别的日志
func Errorfc(ctx context.Context, format string, v ...interface{}) {
	defaultLogger.log(ctx, ERROR, format, v...)
}

// Fatalfc 带上下文记录FATAL级别的日志并退出程序
func Fatalfc(ctx context.Context, format string, v ...interface{}) {
	defaultLogger.log(ctx, FATAL, format, v...)
}

// ------------------------ 数据库日志方法 ------------------------

// DatabaseDebugf 记录数据库DEBUG级别的日志
func DatabaseDebugf(format string, v ...interface{}) {
	if databaseLogger != nil {
		databaseLogger.log(context.TODO(), DEBUG, format, v...)
	} else {
		// 降级使用默认日志器
		defaultLogger.log(context.TODO(), DEBUG, format, v...)
	}
}

// DatabaseInfof 记录数据库INFO级别的日志
func DatabaseInfof(format string, v ...interface{}) {
	if databaseLogger != nil {
		databaseLogger.log(context.TODO(), INFO, format, v...)
	} else {
		// 降级使用默认日志器
		defaultLogger.log(context.TODO(), INFO, format, v...)
	}
}

// DatabaseWarnf 记录数据库WARN级别的日志
func DatabaseWarnf(format string, v ...interface{}) {
	if databaseLogger != nil {
		databaseLogger.log(context.TODO(), WARN, format, v...)
	} else {
		// 降级使用默认日志器
		defaultLogger.log(context.TODO(), WARN, format, v...)
	}
}

// DatabaseErrorf 记录数据库ERROR级别的日志
func DatabaseErrorf(format string, v ...interface{}) {
	if databaseLogger != nil {
		databaseLogger.log(context.TODO(), ERROR, format, v...)
	} else {
		// 降级使用默认日志器
		defaultLogger.log(context.TODO(), ERROR, format, v...)
	}
}

// DatabaseFatalf 记录数据库FATAL级别的日志并退出程序
func DatabaseFatalf(format string, v ...interface{}) {
	if databaseLogger != nil {
		databaseLogger.log(context.TODO(), FATAL, format, v...)
	} else {
		// 降级使用默认日志器
		defaultLogger.log(context.TODO(), FATAL, format, v...)
	}
}

// DatabaseDebugfc 带上下文记录数据库DEBUG级别的日志
func DatabaseDebugfc(ctx context.Context, format string, v ...interface{}) {
	if databaseLogger != nil {
		databaseLogger.log(ctx, DEBUG, format, v...)
	} else {
		// 降级使用默认日志器
		defaultLogger.log(ctx, DEBUG, format, v...)
	}
}

// DatabaseInfofc 带上下文记录数据库INFO级别的日志
func DatabaseInfofc(ctx context.Context, format string, v ...interface{}) {
	if databaseLogger != nil {
		databaseLogger.log(ctx, INFO, format, v...)
	} else {
		// 降级使用默认日志器
		defaultLogger.log(ctx, INFO, format, v...)
	}
}

// DatabaseWarnfc 带上下文记录数据库WARN级别的日志
func DatabaseWarnfc(ctx context.Context, format string, v ...interface{}) {
	if databaseLogger != nil {
		databaseLogger.log(ctx, WARN, format, v...)
	} else {
		// 降级使用默认日志器
		defaultLogger.log(ctx, WARN, format, v...)
	}
}

// DatabaseErrorfc 带上下文记录数据库ERROR级别的日志
func DatabaseErrorfc(ctx context.Context, format string, v ...interface{}) {
	if databaseLogger != nil {
		databaseLogger.log(ctx, ERROR, format, v...)
	} else {
		// 降级使用默认日志器
		defaultLogger.log(ctx, ERROR, format, v...)
	}
}

// DatabaseFatalfc 带上下文记录数据库FATAL级别的日志并退出程序
func DatabaseFatalfc(ctx context.Context, format string, v ...interface{}) {
	if databaseLogger != nil {
		databaseLogger.log(ctx, FATAL, format, v...)
	} else {
		// 降级使用默认日志器
		defaultLogger.log(ctx, FATAL, format, v...)
	}
}
