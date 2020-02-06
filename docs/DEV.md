## Manual installation of faasd for development

### Pre-reqs

* Linux

    PC / Cloud - any Linux that containerd works on should be fair game, but faasd is tested with Ubuntu 18.04

    For Raspberry Pi Raspbian Stretch or newer also works fine

    For MacOS users try [multipass.run](https://multipass.run) or [Vagrant](https://www.vagrantup.com/)

    For Windows users, install [Git Bash](https://git-scm.com/downloads) along with multipass or vagrant. You can also use WSL1 or WSL2 which provides a Linux environment.

    You will also need [containerd v1.3.2](https://github.com/containerd/containerd) and the [CNI plugins v0.8.5](https://github.com/containernetworking/plugins)

    [faas-cli](https://github.com/openfaas/faas-cli) is optional, but recommended.

### Get containerd

You have three options - binaries for PC, binaries for armhf, or build from source.

* Install containerd `x86_64` only

```sh
export VER=1.3.2
curl -sLSf https://github.com/containerd/containerd/releases/download/v$VER/containerd-$VER.linux-amd64.tar.gz > /tmp/containerd.tar.gz \
  && sudo tar -xvf /tmp/containerd.tar.gz -C /usr/local/bin/ --strip-components=1

containerd -version
```

* Or get my containerd binaries for armhf

Building containerd on armhf is extremely slow.

```sh
curl -sSL https://github.com/alexellis/containerd-armhf/releases/download/v1.3.2/containerd.tgz | sudo tar -xvz --strip-components=2 -C /usr/local/bin/
```

* Or clone / build / install [containerd](https://github.com/containerd/containerd) from source:

```sh
export GOPATH=$HOME/go/
mkdir -p $GOPATH/src/github.com/containerd
cd $GOPATH/src/github.com/containerd
git clone https://github.com/containerd/containerd
cd containerd
git fetch origin --tags
git checkout v1.3.2

make
sudo make install

containerd --version
```

Kill any old containerd version:

```sh
# Kill any old version
sudo killall containerd
sudo systemctl disable containerd
```

Start containerd in a new terminal:

```sh
sudo containerd &
```
#### Enable forwarding

> This is required to allow containers in containerd to access the Internet via your computer's primary network interface.

```sh
sudo /sbin/sysctl -w net.ipv4.conf.all.forwarding=1
```

Make the setting permanent:

```sh
echo "net.ipv4.conf.all.forwarding=1" | sudo tee -a /etc/sysctl.conf
```

### Hacking (build from source)

#### Get build packages

```sh
sudo apt update \
  && sudo apt install -qy \
    runc \
    bridge-utils
```

You may find alternatives for CentOS and other distributions.

#### Install Go 1.13 (x86_64)

```sh
curl -sSLf https://dl.google.com/go/go1.13.6.linux-amd64.tar.gz > go.tgz
sudo rm -rf /usr/local/go/
sudo mkdir -p /usr/local/go/
sudo tar -xvf go.tgz -C /usr/local/go/ --strip-components=1

export GOPATH=$HOME/go/
export PATH=$PATH:/usr/local/go/bin/

go version
```

#### Or on Raspberry Pi (armhf)

```sh
curl -SLsf https://dl.google.com/go/go1.13.6.linux-armv6l.tar.gz > go.tgz
sudo rm -rf /usr/local/go/
sudo mkdir -p /usr/local/go/
sudo tar -xvf go.tgz -C /usr/local/go/ --strip-components=1

export GOPATH=$HOME/go/
export PATH=$PATH:/usr/local/go/bin/

go version
```

#### Install the CNI plugins:

* For PC run `export ARCH=amd64`
* For RPi/armhf run `export ARCH=arm`
* For arm64 run `export ARCH=arm64`

Then run:

```sh
export ARCH=amd64
export CNI_VERSION=v0.8.5

sudo mkdir -p /opt/cni/bin
curl -sSL https://github.com/containernetworking/plugins/releases/download/${CNI_VERSION}/cni-plugins-linux-${ARCH}-${CNI_VERSION}.tgz | sudo tar -xz -C /opt/cni/bin
```

Run or install faasd, which brings up the gateway and Prometheus as containers

```sh
cd $GOPATH/src/github.com/openfaas/faasd
go build

# Install with systemd
# sudo ./faasd install

# Or run interactively
# sudo ./faasd up
```

#### Build and run `faasd` (binaries)

```sh
# For x86_64
sudo curl -fSLs "https://github.com/openfaas/faasd/releases/download/0.7.4/faasd" \
    -o "/usr/local/bin/faasd" \
    && sudo chmod a+x "/usr/local/bin/faasd"

# armhf
sudo curl -fSLs "https://github.com/openfaas/faasd/releases/download/0.7.4/faasd-armhf" \
    -o "/usr/local/bin/faasd" \
    && sudo chmod a+x "/usr/local/bin/faasd"

# arm64
sudo curl -fSLs "https://github.com/openfaas/faasd/releases/download/0.7.4/faasd-arm64" \
    -o "/usr/local/bin/faasd" \
    && sudo chmod a+x "/usr/local/bin/faasd"
```

#### At run-time

Look in `hosts` in the current working folder or in `/var/lib/faasd/` to get the IP for the gateway or Prometheus

```sh
127.0.0.1      localhost
10.62.0.1      faasd-provider

10.62.0.2      prometheus
10.62.0.3      gateway
10.62.0.4      nats
10.62.0.5      queue-worker
```

The IP addresses are dynamic and may change on every launch.

Since faasd-provider uses containerd heavily it is not running as a container, but as a stand-alone process. Its port is available via the bridge interface, i.e. `openfaas0`

* Prometheus will run on the Prometheus IP plus port 8080 i.e. http://[prometheus_ip]:9090/targets

* faasd-provider runs on 10.62.0.1:8081, i.e. directly on the host, and accessible via the bridge interface from CNI.

* Now go to the gateway's IP address as shown above on port 8080, i.e. http://[gateway_ip]:8080 - you can also use this address to deploy OpenFaaS Functions via the `faas-cli`. 

* basic-auth

    You will then need to get the basic-auth password, it is written to `/var/lib/faasd/secrets/basic-auth-password` if you followed the above instructions.
The default Basic Auth username is `admin`, which is written to `/var/lib/faasd/secrets/basic-auth-user`, if you wish to use a non-standard user then create this file and add your username (no newlines or other characters) 

#### Installation with systemd

* `faasd install` - install faasd and containerd with systemd, this must be run from `$GOPATH/src/github.com/openfaas/faasd`
* `journalctl -u faasd -f` - faasd service logs
* `journalctl -u faasd-provider -f` - faasd-provider service logs
