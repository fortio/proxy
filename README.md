[![Go Reference](https://pkg.go.dev/badge/fortio.org/proxy.svg)](https://pkg.go.dev/fortio.org/proxy)
[![Go Report Card](https://goreportcard.com/badge/fortio.org/proxy)](https://goreportcard.com/report/fortio.org/proxy)
[![GitHub Release](https://img.shields.io/github/release/fortio/proxy.svg?style=flat)](https://github.com/fortio/proxy/releases/)
[![govulncheck](https://img.shields.io/badge/govulncheck-No%20vulnerabilities-success)](https://github.com/fortio/proxy/actions/workflows/gochecks.yml)
[![golangci-lint](https://img.shields.io/badge/golangci%20lint-No%20issue-success)](https://github.com/fortio/proxy/actions/workflows/gochecks.yml)


# Fortio proxy

Fortio simple TLS/ingress autocert proxy

Front end for running fortio report for instance standalone with TLS / Autocert and routing rules to multiplex multiple service behind a common TLS ingress (works with and allows multiplexing of grpc and h2c servers too)

Any -certs-domains ending with `.ts.net` will be handled by the Tailscale cert client (see https://tailscale.com/kb/1153/enabling-https). Or you can now specify `-tailscale` and it will get the local server name and domain automatically using the tailscale go client api.

# Install

using golang 1.20+ (improved ReverseProxy api and security from 1.18)

```shell
go install fortio.org/proxy@latest
sudo setcap CAP_NET_BIND_SERVICE=+eip $(which proxy)
```

If you don't need or want the tailscale support, add `-tags no_tailscale` for a much smaller binary.

You can also download one of the many binary [releases](https://github.com/fortio/proxy/releases)

We publish a multi architecture docker image (linux/amd64, linux/arm64) `docker run fortio/proxy`

# Usage

See example of setup in https://github.com/fortio/demo-deployment

You can define routing rules using host or prefix matching, for instance:

```json
[
  {
    "host": "grpc.fortio.org",
    "destination": "http://127.0.0.1:8079"
  },
  {
    "prefix": "/fgrpc.PingServer",
    "destination": "http://127.0.0.1:8079"
  },
  {
    "prefix": "/grpc.health.v1.Health/Check",
    "destination": "http://127.0.0.1:8079"
  },
  {
    "host": "*",
    "destination": "http://127.0.0.1:8080"
  }
]
```

And which domains/common names you will accept and request certificates for (coma separated list in `-certs-domains` flag or dynamic config directory)

Optionally you can also configure `debug-host` for a Host (header, Authority in h2) that will serve a secured variant of fortio's debug handler for these requests: you can see it on [https://debug.fortio.org/a/random/test](https://debug.fortio.org/a/random/test)

There is a simpler config for single/default route:
If you want to setup TLS and forward everything to local (h2c) http server running on port 3000
```
go run fortio.org/proxy@latest -certs-domains ...your..server..full..name -h2 -default-route localhost:3000
```
(`http://` prefix can be omitted in the default route only)
