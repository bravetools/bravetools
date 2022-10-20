#!/bin/bash

set -x

export DEBIAN_FRONTEND=noninteractive

sudo apt update -y
sudo apt install -y zfsutils-linux

wget -q https://golang.org/dl/go1.19.2.linux-amd64.tar.gz
tar xf go1.19.2.linux-amd64.tar.gz
sudo mv go /usr/local/go-1.19
sudo touch /etc/profile.d/Z99-go-1.19.sh
rm go1.19.2.linux-amd64.tar.gz

echo 'export GOROOT=/usr/local/go-1.19' | sudo tee -a /etc/profile.d/Z99-go-1.19.sh
echo 'export GOPATH=$HOME/workspace' | sudo tee -a /home/vagrant/.bashrc
echo 'export GOROOT=/usr/local/go-1.19' | sudo tee -a /home/vagrant/.bashrc
echo 'export PATH=$PATH:$GOROOT/bin' | sudo tee -a /home/vagrant/.bashrc

sudo usermod -aG lxd vagrant

sudo chown -R vagrant:vagrant /home/vagrant/workspace
