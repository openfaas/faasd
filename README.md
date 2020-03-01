# faasd - serverless with containerd and CNI  🐳

[![Build Status](https://travis-ci.com/openfaas/faasd.svg?branch=master)](https://travis-ci.com/openfaas/faasd)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![OpenFaaS](https://img.shields.io/badge/openfaas-serverless-blue.svg)](https://www.openfaas.com)

faasd is the same OpenFaaS experience and ecosystem, but without Kubernetes. Functions and microservices can be deployed anywhere with reduced overheads whilst retaining the portability of containers and cloud-native tooling.

## About faasd

* is a single Golang binary
* can be set-up and left alone to run your applications
* is multi-arch, so works on Intel `x86_64` and ARM out the box
* uses the same core components and ecosystem of OpenFaaS

![demo](https://pbs.twimg.com/media/EPNQz00W4AEwDxM?format=jpg&name=small)

> Demo of faasd running in KVM

## What does faasd deploy?

* faasd - itself, and its [faas-provider](https://github.com/openfaas/faas-provider) for containerd - CRUD for functions and services, implements the OpenFaaS REST API
* [Prometheus](https://github.com/prometheus/prometheus) - for monitoring of services, metrics, scaling and dashboards
* [OpenFaaS Gateway](https://github.com/openfaas/faas/tree/master/gateway) - the UI portal, CLI, and other OpenFaaS tooling can talk to this.
* [OpenFaaS queue-worker for NATS](https://github.com/openfaas/nats-queue-worker) - run your invocations in the background without adding any code. See also: [asynchronous invocations](https://docs.openfaas.com/reference/triggers/#async-nats-streaming)
* [NATS](https://nats.io) for asynchronous processing and queues

You'll also need:

* [CNI](https://github.com/containernetworking/plugins)
* [containerd](https://github.com/containerd/containerd)
* [runc](https://github.com/opencontainers/runc)

You can use the standard [faas-cli](https://github.com/openfaas/faas-cli) along with pre-packaged functions from *the Function Store*, or build your own using any OpenFaaS template.

## Tutorials

### Get started on DigitalOcean, or any other IaaS

If your IaaS supports `user_data` aka "cloud-init", then this guide is for you. If not, then checkout the approach and feel free to run each step manually.

* [Build a Serverless appliance with faasd](https://blog.alexellis.io/deploy-serverless-faasd-with-cloud-init/)

### Run locally on MacOS, Linux, or Windows with Multipass.run

* [Get up and running with your own faasd installation on your Mac/Ubuntu or Windows with cloud-config](https://gist.github.com/alexellis/6d297e678c9243d326c151028a3ad7b9)

### Get started on armhf / Raspberry Pi

You can run this tutorial on your Raspberry Pi, or adapt the steps for a regular Linux VM/VPS host.

* [faasd - lightweight Serverless for your Raspberry Pi](https://blog.alexellis.io/faasd-for-lightweight-serverless/)

### Terraform for DigitalOcean

Automate everything within < 60 seconds and get a public URL and IP address back. Customise as required, or adapt to your preferred cloud such as AWS EC2.

* [Provision faasd 0.7.5 on DigitalOcean with Terraform 0.12.0](https://gist.github.com/alexellis/fd618bd2f957eb08c44d086ef2fc3906)

### A note on private repos / registries

To use private image repos, `~/.docker/config.json` needs to be copied to `/var/lib/faasd/.docker/config.json`.

If you'd like to set up your own private registry, [see this tutorial](https://blog.alexellis.io/get-a-tls-enabled-docker-registry-in-5-minutes/).

Beware that running `docker login` on MacOS and Windows may create an empty file with your credentials stored in the system helper.

Alternatively, use you can use the `registry-login` command from the OpenFaaS Cloud bootstrap tool (ofc-bootstrap):

```bash
curl -sLSf https://raw.githubusercontent.com/openfaas-incubator/ofc-bootstrap/master/get.sh | sudo sh

ofc-bootstrap registry-login --username <your-registry-username> --password-stdin
# (the enter your password and hit return)
```
The file will be created in `./credentials/`

### Logs for functions

You can view the logs of functions using `journalctl`:

```bash
journalctl -t openfaas-fn:FUNCTION_NAME


faas-cli store deploy figlet
journalctl -t openfaas-fn:figlet -f &
echo logs | faas-cli invoke figlet
```

### Manual / developer instructions

See [here for manual / developer instructions](docs/DEV.md)

## Getting help

### Docs

The [OpenFaaS docs](https://docs.openfaas.com/) provide a wealth of information and are kept up to date with new features.

### Function and template store

For community functions see `faas-cli store --help`

For templates built by the community see: `faas-cli template store list`, you can also use the `dockerfile` template if you just want to migrate an existing service without the benefits of using a template.

### Workshop

[The OpenFaaS workshop](https://github.com/openfaas/workshop/) is a set of 12 self-paced labs and provides a great starting point

### Community support

An active community of almost 3000 users awaits you on Slack. Over 250 of those users are also contributors and help maintain the code.

* [Join Slack](https://slack.openfaas.io/)

## Backlog

### Supported operations

* `faas login`
* `faas up`
* `faas list`
* `faas describe`
* `faas deploy --update=true --replace=false`
* `faas invoke --async`
* `faas invoke`
* `faas rm`
* `faas store list/deploy/inspect`
* `faas version`
* `faas namespace`
* `faas secret`
* `faas logs`

Scale from and to zero is also supported. On a Dell XPS with a small, pre-pulled image unpausing an existing task took 0.19s and starting a task for a killed function took 0.39s. There may be further optimizations to be gained.

Other operations are pending development in the provider such as:

* `faas auth` - supported for Basic Authentication, but OAuth2 & OIDC require a patch

## Todo

Pending:

* [ ] Add support for using container images in third-party public registries
* [ ] Add support for using container images in private third-party registries
* [ ] Monitor and restart any of the core components at runtime if the container stops
* [ ] Bundle/package/automate installation of containerd - [see bootstrap from k3s](https://github.com/rancher/k3s)
* [ ] Provide ufw rules / example for blocking access to everything but a reverse proxy to the gateway container
* [ ] Provide [simple Caddyfile example](https://blog.alexellis.io/https-inlets-local-endpoints/) in the README showing how to expose the faasd proxy on port 80/443 with TLS

Done:

* [x] Provide a cloud-config.txt file for automated deployments of `faasd`
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

