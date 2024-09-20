#!/bin/bash

if [ "$EUID" -ne 0 ]; then
    echo "Please run as root or with sudo"
    exit
fi

echo "1. Downloading OCI image, and installing pre-requisites"
echo ""
if [ ! -x "$(command -v arkade)" ]; then
    curl -sLS https://get.arkade.dev | sh
fi

PATH=$PATH:$HOME/.arkade/bin

tmpdir=$(mktemp -d)

arkade oci install --path ${tmpdir} \
  ghcr.io/openfaasltd/faasd-pro:latest

cd ${tmpdir}
./install.sh ./

echo ""
echo "2. You now need to activate your license via GitHub"
echo ""
echo "sudo -E faasd github login"
echo "sudo -E faasd activate"
echo ""
echo ""
echo "3. Then perform the final installation steps"
echo ""
echo "sudo -E sh -c \"cd ${tmpdir}/var/lib/faasd && faasd install\""
echo ""
echo "4. Additional OS packages are sometimes required, with one of the below:"
echo ""
echo "apt install -qy runc bridge-utils iptables"
echo ""
echo "yum install runc iptables-services"
echo ""
echo "pacman -Sy runc bridge-utils"

