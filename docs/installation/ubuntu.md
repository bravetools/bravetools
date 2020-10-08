---
layout: default
description: Instructions for installing Bravetools on Ubuntu
keywords: requirements, apt, installation, ubuntu, install, uninstall, upgrade, update
title: Install Bravetools on Ubuntu
parent: Install Bravetools
nav_order: 1
---

To get started with Bravetools on Ubuntu, make sure you [meet the prerequisites](#prerequisites), then [install Bravetools](#install-bravetools)

## Prerequisites

Ensure that your user is part of the `lxd group`:
```bash
sudo usermod --append --groups lxd USER
```

You may also need to install `zfsutils`:
```bash
sudo apt install zfsutils-linux
```

### OS requirements

To install Bravetools, you need the 64-bit version of one of these Ubuntu
versions:

- Ubuntu Focal 20.04 (LTS)
- Ubuntu Eoan 19.10
- Ubuntu Bionic 18.04 (LTS)

### Software requirements

Default installation of Bravetools runs on top of the [LXD](https://linuxcontainers.org/lxd/introduction/) daemon.

### Supported storage backends

Bravetools supports `zfs` (default) and `btrfs` file systems.

## Install Bravetools

Latest stable release of Bravetools can be installed by:

```bash
git clone https://github.com/beringresearch/bravetools
cd bravetools
make ubuntu
```

If this is your first time setting up Bravetools, run `brave init` to initialise the required profile, storage pool, and LXD bridge.