// internal/server/server_test.go
package server

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/teddysun/xray-plugin/internal/config"
	"github.com/xtls/xray-core/app/proxyman"
	"github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/proxy/dokodemo"
	"github.com/xtls/xray-core/transport/internet"
	"github.com/xtls/xray-core/transport/internet/finalmask/udphop"
	"github.com/xtls/xray-core/transport/internet/grpc"
	"github.com/xtls/xray-core/transport/internet/hysteria"
	"github.com/xtls/xray-core/transport/internet/tls"
	"github.com/xtls/xray-core/transport/internet/websocket"
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
			Auth:              "test-secret",
			UpMbps:            100,
			DownMbps:          100,
			MasqType:          "string",
			MasqString:        "Hello World",
			MasqStringStatus:  200,
			MasqStringHeaders: map[string]string{"Content-Type": "text/plain"},
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

// mustType 断言类型，类型不符时以明确信息结束测试，而不是 panic。
func mustType[T any](t *testing.T, v interface{}) T {
	t.Helper()
	x, ok := v.(T)
	if !ok {
		t.Fatalf("unexpected type %T", v)
	}
	return x
}

// 检查各传输及客户端、服务器配置。
func TestGenerateConfigTransports(t *testing.T) {
	for _, mode := range []string{"websocket", "grpc", "hysteria"} {
		for _, server := range []bool{false, true} {
			role := "client"
			if server {
				role = "server"
			}
			t.Run(mode+"/"+role, func(t *testing.T) {
				cfg := config.DefaultConfig()
				cfg.Mode = mode
				cfg.Server = server
				cfg.LocalAddr = "127.0.0.1"
				cfg.LocalPort = "1080"
				cfg.RemoteAddr = "127.0.0.1"
				cfg.RemotePort = "8388"
				cfg.Host = "example.com"
				cfg.TLS.Enabled = false
				generated, err := generateConfig(cfg)
				if err != nil {
					t.Fatal(err)
				}
				inbound, err := generated.GetInbound()[0].GetProxySettings().GetInstance()
				if err != nil {
					t.Fatal(err)
				}
				door := mustType[*dokodemo.Config](t, inbound)
				if len(door.GetAllowedNetworks()) != 1 || door.GetAllowedNetworks()[0] != net.Network_TCP {
					t.Fatalf("unexpected inbound networks: %v", door.GetAllowedNetworks())
				}
				if door.GetPredefinedAddress() != net.LocalHostIP {
					t.Fatalf("unexpected rewrite address: %v", door.GetRewriteAddress())
				}
				var stream *internet.StreamConfig
				if server {
					receiver, err := generated.GetInbound()[0].GetReceiverSettings().GetInstance()
					if err != nil {
						t.Fatal(err)
					}
					stream = mustType[*proxyman.ReceiverConfig](t, receiver).GetStreamSettings()
				} else {
					sender, err := generated.GetOutbound()[0].GetSenderSettings().GetInstance()
					if err != nil {
						t.Fatal(err)
					}
					stream = mustType[*proxyman.SenderConfig](t, sender).GetStreamSettings()
				}
				if stream.GetProtocolName() != mode || len(stream.GetTransportSettings()) != 1 || stream.GetTransportSettings()[0].GetProtocolName() != mode {
					t.Fatalf("unexpected transport: %v", stream)
				}
				transport, err := stream.GetTransportSettings()[0].GetSettings().GetInstance()
				if err != nil {
					t.Fatal(err)
				}
				switch mode {
				case "websocket":
					ws := mustType[*websocket.Config](t, transport)
					if ws.GetHost() != cfg.Host || ws.GetPath() != cfg.Path || ws.GetHeader()["host"] != cfg.Host {
						t.Fatalf("unexpected websocket: %v", ws)
					}
				case "grpc":
					if mustType[*grpc.Config](t, transport).GetServiceName() != cfg.ServiceName {
						t.Fatal("unexpected grpc service")
					}
				case "hysteria":
					if mustType[*hysteria.Config](t, transport).GetUdpIdleTimeout() != 60 || stream.GetQuicParams() == nil {
						t.Fatal("unexpected hysteria settings")
					}
				}
				if len(stream.GetUdpmasks()) != 0 {
					t.Fatalf("unexpected UDP mask: %v", stream)
				}
				// Hysteria 客户端始终启用 TLS，其余情况在未指定 tls 时不启用
				if mode == "hysteria" && !server {
					if len(stream.GetSecuritySettings()) != 1 || stream.GetSecurityType() != stream.GetSecuritySettings()[0].GetType() {
						t.Fatalf("missing TLS for hysteria client: %v", stream)
					}
				} else if stream.GetSecurityType() != "" || len(stream.GetSecuritySettings()) != 0 {
					t.Fatalf("unexpected TLS: %v", stream)
				}
			})
		}
	}
}

// TestGenerateConfigHysteriaUDPHop 检查客户端 UDP 端口跳跃掩码（仅在端口范围有效时生成）。
func TestGenerateConfigHysteriaUDPHop(t *testing.T) {
	for _, server := range []bool{false, true} {
		for _, ports := range [][2]int{{0, 0}, {0, 20100}, {20000, 0}, {20000, 20000}, {20100, 20000}, {20000, 20100}} {
			cfg := config.DefaultConfig()
			cfg.Mode, cfg.Server = "hysteria", server
			cfg.LocalPort, cfg.RemotePort = "1080", "443"
			cfg.Hysteria.MinPort, cfg.Hysteria.MaxPort = ports[0], ports[1]
			generated, err := generateConfig(cfg)
			if err != nil {
				t.Fatal(err)
			}
			var stream *internet.StreamConfig
			if server {
				receiver, err := generated.GetInbound()[0].GetReceiverSettings().GetInstance()
				if err != nil {
					t.Fatal(err)
				}
				stream = mustType[*proxyman.ReceiverConfig](t, receiver).GetStreamSettings()
			} else {
				sender, err := generated.GetOutbound()[0].GetSenderSettings().GetInstance()
				if err != nil {
					t.Fatal(err)
				}
				stream = mustType[*proxyman.SenderConfig](t, sender).GetStreamSettings()
			}
			if server || ports[0] == 0 || ports[1] <= ports[0] {
				if len(stream.GetUdpmasks()) != 0 {
					t.Fatalf("unexpected mask: server=%v ports=%v", server, ports)
				}
				continue
			}
			if len(stream.GetUdpmasks()) != 1 {
				t.Fatal("missing udphop mask")
			}
			mask, err := stream.GetUdpmasks()[0].GetInstance()
			if err != nil {
				t.Fatal(err)
			}
			hop := mustType[*udphop.Config](t, mask)
			if !hop.GetRemote() || !hop.GetLocal() || hop.GetRemoteOnce() || hop.GetIntervalMin() != 30 || hop.GetIntervalMax() != 30 || len(hop.GetRemoteIPs()) != 0 || len(hop.GetRemotePorts()) != ports[1]-ports[0]+1 {
				t.Fatalf("unexpected udphop: %v", hop)
			}
			for i, port := range hop.GetRemotePorts() {
				if int(port) != ports[0]+i {
					t.Fatalf("unexpected port: %d", port)
				}
			}
		}
	}
}

// TestGenerateConfigHysteriaTLS 检查 Hysteria 客户端始终带 TLS 配置，SNI 与证书保持不变。
func TestGenerateConfigHysteriaTLS(t *testing.T) {
	certPath, _ := createTestCert(t)
	for _, enabled := range []bool{false, true} {
		for _, host := range []string{"", "example.com"} {
			for _, certificate := range []string{"none", "file", "raw"} {
				cfg := config.DefaultConfig()
				cfg.Mode, cfg.Host = "hysteria", host
				cfg.LocalPort, cfg.RemotePort = "1080", "443"
				cfg.TLS.Enabled = enabled
				var wantCert []byte
				switch certificate {
				case "file":
					cfg.TLS.Cert = certPath
					var err error
					wantCert, err = os.ReadFile(certPath)
					if err != nil {
						t.Fatal(err)
					}
				case "raw":
					cfg.TLS.CertRaw = "test certificate"
					wantCert = []byte(cfg.TLS.CertRaw)
				}
				generated, err := generateConfig(cfg)
				if err != nil {
					t.Fatal(err)
				}
				sender, err := generated.GetOutbound()[0].GetSenderSettings().GetInstance()
				if err != nil {
					t.Fatal(err)
				}
				stream := mustType[*proxyman.SenderConfig](t, sender).GetStreamSettings()
				// 无论是否指定 tls 选项，Hysteria 客户端都必须带 TLS 配置
				if len(stream.GetSecuritySettings()) != 1 {
					t.Fatalf("missing TLS settings (tls option = %v)", enabled)
				}
				security, err := stream.GetSecuritySettings()[0].GetInstance()
				if err != nil {
					t.Fatal(err)
				}
				securityConfig := mustType[*tls.Config](t, security)
				if securityConfig.GetServerName() != host || stream.GetSecurityType() != stream.GetSecuritySettings()[0].GetType() {
					t.Fatalf("unexpected TLS: %v", securityConfig)
				}
				if len(securityConfig.GetPinnedPeerCertSha256()) != 0 || len(securityConfig.GetVerifyPeerCertByName()) != 0 {
					t.Fatal("unexpected verification override")
				}
				if wantCert == nil {
					if len(securityConfig.GetCertificate()) != 0 {
						t.Fatal("unexpected certificate")
					}
				} else if len(securityConfig.GetCertificate()) != 1 || securityConfig.GetCertificate()[0].GetUsage() != tls.Certificate_AUTHORITY_VERIFY || string(securityConfig.GetCertificate()[0].GetCertificate()) != string(wantCert) {
					t.Fatal("unexpected authority certificate")
				}
			}
		}
	}
}
