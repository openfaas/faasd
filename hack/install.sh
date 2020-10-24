#!/bin/bash

# Copyright OpenFaaS Author(s) 2020

#########################
# Repo specific content #
#########################

export OWNER="openfaas"
export REPO="faasd"

version=""

echo "Finding latest version from GitHub"
version=$(curl -sI https://github.com/$OWNER/$REPO/releases/latest | grep -i location | awk -F"/" '{ printf "%s", $NF }' | tr -d '\r')
echo "$version"

if [ ! $version ]; then
  echo "Failed while attempting to get latest version"
  exit 1
fi

SUDO=sudo
if [ "$(id -u)" -eq 0 ]; then
  SUDO=
fi

verify_system() {
  if ! [ -d /run/systemd ]; then
    fatal 'Can not find systemd to use as a process supervisor for faasd'
  fi
}

has_yum() {
  [ -n "$(command -v yum)" ]
}

has_apt_get() {
  [ -n "$(command -v apt-get)" ]
}

install_required_packages() {
  if $(has_apt_get); then
    $SUDO apt-get update -y
    $SUDO apt-get install -y curl runc bridge-utils
  elif $(has_yum); then
    $SUDO yum check-update -y
    $SUDO yum install -y curl runc
  else
    fatal "Could not find apt-get or yum. Cannot install dependencies on this OS."
    exit 1
  fi
}

install_cni_plugins() {
  cni_version=v0.8.5
  suffix=""
  arch=$(uname -m)
  case $arch in
  x86_64 | amd64)
    suffix=amd64
    ;;
  aarch64)
    suffix=arm64
    ;;
  arm*)
    suffix=arm
    ;;
  *)
    fatal "Unsupported architecture $arch"
    ;;
  esac

  $SUDO mkdir -p /opt/cni/bin
  curl -sSL https://github.com/containernetworking/plugins/releases/download/${cni_version}/cni-plugins-linux-${suffix}-${cni_version}.tgz | $SUDO tar -xvz -C /opt/cni/bin
}

install_containerd() {
  arch=$(uname -m)
  case $arch in
  x86_64 | amd64)
    curl -sLSf https://github.com/containerd/containerd/releases/download/v1.3.7/containerd-1.3.7-linux-amd64.tar.gz | $SUDO tar -xvz --strip-components=1 -C /usr/local/bin/
    ;;
  armv7l)
    curl -sSL https://github.com/alexellis/containerd-arm/releases/download/v1.3.5/containerd-1.3.5-linux-armhf.tar.gz | $SUDO tar -xvz --strip-components=1 -C /usr/local/bin/
    ;;
  aarch64)
    curl -sSL https://github.com/alexellis/containerd-arm/releases/download/v1.3.5/containerd-1.3.5-linux-arm64.tar.gz | $SUDO tar -xvz --strip-components=1 -C /usr/local/bin/
    ;;
  *)
    fatal "Unsupported architecture $arch"
    ;;
  esac

  $SUDO curl -SLfs https://raw.githubusercontent.com/containerd/containerd/v1.3.5/containerd.service --output /etc/systemd/system/containerd.service

  $SUDO systemctl enable containerd
  $SUDO systemctl start containerd

  sleep 5
}

install_faasd() {
  arch=$(uname -m)
  case $arch in
  x86_64 | amd64)
    suffix=""
    ;;
  aarch64)
    suffix=-arm64
    ;;
  armv7l)
    suffix=-armhf
    ;;
  *)
    echo "Unsupported architecture $arch"
    exit 1
    ;;
  esac

  $SUDO curl -fSLs "https://github.com/openfaas/faasd/releases/download/${version}/faasd${suffix}" --output "/usr/local/bin/faasd"
  $SUDO chmod a+x "/usr/local/bin/faasd"

  mkdir -p /tmp/faasd-${version}-installation/hack
  cd /tmp/faasd-${version}-installation
  $SUDO curl -fSLs "https://raw.githubusercontent.com/openfaas/faasd/${version}/docker-compose.yaml" --output "docker-compose.yaml"
  $SUDO curl -fSLs "https://raw.githubusercontent.com/openfaas/faasd/${version}/prometheus.yml" --output "prometheus.yml"
  $SUDO curl -fSLs "https://raw.githubusercontent.com/openfaas/faasd/${version}/resolv.conf" --output "resolv.conf"
  $SUDO curl -fSLs "https://raw.githubusercontent.com/openfaas/faasd/${version}/hack/faasd-provider.service" --output "hack/faasd-provider.service"
  $SUDO curl -fSLs "https://raw.githubusercontent.com/openfaas/faasd/${version}/hack/faasd.service" --output "hack/faasd.service"
  $SUDO /usr/local/bin/faasd install

  sleep 5
}

verify_system
install_required_packages

/sbin/sysctl -w net.ipv4.conf.all.forwarding=1
echo "net.ipv4.conf.all.forwarding=1" | $SUDO tee -a /etc/sysctl.conf

install_cni_plugins
install_containerd
install_faasd
