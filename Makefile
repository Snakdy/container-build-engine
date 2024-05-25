.PHONY: build test

build:
	go run main.go build --config fixtures/v1/pipeline-ubi.yaml --save /tmp/test.tar --v=10
	docker load < /tmp/test.tar
	docker inspect image

test:
	go test ./...