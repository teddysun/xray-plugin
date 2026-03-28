// pkg/build/build.go
package build

import (
	"os"
	"runtime"
)

// GetHomeDir 获取用户主目录
func GetHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// 回退方案
		if runtime.GOOS == "windows" {
			return os.Getenv("USERPROFILE")
		}
		return os.Getenv("HOME")
	}
	return home
}
