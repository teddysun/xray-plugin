// internal/log/logger.go
package log

import (
	"fmt"
	"os"
	"time"
)

// Level 日志级别
type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
	LevelFatal Level = "fatal"
	LevelNone  Level = "none"
)

// severityMap 将日志级别映射到 Xray 格式
var severityMap = map[Level]string{
	LevelDebug: "Debug",
	LevelInfo:  "Info",
	LevelWarn:  "Warning",
	LevelError: "Error",
	LevelFatal: "Error", // Fatal 也映射为 Error，因为 Xray 没有 Fatal 级别
}

// Logger 日志记录器
type Logger struct {
	level Level
}

// NewLogger 创建日志记录器
func NewLogger(level string) *Logger {
	return &Logger{
		level: Level(level),
	}
}

// log 内部日志输出方法
func (l *Logger) log(level Level, v ...interface{}) {
	if l.level == LevelNone {
		return
	}
	
	// 检查日志级别
	if !l.shouldLog(level) {
		return
	}
	
	severity := severityMap[level]
	timestamp := time.Now().Format("2006/01/02 15:04:05.000000")
	
	fmt.Fprintf(os.Stderr, "%s [%s] %s\n", timestamp, severity, fmt.Sprint(v...))
}

// shouldLog 检查是否应该记录该级别的日志
func (l *Logger) shouldLog(level Level) bool {
	levelOrder := map[Level]int{
		LevelDebug: 0,
		LevelInfo:  1,
		LevelWarn:  2,
		LevelError: 3,
		LevelFatal: 4,
		LevelNone:  5,
	}
	
	return levelOrder[level] >= levelOrder[l.level]
}

// Debug 调试日志
func (l *Logger) Debug(v ...interface{}) {
	l.log(LevelDebug, v...)
}

// Info 信息日志
func (l *Logger) Info(v ...interface{}) {
	l.log(LevelInfo, v...)
}

// Warn 警告日志
func (l *Logger) Warn(v ...interface{}) {
	l.log(LevelWarn, v...)
}

// Error 错误日志
func (l *Logger) Error(v ...interface{}) {
	l.log(LevelError, v...)
}

// Fatal 致命错误
func (l *Logger) Fatal(v ...interface{}) {
	l.log(LevelFatal, v...)
	os.Exit(1)
}
