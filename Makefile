
test:
	go test -race ./...
	go run -race . -version
	go run -race . -h2 -config sampleConfig/ -redirect-port :8081 -https-port :8443 -http-port :8001


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
		 -routes.json '[{"host":"*", "destination":"http://localhost:8080/"}]'

dev:
	go run -race . -http-port 8001 -https-port disabled -redirect-port disabled\
		 -routes.json '[{"host":"*", "destination":"http://localhost:8080/"}]'
