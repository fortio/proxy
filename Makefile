
test:
	go run -race . -version
	go run -race . -config sampleConfig/ -port :8443
