# Fortio proxy

Fortio simple TLS/ingress autocert proxy

Front end for running fortio report for instance standalone with TLS / Autocert and routing rules to multiplex multiple service behind a common TLS ingress (works with and allows multiplexing of grpc and h2c servers too)

# Install

using golang 1.18+

```shell
go install fortio.org/proxy@latest
sudo setcap CAP_NET_BIND_SERVICE=+eip $(which proxy)
```

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
