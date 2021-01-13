# faasd backlog and features

## Supported operations

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
* `faas auth` - supported for Basic Authentication and OpenFaaS PRO with OIDC and Single-sign On.

Scale from and to zero is also supported. On a Dell XPS with a small, pre-pulled image unpausing an existing task took 0.19s and starting a task for a killed function took 0.39s. There may be further optimizations to be gained.

## Constraints vs OpenFaaS on Kubernetes

faasd suits certain use-cases as mentioned in the README file, for those who want a solution which can scale out horizontally with minimum effort, Kubernetes or K3s is a valid option.

### One replica per function

Functions only support one replica, so cannot scale horizontally, but can scale vertically.

Workaround: deploy one uniquely named function per replica.

### Scale from zero may give a non-200

When scaling from zero there is no health check implemented, so the request may arrive before your HTTP server is ready to serve a request, and therefore give a non-200 code.

Workaround: Do not scale to zero, or have your client retry HTTP calls.

### No clustering is available

No clustering is available in faasd, however you can still apply fault-tolerance and high availability techniques.

Workaround: deploy multiple faasd instances and use a hardware or software load-balancer. Take regular VM/host snapshots or backups.

### No rolling updates

When running `faas-cli deploy`, your old function is removed before the new one is started. This may cause a small amount of downtime, depending on the timeouts and grace periods you set.

Workaround: deploy uniquely named functions per version, and switch an Ingress or Reverse Proxy record to point at a new version once it is ready.

## Known issues

### Non 200 HTTP status from the gateway upon first use

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

