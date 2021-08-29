
GOPATH:=$(shell go env GOPATH)
.PHONY: init
init:
	go get -u github.com/joho/godotenv
	go get -u gorm.io/gorm
	go get -u gorm.io/driver/mysql
	go get -u github.com/go-kit/kit
	go get -u github.com/streadway/amqp
	
.PHONY: proto
proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative pb/driver.proto



.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -a -installsuffix cgo -ldflags '-s' -o bin/drivermanagement cmd/main.go

.PHONY: test
test:
	go test -v ./... -cover

.PHONY: docker
docker:
	docker build . -t drivermanagement:0.0.4 -t drivermanagement:latest
	docker tag drivermanagement:0.0.4 jadilet/drivermanagement:0.0.4
	docker tag drivermanagement:0.0.4 jadilet/drivermanagement:latest
	docker push jadilet/drivermanagement:0.0.4
	docker push jadilet/drivermanagement:latest
