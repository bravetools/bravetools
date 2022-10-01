---
layout: default
description: Instructions for installing Bravetools on macOS
keywords: requirements, apt, installation, ubuntu, install, uninstall, upgrade, update
title: Install Bravetools on macOS
parent: Install Bravetools
nav_order: 2
---

To get started with Bravetools on macOS, make sure you [meet the prerequisites](#prerequisites), then [install Bravetools](#install-bravetools)

## Prerequisites

### OS requirements

Your Mac must meet the following requirements to successfully install Bravetools:

- **macOS must be version 10.13 or newer**. That is, Catalina, Mojave, or High Sierra. We recommend upgrading to the latest version of macOS.
- At least 8 GB of RAM.

### Software requirements

Default installation of Bravetools runs on top of the [LXD](https://linuxcontainers.org/lxd/introduction/) daemon. Although LXD client is supported on macOS, LXC (the underlying container technology), is a feature of the Linux kernel and is not available natively on macOS.

On macOS, Bravetool requires [multipass](multipass.run). Multipass uses Mac's native hyperkit technology to provision fast and lightweight Ubuntu virtual machines, which are seamlessly used by Bravetools behind the scenes.


## Install Bravetools

### Stable release
1. Download the [latest stable release](https://github.com/bravetools/bravetools/releases) for your host platform
2. Add the `brave` binary to your `$PATH`.
3. Run `brave init` to get going.

### Development release
Bleeding edge release of Bravetools can be installed by:

```bash
git clone https://github.com/beringresearch/bravetools
cd bravetools
make darwin
```

If this is your first time setting up Bravetools, run `brave init` to initialise the required profile, storage pool, and LXD bridge.
