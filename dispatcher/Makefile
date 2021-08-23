
GOPATH:=$(shell go env GOPATH)
.PHONY: init
init:
	go get -u github.com/joho/godotenv
	go get -u github.com/go-kit/kit
	go get -u github.com/streadway/amqp
	
.PHONY: build

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -a -installsuffix cgo -ldflags '-s' -o bin/dispatcher cmd/main.go

.PHONY: test
test:
	go test -v ./... -cover

.PHONY: docker
docker:
	docker build . -t dispatcher:0.0.2
	docker tag dispatcher:0.0.2 jadilet/dispatcher:0.0.2
	docker push jadilet/dispatcher:0.0.2
