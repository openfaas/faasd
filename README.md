# faasd - a lightweight & portable faas engine

[![Build Status](https://github.com/openfaas/faasd/workflows/build/badge.svg?branch=master)](https://github.com/openfaas/faasd/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![OpenFaaS](https://img.shields.io/badge/openfaas-serverless-blue.svg)](https://www.openfaas.com)
![Downloads](https://img.shields.io/github/downloads/openfaas/faasd/total)

faasd is [OpenFaaS](https://github.com/openfaas/) reimagined, but without the cost and complexity of Kubernetes. It runs on a single host with very modest requirements, making it fast and easy to manage. Under the hood it uses [containerd](https://containerd.io/) and [Container Networking Interface (CNI)](https://github.com/containernetworking/cni) along with the same core OpenFaaS components from the main project.

## When should you use faasd over OpenFaaS on Kubernetes?

* You have a cost sensitive project - run faasd on a 5-10 USD VPS or on your Raspberry Pi
* When you just need a few functions or microservices, without the cost of a cluster
* When you don't have the bandwidth to learn or manage Kubernetes
* To deploy embedded apps in IoT and edge use-cases
* To shrink-wrap applications for use with a customer or client

faasd does not create the same maintenance burden you'll find with maintaining, upgrading, and securing a Kubernetes cluster. You can deploy it and walk away, in the worst case, just deploy a new VM and deploy your functions again.

## About faasd

* is a single Golang binary
* uses the same core components and ecosystem of OpenFaaS
* is multi-arch, so works on Intel `x86_64` and ARM out the box
* can be set-up and left alone to run your applications

![demo](https://pbs.twimg.com/media/EPNQz00W4AEwDxM?format=jpg&name=small)

> Demo of faasd running in KVM

## Walk-through of faasd

faasd is OpenFaaS, so most things work the same, but you will need to pick one of the guides in the section below for deployment.

* For reference: [OpenFaaS docs](https://docs.openfaas.com)
* For use-cases and tutorials: [OpenFaaS blog](https://openfaas.com/blog/)
* For self-paced learning: [OpenFaaS workshop](https://github.com/openfaas/workshop/)

## Deploy faasd

The easiest way to deploy faasd is with cloud-init, we give several examples below, and post IaaS platforms will accept "user-data" pasted into their UI, or via their API.

If you don't use cloud-init, or have already created your Linux server you can use the installation script. This approach also works for Raspberry Pi:

```bash
git clone https://github.com/openfaas/faasd
cd faasd

./hack/install.sh
```

For trying out fasad on MacOS or Windows, we recommend using multipass and its cloud-init option.

### Run locally on MacOS, Linux, or Windows with multipass

* [Get up and running with your own faasd installation on your Mac/Ubuntu or Windows with cloud-config](/docs/MULTIPASS.md)

### DigitalOcean tutorial with Terraform and TLS

The terraform can be adapted for any IaaS provider:

* [Bring a lightweight Serverless experience to DigitalOcean with Terraform and faasd](https://www.openfaas.com/blog/faasd-tls-terraform/)

See also: [Build a Serverless appliance with faasd and cloud-init](https://blog.alexellis.io/deploy-serverless-faasd-with-cloud-init/)

### Get started on armhf / Raspberry Pi

You can run this tutorial on your Raspberry Pi, or adapt the steps for a regular Linux VM/VPS host.

* [faasd - lightweight Serverless for your Raspberry Pi](https://blog.alexellis.io/faasd-for-lightweight-serverless/)

### Terraform for DigitalOcean

Automate everything within < 60 seconds and get a public URL and IP address back. Customise as required, or adapt to your preferred cloud such as AWS EC2.

* [Provision faasd 0.9.10 on DigitalOcean with Terraform 0.12.0](docs/bootstrap/README.md)

* [Provision faasd on DigitalOcean with built-in TLS support](docs/bootstrap/digitalocean-terraform/README.md)

## Operational concerns

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

> Note for the GitHub container registry, you should use `ghcr.io` Container Registry and not the previous generation of "Docker Package Registry". [See notes on migrating](https://docs.github.com/en/free-pro-team@latest/packages/getting-started-with-github-container-registry/migrating-to-github-container-registry-for-docker-images)

### Logs for functions

You can view the logs of functions using `journalctl`:

```bash
journalctl -t openfaas-fn:FUNCTION_NAME


faas-cli store deploy figlet
journalctl -t openfaas-fn:figlet -f &
echo logs | faas-cli invoke figlet
```

### Logs for the core services

Core services as defined in the docker-compose.yaml file are deployed as containers by faasd.

View the logs for a component by giving its NAME:

```bash
journalctl -t default:NAME

journalctl -t default:gateway

journalctl -t default:queue-worker
```

You can also use `-f` to follow the logs, or `--lines` to tail a number of lines, or `--since` to give a timeframe.

### Exposing core services

The OpenFaaS stack is made up of several core services including NATS and Prometheus. You can expose these through the `docker-compose.yaml` file located at `/var/lib/faasd`.

Expose the gateway to all adapters:

```yaml
  gateway:
    ports:
       - "8080:8080"
```

Expose Prometheus only to 127.0.0.1:

```yaml
  prometheus:
    ports:
       - "127.0.0.1:9090:9090"
```

### Upgrading faasd

To upgrade `faasd` either re-create your VM using Terraform, or simply replace the faasd binary with a newer one.

```bash
systemctl stop faasd-provider
systemctl stop faasd

# Replace /usr/local/bin/faasd with the desired release

# Replace /var/lib/faasd/docker-compose.yaml with the matching version for
# that release.
# Remember to keep any custom patches you make such as exposing additional 
# ports, or updating timeout values

systemctl start faasd
systemctl start faasd-provider
```

You could also perform this task over SSH, or use a configuration management tool.

> Note: if you are using Caddy or Let's Encrypt for free SSL certificates, that you may hit rate-limits for generating new certificates if you do this too often within a given week.

### Memory limits for functions

Memory limits for functions are supported. When the limit is exceeded the function will be killed.

Example:

```yaml
functions:
  figlet:
    skip_build: true
    image: functions/figlet:latest
    limits:
      memory: 20Mi
```

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

### Manual / developer instructions

See [here for manual / developer instructions](docs/DEV.md)

## Getting help

### Docs

The [OpenFaaS docs](https://docs.openfaas.com/) provide a wealth of information and are kept up to date with new features.

### Function and template store

For community functions see `faas-cli store --help`

For templates built by the community see: `faas-cli template store list`, you can also use the `dockerfile` template if you just want to migrate an existing service without the benefits of using a template.

### Training and courses

#### LinuxFoundation training course

The founder of faasd and OpenFaaS has written a training course for the LinuxFoundation which also covers how to use OpenFaaS on Kubernetes. Much of the same concepts can be applied to faasd, and the course is free:

* [Introduction to Serverless on Kubernetes](https://www.edx.org/course/introduction-to-serverless-on-kubernetes)

#### Community workshop

[The OpenFaaS workshop](https://github.com/openfaas/workshop/) is a set of 12 self-paced labs and provides a great starting point for learning the features of openfaas. Not all features will be available or usable with faasd.

### Community support

An active community of almost 3000 users awaits you on Slack. Over 250 of those users are also contributors and help maintain the code.

* [Join Slack](https://slack.openfaas.io/)

## Roadmap

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

* `faas auth` - supported for Basic Authentication, but SSO, OAuth2 & OIDC may require a patch

### Backlog

Should have:

* [ ] Resolve core services from functions by populating/sharing `/etc/hosts` between `faasd` and `faasd-provider`
* [ ] Docs or examples on how to use the various connectors and connector-sdk
* [ ] Monitor and restart any of the core components at runtime if the container stops
* [ ] Asynchronous deletion instead of synchronous

Nice to Have:
* [ ] Total memory limits - if a node has 1GB of RAM, don't allow more than 1000MB of RAM to be reserved via limits
* [ ] Offer live rolling-updates, with zero downtime - requires moving to IDs vs. names for function containers
* [ ] Multiple replicas per function

### Known-issues

### Completed

* [x] Provide a cloud-init configuration for faasd bootstrap
* [x] Configure core services from a docker-compose.yaml file
* [x] Store and fetch logs from the journal
* [x] Add support for using container images in third-party public registries
* [x] Add support for using container images in private third-party registries
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
* [x] Optionally expose core services from the docker-compose.yaml file, locally or to all adapters.
* [x] ~~[containerd can't pull image from Github Docker Package Registry](https://github.com/containerd/containerd/issues/3291)~~ ghcr.io support
* [x] Provide [simple Caddyfile example](https://blog.alexellis.io/https-inlets-local-endpoints/) in the README showing how to expose the faasd proxy on port 80/443 with TLS
* [x] Annotation support
* [x] Hard memory limits for functions
* [x] Terraform for DigitalOcean
* [x] [Store and retrieve annotations in function spec](https://github.com/openfaas/faasd/pull/86) - in progress
* [x] An installer for faasd and dependencies - runc, containerd

WIP:

* [ ] Terraform for AWS
