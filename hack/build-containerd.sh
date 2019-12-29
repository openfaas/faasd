#!/bin/bash

export ARCH="amd64"
echo "Downloading Go"

curl -SLsf https://dl.google.com/go/go1.12.14.linux-$ARCH.tar.gz --output /tmp/go.tgz
sudo rm -rf /usr/local/go/
sudo mkdir -p /usr/local/go/
sudo tar -xvf /tmp/go.tgz -C /usr/local/go/ --strip-components=1

export GOPATH=$HOME/go/
export PATH=$PATH:/usr/local/go/bin/

go version

echo "Building containerd"

mkdir -p $GOPATH/src/github.com/containerd
cd $GOPATH/src/github.com/containerd
git clone https://github.com/containerd/containerd

cd containerd
git fetch origin --tags
git checkout v1.3.2

make
sudo make install

sudo containerd --version
