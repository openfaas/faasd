#!/bin/bash


# See pre-reqs:
# https://github.com/alexellis/containerd-arm

export ARCH="arm64"

if [ ! -d "/usr/local/go/bin" ]; then
    curl -sLS https://get.arkade.dev | sudo sh
    sudo -E arkade system install go
else
    echo "Go already present, skipping."
fi

export GOPATH=$HOME/go/
export PATH=$PATH:/usr/local/go/bin/

go version

echo "Building containerd"

mkdir -p $GOPATH/src/github.com/containerd
cd $GOPATH/src/github.com/containerd
git clone https://github.com/containerd/containerd

cd containerd
git fetch origin --tags
git checkout v1.7.18

make
sudo make install

sudo containerd --version
