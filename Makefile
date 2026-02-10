.PHONY: run build clean tidy

APP_NAME=calendar-reg-main-api
BUILD_DIR=./bin

run:
	go run cmd/api/main.go

build:
	go build -o $(BUILD_DIR)/$(APP_NAME) cmd/api/main.go

clean:
	rm -rf $(BUILD_DIR)

tidy:
	go mod tidy
