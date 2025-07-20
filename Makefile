APP_NAME := virginizer

BIN := $(APP_NAME)
BIN_MAC := $(BIN)_darwin
BIN_LINUX := $(BIN)_linux
BIN_WINDOWS := $(BIN).exe

EXE := $(APP_NAME)
IMAGE := $(APP_NAME)
TARGET := distr
VERSION := $(shell git describe --tags --always --dirty)
TAG := $(VERSION)
PWD := $(shell pwd)
NOW := $(shell date +"%m-%d-%Y")

build:
	go test -cover ./...
	go build -o $(BIN) -v -ldflags "-X main.Version=$(VERSION) -X main.Build=$(NOW)"

distr: build
	rm -rf $(TARGET)
	mkdir $(TARGET)
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o $(TARGET)/$(BIN_LINUX) -v -ldflags "-X main.Version=$(VERSION) -X main.Build=$(NOW)"
	gzip $(TARGET)/$(BIN_LINUX)
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o $(TARGET)/$(BIN_MAC) -v -ldflags "-X main.Version=$(VERSION) -X main.Build=$(NOW)"
	gzip $(TARGET)/$(BIN_MAC)
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ go build -o $(TARGET)/$(BIN_WINDOWS) -v -ldflags "-X main.Version=$(VERSION) -X main.Build=$(NOW)"
	gzip $(TARGET)/$(BIN_WINDOWS)

test:
	go test ./...

clean:
	go clean
	rm -f $(BIN)
	rm -f $(BIN_MAC)
	rm -f $(BIN_LINUX)
	rm -f $(BIN_WINDOWS)
