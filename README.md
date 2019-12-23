# faasd - serverless with containerd

[![Build
Status](https://travis-ci.com/alexellis/faasd.svg?branch=master)](https://travis-ci.com/alexellis/faasd)

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

Other operations are pending development in the provider.

### Pre-reqs

* Linux - ideally Ubuntu, which is used for testing.
* Installation steps as per [faas-containerd](https://github.com/alexellis/faas-containerd) for building and for development
* [faas-cli](https://github.com/openfaas/faas-cli) (optional)

## Backlog

* Use CNI to create network namespaces and adapters
* Inject / manage IPs between core components for service to service communication - i.e. so Prometheus can scrape the OpenFaaS gateway
* Monitor and restart any of the core components, if they crash
* Configure `basic_auth` to protect the OpenFaaS gateway and faas-containerd HTTP API
* Self-install / create systemd service on start-up using [go-systemd](https://github.com/coreos/go-systemd)
* Bundle/package/automate installation of containerd - [see bootstrap from k3s](https://github.com/rancher/k3s)
* Create [faasd.service](https://github.com/rancher/k3s/blob/master/k3s.service)


## Hacking

First run faas-containerd

```sh
cd $GOPATH/src/github.com/alexellis/faas-containerd
go build && sudo ./faas-containerd
```

Then run faasd, which brings up the gateway and Prometheus as containers

```sh
cd $GOPATH/src/github.com/alexellis/faasd
go build && sudo ./faasd
```

Look in `hosts` in the current working folder to get the IP for the gateway or Prometheus

```sh
127.0.0.1       localhost
172.19.0.1      faas-containerd
172.19.0.2      prometheus

172.19.0.3      gateway
```

Since faas-containerd uses containerd heavily it is not running as a container, but as a stand-alone process. Its port is available via the bridge interface, i.e. netns0.

Now go to the gateway's IP address as shown above on port 8080, i.e. http://172.19.0.3:8080 - you can also use this address to deploy OpenFaaS Functions via the `faas-cli`. 

Removing containers:

```sh
echo faas-containerd gateway prometheus |xargs sudo ctr task rm -f

echo faas-containerd gateway prometheus |xargs sudo ctr container rm

echo faas-containerd gateway prometheus |xargs sudo ctr snapshot rm
```
