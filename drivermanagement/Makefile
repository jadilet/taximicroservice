
GOPATH:=$(shell go env GOPATH)
.PHONY: init
init:
	go get -u github.com/joho/godotenv
	go get -u gorm.io/gorm
	go get -u gorm.io/driver/mysql
	go get -u github.com/go-kit/kit
	go get -u github.com/streadway/amqp
	
.PHONY: build

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -a -installsuffix cgo -ldflags '-s' -o bin/drivermanagement cmd/main.go

.PHONY: test
test:
	go test -v ./... -cover

.PHONY: docker
docker:
	docker build . -t drivermanagement:latest
