// internal/config/config.go
package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config 主配置结构体
type Config struct {
	Server      bool   `json:"server"`
	LocalAddr   string `json:"localAddr"`
	LocalPort   string `json:"localPort"`
	RemoteAddr  string `json:"remoteAddr"`
	RemotePort  string `json:"remotePort"`
	Mode        string `json:"mode"` // websocket, grpc, hysteria
	Path        string `json:"path"`
	Host        string `json:"host"`
	ServiceName string `json:"serviceName"`
	Mux         int    `json:"mux"`
	FastOpen    bool   `json:"fastOpen"`
	VPN         bool   `json:"vpn"`
	Fwmark      int    `json:"fwmark"`
	LogLevel    string `json:"logLevel"`
	TLS         TLSConfig `json:"tls"`
	
	// Hysteria 专用配置
	Hysteria HysteriaConfig `json:"hysteria"`
}

// TLSConfig TLS 配置
type TLSConfig struct {
	Enabled bool   `json:"enabled"`
	Cert    string `json:"cert"`
	Key     string `json:"key"`
	CertRaw string `json:"certRaw"`
}

// HysteriaConfig Hysteria 协议专用配置
type HysteriaConfig struct {
	// 认证
	Auth string `json:"auth"`
	
	// 速度限制 (bps, 0 = 无限制)
	UpMbps   int `json:"upMbps"`
	DownMbps int `json:"downMbps"`
	
	// UDP 端口范围 (服务器模式)
	MinPort int `json:"minPort"`
	MaxPort int `json:"maxPort"`
	
	// 拥塞控制: "brutal", "cubic", "none" (默认 brutal)
	Congestion string `json:"congestion"`
	
	// 流控窗口参数 (字节)
	InitStreamReceiveWindow uint64 `json:"initStreamReceiveWindow"`
	MaxStreamReceiveWindow  uint64 `json:"maxStreamReceiveWindow"`
	InitConnReceiveWindow   uint64 `json:"initConnReceiveWindow"`
	MaxConnReceiveWindow    uint64 `json:"maxConnReceiveWindow"`
	
	// 连接参数
	MaxIdleTimeout  int64 `json:"maxIdleTimeout"`  // 毫秒
	KeepAlivePeriod int64 `json:"keepAlivePeriod"` // 毫秒
	
	// 服务器连接限制
	MaxIncomingStreams int64 `json:"maxIncomingStreams"`
	
	// 伪装配置 (masquerade)
	MasqType           string            `json:"masqType"`           // "file", "proxy", "string"
	MasqFile           string            `json:"masqFile"`           // 伪装文件路径
	MasqURL            string            `json:"masqURL"`            // 伪装 URL
	MasqURLRewriteHost bool              `json:"masqURLRewriteHost"` // 是否重写 Host
	MasqURLInsecure    bool              `json:"masqURLInsecure"`    // 跳过 TLS 验证
	MasqString         string            `json:"masqString"`         // 伪装响应字符串
	MasqStringHeaders  map[string]string `json:"masqStringHeaders"`  // 伪装响应头
	MasqStringStatus   int               `json:"masqStringStatus"`   // 伪装响应状态码
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		LocalAddr:  "127.0.0.1",
		LocalPort:  "1984",
		RemoteAddr: "127.0.0.1",
		RemotePort: "1080",
		Mode:       "websocket",
		Path:       "/",
		Host:       "cloudflare.com",
		LogLevel:   "info",
		Hysteria:   DefaultHysteriaConfig(),
	}
}

// DefaultHysteriaConfig 返回 Hysteria 默认配置
func DefaultHysteriaConfig() HysteriaConfig {
	return HysteriaConfig{
		// 速度限制: 0 表示无限制
		UpMbps:   0,
		DownMbps: 0,

		// UDP 端口范围: 默认不指定 (使用随机端口)
		MinPort: 0,
		MaxPort: 0,

		// 拥塞控制: 默认使用 brutal (Hysteria 特色)
		Congestion: "brutal",

		// 流控窗口: 使用 Xray-core 内部默认值
		// 如需自定义，参考值: 2MB-8MB
		InitStreamReceiveWindow: 0, // 使用内核默认 (2MB)
		MaxStreamReceiveWindow:  0, // 使用内核默认 (8MB)
		InitConnReceiveWindow:   0, // 使用内核默认 (2MB)
		MaxConnReceiveWindow:    0, // 使用内核默认 (8MB)

		// 连接参数: 使用内核默认值
		MaxIdleTimeout:  0, // 使用内核默认 (30s)
		KeepAlivePeriod: 0, // 使用内核默认 (10s)

		// 服务器连接限制: 默认 1024
		MaxIncomingStreams: 1024,

		// 伪装: 默认 404
		MasqType:           "404",
		MasqStringStatus:   200,
		MasqStringHeaders:  nil,
	}
}

// Validate 验证配置有效性
func (c *Config) Validate() error {
	if c.LocalPort == "" {
		return fmt.Errorf("localPort is required")
	}
	if c.RemotePort == "" {
		return fmt.Errorf("remotePort is required")
	}
	if c.Mode == "" {
		return fmt.Errorf("mode is required")
	}
	
	switch c.Mode {
	case "websocket", "grpc", "hysteria":
		// 支持的协议
	default:
		return fmt.Errorf("unsupported mode: %s", c.Mode)
	}
	
	if c.TLS.Enabled && c.Host == "" {
		return fmt.Errorf("host is required when TLS is enabled")
	}
	
	return nil
}

// FromFile 从 JSON 文件加载配置
func FromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	cfg := DefaultConfig()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	
	return cfg, nil
}

// ToFile 保存配置到 JSON 文件
func (c *Config) ToFile(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	return nil
}
