#!/bin/bash

# Copyright OpenFaaS Ltd 2025

set -e # stop on error
set -o pipefail

export NEEDRESTART_MODE=a
export DEBIAN_FRONTEND=noninteractive

if [ "$EUID" -ne 0 ]; then
    echo "Please run as root or with sudo"
    exit
fi

has_yum() {
  [ -n "$(command -v yum)" ]
}

has_apt_get() {
  [ -n "$(command -v apt-get)" ]
}

has_pacman() {
  [ -n "$(command -v pacman)" ]
}

install_required_packages() {
  if $(has_apt_get); then

    echo iptables-persistent iptables-persistent/autosave_v4 boolean false | sudo debconf-set-selections
    echo iptables-persistent iptables-persistent/autosave_v6 boolean false | sudo debconf-set-selections

    # Debian bullseye is missing iptables. Added to required packages
    # to get it working in raspberry pi. No such known issues in
    # other distros. Hence, adding only to this block.
    # reference: https://github.com/openfaas/faasd/pull/237
    apt-get update -yq
    apt-get install -yq curl runc bridge-utils iptables iptables-persistent
  elif $(has_yum); then
    yum check-update -y
    yum install -y curl runc iptables-services which
  elif $(has_pacman); then
    pacman -Syy
    pacman -Sy curl runc bridge-utils
  else
    fatal "Could not find apt-get, yum, or pacman. Cannot install dependencies on this OS."
    exit 1
  fi
}

echo "OpenFaaS Edge combines faasd with OpenFaaS Standard"
echo ""
echo ""

echo "1. Installing required OS packages, set SKIP_OS=1 to skip this step"
echo ""

if [ -z "$SKIP_OS" ]; then
    install_required_packages
fi

echo "2. Downloading OCI image, and installing pre-requisites"
echo ""
if [ ! -x "$(command -v arkade)" ]; then
    # For Centos, RHEL, Fedora, Amazon Linux, and Oracle Linux, use BINLOCATION=/usr/bin/

    if $(has_yum); then
      BINLOCATION=/usr/bin/
    fi

    curl -sLS https://get.arkade.dev | BINLOCATION=${BINLOCATION} sh
fi

PATH=$PATH:$HOME/.arkade/bin

tmpdir=$(mktemp -d)

# Ensure all existing services are stopped when installing over an 
# existing faasd installation
systemctl stop faasd || :
systemctl stop faasd-provider || :
systemctl stop containerd || :
killall -9 containerd-shim-runc-v2 || :
killall -9 faasd || :

# crane, or docker can also be used to download the OCI image and to extract it

# Rather than the :latest tag, a specific tag can be given
# Use "crane ls ghcr.io/openfaasltd/faasd-pro" to see available tags

${BINLOCATION}arkade oci install --path ${tmpdir} \
  ghcr.io/openfaasltd/faasd-pro:latest

cd ${tmpdir}
./install.sh ./

echo ""
echo "3.1 Commercial users can create their license key as follows:"
echo ""
echo "sudo mkdir -p /var/lib/faasd/secrets"
echo "sudo nano /var/lib/faasd/secrets/openfaas_license"
echo ""
echo "3.2 For personal, non-commercial use only, GitHub Sponsors of @openfaas (25USD+) can run:"
echo ""
echo "sudo -E faasd github login"
echo "sudo -E faasd activate"
echo ""
echo "4. Then perform the final installation steps"
echo ""
echo "sudo -E sh -c \"cd ${tmpdir}/var/lib/faasd && faasd install\""
echo ""
echo "5. Refer to the complete handbook and supplementary documentation at:"
echo ""
echo "http://store.openfaas.com/l/serverless-for-everyone-else?layout=profile"
echo ""
echo "https://docs.openfaas.com/edge/overview"
echo ""
