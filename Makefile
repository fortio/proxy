
test:
	go test -race ./...
	go run -race . -version
	go run -race . -config sampleConfig/ -redirect-port :8080 -https-port :8443

dev-grpc:
	go run -race . -http-port 8001 -https-port disabled -redirect-port disabled\
		 -routes.json '[{"host":"*", "destination":"h2c://localhost:8079/"}]'

dev-h2c:
	go run -race . -http-port 8001 -https-port disabled -redirect-port disabled\
		 -routes.json '[{"host":"*", "destination":"h2c://localhost:8080/"}]'

dev:
	go run -race . -http-port 8001 -https-port disabled -redirect-port disabled\
		 -routes.json '[{"host":"*", "destination":"http://localhost:8080/"}]'
