
test:
	go test -race ./...
	go test -race -tags no_tailscale ./...
	go run -race . version

test-local:
	go run -race . -h2 -config-dir sampleConfig/ -redirect-port :8081 -https-port :8443 -http-port :8001


docker-test:
	GOOS=linux go build -tags no_tailscale
	docker build . --tag fortio/proxy:test
	docker run -v `pwd`/sampleConfig:/etc/fortio-proxy-config fortio/proxy:test

dev-prefix:
	go run -race . -h2 -http-port 8001 -https-port disabled -redirect-port disabled\
		-loglevel debug \
		-routes.json '[{"prefix":"/fgrpc", "destination":"http://localhost:8079/"}, {"host":"*", "destination":"http://localhost:8080/"}]'

dev-prefix-only:
	go run -race . -http-port 8001 -https-port disabled -redirect-port disabled\
		-loglevel debug \
		-routes.json '[{"prefix":"/debug", "destination":"http://localhost:8080/"}]'

dev-grpc:
	go run -race . -h2 -http-port 8001 -https-port disabled -redirect-port disabled\
		-loglevel debug \
		-routes.json '[{"host":"*", "destination":"http://localhost:8079/"}]'

dev-h2c:
	go run -race . -h2 -http-port 8001 -https-port disabled -redirect-port disabled\
		-debug-host "debug.fortio.org" \
		 -routes.json '[{"host":"*", "destination":"http://localhost:8080/"}]'

TAILSCALE_SERVERNAME=$(shell tailscale status --json | jq -r '.Self.DNSName | sub("\\.$$"; "")')
dev-tailscale:
	@echo "Visit https://$(TAILSCALE_SERVERNAME)/"
	go run -race . -loglevel debug -hostid local -certs-domains $(TAILSCALE_SERVERNAME) -debug-host $(TAILSCALE_SERVERNAME)

dev:
	# Run: curl -H "Host: debug.fortio.org" http://localhost:8001/debug
	# and curl -H "Host: debug.fortio.org" http://localhost:8000/foo (no redirect with that host header)
	go run -race . -http-port 8001 -https-port disabled -redirect-port 8000 -hostid "$(shell hostname)-test" \
		-debug-host "debug.fortio.org" -routes.json '[{"host":"*", "destination":"http://localhost:8080/"}]'

lint: .golangci.yml
	golangci-lint run

.golangci.yml: Makefile
	curl -fsS -o .golangci.yml https://raw.githubusercontent.com/fortio/workflows/main/golangci.yml

.PHONY: lint
