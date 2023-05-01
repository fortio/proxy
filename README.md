[![Go Reference](https://pkg.go.dev/badge/fortio.org/proxy.svg)](https://pkg.go.dev/fortio.org/proxy)
[![Go Report Card](https://goreportcard.com/badge/fortio.org/proxy)](https://goreportcard.com/report/fortio.org/proxy)
[![GitHub Release](https://img.shields.io/github/release/fortio/proxy.svg?style=flat)](https://github.com/fortio/proxy/releases/)
# Fortio proxy

Fortio simple TLS/ingress autocert proxy

Front end for running fortio report for instance standalone with TLS / Autocert and routing rules to multiplex multiple service behind a common TLS ingress (works with and allows multiplexing of grpc and h2c servers too)

# Install

using golang 1.20+ (improved ReverseProxy api and security from 1.18)

```shell
go install fortio.org/proxy@latest
sudo setcap CAP_NET_BIND_SERVICE=+eip $(which proxy)
```

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
