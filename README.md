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

## Try faasd for the first time

faasd is OpenFaaS, so many things you read in the docs or in blog posts will work the same way.

Use-cases and tutorials:

* [Deploy via GitHub Actions](https://www.openfaas.com/blog/openfaas-functions-with-github-actions/)
* [Scrape and automate websites with Puppeteer](https://www.openfaas.com/blog/puppeteer-scraping/)
* [Serverless Node.js that you can run anywhere](https://www.openfaas.com/blog/serverless-nodejs/)
* [Build a Flask microservice with OpenFaaS](https://www.openfaas.com/blog/openfaas-flask/)

Additional resources:

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

## faasd handbook - everything you need to know to run functions without Kubernetes (early access)

faasd is a portable, and open source serverless engine. It runs a number of core services for its REST API, for background processing, and for metrics. The project schedules functions with containerd directly, and supports scale to and from zero, but without the need for clustering or Kubernetes.

It makes for a quick and easy way to start hosting APIs and websites, benefiting from containers and cloud native technology without having to manage Kubernetes, or pay significant hosting costs.

This handbook is written for those deploying faasd to self-hosted or cloud infrastructure. Whilst OpenFaaS has reference documentation, here we focus on everything you need to know about faasd itself.

Topics include:

* Should you deploy to a VPS or Raspberry Pi?
* Deploying your server with bash, cloud-init or terraform
* Using a private container registry
* Building your first function, and customising templates
* Monitoring your functions with Grafana and Prometheus
* Scheduling invocations and background jobs
* Tuning timeouts, parallelism, running tasks in the background
* Upgrading faasd
* Setting memory limits for functions
* Exposing the core services like Prometheus and NATS

> faasd users can upgrade to Kubernetes when the need presents itself and can bring their functions with them.

Get early access in The OpenFaaS GitHub Sponsors Portal: [The Treasure Trove](https://faasd.exit.openfaas.pro/function/trove/)

* [Become an OpenFaaS Sponsor to gain access](https://github.com/sponsors/openfaas/)

## Finding logs

### Logs for functions

You can view the logs of functions using `journalctl`:

```bash
journalctl -t openfaas-fn:FUNCTION_NAME


faas-cli store deploy figlet
journalctl -t openfaas-fn:figlet -f &
echo logs | faas-cli invoke figlet
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

Commercial users and solo business owners should become OpenFaaS GitHub Sponsors to receive regular email updates on changes, tutorials and new features.

If you are learning faasd, or want to share your use-case, you can join the OpenFaaS Slack community.

* [Become an OpenFaaS GitHub Sponsor](https://github.com/sponsors/openfaas/)
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

### Known-issues

#### Non 200 HTTP status code upon first use

This issue appears to happen sporadically and only for some users.

If you get a non 200 HTTP code from the gateway, or caddy after installing faasd, check the logs of faasd:

```bash
sudo journalctl -t faasd
```

If you see the following error:

```
unable to dial to 10.62.0.5:8080, error: dial tcp 10.62.0.5:8080: connect: no route to host
```

Restart the faasd service with:

```bash
sudo systemctl restart faasd
```

### Backlog and features

For completed features, WIP and upcoming roadmap see:

See [ROADMAP.md](docs/ROADMAP.md)
