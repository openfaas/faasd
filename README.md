# faasd - a lightweight and portable version of OpenFaaS

faasd is [OpenFaaS](https://github.com/openfaas/) reimagined, but without the cost and complexity of Kubernetes. It runs on a single host with very modest requirements, making it fast and easy to manage. Under the hood it uses [containerd](https://containerd.io/) and [Container Networking Interface (CNI)](https://github.com/containernetworking/cni) along with the same core OpenFaaS components from the main project.

![faasd logo](docs/media/social.png)

## Features & Benefits

- **Lightweight** - faasd is a single Go binary, which runs as a systemd service making it easy to manage
- **Portable** - it runs on any Linux host with containerd and CNI, on as little as 2x vCPU and 2GB RAM - x86_64 and Arm64 supported
- **Easy to manage** - unlike Kubernetes, its API is stable and requires little maintenance
- **Low cost** - it's licensed per installation, so you can invoke your functions as much as you need, without additional cost
- **Stateful containers** - faasd supports stateful containers with persistent volumes such as PostgreSQL, Grafana, Prometheus, etc
- **Built on OpenFaaS** - uses the same containers that power OpenFaaS on Kubernetes for the Gateway, Queue-Worker, Event Connectors, Dashboards, Scale To Zero, etc
- **Ideal for internal business use** - use it to build internal tools, automate tasks, and integrate with existing systems
- **Deploy it for a customer** - package your functions along with OpenFaaS Edge into a VM image, and deploy it to your customers to run in their own datacenters

faasd does not create the same maintenance burden you'll find with installing, upgrading, and securing a Kubernetes cluster. You can deploy it and walk away, in the worst case, just deploy a new VM and deploy your functions again.

You can learn more about supported OpenFaaS features in the [ROADMAP.md](/docs/ROADMAP.md)

## Getting Started

There are two versions of faasd:

* faasd CE - for non-commercial, personal use only licensed under [the faasd CE EULA](/EULA.md)
* OpenFaaS Edge (faasd-pro) - fully licensed for commercial use

You can install either edition using the instructions in the [OpenFaaS docs](https://docs.openfaas.com/deployment/edge/).

You can request a license for [OpenFaaS Edge using this form](https://forms.gle/g6oKLTG29mDTSk5k9)

## Further resources

There are many blog posts and documentation pages about OpenFaaS on Kubernetes, which also apply to faasd.

Videos and overviews:

* [Exploring of serverless use-cases from commercial and personal users (YouTube)](https://www.youtube.com/watch?v=mzuXVuccaqI)
* [Meet faasd. Look Ma’ No Kubernetes! (YouTube)](https://www.youtube.com/watch?v=ZnZJXI377ak)

Use-cases and tutorials:

* [Serverless Node.js that you can run anywhere](https://www.openfaas.com/blog/serverless-nodejs/)
* [Simple Serverless with Golang Functions and Microservices](https://www.openfaas.com/blog/golang-serverless/)
* [Build a Flask microservice with OpenFaaS](https://www.openfaas.com/blog/openfaas-flask/)
* [Get started with Java 11 and Vert.x on Kubernetes with OpenFaaS](https://www.openfaas.com/blog/get-started-with-java-openjdk11/)
* [Deploy to faasd via GitHub Actions](https://www.openfaas.com/blog/openfaas-functions-with-github-actions/)
* [Scrape and automate websites with Puppeteer](https://www.openfaas.com/blog/puppeteer-scraping/)

Additional resources:

* The official handbook - [Serverless For Everyone Else](https://openfaas.gumroad.com/l/serverless-for-everyone-else)
* For reference: [OpenFaaS docs](https://docs.openfaas.com)
* For use-cases and tutorials: [OpenFaaS blog](https://openfaas.com/blog/)
* For self-paced learning: [OpenFaaS workshop](https://github.com/openfaas/workshop/)

### Deployment tutorials

* [Use multipass on Windows, MacOS or Linux](/docs/MULTIPASS.md)
* [Deploy to DigitalOcean with Terraform and TLS](https://www.openfaas.com/blog/faasd-tls-terraform/)
* [Deploy to any IaaS with cloud-init](https://blog.alexellis.io/deploy-serverless-faasd-with-cloud-init/)
* [Deploy faasd to your Raspberry Pi](https://blog.alexellis.io/faasd-for-lightweight-serverless/)

Terraform scripts:

* [Provision faasd on DigitalOcean with Terraform](docs/bootstrap/README.md)
* [Provision faasd with TLS on DigitalOcean with Terraform](docs/bootstrap/digitalocean-terraform/README.md)


### Training / Handbook

You can find various resources to learn about faasd for free, however the official handbook is the most comprehensive guide to getting started with faasd and OpenFaaS.

["Serverless For Everyone Else"](https://openfaas.gumroad.com/l/serverless-for-everyone-else) is the official handbook and was written to contribute funds towards the upkeep and maintenance of the project.


<a href="https://openfaas.gumroad.com/l/serverless-for-everyone-else">
<img src="https://www.alexellis.io/serverless.png" width="40%"></a>

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
* Self-hosting on your Raspberry Pi
* Adding a database for storage with InfluxDB and Postgresql
* Troubleshooting and logs
* CI/CD with GitHub Actions and multi-arch
* Taking things further, community and case-studies

View sample pages, reviews and testimonials on Gumroad:

["Serverless For Everyone Else"](https://openfaas.gumroad.com/l/serverless-for-everyone-else)