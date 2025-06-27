module fortio.org/proxy

go 1.24.0

require (
	fortio.org/cli v1.10.0
	fortio.org/dflag v1.8.1
	fortio.org/fortio v1.69.5
	fortio.org/log v1.17.2
	fortio.org/scli v1.16.1
	golang.org/x/crypto v0.39.0
	golang.org/x/net v0.41.0
	tailscale.com v1.84.3
)

require github.com/coder/websocket v1.8.12 // indirect

// Note most of these are coming in because of tailscale, if you want a smaller
// binary build with -tags no_tailscale
require (
	filippo.io/edwards25519 v1.1.0 // indirect
	fortio.org/safecast v1.0.0 // indirect
	fortio.org/sets v1.3.0 // indirect
	fortio.org/struct2env v0.4.2 // indirect
	fortio.org/version v1.0.4 // indirect
	github.com/akutz/memconn v0.1.0 // indirect
	github.com/alexbrainman/sspi v0.0.0-20231016080023-1a75b4708caa // indirect
	github.com/dblohm7/wingoes v0.0.0-20240119213807-a09d6be7affa // indirect
	github.com/fsnotify/fsnotify v1.8.0 // indirect
	github.com/fxamacker/cbor/v2 v2.7.0 // indirect
	github.com/go-json-experiment/json v0.0.0-20250223041408-d3c622f1b874 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hdevalence/ed25519consensus v0.2.0 // indirect
	github.com/jsimonetti/rtnetlink v1.4.0 // indirect
	github.com/kortschak/goroutine v1.1.2 // indirect
	github.com/mdlayher/netlink v1.7.3-0.20250113171957-fbb4dce95f42 // indirect
	github.com/mdlayher/socket v0.5.0 // indirect
	github.com/mitchellh/go-ps v1.0.0 // indirect
	github.com/tailscale/go-winio v0.0.0-20231025203758-c4f33415bf55 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	go4.org/mem v0.0.0-20240501181205-ae6ca9944745 // indirect
	go4.org/netipx v0.0.0-20231129151722-fdeea329fbba // indirect
	golang.org/x/crypto/x509roots/fallback v0.0.0-20250406160420-959f8f3db0fb // indirect
	golang.org/x/exp v0.0.0-20250210185358-939b2ce775ac // indirect
	golang.org/x/sync v0.15.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	golang.zx2c4.com/wireguard/windows v0.5.3 // indirect
)
