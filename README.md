# Xray Plugin (Refactored)

重构后的 xray-plugin，基于 Xray-core 的 Shadowsocks SIP003 插件。

与原版完全兼容的命令行接口和功能，采用模块化架构设计。

## 特性

- **完整兼容**: 与原版 xray-plugin 命令行接口完全一致
- **模块化设计**: 分离配置、服务器、日志等逻辑
- **多协议支持**: WebSocket、gRPC、Hysteria
- **Hysteria 服务器模式**: 支持 Hysteria 2 服务器端部署
- **TLS 支持**: 自动证书管理和自定义证书
- **伪装支持**: Hysteria 服务器支持多种伪装模式（404、文件、反向代理、自定义响应）
- **系统信息**: 使用 gopsutil 获取详细的系统版本和架构信息
- **跨平台**: 支持 Linux、Windows
- **单元测试**: 配置解析全覆盖测试

## 项目结构

```
xray-plugin-refactored/
├── cmd/xray-plugin/          # 主入口
│   └── main.go
├── internal/
│   ├── config/               # 配置管理
│   │   ├── config.go         # 配置结构体和验证
│   │   ├── parse.go          # 参数解析（CLI、JSON、环境变量）
│   │   └── *_test.go         # 单元测试
│   ├── server/               # Xray 服务器实现
│   │   └── server.go         # 完整的 Xray-core 集成
│   └── log/                  # 日志系统（与原版格式一致）
│       └── logger.go
├── pkg/
│   ├── build/                # 构建信息
│   │   └── build.go
│   └── version/              # 版本信息
│       └── version.go
├── go.mod
└── README.md
```

## 构建

```bash
# 基本构建
go build -o xray-plugin ./cmd/xray-plugin

# 优化构建（减小二进制体积）
go build -v -trimpath -buildvcs=false -ldflags "-s -w -buildid=" -o xray-plugin ./cmd/xray-plugin

# 指定版本号构建
go build -ldflags "-X main.VERSION=v1.0.0" -o xray-plugin ./cmd/xray-plugin

# 运行测试
go test ./...
```

## 使用

### 命令行参数

```bash
# 显示版本
./xray-plugin -version
./xray-plugin version

# 客户端模式（WebSocket）
./xray-plugin -localPort 1984 -remotePort 443 -mode websocket -host example.com

# 服务器模式（WebSocket + TLS）
./xray-plugin -server -localPort 443 -mode websocket -tls -host example.com

# gRPC 模式
./xray-plugin -mode grpc -serviceName myservice -tls -host example.com

# Hysteria 模式（客户端）
./xray-plugin -mode hysteria -host example.com

# Hysteria 模式（服务器）
./xray-plugin -server -localPort 443 -mode hysteria -tls -host example.com
```

### 完整参数列表

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `-server` | 服务器模式 | false |
| `-localAddr` | 本地监听地址 | 127.0.0.1 |
| `-localPort` | 本地监听端口 | 1984 |
| `-remoteAddr` | 远程地址 | 127.0.0.1 |
| `-remotePort` | 远程端口 | 1080 |
| `-mode` | 传输模式 (websocket/grpc/hysteria) | websocket |
| `-path` | WebSocket 路径 | / |
| `-host` | TLS/WebSocket 主机名 | cloudflare.com |
| `-serviceName` | gRPC 服务名 | |
| `-mux` | WebSocket 并发连接数 | 0 |
| `-fastOpen` | 启用 TCP Fast Open | false |
| `-tls` | 启用 TLS | false |
| `-cert` | TLS 证书路径 | |
| `-key` | TLS 密钥路径 | |
| `-certRaw` | TLS 证书内容 (PEM) | |
| `-loglevel` | 日志级别 (debug/info/error/none) | info |
| `-version` | 显示版本 | |
| `-V` | VPN 模式 | false |
| `-fwmark` | SO_MARK 值 | 0 |
| `-config` | 配置文件路径 (JSON) | |

### Hysteria 模式

Hysteria 是一个基于 UDP 的高性能代理协议，支持带宽控制和伪装功能。

#### 服务器端配置

**最简配置（仅需认证密钥）：**

```bash
ss-server -c config.json -p 443 --plugin xray-plugin \
  --plugin-opts "server;mode=hysteria;tls;host=mydomain.com;hysteriaAuth=secret123"
```

**带速度限制：**

```bash
ss-server -c config.json -p 443 --plugin xray-plugin \
  --plugin-opts "server;mode=hysteria;tls;host=mydomain.com;hysteriaAuth=secret123;hysteriaUp=100;hysteriaDown=100"
```

**带伪装（反向代理）：**

```bash
ss-server -c config.json -p 443 --plugin xray-plugin \
  --plugin-opts "server;mode=hysteria;tls;host=mydomain.com;hysteriaAuth=secret123;hysteriaMasqType=proxy;hysteriaMasqURL=https://example.com"
```

**完整服务器参数：**

```bash
ss-server -c config.json -p 443 --plugin xray-plugin \
  --plugin-opts "server;mode=hysteria;tls;host=mydomain.com;hysteriaAuth=secret123;hysteriaUp=100;hysteriaDown=100;hysteriaCongestion=brutal;hysteriaMaxStreams=4096;hysteriaMasqType=proxy;hysteriaMasqURL=https://example.com"
```

#### 客户端配置

**基本配置：**

```bash
ss-local -c config.json -p 443 --plugin xray-plugin \
  --plugin-opts "mode=hysteria;tls;host=mydomain.com;hysteriaAuth=secret123"
```

**带 UDP 端口跳跃：**

```bash
ss-local -c config.json -p 443 --plugin xray-plugin \
  --plugin-opts "mode=hysteria;tls;host=mydomain.com;hysteriaAuth=secret123;hysteriaMinPort=20000;hysteriaMaxPort=30000"
```

#### Hysteria 配置参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `hysteriaAuth` | 认证密钥（必须）| - |
| `hysteriaUp` | 上行速度限制 (Mbps) | 0 (无限制) |
| `hysteriaDown` | 下行速度限制 (Mbps) | 0 (无限制) |
| `hysteriaCongestion` | 拥塞控制 (brutal/bbr/reno/force-brutal) | brutal |
| `hysteriaMaxStreams` | 最大并发流数（服务器）| 1024 |
| `hysteriaMinPort` | UDP 端口跳跃起始端口（客户端）| 0 (不启用) |
| `hysteriaMaxPort` | UDP 端口跳跃结束端口（客户端）| 0 (不启用) |

#### 伪装配置参数

Hysteria 服务器支持对未认证请求返回伪装响应，降低被识别风险。

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `hysteriaMasqType` | 伪装类型 (404/file/proxy/string) | 404 |
| `hysteriaMasqFile` | 伪装文件目录路径（type=file）| |
| `hysteriaMasqURL` | 反向代理目标 URL（type=proxy）| |
| `hysteriaMasqURLRewriteHost` | 反向代理时重写 Host 头 | false |
| `hysteriaMasqURLInsecure` | 反向代理时跳过 TLS 验证 | false |
| `hysteriaMasqString` | 自定义响应内容（type=string）| |
| `hysteriaMasqStatus` | 自定义响应状态码（type=string）| 200 |

**伪装类型说明：**

| 类型 | 说明 |
|------|------|
| `404` | 返回 404 Not Found（默认） |
| `file` | 作为静态文件服务器，提供指定目录下的文件 |
| `proxy` | 反向代理到指定 URL，伪装为正常网站 |
| `string` | 返回自定义字符串内容，可指定状态码和 Header |

### JSON 配置文件

#### WebSocket 配置

```json
{
  "server": true,
  "localAddr": "0.0.0.0",
  "localPort": "443",
  "remoteAddr": "127.0.0.1",
  "remotePort": "8388",
  "mode": "websocket",
  "path": "/ws",
  "host": "example.com",
  "mux": 8,
  "tls": {
    "enabled": true,
    "cert": "/path/to/cert.pem",
    "key": "/path/to/key.pem"
  },
  "logLevel": "info"
}
```

#### Hysteria 服务器配置（最简）

```json
{
  "server": true,
  "localAddr": "0.0.0.0",
  "localPort": "443",
  "remoteAddr": "127.0.0.1",
  "remotePort": "8388",
  "mode": "hysteria",
  "host": "mydomain.com",
  "logLevel": "info",
  "tls": {
    "enabled": true,
    "cert": "/path/to/cert.pem",
    "key": "/path/to/key.pem"
  },
  "hysteria": {
    "auth": "your-secret-key"
  }
}
```

#### Hysteria 服务器配置（完整）

```json
{
  "server": true,
  "localAddr": "0.0.0.0",
  "localPort": "443",
  "remoteAddr": "127.0.0.1",
  "remotePort": "8388",
  "mode": "hysteria",
  "host": "mydomain.com",
  "logLevel": "info",
  "tls": {
    "enabled": true,
    "cert": "/path/to/cert.pem",
    "key": "/path/to/key.pem"
  },
  "hysteria": {
    "auth": "your-secret-key",
    "upMbps": 100,
    "downMbps": 100,
    "congestion": "brutal",
    "maxIncomingStreams": 4096,
    "masqType": "proxy",
    "masqURL": "https://example.com",
    "masqURLRewriteHost": true
  }
}
```

#### Hysteria 客户端配置

```json
{
  "server": false,
  "localAddr": "127.0.0.1",
  "localPort": "1080",
  "remoteAddr": "mydomain.com",
  "remotePort": "443",
  "mode": "hysteria",
  "host": "mydomain.com",
  "logLevel": "info",
  "tls": {
    "enabled": true
  },
  "hysteria": {
    "auth": "your-secret-key",
    "upMbps": 50,
    "downMbps": 100,
    "minPort": 20000,
    "maxPort": 30000
  }
}
```

### Shadowsocks 插件模式

```bash
# WebSocket + TLS
ss-server -c config.json -p 443 --plugin xray-plugin \
  --plugin-opts "server;mode=websocket;tls;host=example.com"

# Hysteria（最简）
ss-server -c config.json -p 443 --plugin xray-plugin \
  --plugin-opts "server;mode=hysteria;tls;host=mydomain.com;hysteriaAuth=secret123"

# Hysteria（带速度限制和伪装）
ss-server -c config.json -p 443 --plugin xray-plugin \
  --plugin-opts "server;mode=hysteria;tls;host=mydomain.com;hysteriaAuth=secret123;hysteriaUp=100;hysteriaDown=100;hysteriaMasqType=proxy;hysteriaMasqURL=https://example.com"

# gRPC + TLS
ss-server -c config.json -p 443 --plugin xray-plugin \
  --plugin-opts "server;mode=grpc;serviceName=myservice;tls;host=example.com"
```

## 支持的传输协议

| 协议 | 模式 | 说明 |
|------|------|------|
| WebSocket | `websocket` | 基于 HTTP 的 WebSocket 传输 |
| WebSocket + TLS | `websocket` + `tls` | 加密 WebSocket 传输 |
| gRPC | `grpc` | 基于 HTTP/2 的 gRPC 传输 |
| gRPC + TLS | `grpc` + `tls` | 加密 gRPC 传输 |
| Hysteria | `hysteria` | 基于 UDP 的高性能代理协议（服务器/客户端双端支持）|

## 测试

```bash
# 运行所有测试
go test ./...

# 运行配置包测试
go test ./internal/config/...

# 运行服务器测试
go test ./internal/server/...

# 查看覆盖率
go test -cover ./...

# 详细输出
go test -v ./...
```

## 依赖

- Go 1.26+
- github.com/xtls/xray-core v1.260327.0
- github.com/shirou/gopsutil/v4 v4.25.12
- golang.org/x/sys v0.42.0

## 与原版对比

| 特性 | 原版 | 重构版 |
|------|------|--------|
| 命令行接口 | ✅ | ✅ 完全兼容 |
| 日志格式 | ✅ | ✅ 完全一致 |
| 系统信息 | gopsutil | ✅ gopsutil |
| Windows 支持 | ✅ | ✅ 完整支持 |
| ARM 架构检测 | ✅ | ✅ 完整支持 |
| Hysteria 服务器 | ❌ | ✅ 新增支持 |
| 伪装 (Masquerade) | ❌ | ✅ 404/文件/代理/自定义 |
| Xray-core 集成 | 内联 | ✅ 模块化 |
| 单元测试 | 无 | ✅ 28个测试 |
| 配置文件 | 无 | ✅ JSON 支持 |

## 许可证

MIT License

## 致谢

基于 [Xray-core](https://github.com/xtls/xray-core) 和原版 [xray-plugin](https://github.com/teddysun/xray-plugin) 开发。
