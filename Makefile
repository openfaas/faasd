Version := $(shell git describe --tags --dirty)
GitCommit := $(shell git rev-parse HEAD)
LDFLAGS := "-s -w -X main.Version=$(Version) -X main.GitCommit=$(GitCommit)"
CONTAINERD_VER := 1.6.8
CNI_VERSION := v0.9.1
ARCH := amd64

export GO111MODULE=on

.PHONY: all
all: test dist hashgen

.PHONY: publish
publish: dist hashgen

local:
	CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o bin/faasd

.PHONY: test
test:
	CGO_ENABLED=0 GOOS=linux go test -mod=vendor -ldflags $(LDFLAGS) ./...

.PHONY: dist
dist:
	CGO_ENABLED=0 GOOS=linux go build -mod=vendor -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/faasd
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -mod=vendor -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/faasd-armhf
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -mod=vendor -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/faasd-arm64

.PHONY: hashgen
hashgen:
	for f in bin/faasd*; do shasum -a 256 $$f > $$f.sha256; done

.PHONY: prepare-test
prepare-test:
	curl -sLSf https://github.com/containerd/containerd/releases/download/v$(CONTAINERD_VER)/containerd-$(CONTAINERD_VER)-linux-amd64.tar.gz > /tmp/containerd.tar.gz && sudo tar -xvf /tmp/containerd.tar.gz -C /usr/local/bin/ --strip-components=1
	curl -SLfs https://raw.githubusercontent.com/containerd/containerd/v1.6.8/containerd.service | sudo tee /etc/systemd/system/containerd.service
	sudo systemctl daemon-reload && sudo systemctl start containerd
	sudo /sbin/sysctl -w net.ipv4.conf.all.forwarding=1
	sudo mkdir -p /opt/cni/bin
	curl -sSL https://github.com/containernetworking/plugins/releases/download/$(CNI_VERSION)/cni-plugins-linux-$(ARCH)-$(CNI_VERSION).tgz | sudo tar -xz -C /opt/cni/bin
	sudo cp bin/faasd /usr/local/bin/
	sudo /usr/local/bin/faasd install
	sudo systemctl status -l containerd --no-pager
	sudo journalctl -u faasd-provider --no-pager
	sudo systemctl status -l faasd-provider --no-pager
	sudo systemctl status -l faasd --no-pager
	curl -sSLf https://cli.openfaas.com | sudo sh
	echo "Sleeping for 2m" && sleep 120 && sudo journalctl -u faasd --no-pager

.PHONY: test-e2e
test-e2e:
	sudo cat /var/lib/faasd/secrets/basic-auth-password | /usr/local/bin/faas-cli login --password-stdin
	/usr/local/bin/faas-cli store deploy figlet --env write_timeout=1s --env read_timeout=1s --label testing=true
	sleep 5
	/usr/local/bin/faas-cli list -v
	/usr/local/bin/faas-cli describe figlet | grep testing
	uname | /usr/local/bin/faas-cli invoke figlet
	uname | /usr/local/bin/faas-cli invoke figlet --async
	sleep 10
	/usr/local/bin/faas-cli list -v
	/usr/local/bin/faas-cli remove figlet
	sleep 3
	/usr/local/bin/faas-cli list
	sleep 3
	journalctl -t openfaas-fn:figlet --no-pager

# Removed due to timing issue in CI on GitHub Actions
#	/usr/local/bin/faas-cli logs figlet --since 15m --follow=false | grep Forking

verify-compose:
	@echo Verifying docker-compose.yaml images in remote registries && \
	arkade chart verify --verbose=$(VERBOSE) -f ./docker-compose.yaml