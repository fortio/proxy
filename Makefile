
test:
	go test -race ./...
	go run -race . -version

test-local:
	go run -race . -h2 -config sampleConfig/ -redirect-port :8081 -https-port :8443 -http-port :8001


docker-test:
	GOOS=linux go build
	docker build . --tag fortio/proxy:test
	docker run -v `pwd`/sampleConfig:/etc/fortio-proxy-config fortio/proxy:test

dev-prefix:
	go run -race . -h2 -http-port 8001 -https-port disabled -redirect-port disabled\
		-loglevel debug \
		-routes.json '[{"prefix":"/fgrpc", "destination":"http://localhost:8079/"}, {"host":"*", "destination":"http://localhost:8080/"}]'

dev-grpc:
	go run -race . -h2 -http-port 8001 -https-port disabled -redirect-port disabled\
		-loglevel debug \
		-routes.json '[{"host":"*", "destination":"http://localhost:8079/"}]'

dev-h2c:
	go run -race . -h2 -http-port 8001 -https-port disabled -redirect-port disabled\
		-debug-host "debug.fortio.org" \
		 -routes.json '[{"host":"*", "destination":"http://localhost:8080/"}]'

dev:
	# Run: curl -H "Host: debug.fortio.org" http://localhost:8001/debug
	# and curl -H "Host: debug.fortio.org" http://localhost:8000/foo (no redirect with that host header)
	go run -race . -http-port 8001 -https-port disabled -redirect-port 8000 -hostid "$(shell hostname)-test" \
		-debug-host "debug.fortio.org" -routes.json '[{"host":"*", "destination":"http://localhost:8080/"}]'
