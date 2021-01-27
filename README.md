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


## "Serverless For Everyone Else" is the official handbook for faasd

<a href="https://gumroad.com/l/serverless-for-everyone-else">
<img src="https://static-2.gumroad.com/res/gumroad/2028406193591/asset_previews/714aad765f8246463fafb64fcd3be4ea/retina/104810333-b628f280-57eb-11eb-8be9-a2f6c773346b.png" width="40%"></a>

You'll learn how to deploy code in any language, lift and shift Dockerfiles, run requests in queues, write background jobs and to integrate with databases. faasd packages the same code as OpenFaaS, so you get built-in metrics for your HTTP endpoints, a user-friendly CLI, pre-packaged functions and templates from the store and a UI.

Topics include:

* Should you deploy to a VPS or Raspberry Pi?
* Deploying your server with bash, cloud-init or terraform
* Using a private container registry
* Finding functions in the store
* Building your first function with Node.js
* Using environment variables for configuration
* Using secrets from functions, and enabling authentication tokens
* Customising templates
* Monitoring your functions with Grafana and Prometheus
* Scheduling invocations and background jobs
* Tuning timeouts, parallelism, running tasks in the background
* Adding TLS to faasd and custom domains for functions
* Adding a database for storage with InfluxDB
* Troubleshooting and logs
* CI/CD with GitHub Actions and multi-arch
* Taking things further, community and case-studies

View sample pages, reviews and testimonials on Gumroad:

["Serverless For Everyone Else"](https://gumroad.com/l/serverless-for-everyone-else)


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

For trying out fasad on MacOS or Windows, we recommend using multipass.

If you don't use cloud-init, or have already created your Linux server you can use the installation script as per below:

```bash
git clone https://github.com/openfaas/faasd
cd faasd

./hack/install.sh
```

> This approach also works for Raspberry Pi

It's recommended that you do not install Docker on the same host as faasd, since 1) they may both use different versions of containerd and 2) docker's networking rules can disrupt faasd's networking. When using faasd - make your faasd server a faasd server, and build container image on your laptop or in a CI pipeline.

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

* [Provision faasd 0.10.0 on DigitalOcean with Terraform 0.12.0](docs/bootstrap/README.md)

* [Provision faasd on DigitalOcean with built-in TLS support](docs/bootstrap/digitalocean-terraform/README.md)

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

### Instructions for hacking on faasd itself

See [here for manual / developer instructions](docs/DEV.md)

## Getting help

### faasd handbook

"Serverless For Everyone Else" is the complete guide and documentation for faasd. If you're looking for how to do something, it's likely that the book covers it.

* [Find out more on Gumroad](https://gumroad.com/l/serverless-for-everyone-else)

### Reference docs for Kubernetes

The [OpenFaaS docs](https://docs.openfaas.com/) provide a wealth of information for OpenFaaS on Kubernetes, and are likely to be useful for you, even using faasd.

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

### Backlog, features and known issues

For completed features, WIP and upcoming roadmap see:

See [ROADMAP.md](docs/ROADMAP.md)
