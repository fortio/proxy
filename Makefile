
test:
	go run -race . -version
	go run -race . -config sampleConfig/ -redirect-port :8080 -https-port :8443

dev:
	go run -race . -http-port 8001 -https-port disabled -redirect-port disabled -routes.json '[{"host":"*", "destination":"http://localhost:8079/"}]'
