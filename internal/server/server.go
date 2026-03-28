// internal/server/server.go
package server

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"google.golang.org/protobuf/proto"

	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/app/dispatcher"
	vlog "github.com/xtls/xray-core/app/log"
	"github.com/xtls/xray-core/app/proxyman"
	_ "github.com/xtls/xray-core/app/proxyman/inbound"
	_ "github.com/xtls/xray-core/app/proxyman/outbound"
	clog "github.com/xtls/xray-core/common/log"
	"github.com/xtls/xray-core/common/errors"
	"github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/common/platform/filesystem"
	"github.com/xtls/xray-core/common/protocol"
	"github.com/xtls/xray-core/common/serial"
	"github.com/xtls/xray-core/proxy/dokodemo"
	"github.com/xtls/xray-core/proxy/freedom"
	"github.com/xtls/xray-core/transport/internet"
	"github.com/xtls/xray-core/transport/internet/grpc"
	"github.com/xtls/xray-core/transport/internet/hysteria"
	"github.com/xtls/xray-core/transport/internet/tls"
	"github.com/xtls/xray-core/transport/internet/websocket"

	"github.com/teddysun/xray-plugin/internal/config"
	"github.com/teddysun/xray-plugin/pkg/build"
)

// Server 服务器接口
type Server interface {
	Start() error
	Close() error
}

// xrayServer Xray 服务器实现
type xrayServer struct {
	instance core.Server
	config   *config.Config
}

// NewServer 创建服务器实例
func NewServer(cfg *config.Config) (Server, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	coreConfig, err := generateConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to generate config: %w", err)
	}

	instance, err := core.New(coreConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create xray instance: %w", err)
	}

	return &xrayServer{
		instance: instance,
		config:   cfg,
	}, nil
}

// Start 启动服务器
func (s *xrayServer) Start() error {
	return s.instance.Start()
}

// Close 关闭服务器
func (s *xrayServer) Close() error {
	return s.instance.Close()
}

// generateConfig 生成 Xray 核心配置
func generateConfig(cfg *config.Config) (*core.Config, error) {
	lport, err := net.PortFromString(cfg.LocalPort)
	if err != nil {
		return nil, errors.New("invalid localPort:", cfg.LocalPort).Base(err)
	}

	rport, err := strconv.ParseUint(cfg.RemotePort, 10, 32)
	if err != nil {
		return nil, errors.New("invalid remotePort:", cfg.RemotePort).Base(err)
	}

	outboundProxy := serial.ToTypedMessage(&freedom.Config{
		DestinationOverride: &freedom.DestinationOverride{
			Server: &protocol.ServerEndpoint{
				Address: net.NewIPOrDomain(net.ParseAddress(cfg.RemoteAddr)),
				Port:    uint32(rport),
			},
		},
	})

	var transportSettings proto.Message
	var connectionReuse bool
	var protocolName string
	var quicParams *internet.QuicParams
	tlsEnabled := cfg.TLS.Enabled

	switch cfg.Mode {
	case "websocket":
		var ed uint32
		if u, err := url.Parse(cfg.Path); err == nil {
			if q := u.Query(); q.Get("ed") != "" {
				edVal, _ := strconv.Atoi(q.Get("ed"))
				ed = uint32(edVal)
				q.Del("ed")
				u.RawQuery = q.Encode()
				cfg.Path = u.String()
			}
		}
		transportSettings = &websocket.Config{
			Path: cfg.Path,
			Host: cfg.Host,
			Header: map[string]string{
				"host": cfg.Host,
			},
			Ed: ed,
		}
		if cfg.Mux != 0 {
			connectionReuse = true
		}
		protocolName = "websocket"

	case "grpc":
		transportSettings = &grpc.Config{
			ServiceName: cfg.ServiceName,
		}
		protocolName = "grpc"

	case "hysteria":
		// Hysteria 传输配置
		// xray-core v1.260327.0 精简了 hysteria.Config，大部分参数移到 QuicParams
		hysCfg := &hysteria.Config{
			Version: 2,
		}

		// 认证 (必须)
		if cfg.Hysteria.Auth != "" {
			hysCfg.Auth = cfg.Hysteria.Auth
		}

		// UDP 会话空闲超时 (秒, 用于 Hysteria UDP relay 清理)
		if cfg.Hysteria.MaxIdleTimeout > 0 {
			hysCfg.UdpIdleTimeout = cfg.Hysteria.MaxIdleTimeout / 1000 // 毫秒→秒
		} else {
			hysCfg.UdpIdleTimeout = 60 // 默认 60 秒
		}

		// 伪装配置 (服务器模式)
		if cfg.Server {
			hysCfg.MasqType = cfg.Hysteria.MasqType
			hysCfg.MasqFile = cfg.Hysteria.MasqFile
			hysCfg.MasqUrl = cfg.Hysteria.MasqURL
			hysCfg.MasqUrlRewriteHost = cfg.Hysteria.MasqURLRewriteHost
			hysCfg.MasqUrlInsecure = cfg.Hysteria.MasqURLInsecure
			hysCfg.MasqString = cfg.Hysteria.MasqString
			hysCfg.MasqStringHeaders = cfg.Hysteria.MasqStringHeaders
			if cfg.Hysteria.MasqStringStatus > 0 {
				hysCfg.MasqStringStatusCode = int32(cfg.Hysteria.MasqStringStatus)
			}
		}

		transportSettings = hysCfg
		protocolName = "hysteria"

		// 构建 QuicParams (速度限制、拥塞控制、流控窗口等参数)
		qp := &internet.QuicParams{}

		// 拥塞控制 (默认 brutal)
		if cfg.Hysteria.Congestion != "" {
			qp.Congestion = cfg.Hysteria.Congestion
		} else {
			qp.Congestion = "brutal"
		}

		// 速度限制 (Mbps → bytes/s)
		qp.BrutalUp = uint64(cfg.Hysteria.UpMbps) * 1024 * 1024 / 8
		qp.BrutalDown = uint64(cfg.Hysteria.DownMbps) * 1024 * 1024 / 8

		// 服务器: 最大入站流数
		if cfg.Server {
			if cfg.Hysteria.MaxIncomingStreams > 0 {
				qp.MaxIncomingStreams = cfg.Hysteria.MaxIncomingStreams
			} else {
				qp.MaxIncomingStreams = 1024
			}
		}

		// 客户端: UDP 端口跳跃
		if !cfg.Server && cfg.Hysteria.MinPort > 0 && cfg.Hysteria.MaxPort > cfg.Hysteria.MinPort {
			ports := make([]uint32, 0, cfg.Hysteria.MaxPort-cfg.Hysteria.MinPort+1)
			for port := cfg.Hysteria.MinPort; port <= cfg.Hysteria.MaxPort; port++ {
				ports = append(ports, uint32(port))
			}
			qp.UdpHop = &internet.UdpHop{Ports: ports}
		}

		// 流控窗口 (字节)
		if cfg.Hysteria.InitStreamReceiveWindow > 0 {
			qp.InitStreamReceiveWindow = cfg.Hysteria.InitStreamReceiveWindow
		}
		if cfg.Hysteria.MaxStreamReceiveWindow > 0 {
			qp.MaxStreamReceiveWindow = cfg.Hysteria.MaxStreamReceiveWindow
		}
		if cfg.Hysteria.InitConnReceiveWindow > 0 {
			qp.InitConnReceiveWindow = cfg.Hysteria.InitConnReceiveWindow
		}
		if cfg.Hysteria.MaxConnReceiveWindow > 0 {
			qp.MaxConnReceiveWindow = cfg.Hysteria.MaxConnReceiveWindow
		}

		// 超时参数 (内部配置毫秒 → QuicParams 秒)
		if cfg.Hysteria.MaxIdleTimeout > 0 {
			qp.MaxIdleTimeout = cfg.Hysteria.MaxIdleTimeout / 1000
		}
		if cfg.Hysteria.KeepAlivePeriod > 0 {
			qp.KeepAlivePeriod = cfg.Hysteria.KeepAlivePeriod / 1000
		}

		quicParams = qp

		// Hysteria 客户端使用自己的 TLS 实现，但服务器模式需要标准 TLS
		if !cfg.Server {
			tlsEnabled = false
		}

	default:
		return nil, errors.New("unsupported mode:", cfg.Mode)
	}

	streamConfig := internet.StreamConfig{
		ProtocolName: protocolName,
		TransportSettings: []*internet.TransportConfig{{
			ProtocolName: protocolName,
			Settings:     serial.ToTypedMessage(transportSettings),
		}},
		QuicParams: quicParams,
	}

	if cfg.FastOpen || cfg.Fwmark != 0 {
		socketConfig := &internet.SocketConfig{}
		if cfg.FastOpen {
			socketConfig.Tfo = 256
		}
		if cfg.Fwmark != 0 {
			socketConfig.Mark = int32(cfg.Fwmark)
		}
		streamConfig.SocketSettings = socketConfig
	}

	if tlsEnabled {
		tlsConfig := tls.Config{ServerName: cfg.Host}

		if cfg.Server {
			certificate := tls.Certificate{}
			certPath := cfg.TLS.Cert
			keyPath := cfg.TLS.Key

			if certPath == "" && cfg.TLS.CertRaw == "" {
				certPath = fmt.Sprintf("%s/.acme.sh/%s/fullchain.cer", build.GetHomeDir(), cfg.Host)
			}

			certificate.Certificate, err = readCertificate(certPath, cfg.TLS.CertRaw)
			if err != nil {
				return nil, errors.New("failed to read cert").Base(err)
			}

			if keyPath == "" {
				keyPath = fmt.Sprintf("%[1]s/.acme.sh/%[2]s/%[2]s.key", build.GetHomeDir(), cfg.Host)
			}
			certificate.Key, err = filesystem.ReadFile(keyPath)
			if err != nil {
				return nil, errors.New("failed to read key file").Base(err)
			}
			tlsConfig.Certificate = []*tls.Certificate{&certificate}

		} else if cfg.TLS.Cert != "" || cfg.TLS.CertRaw != "" {
			certificate := tls.Certificate{Usage: tls.Certificate_AUTHORITY_VERIFY}
			certificate.Certificate, err = readCertificate(cfg.TLS.Cert, cfg.TLS.CertRaw)
			if err != nil {
				return nil, errors.New("failed to read cert").Base(err)
			}
			tlsConfig.Certificate = []*tls.Certificate{&certificate}
		}

		streamConfig.SecurityType = serial.GetMessageType(&tlsConfig)
		streamConfig.SecuritySettings = []*serial.TypedMessage{
			serial.ToTypedMessage(&tlsConfig),
		}
	}

	apps := []*serial.TypedMessage{
		serial.ToTypedMessage(&dispatcher.Config{}),
		serial.ToTypedMessage(&proxyman.InboundConfig{}),
		serial.ToTypedMessage(&proxyman.OutboundConfig{}),
		serial.ToTypedMessage(buildLogConfig(cfg.LogLevel)),
	}

	if cfg.Server {
		proxyAddress := net.LocalHostIP
		if connectionReuse {
			proxyAddress = net.ParseAddress("v1.mux.cool")
		}

		localAddrs := parseLocalAddr(cfg.LocalAddr)
		inbounds := make([]*core.InboundHandlerConfig, len(localAddrs))

		for i, addr := range localAddrs {
			inbounds[i] = &core.InboundHandlerConfig{
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortList: &net.PortList{
						Range: []*net.PortRange{net.SinglePortRange(lport)},
					},
					Listen:         net.NewIPOrDomain(net.ParseAddress(addr)),
					StreamSettings: &streamConfig,
				}),
				ProxySettings: serial.ToTypedMessage(&dokodemo.Config{
					Address:  net.NewIPOrDomain(proxyAddress),
					Networks: []net.Network{net.Network_TCP},
				}),
			}
		}

		return &core.Config{
			Inbound: inbounds,
			Outbound: []*core.OutboundHandlerConfig{{
				ProxySettings: outboundProxy,
			}},
			App: apps,
		}, nil

	} else {
		senderConfig := proxyman.SenderConfig{StreamSettings: &streamConfig}
		if connectionReuse {
			senderConfig.MultiplexSettings = &proxyman.MultiplexingConfig{
				Enabled:     true,
				Concurrency: int32(cfg.Mux),
			}
		}
		return &core.Config{
			Inbound: []*core.InboundHandlerConfig{{
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortList: &net.PortList{
						Range: []*net.PortRange{net.SinglePortRange(lport)},
					},
					Listen: net.NewIPOrDomain(net.ParseAddress(cfg.LocalAddr)),
				}),
				ProxySettings: serial.ToTypedMessage(&dokodemo.Config{
					Address:  net.NewIPOrDomain(net.LocalHostIP),
					Networks: []net.Network{net.Network_TCP},
				}),
			}},
			Outbound: []*core.OutboundHandlerConfig{{
				SenderSettings: serial.ToTypedMessage(&senderConfig),
				ProxySettings:  outboundProxy,
			}},
			App: apps,
		}, nil
	}
}

// readCertificate 读取证书
func readCertificate(certPath, certRaw string) ([]byte, error) {
	if certRaw != "" {
		return []byte(certRaw), nil
	}
	return filesystem.ReadFile(certPath)
}

// buildLogConfig 构建日志配置
func buildLogConfig(logLevel string) *vlog.Config {
	config := &vlog.Config{
		ErrorLogType:  vlog.LogType_Console,
		ErrorLogLevel: clog.Severity_Warning,
		AccessLogType: vlog.LogType_Console,
	}

	switch strings.ToLower(logLevel) {
	case "debug":
		config.ErrorLogLevel = clog.Severity_Debug
	case "info":
		config.ErrorLogLevel = clog.Severity_Info
	case "warn", "warning":
		config.ErrorLogLevel = clog.Severity_Warning
	case "error":
		config.ErrorLogLevel = clog.Severity_Error
	case "none":
		config.ErrorLogType = vlog.LogType_None
		config.AccessLogType = vlog.LogType_None
	}

	return config
}

// parseLocalAddr 解析本地地址
func parseLocalAddr(localAddr string) []string {
	if localAddr == "" {
		return []string{"127.0.0.1"}
	}
	addrs := strings.Split(localAddr, "|")
	if len(addrs) == 0 {
		return []string{"127.0.0.1"}
	}
	return addrs
}
