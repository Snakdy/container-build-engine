.PHONY: build
build:
	go run main.go build --config fixtures/v1/pipeline-ordered.yaml --save /tmp/test.tar --v=10
	docker load < /tmp/test.tar