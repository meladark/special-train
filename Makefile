APP_NAME=antibruteforce
CMD_DIR=./cmd/server
BUILD_DIR=./bin

.PHONY: build run test clean

build:
	go build -o $(BUILD_DIR)/$(APP_NAME) $(CMD_DIR)

run:
	PORT=8888 go run $(CMD_DIR)

start: build
	$(BUILD_DIR)/$(APP_NAME)

test:
	go test ./... -v

clean:
	rm -rf $(BUILD_DIR)
	go clean
