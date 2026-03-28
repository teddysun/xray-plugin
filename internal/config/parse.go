// internal/config/parse.go
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Args 参数映射类型
type Args map[string][]string

// Get 获取第一个值
func (a Args) Get(key string) (string, bool) {
	if a == nil {
		return "", false
	}
	vals, ok := a[key]
	if !ok || len(vals) == 0 {
		return "", false
	}
	return vals[0], true
}

// Add 添加值
func (a Args) Add(key, value string) {
	a[key] = append(a[key], value)
}

// ParseEnv 从环境变量解析 Shadowsocks 插件选项
func ParseEnv() (Args, error) {
	opts := make(Args)
	
	ssRemoteHost := os.Getenv("SS_REMOTE_HOST")
	ssRemotePort := os.Getenv("SS_REMOTE_PORT")
	ssLocalHost := os.Getenv("SS_LOCAL_HOST")
	ssLocalPort := os.Getenv("SS_LOCAL_PORT")
	
	if ssRemoteHost == "" || ssRemotePort == "" || ssLocalHost == "" || ssLocalPort == "" {
		return opts, nil
	}
	
	opts.Add("remoteAddr", ssRemoteHost)
	opts.Add("remotePort", ssRemotePort)
	opts.Add("localAddr", ssLocalHost)
	opts.Add("localPort", ssLocalPort)
	
	ssPluginOptions := os.Getenv("SS_PLUGIN_OPTIONS")
	if ssPluginOptions != "" {
		otherOpts, err := ParsePluginOptions(ssPluginOptions)
		if err != nil {
			return nil, err
		}
		for k, v := range otherOpts {
			opts[k] = v
		}
	}
	
	return opts, nil
}

// ParsePluginOptions 解析插件选项字符串
// 格式: "server;mode=websocket;tls;host=example.com"
func ParsePluginOptions(s string) (Args, error) {
	opts := make(Args)
	
	for len(s) > 0 {
		var key, value string
		
		// 查找分隔符
		sepIdx := strings.IndexByte(s, ';')
		if sepIdx == -1 {
			sepIdx = len(s)
		}
		
		part := s[:sepIdx]
		
		// 检查是否有 '='
		if eqIdx := strings.IndexByte(part, '='); eqIdx > 0 {
			key = part[:eqIdx]
			value = part[eqIdx+1:]
		} else {
			key = part
			value = "true" // 布尔标志
		}
		
		if key != "" {
			opts.Add(key, value)
		}
		
		if sepIdx >= len(s) {
			break
		}
		s = s[sepIdx+1:]
	}
	
	return opts, nil
}

// ArgsToConfig 将 Args 转换为 Config
func ArgsToConfig(args Args) (*Config, error) {
	cfg := DefaultConfig()
	
	// 检查是否为服务器模式
	_, isServer := args.Get("server")
	
	if v, ok := args.Get("localAddr"); ok {
		cfg.LocalAddr = v
	}
	if v, ok := args.Get("localPort"); ok {
		cfg.LocalPort = v
	}
	if v, ok := args.Get("remoteAddr"); ok {
		cfg.RemoteAddr = v
	}
	if v, ok := args.Get("remotePort"); ok {
		cfg.RemotePort = v
	}
	
	// SIP003 服务器模式：使用 remote 地址作为监听地址
	// 因为 SS_REMOTE_HOST/PORT 才是服务器外部地址
	if isServer {
		if cfg.RemoteAddr != "" {
			cfg.LocalAddr = cfg.RemoteAddr
		}
		if cfg.RemotePort != "" {
			cfg.LocalPort = cfg.RemotePort
		}
	}
	
	if v, ok := args.Get("mode"); ok {
		cfg.Mode = v
	}
	if v, ok := args.Get("path"); ok {
		cfg.Path = v
	}
	if v, ok := args.Get("host"); ok {
		cfg.Host = v
	}
	if v, ok := args.Get("serviceName"); ok {
		cfg.ServiceName = v
	}
	if v, ok := args.Get("loglevel"); ok {
		cfg.LogLevel = v
	}
	if v, ok := args.Get("mux"); ok {
		if mux, err := strconv.Atoi(v); err == nil {
			cfg.Mux = mux
		}
	}
	if _, ok := args.Get("server"); ok {
		cfg.Server = true
	}
	if _, ok := args.Get("fastOpen"); ok {
		cfg.FastOpen = true
	}
	if _, ok := args.Get("tls"); ok {
		cfg.TLS.Enabled = true
	}
	if v, ok := args.Get("cert"); ok {
		cfg.TLS.Cert = v
	}
	if v, ok := args.Get("key"); ok {
		cfg.TLS.Key = v
	}
	if v, ok := args.Get("certRaw"); ok {
		cfg.TLS.CertRaw = v
	}
	if v, ok := args.Get("fwmark"); ok {
		if fwmark, err := strconv.Atoi(v); err == nil {
			cfg.Fwmark = fwmark
		}
	}
	if _, ok := args.Get("__android_vpn"); ok {
		cfg.VPN = true
	}
	
	// Hysteria 配置
	if v, ok := args.Get("hysteriaAuth"); ok {
		cfg.Hysteria.Auth = v
	}
	if v, ok := args.Get("hysteriaUp"); ok {
		if up, err := strconv.Atoi(v); err == nil {
			cfg.Hysteria.UpMbps = up
		}
	}
	if v, ok := args.Get("hysteriaDown"); ok {
		if down, err := strconv.Atoi(v); err == nil {
			cfg.Hysteria.DownMbps = down
		}
	}
	if v, ok := args.Get("hysteriaMinPort"); ok {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Hysteria.MinPort = port
		}
	}
	if v, ok := args.Get("hysteriaMaxPort"); ok {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Hysteria.MaxPort = port
		}
	}
	if v, ok := args.Get("hysteriaCongestion"); ok {
		cfg.Hysteria.Congestion = v
	}
	if v, ok := args.Get("hysteriaMaxStreams"); ok {
		if streams, err := strconv.ParseInt(v, 10, 64); err == nil {
			cfg.Hysteria.MaxIncomingStreams = streams
		}
	}
	if v, ok := args.Get("hysteriaMasqType"); ok {
		cfg.Hysteria.MasqType = v
	}
	if v, ok := args.Get("hysteriaMasqFile"); ok {
		cfg.Hysteria.MasqFile = v
	}
	if v, ok := args.Get("hysteriaMasqURL"); ok {
		cfg.Hysteria.MasqURL = v
	}
	if _, ok := args.Get("hysteriaMasqURLRewriteHost"); ok {
		cfg.Hysteria.MasqURLRewriteHost = true
	}
	if _, ok := args.Get("hysteriaMasqURLInsecure"); ok {
		cfg.Hysteria.MasqURLInsecure = true
	}
	if v, ok := args.Get("hysteriaMasqString"); ok {
		cfg.Hysteria.MasqString = v
	}
	if v, ok := args.Get("hysteriaMasqStatus"); ok {
		if status, err := strconv.Atoi(v); err == nil {
			cfg.Hysteria.MasqStringStatus = status
		}
	}
	
	return cfg, nil
}

// EncodePluginOptions 将 Config 编码为插件选项字符串
func EncodePluginOptions(cfg *Config) string {
	var parts []string

	if cfg.Server {
		parts = append(parts, "server")
	}
	if cfg.Mode != "" && cfg.Mode != "websocket" {
		parts = append(parts, fmt.Sprintf("mode=%s", cfg.Mode))
	}
	if cfg.TLS.Enabled {
		parts = append(parts, "tls")
	}
	if cfg.Host != "" {
		parts = append(parts, fmt.Sprintf("host=%s", cfg.Host))
	}
	if cfg.Path != "" && cfg.Path != "/" {
		parts = append(parts, fmt.Sprintf("path=%s", cfg.Path))
	}
	if cfg.Mux > 0 {
		parts = append(parts, fmt.Sprintf("mux=%d", cfg.Mux))
	}
	if cfg.FastOpen {
		parts = append(parts, "fastOpen")
	}
	if cfg.LogLevel != "" && cfg.LogLevel != "info" {
		parts = append(parts, fmt.Sprintf("loglevel=%s", cfg.LogLevel))
	}

	// Hysteria 配置
	if cfg.Hysteria.Auth != "" {
		parts = append(parts, fmt.Sprintf("hysteriaAuth=%s", cfg.Hysteria.Auth))
	}
	if cfg.Hysteria.UpMbps > 0 {
		parts = append(parts, fmt.Sprintf("hysteriaUp=%d", cfg.Hysteria.UpMbps))
	}
	if cfg.Hysteria.DownMbps > 0 {
		parts = append(parts, fmt.Sprintf("hysteriaDown=%d", cfg.Hysteria.DownMbps))
	}
	if cfg.Hysteria.MinPort > 0 {
		parts = append(parts, fmt.Sprintf("hysteriaMinPort=%d", cfg.Hysteria.MinPort))
	}
	if cfg.Hysteria.MaxPort > 0 {
		parts = append(parts, fmt.Sprintf("hysteriaMaxPort=%d", cfg.Hysteria.MaxPort))
	}
	if cfg.Hysteria.Congestion != "" {
		parts = append(parts, fmt.Sprintf("hysteriaCongestion=%s", cfg.Hysteria.Congestion))
	}
	if cfg.Hysteria.MaxIncomingStreams > 0 {
		parts = append(parts, fmt.Sprintf("hysteriaMaxStreams=%d", cfg.Hysteria.MaxIncomingStreams))
	}

	return strings.Join(parts, ";")
}
