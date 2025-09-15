APP_NAME=antibruteforce
CMD_DIR=./cmd/server
BUILD_DIR=./bin

.PHONY: build run test clean redis start clean docker stop_docker integration_test 

build:
	go build -o $(BUILD_DIR)/$(APP_NAME) $(CMD_DIR)

start_redis:
	docker run -d --name redis-6380 -p 6380:6379 redis:7.2

stop_redis:
	docker stop redis-6380

run: start_redis
	REDIS_ADDR=127.0.0.1:6380 PORT=8888 go run $(CMD_DIR)

start: build start_redis
	REDIS_ADDR=127.0.0.1:6380 $(BUILD_DIR)/$(APP_NAME)

test:
	go test ./... -v -race -count 10

clean:
	rm -rf $(BUILD_DIR)
	go clean

docker:
	docker compose up --build -d

stop_docker:
	docker compose down

integration_test: docker
	go test -v ./tests/integration_test.go
	@make stop_docker

lint:
	golangci-lint run

CI: clean build lint test integration_test