include $(CURDIR)/hack/multipass.mk

Version := $(shell git describe --tags --dirty)
GitCommit := $(shell git rev-parse HEAD)
LDFLAGS := "-s -w -X main.Version=$(Version) -X main.GitCommit=$(GitCommit)"
CONTAINERD_VER := 1.3.2
CNI_VERSION := v0.8.5
ARCH := amd64
GO_VERSION := 1.12.1
GO := /usr/local/go/bin/go
GCC := /usr/bin/gcc

$(GO):
	wget --quiet https://dl.google.com/go/go$(GO_VERSION).linux-amd64.tar.gz
	sudo tar -xvf go$(GO_VERSION).linux-amd64.tar.gz
	sudo mv go /usr/local

$(GCC):
	sudo apt update
	sudo apt install -y gcc

##########################################################
##@ MAIN
##########################################################
.PHONY: all dist prepare-test test-e2e

all: ## build all
all: local

local: ## build local binary
	CGO_ENABLED=0 GOOS=linux $(GO) build -o bin/faasd

dist: ## Package for distribution
	CGO_ENABLED=0 GOOS=linux go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/faasd
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/faasd-armhf
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/faasd-arm64

install: $(GO) $(GCC) local
	sudo mv bin/faasd /usr/local/bin/faasd
	sudo systemctl restart faasd

prepare-test:
	curl -sLSf https://github.com/containerd/containerd/releases/download/v$(CONTAINERD_VER)/containerd-$(CONTAINERD_VER).linux-amd64.tar.gz > /tmp/containerd.tar.gz && sudo tar -xvf /tmp/containerd.tar.gz -C /usr/local/bin/ --strip-components=1
	curl -SLfs https://raw.githubusercontent.com/containerd/containerd/v1.3.2/containerd.service | sudo tee /etc/systemd/system/containerd.service
	sudo systemctl daemon-reload && sudo systemctl start containerd
	sudo /sbin/sysctl -w net.ipv4.conf.all.forwarding=1
	sudo mkdir -p /opt/cni/bin
	curl -sSL https://github.com/containernetworking/plugins/releases/download/$(CNI_VERSION)/cni-plugins-linux-$(ARCH)-$(CNI_VERSION).tgz | sudo tar -xz -C /opt/cni/bin
	sudo cp $(GOPATH)/src/github.com/alexellis/faasd/bin/faasd /usr/local/bin/
	cd $(GOPATH)/src/github.com/alexellis/faasd/ && sudo /usr/local/bin/faasd install
	sudo systemctl status -l containerd --no-pager
	sudo journalctl -u faasd-provider --no-pager
	sudo systemctl status -l faasd-provider --no-pager
	sudo systemctl status -l faasd --no-pager
	curl -sSLf https://cli.openfaas.com | sudo sh
	sleep 120 && sudo journalctl -u faasd --no-pager

test-e2e: ## run e2e tests
	sudo cat /var/lib/faasd/secrets/basic-auth-password | /usr/local/bin/faas-cli login --password-stdin
	/usr/local/bin/faas-cli store deploy figlet --env write_timeout=1s --env read_timeout=1s
	sleep 2
	/usr/local/bin/faas-cli list -v
	uname | /usr/local/bin/faas-cli invoke figlet
	uname | /usr/local/bin/faas-cli invoke figlet --async
	sleep 10
	/usr/local/bin/faas-cli list -v
	/usr/local/bin/faas-cli remove figlet
	sleep 3
	/usr/local/bin/faas-cli list

##########################################################
##@ UTIL
##########################################################
.PHONY: help

help: ## show help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m 	%s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)