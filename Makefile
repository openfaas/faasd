Version := $(shell git describe --tags --dirty)
GitCommit := $(shell git rev-parse HEAD)
LDFLAGS := "-s -w -X main.Version=$(Version) -X main.GitCommit=$(GitCommit)"
CONTAINERD_VER := 1.3.2
FAASC_VER := 0.4.0

.PHONY: all
all: local

local:
	CGO_ENABLED=0 GOOS=linux go build -o bin/faasd

.PHONY: dist
dist:
	CGO_ENABLED=0 GOOS=linux go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/faasd
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/faasd-armhf
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/faasd-arm64

.PHONY: prepare-test
prepare-test:
	curl -sLSf https://github.com/containerd/containerd/releases/download/v$(CONTAINERD_VER)/containerd-$(CONTAINERD_VER).linux-amd64.tar.gz > /tmp/containerd.tar.gz && sudo tar -xvf /tmp/containerd.tar.gz -C /usr/local/bin/ --strip-components=1
	curl -SLfs https://raw.githubusercontent.com/containerd/containerd/v1.3.2/containerd.service | sudo tee /etc/systemd/system/containerd.service
	sudo systemctl daemon-reload && sudo systemctl start containerd
	sudo curl -fSLs "https://github.com/genuinetools/netns/releases/download/v0.5.3/netns-linux-amd64" --output "/usr/local/bin/netns" && sudo chmod a+x "/usr/local/bin/netns"
	sudo /sbin/sysctl -w net.ipv4.conf.all.forwarding=1
	sudo curl -sSLf "https://github.com/alexellis/faas-containerd/releases/download/$(FAASC_VER)/faas-containerd" --output "/usr/local/bin/faas-containerd" && sudo chmod a+x "/usr/local/bin/faas-containerd" || :
	sudo cp $(GOPATH)/src/github.com/alexellis/faasd/bin/faasd /usr/local/bin/
	cd $(GOPATH)/src/github.com/alexellis/faasd/ && sudo /usr/local/bin/faasd install
	sudo systemctl status -l containerd --no-pager
	sudo journalctl -u faas-containerd --no-pager
	sudo systemctl status -l faas-containerd --no-pager
	sudo systemctl status -l faasd --no-pager
	curl -sSLf https://cli.openfaas.com | sudo sh
	sleep 120 && sudo journalctl -u faasd --no-pager

.PHONY: test-e2e
test-e2e:
	sudo cat /run/faasd/secrets/basic-auth-password | /usr/local/bin/faas-cli login --password-stdin
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
