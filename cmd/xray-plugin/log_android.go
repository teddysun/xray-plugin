//go:build android

package main

/*
#cgo LDFLAGS: -landroid -llog

#include <android/log.h>
#include <stdlib.h>
#include <string.h>
*/
import "C"

import (
	"fmt"
	"unsafe"

	alog "github.com/xtls/xray-core/app/log"
	"github.com/xtls/xray-core/common"
	"github.com/xtls/xray-core/common/log"
	"github.com/xtls/xray-core/common/serial"
)

var ctag = C.CString("xray")

// androidLogger 将 xray-core 日志输出到 Android logcat
type androidLogger struct{}

func (l *androidLogger) Handle(msg log.Message) {
	var priority C.int
	var message string

	switch msg := msg.(type) {
	case *log.GeneralMessage:
		switch msg.Severity {
		case log.Severity_Error:
			priority = C.ANDROID_LOG_ERROR
		case log.Severity_Warning:
			priority = C.ANDROID_LOG_WARN
		case log.Severity_Info:
			priority = C.ANDROID_LOG_INFO
		case log.Severity_Debug:
			priority = C.ANDROID_LOG_DEBUG
		default:
			priority = C.ANDROID_LOG_VERBOSE
		}
		message = serial.ToString(msg.Content)
	default:
		priority = C.ANDROID_LOG_INFO
		message = msg.String()
	}

	cstr := C.CString(message)
	defer C.free(unsafe.Pointer(cstr))
	C.__android_log_write(C.int(priority), ctag, cstr)
}

// logInit 注册 Android logcat 日志处理器，替代 xray-core 默认的 Console 输出
func logInit() {
	common.Must(alog.RegisterHandlerCreator(alog.LogType_Console,
		func(_ alog.LogType, _ alog.HandlerCreatorOptions) (log.Handler, error) {
			return &androidLogger{}, nil
		}))
}

// logPlatform 使用 Android logcat 输出平台日志
func logPlatform(v ...interface{}) {
	cstr := C.CString(fmt.Sprintln(v...))
	defer C.free(unsafe.Pointer(cstr))
	C.__android_log_write(C.ANDROID_LOG_INFO, ctag, cstr)
}
