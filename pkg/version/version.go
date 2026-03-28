// pkg/version/version.go
package version

import (
	"runtime"
	"strings"

	"github.com/shirou/gopsutil/v4/host"
	"golang.org/x/sys/cpu"
)

// Info 版本信息
type Info struct {
	Version   string
	OSVersion string
	OSKernel  string
	OSType    string
	OSArch    string
	GoVersion string
}

// GetInfo 获取版本信息
func GetInfo(version string) *Info {
	osVersion, osKernel := getOSVersion()

	return &Info{
		Version:   version,
		OSVersion: osVersion,
		OSKernel:  osKernel,
		OSType:    runtime.GOOS,
		OSArch:    getArch(),
		GoVersion: runtime.Version(),
	}
}

// getArch 获取详细的架构信息（包含 ARM 兼容性说明）
func getArch() string {
	arch := runtime.GOARCH

	if arch == "arm64" {
		// 64-bit ARM architecture, known as AArch64, was introduced with ARMv8
		arch += " (ARMv8 compatible)"
	} else if arch == "arm" {
		// 32-bit ARM architecture, which is ARMv7 and lower
		// Check CPU features for floating point hardware support
		if cpu.Initialized {
			if cpu.ARM.HasVFPv3 {
				arch += " (ARMv7 compatible)"
			} else if cpu.ARM.HasVFP {
				arch += " (ARMv6 compatible)"
			} else {
				arch += " (ARMv5 compatible, no hardfloat)"
			}
		}
	}

	return arch
}

// getOSVersion 使用 gopsutil 获取操作系统版本和内核信息
func getOSVersion() (osVersion, osKernel string) {
	if platform, _, version, err := host.PlatformInformation(); err == nil && platform != "" {
		osVersion = platform
		if version != "" {
			osVersion += " " + version
		}
	}

	if version, err := host.KernelVersion(); err == nil && version != "" {
		osKernel = version
	}

	if arch, err := host.KernelArch(); err == nil && arch != "" {
		if strings.HasSuffix(arch, "64") && osVersion != "" {
			osVersion += " (64 bit)"
		}
		if osKernel != "" {
			osKernel += " (" + arch + ")"
		}
	}

	return
}
