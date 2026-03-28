// internal/config/config_test.go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	
	if cfg.LocalAddr != "127.0.0.1" {
		t.Errorf("Expected LocalAddr to be '127.0.0.1', got '%s'", cfg.LocalAddr)
	}
	
	if cfg.LocalPort != "1984" {
		t.Errorf("Expected LocalPort to be '1984', got '%s'", cfg.LocalPort)
	}
	
	if cfg.Mode != "websocket" {
		t.Errorf("Expected Mode to be 'websocket', got '%s'", cfg.Mode)
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &Config{
				LocalPort:  "1984",
				RemotePort: "1080",
				Mode:       "websocket",
			},
			wantErr: false,
		},
		{
			name: "missing localPort",
			cfg: &Config{
				RemotePort: "1080",
				Mode:       "websocket",
			},
			wantErr: true,
		},
		{
			name: "missing remotePort",
			cfg: &Config{
				LocalPort: "1984",
				Mode:      "websocket",
			},
			wantErr: true,
		},
		{
			name: "missing mode",
			cfg: &Config{
				LocalPort:  "1984",
				RemotePort: "1080",
			},
			wantErr: true,
		},
		{
			name: "unsupported mode",
			cfg: &Config{
				LocalPort:  "1984",
				RemotePort: "1080",
				Mode:       "tcp",
			},
			wantErr: true,
		},
		{
			name: "TLS without host",
			cfg: &Config{
				LocalPort:  "1984",
				RemotePort: "1080",
				Mode:       "websocket",
				TLS:        TLSConfig{Enabled: true},
			},
			wantErr: true,
		},
		{
			name: "valid hysteria config",
			cfg: &Config{
				LocalPort:  "1984",
				RemotePort: "1080",
				Mode:       "hysteria",
				Hysteria:   HysteriaConfig{Auth: "secret123"},
			},
			wantErr: false,
		},
		{
			name: "hysteria server config",
			cfg: &Config{
				Server:     true,
				LocalPort:  "443",
				RemotePort: "8388",
				Mode:       "hysteria",
				Host:       "example.com",
				TLS:        TLSConfig{Enabled: true},
				Hysteria: HysteriaConfig{
					Auth:       "secret",
					UpMbps:     100,
					DownMbps:   100,
					MinPort:    20000,
					MaxPort:    30000,
					Congestion: "brutal",
				},
			},
			wantErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigFromFile(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()
	
	// 创建测试配置文件
	configContent := `{
		"server": true,
		"localAddr": "0.0.0.0",
		"localPort": "443",
		"mode": "websocket",
		"host": "example.com",
		"tls": {
			"enabled": true
		}
	}`
	
	configPath := filepath.Join(tmpDir, "test_config.json")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	
	// 加载配置
	cfg, err := FromFile(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	// 验证
	if !cfg.Server {
		t.Error("Expected Server to be true")
	}
	
	if cfg.LocalPort != "443" {
		t.Errorf("Expected LocalPort to be '443', got '%s'", cfg.LocalPort)
	}
	
	if !cfg.TLS.Enabled {
		t.Error("Expected TLS.Enabled to be true")
	}
}

func TestHysteriaConfigFromFile(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()
	
	// 创建 Hysteria 测试配置文件
	configContent := `{
		"server": true,
		"localAddr": "0.0.0.0",
		"localPort": "8443",
		"remoteAddr": "127.0.0.1",
		"remotePort": "8388",
		"mode": "hysteria",
		"host": "example.com",
		"logLevel": "debug",
		"tls": {
			"enabled": true,
			"cert": "/path/to/cert.pem",
			"key": "/path/to/key.pem"
		},
		"hysteria": {
			"auth": "test-secret-key",
			"upMbps": 100,
			"downMbps": 100,
			"minPort": 20000,
			"maxPort": 20100,
			"congestion": "brutal",
			"maxIncomingStreams": 4096,
			"masqType": "file",
			"masqFile": "/path/to/masq.html"
		}
	}`
	
	configPath := filepath.Join(tmpDir, "hysteria_config.json")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	
	// 加载配置
	cfg, err := FromFile(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	// 验证基本配置
	if !cfg.Server {
		t.Error("Expected Server to be true")
	}
	if cfg.Mode != "hysteria" {
		t.Errorf("Expected Mode to be 'hysteria', got '%s'", cfg.Mode)
	}
	
	// 验证 Hysteria 配置
	if cfg.Hysteria.Auth != "test-secret-key" {
		t.Errorf("Expected Hysteria.Auth='test-secret-key', got '%s'", cfg.Hysteria.Auth)
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
	if cfg.Hysteria.MaxPort != 20100 {
		t.Errorf("Expected Hysteria.MaxPort=20100, got %d", cfg.Hysteria.MaxPort)
	}
	if cfg.Hysteria.Congestion != "brutal" {
		t.Errorf("Expected Hysteria.Congestion='brutal', got '%s'", cfg.Hysteria.Congestion)
	}
	if cfg.Hysteria.MaxIncomingStreams != 4096 {
		t.Errorf("Expected Hysteria.MaxIncomingStreams=4096, got %d", cfg.Hysteria.MaxIncomingStreams)
	}
	if cfg.Hysteria.MasqType != "file" {
		t.Errorf("Expected Hysteria.MasqType='file', got '%s'", cfg.Hysteria.MasqType)
	}
	if cfg.Hysteria.MasqFile != "/path/to/masq.html" {
		t.Errorf("Expected Hysteria.MasqFile='/path/to/masq.html', got '%s'", cfg.Hysteria.MasqFile)
	}
}

func TestConfigToFile(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "output_config.json")
	
	cfg := &Config{
		Server:     true,
		LocalAddr:  "0.0.0.0",
		LocalPort:  "443",
		Mode:       "grpc",
		ServiceName: "xray",
	}
	
	if err := cfg.ToFile(configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}
	
	// 验证文件存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}
	
	// 加载回来验证
	loaded, err := FromFile(configPath)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}
	
	if loaded.Mode != "grpc" {
		t.Errorf("Expected Mode to be 'grpc', got '%s'", loaded.Mode)
	}
}
