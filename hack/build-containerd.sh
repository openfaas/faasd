#!/bin/bash


# See pre-reqs:
# https://github.com/alexellis/containerd-arm

export ARCH="arm64"

if [ ! -d "/usr/local/go/bin" ]; then
    echo "Downloading Go.."
    
    curl -SLsf https://golang.org/dl/go1.16.6.linux-$ARCH.tar.gz --output /tmp/go.tgz
    sudo rm -rf /usr/local/go/
    sudo mkdir -p /usr/local/go/
    sudo tar -xvf /tmp/go.tgz -C /usr/local/go/ --strip-components=1
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
git checkout v1.6.8

make
sudo make install

sudo containerd --version
