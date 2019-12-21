# faasd - serverless with containerd

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

## Backlog

* Use CNI to create network namespaces and adapters
* Inject / manage IPs between core components for service to service communication - i.e. so Prometheus can scrape the OpenFaaS gateway
* Monitor and restart any of the core components, if they crash
* Configure `basic_auth` to protect the OpenFaaS gateway and faas-containerd HTTP API
* Self-install / create systemd service on start-up using [go-systemd](https://github.com/coreos/go-systemd)
* Bundle/package/automate installation of containerd - [see bootstrap from k3s](https://github.com/rancher/k3s)
* Create [faasd.service](https://github.com/rancher/k3s/blob/master/k3s.service)


Hacking:

```sh
echo faas-containerd gateway prometheus |xargs sudo ctr task rm -f

echo faas-containerd gateway prometheus |xargs sudo ctr container rm

echo faas-containerd gateway prometheus |xargs sudo ctr snapshot rm
```