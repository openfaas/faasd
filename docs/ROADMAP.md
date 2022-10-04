# faasd backlog and features

## Supported operations

* `faas-cli login`
* `faas-cli up`
* `faas-cli list`
* `faas-cli describe`
* `faas-cli deploy --update=true --replace=false`
* `faas-cli invoke --async`
* `faas-cli invoke`
* `faas-cli rm`
* `faas-cli store list/deploy/inspect`
* `faas-cli version`
* `faas-cli namespace`
* `faas-cli secret`
* `faas-cli logs`
* `faas-cli auth` - supported for Basic Authentication and OpenFaaS Pro with OIDC and Single-sign On.

Scale from and to zero is also supported. On a Dell XPS with a small, pre-pulled image unpausing an existing task took 0.19s and starting a task for a killed function took 0.39s. There may be further optimizations to be gained.

The OpenFaaS REST API is supported by faasd, learn more in the [manual](https://store.openfaas.com/l/serverless-for-everyone-else) under "Can I get an API with that?"

## Constraints vs OpenFaaS on Kubernetes

faasd suits certain use-cases as mentioned in the [README.md](/README.md) file, for those who want a solution which can scale out horizontally with minimum effort, Kubernetes or K3s is a valid option.

Which is right for you? [Read a comparison in the OpenFaaS docs](https://docs.openfaas.com/deployment/)

### One replica per function

Functions only support one replica for each function, so that means horizontal scaling is not available.

It can scale vertically, and this may be a suitable alternative for many use-cases. See the [YAML reference for how to configure limits](https://docs.openfaas.com/reference/yaml/).

Workaround: deploy multiple, dynamically named functions `scraper-1`, `scraper-2`, `scraper-3` and set up a reverse proxy rule to load balance i.e. `scraper.example.com => [/function/scraper-1, /function/scraper-2, /function/scraper-3]`.

### Scale from zero may give a non-200

faasd itself does not implement a health check to determine if a function is ready for traffic. Since faasd doesn't support auto-scaling, this is unlikely to affect you.

Workaround: Have your client retry HTTP calls, or don't scale to zero.

### Single node, no clustering

faasd is operates on a single-node model. If this is an issue for you, but you have resource constraints, you will need to use OpenFaaS on Kubernetes.

There are no plans to add any form of clustering or multi-node support to faasd. See also: [HA / resilience in faasd #225](https://github.com/openfaas/faasd/issues/225)

What about HA and fault tolerance?

To achieve fault tolerance, you could put two faasd instances behind a load balancer or proxy, but you will need to deploy the same set of functions to each.

An alternative would be to take regular VM backups or snapshots.

### No rolling updates are available today

When running `faas-cli deploy`, your old function is removed before the new one is started. This may cause a period of downtime, depending on the timeouts and grace periods you set.

Workaround: deploy uniquely named functions i.e. `scraper-1` and `scraper-2` with a reverse proxy rule that maps `/function/scraper` to the active version.

## Known issues

### Troubleshooting

There is a very detailed chapter on troubleshooting in the eBook [Serverless For Everyone Else](https://store.openfaas.com/l/serverless-for-everyone-else)

### Your function timed-out at 60 seconds

This is no longer an issue, see the manual for how to configure a longer timeout, updated 3rd October 2022.

### Non 200 HTTP status from the gateway upon reboot

This issue appears to happen sporadically and only for some users.

If you get a non 200 HTTP code from the gateway, or caddy after installing faasd, check the logs of faasd:

```bash
sudo journalctl -u faasd
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

* [ ] Monitor and restart any of the core components at runtime if the container stops
* [ ] Asynchronous function deletion instead of synchronous
* [ ] Asynchronous function start-up instead of synchronous

Nice to Have:

* [ ] Live rolling-updates, with zero downtime - requires moving to IDs vs. names for function containers
* [ ] Total memory limits for a node - if a node has 1GB of RAM, don't allow more than 1000MB of RAM to be reserved via limits
* [ ] Terraform for AWS EC2

Won't have:

* [ ] Clustering
* [ ] Multiple replicas per function

### Completed

* [x] Docs or examples on how to use the various event connectors (Yes in the eBook)
* [x] Resolve core services from functions by populating/sharing `/etc/hosts` between `faasd` and `faasd-provider`
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
* [x] Offer a recommendation or implement a strategy for faasd replication/HA

