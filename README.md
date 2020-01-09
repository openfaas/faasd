# faasd - serverless with containerd

[![Build Status](https://travis-ci.com/alexellis/faasd.svg?branch=master)](https://travis-ci.com/alexellis/faasd)

faasd is a Golang supervisor that bundles OpenFaaS for use with containerd instead of a container orchestrator like Kubernetes or Docker Swarm.

## About faasd:

* faasd is a single Golang binary
* faasd is multi-arch, so works on `x86_64`, armhf and arm64
* faasd downloads, starts and supervises the core components to run OpenFaaS

## What does faasd deploy?

* [faas-containerd](https://github.com/alexellis/faas-containerd/)
* [Prometheus](https://github.com/prometheus/prometheus)
* [the OpenFaaS gateway](https://github.com/openfaas/faas/tree/master/gateway)

You can use the standard [faas-cli](https://github.com/openfaas/faas-cli) with faasd along with pre-packaged functions in the Function Store, or build your own with the template store.

### faas-containerd supports:

* `faas list`
* `faas describe` 
* `faas deploy --update=true --replace=false`
* `faas invoke`
* `faas invoke --async`

Other operations are pending development in the provider.

### Pre-reqs

* Linux

    PC / Cloud - any Linux that containerd works on should be fair game, but faasd is tested with Ubuntu 18.04

    For Raspberry Pi Raspbian Stretch or newer also works fine

    For MacOS users try [multipass.run](https://multipass.run) or [Vagrant](https://www.vagrantup.com/)

    For Windows users, install [Git Bash](https://git-scm.com/downloads) along with multipass or vagrant. You can also use WSL1 or WSL2 which provides a Linux environment.

* Installation steps as per [faas-containerd](https://github.com/alexellis/faas-containerd) for building and for development
    * [containerd v1.3.2](https://github.com/containerd/containerd)
    * [CNI plugins v0.8.4](https://github.com/containernetworking/plugins)

* [faas-cli](https://github.com/openfaas/faas-cli) (optional)

## Backlog

Pending:

* [ ] Monitor and restart any of the core components at runtime if the container stops
* [ ] Bundle/package/automate installation of containerd - [see bootstrap from k3s](https://github.com/rancher/k3s)
* [ ] Provide ufw rules / example for blocking access to everything but a reverse proxy to the gateway container
* [ ] Provide [simple Caddyfile example](https://blog.alexellis.io/https-inlets-local-endpoints/) in the README showing how to expose the faasd proxy on port 80/443 with TLS

Done:

* [x] Inject / manage IPs between core components for service to service communication - i.e. so Prometheus can scrape the OpenFaaS gateway - done via `/etc/hosts` mount
* [x] Add queue-worker and NATS
* [x] Create faasd.service and faas-containerd.service
* [x] Self-install / create systemd service via `faasd install`
* [x] Restart containers upon restart of faasd
* [x] Clear / remove containers and tasks with SIGTERM / SIGINT
* [x] Determine armhf/arm64 containers to run for gateway
* [x] Configure `basic_auth` to protect the OpenFaaS gateway and faas-containerd HTTP API
* [x] Setup custom working directory for faasd `/run/faasd/`
* [x] Use CNI to create network namespaces and adapters

## Tutorial: Get started on armhf / Raspberry Pi

You can run this tutorial on your Raspberry Pi, or adapt the steps for a regular Linux VM/VPS host.

* [faasd - lightweight Serverless for your Raspberry Pi](https://blog.alexellis.io/faasd-for-lightweight-serverless/)

## Hacking (build from source)

Install the CNI plugins:

```sh
export CNI_VERSION=v0.8.4
```

* For PC run `export ARCH=amd64`
* For RPi/armhf run `export ARCH=arm`
* For arm64 run `export ARCH=arm64`

Then run:

```sh
mkdir -p /opt/cni/bin
curl -sSL https://github.com/containernetworking/plugins/releases/download/${CNI_VERSION}/cni-plugins-linux-${ARCH}-${CNI_VERSION}.tgz | tar -xz -C /opt/cni/bin
```

First run faas-containerd

```sh
cd $GOPATH/src/github.com/alexellis/faas-containerd

# You'll need to install containerd and its pre-reqs first
# https://github.com/alexellis/faas-containerd/

sudo ./faas-containerd
```

Then run faasd, which brings up the gateway and Prometheus as containers

```sh
cd $GOPATH/src/github.com/alexellis/faasd
go build

# Install with systemd
# sudo ./faasd install

# Or run interactively
# sudo ./faasd up
```

### Build and run (binaries)

```sh
# For x86_64
sudo curl -fSLs "https://github.com/alexellis/faasd/releases/download/0.4.4/faasd" \
    -o "/usr/local/bin/faasd" \
    && sudo chmod a+x "/usr/local/bin/faasd"

# armhf
sudo curl -fSLs "https://github.com/alexellis/faasd/releases/download/0.4.4/faasd-armhf" \
    -o "/usr/local/bin/faasd" \
    && sudo chmod a+x "/usr/local/bin/faasd"

# arm64
sudo curl -fSLs "https://github.com/alexellis/faasd/releases/download/0.4.4/faasd-arm64" \
    -o "/usr/local/bin/faasd" \
    && sudo chmod a+x "/usr/local/bin/faasd"
```

### At run-time

Look in `hosts` in the current working folder or in `/run/faasd/` to get the IP for the gateway or Prometheus

```sh
127.0.0.1       localhost
10.62.0.1      faas-containerd

10.62.0.2      prometheus
10.62.0.3      gateway
10.62.0.4      nats
10.62.0.5      queue-worker
```

The IP addresses are dynamic and may change on every launch.

Since faas-containerd uses containerd heavily it is not running as a container, but as a stand-alone process. Its port is available via the bridge interface, i.e. openfaas0.

* Prometheus will run on the Prometheus IP plus port 8080 i.e. http://[prometheus_ip]:9090/targets

* faas-containerd runs on 10.62.0.1:8081

* Now go to the gateway's IP address as shown above on port 8080, i.e. http://[gateway_ip]:8080 - you can also use this address to deploy OpenFaaS Functions via the `faas-cli`. 

* basic-auth

    You will then need to get the basic-auth password, it is written to `/run/faasd/secrets/basic-auth-password` if you followed the above instructions.
The default Basic Auth username is `admin`, which is written to `/run/faasd/secrets/basic-auth-user`, if you wish to use a non-standard user then create this file and add your username (no newlines or other characters) 

#### Installation with systemd

* `faasd install` - install faasd and containerd with systemd, this must be run from `$GOPATH/src/github.com/alexellis/faasd`
* `journalctl -u faasd` - faasd systemd logs
* `journalctl -u faas-containerd` - faas-containerd systemd logs

### Appendix

#### Links

https://github.com/renatofq/ctrofb/blob/31968e4b4893f3603e9998f21933c4131523bb5d/cmd/network.go

https://github.com/renatofq/catraia/blob/c4f62c86bddbfadbead38cd2bfe6d920fba26dce/catraia-net/network.go

https://github.com/containernetworking/plugins

https://github.com/containerd/go-cni

