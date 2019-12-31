Version := $(shell git describe --tags --dirty)
GitCommit := $(shell git rev-parse HEAD)
LDFLAGS := "-s -w -X main.Version=$(Version) -X main.GitCommit=$(GitCommit)"

.PHONY: all
all: local

local:
	CGO_ENABLED=0 GOOS=linux go build -o bin/faasd

.PHONY: dist
dist:
	CGO_ENABLED=0 GOOS=linux go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/faasd
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/faasd-armhf
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/faasd-arm64
