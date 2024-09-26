#!/bin/bash

set -x

export DEBIAN_FRONTEND=noninteractive

sudo apt update -y
sudo apt install -y zfsutils-linux

sudo snap refresh lxd --channel=5.21/stable
sudo snap install go --classic

sudo usermod -aG lxd vagrant

sudo chown -R vagrant:vagrant /home/vagrant/workspace
