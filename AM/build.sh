#!/bin/bash

set -e  # Exit immediately if any command fails

echo ">>> Stopping faasd and related services..."
sudo systemctl stop faasd faasd-provider containerd || true

echo ">>> Killing any lingering faasd processes..."
sudo pkill -9 faasd || true

echo ">>> Rebuilding faasd..."
cd /users/am_CU/go/src/github.com/openfaas/faasd
make local
sudo cp bin/faasd /usr/local/bin/

# Install faasd only if it's missing
if [ ! -d "/var/lib/faasd" ]; then
    echo ">>> Running faasd install (only first-time or after cleanup)..."
    sudo faasd install
fi

echo ">>> Restarting containerd..."
sudo systemctl restart containerd

echo ">>> Restarting faasd services..."
sudo systemctl restart faasd faasd-provider
sudo systemctl enable faasd faasd-provider

echo ">>> Installing OpenFaaS CLI..."
# curl -sLS https://cli.openfaas.com | sudo sh
faas-cli version

# Add sleep to ensure faasd is ready before login
echo ">>> Waiting for faasd to be ready..."
sleep 10  # Give faasd time to fully start

echo ">>> Logging into OpenFaaS..."
export OPENFAAS_URL=http://127.0.0.1:8080
PASSWORD=$(sudo cat /var/lib/faasd/secrets/basic-auth-password)
echo $PASSWORD | faas-cli login -u admin --password-stdin

echo ">>> Deploying test function..."
faas-cli store deploy figlet

echo ">>> Invoking test function..."
echo "Hello MooMoo" | faas-cli invoke figlet

echo ">>> Checking faasd version..."
faasd version
