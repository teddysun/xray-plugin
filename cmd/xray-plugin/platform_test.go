//go:build !android

package main

import (
	"testing"
)

func TestLogInit(t *testing.T) {
	// 非 Android 平台上 logInit 应为空操作，不 panic
	logInit()
}

func TestLogPlatform(t *testing.T) {
	// 非 Android 平台上 logPlatform 应为空操作，不 panic
	logPlatform("test message")
	logPlatform("multiple", "args")
	logPlatform()
}

func TestRegisterControlFunc(t *testing.T) {
	// 非 Android 平台上 registerControlFunc 应为空操作，不 panic
	registerControlFunc()
}

func TestPlatformInitOrder(t *testing.T) {
	// 模拟 main() 中的初始化顺序，确保可安全重复调用
	logInit()
	registerControlFunc()
	logInit()
	registerControlFunc()
}
