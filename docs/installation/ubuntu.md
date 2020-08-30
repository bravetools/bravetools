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

### OS requirements

To install Bravetools, you need the 64-bit version of one of these Ubuntu
versions:

- Ubuntu Focal 20.04 (LTS)
- Ubuntu Eoan 19.10
- Ubuntu Bionic 18.04 (LTS)

### Software requirements

Default installation of Bravetools runs on top of the [LXD](https://linuxcontainers.org/lxd/introduction/) daemon. It is recommended that `snap` LXD is installed on the base system. If it's not, Bravetools will install and configure it automatically for you at the time of initialisation.


> **Note**: 
>LXD up to 3.0.x were published as non-`snap` versions. Bravetools will not work with these distributions. The user is encouraged migrating to  `snap`-LXD before continuing with installation.

### Supported storage backends

Bravetools supports `zfs` (default) and `btrfs` file systems.

## Install Bravetools

Latest stable release of Bravetools can be installed by:

```bash
git clone https://github.com/beringresearch/bravetools
cd bravetools
make ubuntu
```