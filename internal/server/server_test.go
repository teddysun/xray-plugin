// internal/server/server_test.go
package server

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/teddysun/xray-plugin/internal/config"
)

// createTestCert 创建测试用的临时证书文件
func createTestCert(t *testing.T) (certPath, keyPath string) {
	tmpDir := t.TempDir()
	certPath = filepath.Join(tmpDir, "test-cert.pem")
	keyPath = filepath.Join(tmpDir, "test-key.pem")

	// 生成自签名测试证书
	certContent := `-----BEGIN CERTIFICATE-----
MIIBkTCB+wIJAKHBfpE
MIIDXTCCAkWgAwIBAgIJAJC1HiIAZAiUMA0GCSqGSIb3Qa6F7SA8
-----END CERTIFICATE-----`

	keyContent := `-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC5V2MXuB4iWj7K
-----END PRIVATE KEY-----`

	if err := os.WriteFile(certPath, []byte(certContent), 0644); err != nil {
		t.Fatalf("Failed to create test cert: %v", err)
	}
	if err := os.WriteFile(keyPath, []byte(keyContent), 0644); err != nil {
		t.Fatalf("Failed to create test key: %v", err)
	}

	return certPath, keyPath
}

func TestNewServerWebsocket(t *testing.T) {
	cfg := &config.Config{
		Server:     true,
		LocalAddr:  "127.0.0.1",
		LocalPort:  "1984",
		RemoteAddr: "127.0.0.1",
		RemotePort: "8388",
		Mode:       "websocket",
		Path:       "/ws",
		Host:       "example.com",
		LogLevel:   "info",
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	if server == nil {
		t.Fatal("Expected server to be non-nil")
	}

	// 测试关闭
	err = server.Close()
	if err != nil {
		t.Errorf("Failed to close server: %v", err)
	}
}

func TestNewServerHysteria(t *testing.T) {
	certPath, keyPath := createTestCert(t)

	cfg := &config.Config{
		Server:     true,
		LocalAddr:  "127.0.0.1",
		LocalPort:  "1984",
		RemoteAddr: "127.0.0.1",
		RemotePort: "8388",
		Mode:       "hysteria",
		Host:       "localhost",
		LogLevel:   "info",
		TLS: config.TLSConfig{
			Enabled: true,
			Cert:    certPath,
			Key:     keyPath,
		},
		Hysteria: config.HysteriaConfig{
			Auth:               "test-secret",
			UpMbps:             100,
			DownMbps:           100,
			MinPort:            20000,
			MaxPort:            20100,
			Congestion:         "brutal",
			MaxIncomingStreams: 4096,
		},
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	if server == nil {
		t.Fatal("Expected server to be non-nil")
	}

	// 测试关闭
	err = server.Close()
	if err != nil {
		t.Errorf("Failed to close server: %v", err)
	}
}

func TestNewServerHysteriaClient(t *testing.T) {
	cfg := &config.Config{
		Server:     false,
		LocalAddr:  "127.0.0.1",
		LocalPort:  "1080",
		RemoteAddr: "example.com",
		RemotePort: "443",
		Mode:       "hysteria",
		Host:       "example.com",
		LogLevel:   "info",
		Hysteria: config.HysteriaConfig{
			Auth:     "test-secret",
			UpMbps:   50,
			DownMbps: 100,
		},
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	if server == nil {
		t.Fatal("Expected server to be non-nil")
	}

	// 测试关闭
	err = server.Close()
	if err != nil {
		t.Errorf("Failed to close server: %v", err)
	}
}

func TestNewServerInvalidMode(t *testing.T) {
	cfg := &config.Config{
		Server:     true,
		LocalAddr:  "127.0.0.1",
		LocalPort:  "1984",
		RemoteAddr: "127.0.0.1",
		RemotePort: "8388",
		Mode:       "invalid",
		LogLevel:   "info",
	}

	_, err := NewServer(cfg)
	if err == nil {
		t.Fatal("Expected error for invalid mode")
	}

	if !strings.Contains(err.Error(), "unsupported mode") {
		t.Errorf("Expected 'unsupported mode' error, got: %v", err)
	}
}

func TestNewServerInvalidPort(t *testing.T) {
	cfg := &config.Config{
		Server:     true,
		LocalAddr:  "127.0.0.1",
		LocalPort:  "invalid",
		RemoteAddr: "127.0.0.1",
		RemotePort: "8388",
		Mode:       "websocket",
		LogLevel:   "info",
	}

	_, err := NewServer(cfg)
	if err == nil {
		t.Fatal("Expected error for invalid port")
	}
}

func TestNewServerHysteriaWithMasq(t *testing.T) {
	certPath, keyPath := createTestCert(t)

	cfg := &config.Config{
		Server:     true,
		LocalAddr:  "127.0.0.1",
		LocalPort:  "1984",
		RemoteAddr: "127.0.0.1",
		RemotePort: "8388",
		Mode:       "hysteria",
		Host:       "localhost",
		LogLevel:   "info",
		TLS: config.TLSConfig{
			Enabled: true,
			Cert:    certPath,
			Key:     keyPath,
		},
		Hysteria: config.HysteriaConfig{
			Auth:               "test-secret",
			UpMbps:             100,
			DownMbps:           100,
			MasqType:           "string",
			MasqString:         "Hello World",
			MasqStringStatus:   200,
			MasqStringHeaders:  map[string]string{"Content-Type": "text/plain"},
		},
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	if server == nil {
		t.Fatal("Expected server to be non-nil")
	}

	// 测试关闭
	err = server.Close()
	if err != nil {
		t.Errorf("Failed to close server: %v", err)
	}
}

func TestNewServerHysteriaWithURLMasq(t *testing.T) {
	certPath, keyPath := createTestCert(t)

	cfg := &config.Config{
		Server:     true,
		LocalAddr:  "127.0.0.1",
		LocalPort:  "1984",
		RemoteAddr: "127.0.0.1",
		RemotePort: "8388",
		Mode:       "hysteria",
		Host:       "localhost",
		LogLevel:   "info",
		TLS: config.TLSConfig{
			Enabled: true,
			Cert:    certPath,
			Key:     keyPath,
		},
		Hysteria: config.HysteriaConfig{
			Auth:               "test-secret",
			UpMbps:             100,
			DownMbps:           100,
			MasqType:           "proxy",
			MasqURL:            "https://example.com/index.html",
			MasqURLRewriteHost: true,
			MasqURLInsecure:    false,
		},
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	if server == nil {
		t.Fatal("Expected server to be non-nil")
	}

	// 测试关闭
	err = server.Close()
	if err != nil {
		t.Errorf("Failed to close server: %v", err)
	}
}
