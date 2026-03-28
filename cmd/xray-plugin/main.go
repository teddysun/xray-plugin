// cmd/xray-plugin/main.go
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/xtls/xray-core/core"

	"github.com/teddysun/xray-plugin/internal/config"
	"github.com/teddysun/xray-plugin/internal/log"
	"github.com/teddysun/xray-plugin/internal/server"
	"github.com/teddysun/xray-plugin/pkg/version"
)

var (
	VERSION = "custom"
	
	// 命令行参数
	serverMode   = flag.Bool("server", false, "Run in server mode")
	localAddr    = flag.String("localAddr", "127.0.0.1", "Local address to listen on")
	localPort    = flag.String("localPort", "1984", "Local port to listen on")
	remoteAddr   = flag.String("remoteAddr", "127.0.0.1", "Remote address to forward")
	remotePort   = flag.String("remotePort", "1080", "Remote port to forward")
	mode         = flag.String("mode", "websocket", "Transport mode: websocket, grpc, hysteria")
	path         = flag.String("path", "/", "WebSocket path")
	host         = flag.String("host", "cloudflare.com", "Host for TLS/WebSocket")
	serviceName  = flag.String("serviceName", "", "gRPC service name")
	mux          = flag.Int("mux", 0, "Concurrent connections for WebSocket")
	fastOpen     = flag.Bool("fastOpen", false, "Enable TCP fast open")
	tlsEnabled   = flag.Bool("tls", false, "Enable TLS")
	cert         = flag.String("cert", "", "TLS certificate path")
	key          = flag.String("key", "", "TLS key path")
	certRaw      = flag.String("certRaw", "", "TLS certificate in PEM format")
	logLevel     = flag.String("loglevel", "info", "Log level: debug, info, error, none")
	showVersion  = flag.Bool("version", false, "Show version")
	vpn          = flag.Bool("V", false, "Run in VPN mode")
	fwmark       = flag.Int("fwmark", 0, "SO_MARK value for outbound sockets")
	configFile   = flag.String("config", "", "Config file path (JSON format)")
)

func main() {
	// 平台初始化（Android: logcat 日志 + VPN fd 保护）
	logInit()
	registerControlFunc()

	flag.Parse()
	
	// 显示版本（支持 -version 或 version 参数）
	if *showVersion || (len(flag.Args()) > 0 && flag.Args()[0] == "version") {
		printVersion()
		return
	}
	
	// 加载配置
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}
	
	// 初始化日志
	logger := log.NewLogger(cfg.LogLevel)
	logger.Info(fmt.Sprintf("Starting xray-plugin %s", VERSION))
	logger.Info(fmt.Sprintf("Mode: %s", cfg.Mode))
	
	// 打印 Xray 版本
	printCoreVersion(logger)
	
	// 创建并启动服务器
	srv, err := server.NewServer(cfg)
	if err != nil {
		logger.Fatal("Failed to create server:", err)
	}
	
	if err := srv.Start(); err != nil {
		logger.Fatal("Failed to start server:", err)
	}
	
	logger.Info("Server started successfully")
	
	// 优雅关闭
	defer func() {
		// 添加短暂延迟，让 listener 先退出 accept 状态
		time.Sleep(100 * time.Millisecond)
		if err := srv.Close(); err != nil {
			logger.Warn("Error closing server:", err)
		}
	}()
	
	// 等待信号
	waitForSignal()
	
	logger.Info("Shutting down...")
}

// loadConfig 加载配置
func loadConfig() (*config.Config, error) {
	// 优先从文件加载
	if *configFile != "" {
		return config.FromFile(*configFile)
	}
	
	// 从环境变量解析（Shadowsocks 插件模式）
	envArgs, err := config.ParseEnv()
	if err != nil {
		return nil, err
	}
	
	// 如果环境变量有配置，使用它
	if len(envArgs) > 0 {
		return config.ArgsToConfig(envArgs)
	}
	
	// 使用命令行参数构建配置
	return &config.Config{
		Server:      *serverMode,
		LocalAddr:   *localAddr,
		LocalPort:   *localPort,
		RemoteAddr:  *remoteAddr,
		RemotePort:  *remotePort,
		Mode:        *mode,
		Path:        *path,
		Host:        *host,
		ServiceName: *serviceName,
		Mux:         *mux,
		FastOpen:    *fastOpen,
		VPN:         *vpn,
		Fwmark:      *fwmark,
		LogLevel:    *logLevel,
		TLS: config.TLSConfig{
			Enabled: *tlsEnabled,
			Cert:    *cert,
			Key:     *key,
			CertRaw: *certRaw,
		},
	}, nil
}

// printVersion 打印版本信息
func printVersion() {
	v := version.GetInfo(VERSION)
	fmt.Printf("xray-plugin %s\n", v.Version)
	fmt.Println("Yet another SIP003 plugin for shadowsocks")
	fmt.Printf("- os/version: %s\n", v.OSVersion)
	fmt.Printf("- os/kernel: %s\n", v.OSKernel)
	fmt.Printf("- os/type: %s\n", v.OSType)
	fmt.Printf("- os/arch: %s\n", v.OSArch)
	fmt.Printf("- go/version: %s\n", v.GoVersion)
}

// printCoreVersion 打印 Xray 核心版本
func printCoreVersion(logger *log.Logger) {
	version := core.VersionStatement()
	for _, s := range version {
		logger.Info(s)
	}
}

// waitForSignal 等待系统信号
func waitForSignal() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
}
