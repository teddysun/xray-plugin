## Yet another SIP003 plugin for shadowsocks, based on [Xray-core](https://github.com/xtls/xray-core)

## Build

* `go build`

## Usage

See command line args for advanced usages.

### Shadowsocks over websocket (HTTP)

On your server

```sh
ss-server -c config.json -p 80 --plugin xray-plugin --plugin-opts "server"
```

On your client

```sh
ss-local -c config.json -p 80 --plugin xray-plugin
```

### Shadowsocks over websocket (HTTPS)

On your server

```sh
ss-server -c config.json -p 443 --plugin xray-plugin --plugin-opts "server;tls;host=mydomain.com"
```

On your client

```sh
ss-local -c config.json -p 443 --plugin xray-plugin --plugin-opts "tls;host=mydomain.com"
```

### Shadowsocks over quic

On your server

```sh
ss-server -c config.json -p 443 --plugin xray-plugin --plugin-opts "server;mode=quic;host=mydomain.com"
```

On your client

```sh
ss-local -c config.json -p 443 --plugin xray-plugin --plugin-opts "mode=quic;host=mydomain.com"
```

### Issue a cert for TLS and QUIC

`xray-plugin` will look for TLS certificates signed by [acme.sh](https://github.com/acmesh-official/acme.sh) by default.
Here's some sample commands for issuing a certificate using CloudFlare.
You can find commands for issuing certificates for other DNS providers at acme.sh.

```sh
wget -O-  https://get.acme.sh | sh
~/.acme.sh/acme.sh --issue --dns dns_cf -d mydomain.com
```

Alternatively, you can specify path to your certificates using option `cert` and `key`.

### Use `certRaw` to pass certificate

Instead of using `cert` to pass the certificate file, `certRaw` could be used to pass in PEM format certificate, that is the content between `-----BEGIN CERTIFICATE-----` and `-----END CERTIFICATE-----` without the line breaks.
