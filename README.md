# faasd - serverless with containerd

[![Build Status](https://travis-ci.com/openfaas/faasd.svg?branch=master)](https://travis-ci.com/openfaas/faasd)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![OpenFaaS](https://img.shields.io/badge/openfaas-serverless-blue.svg)](https://www.openfaas.com)

faasd is a Golang supervisor that bundles OpenFaaS for use with containerd instead of a container orchestrator like Kubernetes or Docker Swarm.

## About faasd:

* faasd is a single Golang binary
* faasd is multi-arch, so works on `x86_64`, armhf and arm64
* faasd downloads, starts and supervises the core components to run OpenFaaS

![demo](https://pbs.twimg.com/media/EPNQz00W4AEwDxM?format=jpg&name=small)

> Demo of faasd running in KVM

## What does faasd deploy?

* faasd - itself, and its [faas-provider](https://github.com/openfaas/faas-provider)
* [Prometheus](https://github.com/prometheus/prometheus)
* [the OpenFaaS gateway](https://github.com/openfaas/faas/tree/master/gateway)

You can use the standard [faas-cli](https://github.com/openfaas/faas-cli) with faasd along with pre-packaged functions in the Function Store, or build your own with the template store.

## Tutorials

### Get started on DigitalOcean or with cloud-init

* [Build a Serverless appliance with cloud-init and faasd](https://blog.alexellis.io/deploy-serverless-faasd-with-cloud-init/)

### Run locally on MacOS, Linux, or Windows with Multipass.run

* [Get up and running with your own faasd installation on your Mac/Ubuntu or Windows with cloud-config](https://gist.github.com/alexellis/6d297e678c9243d326c151028a3ad7b9)

### Get started on armhf / Raspberry Pi

You can run this tutorial on your Raspberry Pi, or adapt the steps for a regular Linux VM/VPS host.

* [faasd - lightweight Serverless for your Raspberry Pi](https://blog.alexellis.io/faasd-for-lightweight-serverless/)

### Manual / developer instructions

See [here for manual / developer instructions](docs/DEV.md)

## Backlog

### faasd supports:

* `faas list`
* `faas describe` 
* `faas deploy --update=true --replace=false`
* `faas invoke`
* `faas rm`
* `faas login`
* `faas store list/deploy/inspect`
* `faas up`
* `faas version`
* `faas invoke --async`
* `faas namespace`

Scale from and to zero is also supported. On a Dell XPS with a small, pre-pulled image unpausing an existing task took 0.19s and starting a task for a killed function took 0.39s. There may be further optimizations to be gained.

Other operations are pending development in the provider such as:

* `faas logs`
* `faas secret`
* `faas auth`

## Todo

Pending:

* [ ] Add support for using container images in third-party public registries
* [ ] Add support for using container images in private third-party registries
* [ ] Monitor and restart any of the core components at runtime if the container stops
* [ ] Bundle/package/automate installation of containerd - [see bootstrap from k3s](https://github.com/rancher/k3s)
* [ ] Provide ufw rules / example for blocking access to everything but a reverse proxy to the gateway container
* [ ] Provide [simple Caddyfile example](https://blog.alexellis.io/https-inlets-local-endpoints/) in the README showing how to expose the faasd proxy on port 80/443 with TLS

Done:

* [x] Inject / manage IPs between core components for service to service communication - i.e. so Prometheus can scrape the OpenFaaS gateway - done via `/etc/hosts` mount
* [x] Add queue-worker and NATS
* [x] Create faasd.service and faasd-provider.service
* [x] Self-install / create systemd service via `faasd install`
* [x] Restart containers upon restart of faasd
* [x] Clear / remove containers and tasks with SIGTERM / SIGINT
* [x] Determine armhf/arm64 containers to run for gateway
* [x] Configure `basic_auth` to protect the OpenFaaS gateway and faasd-provider HTTP API
* [x] Setup custom working directory for faasd `/var/lib/faasd/`
* [x] Use CNI to create network namespaces and adapters

## Appendix

### Links

https://github.com/renatofq/ctrofb/blob/31968e4b4893f3603e9998f21933c4131523bb5d/cmd/network.go

https://github.com/renatofq/catraia/blob/c4f62c86bddbfadbead38cd2bfe6d920fba26dce/catraia-net/network.go

https://github.com/containernetworking/plugins

https://github.com/containerd/go-cni

