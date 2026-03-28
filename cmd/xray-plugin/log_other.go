//go:build !android

package main

// logInit 非 Android 平台无需额外的日志初始化
func logInit() {}

// logPlatform 非 Android 平台无需平台特定日志
func logPlatform(_ ...interface{}) {}
