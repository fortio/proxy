
test:
	go run -race . -version
	go run -race . -config sampleConfig/ -redirect-port :8080 -https-port :8443
