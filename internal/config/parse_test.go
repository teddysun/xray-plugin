// internal/config/parse_test.go
package config

import (
	"os"
	"testing"
)

func TestArgsGet(t *testing.T) {
	args := make(Args)
	args.Add("key1", "value1")
	args.Add("key2", "value2a")
	args.Add("key2", "value2b")
	
	// 测试 Get
	val, ok := args.Get("key1")
	if !ok || val != "value1" {
		t.Errorf("Expected Get('key1') to return ('value1', true), got ('%s', %v)", val, ok)
	}
	
	// 测试不存在的 key
	_, ok = args.Get("nonexistent")
	if ok {
		t.Error("Expected Get('nonexistent') to return false")
	}
	
	// 测试 nil Args
	var nilArgs Args
	_, ok = nilArgs.Get("key")
	if ok {
		t.Error("Expected Get on nil Args to return false")
	}
}

func TestParsePluginOptions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name:  "simple options",
			input: "server;mode=websocket;tls",
			expected: map[string]string{
				"server": "true",
				"mode":   "websocket",
				"tls":    "true",
			},
		},
		{
			name:  "with values",
			input: "mode=grpc;host=example.com;port=443",
			expected: map[string]string{
				"mode": "grpc",
				"host": "example.com",
				"port": "443",
			},
		},
		{
			name:  "empty string",
			input: "",
			expected: map[string]string{},
		},
		{
			name:  "single option",
			input: "server",
			expected: map[string]string{
				"server": "true",
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParsePluginOptions(tt.input)
			if err != nil {
				t.Fatalf("ParsePluginOptions failed: %v", err)
			}
			
			for key, expectedVal := range tt.expected {
				val, ok := result.Get(key)
				if !ok {
					t.Errorf("Expected key '%s' not found", key)
					continue
				}
				if val != expectedVal {
					t.Errorf("For key '%s', expected '%s', got '%s'", key, expectedVal, val)
				}
			}
		})
	}
}

func TestParseEnv(t *testing.T) {
	// 保存原始环境变量
	originalEnv := map[string]string{
		"SS_REMOTE_HOST": os.Getenv("SS_REMOTE_HOST"),
		"SS_REMOTE_PORT": os.Getenv("SS_REMOTE_PORT"),
		"SS_LOCAL_HOST":  os.Getenv("SS_LOCAL_HOST"),
		"SS_LOCAL_PORT":  os.Getenv("SS_LOCAL_PORT"),
	}
	
	// 测试后恢复
	defer func() {
		for k, v := range originalEnv {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()
	
	// 设置测试环境变量
	os.Setenv("SS_REMOTE_HOST", "192.168.1.1")
	os.Setenv("SS_REMOTE_PORT", "8388")
	os.Setenv("SS_LOCAL_HOST", "127.0.0.1")
	os.Setenv("SS_LOCAL_PORT", "1080")
	os.Setenv("SS_PLUGIN_OPTIONS", "mode=websocket;tls")
	
	args, err := ParseEnv()
	if err != nil {
		t.Fatalf("ParseEnv failed: %v", err)
	}
	
	// 验证解析结果
	if val, ok := args.Get("remoteAddr"); !ok || val != "192.168.1.1" {
		t.Errorf("Expected remoteAddr='192.168.1.1', got '%s'", val)
	}
	
	if val, ok := args.Get("localPort"); !ok || val != "1080" {
		t.Errorf("Expected localPort='1080', got '%s'", val)
	}
	
	// 验证插件选项也被解析
	if val, ok := args.Get("mode"); !ok || val != "websocket" {
		t.Errorf("Expected mode='websocket' from plugin options, got '%s'", val)
	}
}

func TestArgsToConfig(t *testing.T) {
	args := make(Args)
	args.Add("localAddr", "0.0.0.0")
	args.Add("localPort", "443")
	args.Add("remoteAddr", "1.2.3.4")
	args.Add("remotePort", "8388")
	args.Add("mode", "grpc")
	args.Add("host", "example.com")
	args.Add("tls", "true")
	
	cfg, err := ArgsToConfig(args)
	if err != nil {
		t.Fatalf("ArgsToConfig failed: %v", err)
	}
	
	if cfg.LocalAddr != "0.0.0.0" {
		t.Errorf("Expected LocalAddr='0.0.0.0', got '%s'", cfg.LocalAddr)
	}
	
	if cfg.Mode != "grpc" {
		t.Errorf("Expected Mode='grpc', got '%s'", cfg.Mode)
	}
	
	if !cfg.TLS.Enabled {
		t.Error("Expected TLS.Enabled to be true")
	}
}

func TestArgsToConfigHysteria(t *testing.T) {
	args := make(Args)
	args.Add("server", "true")
	args.Add("localAddr", "0.0.0.0")
	args.Add("localPort", "8443")
	args.Add("remoteAddr", "127.0.0.1")
	args.Add("remotePort", "8388")
	args.Add("mode", "hysteria")
	args.Add("host", "example.com")
	args.Add("tls", "true")
	args.Add("hysteriaAuth", "secret123")
	args.Add("hysteriaUp", "100")
	args.Add("hysteriaDown", "100")
	args.Add("hysteriaMinPort", "20000")
	args.Add("hysteriaMaxPort", "30000")
	args.Add("hysteriaCongestion", "brutal")
	args.Add("hysteriaMaxStreams", "4096")
	args.Add("hysteriaMasqType", "string")
	args.Add("hysteriaMasqString", "Hello World")
	
	cfg, err := ArgsToConfig(args)
	if err != nil {
		t.Fatalf("ArgsToConfig failed: %v", err)
	}
	
	if cfg.Mode != "hysteria" {
		t.Errorf("Expected Mode='hysteria', got '%s'", cfg.Mode)
	}
	if cfg.Hysteria.Auth != "secret123" {
		t.Errorf("Expected Hysteria.Auth='secret123', got '%s'", cfg.Hysteria.Auth)
	}
	if cfg.Hysteria.UpMbps != 100 {
		t.Errorf("Expected Hysteria.UpMbps=100, got %d", cfg.Hysteria.UpMbps)
	}
	if cfg.Hysteria.DownMbps != 100 {
		t.Errorf("Expected Hysteria.DownMbps=100, got %d", cfg.Hysteria.DownMbps)
	}
	if cfg.Hysteria.MinPort != 20000 {
		t.Errorf("Expected Hysteria.MinPort=20000, got %d", cfg.Hysteria.MinPort)
	}
	if cfg.Hysteria.MaxPort != 30000 {
		t.Errorf("Expected Hysteria.MaxPort=30000, got %d", cfg.Hysteria.MaxPort)
	}
	if cfg.Hysteria.Congestion != "brutal" {
		t.Errorf("Expected Hysteria.Congestion='brutal', got '%s'", cfg.Hysteria.Congestion)
	}
	if cfg.Hysteria.MaxIncomingStreams != 4096 {
		t.Errorf("Expected Hysteria.MaxIncomingStreams=4096, got %d", cfg.Hysteria.MaxIncomingStreams)
	}
	if cfg.Hysteria.MasqType != "string" {
		t.Errorf("Expected Hysteria.MasqType='string', got '%s'", cfg.Hysteria.MasqType)
	}
	if cfg.Hysteria.MasqString != "Hello World" {
		t.Errorf("Expected Hysteria.MasqString='Hello World', got '%s'", cfg.Hysteria.MasqString)
	}
}

func TestEncodePluginOptions(t *testing.T) {
	cfg := &Config{
		Server:   true,
		Mode:     "websocket",
		TLS:      TLSConfig{Enabled: true},
		Host:     "example.com",
		Path:     "/custom",
		Mux:      8,
		FastOpen: true,
	}
	
	result := EncodePluginOptions(cfg)
	
	// 验证包含预期的选项
	expectedParts := []string{"server", "tls", "host=example.com", "path=/custom", "mux=8", "fastOpen"}
	for _, part := range expectedParts {
		if !contains(result, part) {
			t.Errorf("Expected result to contain '%s', got '%s'", part, result)
		}
	}
}

func TestEncodePluginOptionsHysteria(t *testing.T) {
	cfg := &Config{
		Server: true,
		Mode:   "hysteria",
		Host:   "example.com",
		TLS:    TLSConfig{Enabled: true},
		Hysteria: HysteriaConfig{
			Auth:               "secret123",
			UpMbps:             100,
			DownMbps:           100,
			MinPort:            20000,
			MaxPort:            30000,
			Congestion:         "brutal",
			MaxIncomingStreams: 4096,
		},
	}
	
	result := EncodePluginOptions(cfg)
	
	// 验证包含 Hysteria 选项
	expectedParts := []string{
		"server",
		"mode=hysteria",
		"tls",
		"hysteriaAuth=secret123",
		"hysteriaUp=100",
		"hysteriaDown=100",
		"hysteriaMinPort=20000",
		"hysteriaMaxPort=30000",
		"hysteriaCongestion=brutal",
		"hysteriaMaxStreams=4096",
	}
	for _, part := range expectedParts {
		if !contains(result, part) {
			t.Errorf("Expected result to contain '%s', got '%s'", part, result)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
