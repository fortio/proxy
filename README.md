# Fortio proxy

Fortio simple TLS/ingress autocert proxy

Front end for running fortio report for instance standalone with TLS / Autocert

# Install

using golang 1.18+

```
go install fortio.org/proxy@latest
sudo setcap CAP_NET_BIND_SERVICE=+eip `which proxy`
```

See example of setup in https://github.com/fortio/demo-deployment
