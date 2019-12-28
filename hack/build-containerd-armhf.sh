#!/bin/bash

echo "Downloading Go"

curl -SLsf https://dl.google.com/go/go1.12.14.linux-armv6l.tar.gz > go.tgz
sudo rm -rf /usr/local/go/
sudo mkdir -p /usr/local/go/
sudo tar -xvf go.tgz -C /usr/local/go/ --strip-components=1

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
