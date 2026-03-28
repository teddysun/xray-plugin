//go:build !android

package main

// registerControlFunc 非 Android 平台无需 VPN fd 保护
func registerControlFunc() {}
