# faasd backlog and features

## Backlog

Should have:

* [ ] Resolve core services from functions by populating/sharing `/etc/hosts` between `faasd` and `faasd-provider`
* [ ] Docs or examples on how to use the various connectors and connector-sdk
* [ ] Monitor and restart any of the core components at runtime if the container stops
* [ ] Asynchronous deletion instead of synchronous

Nice to Have:

* [ ] Terraform for AWS (in-progress)
* [ ] Total memory limits - if a node has 1GB of RAM, don't allow more than 1000MB of RAM to be reserved via limits
* [ ] Offer live rolling-updates, with zero downtime - requires moving to IDs vs. names for function containers
* [ ] Multiple replicas per function

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
