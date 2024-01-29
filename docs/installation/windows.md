---
layout: default
description: Instructions for installing Bravetools on Windows
keywords: requirements, installation, windows, install, uninstall, upgrade, update
title: Install Bravetools on Windows
parent: Install Bravetools
nav_order: 3
---

To get started with Bravetools on Windows, make sure you [meet the prerequisites](#prerequisites), then [install Bravetools](#install-bravetools)

## Prerequisites

### OS requirements

  - Windows 10 64-bit: Pro, Enterprise, or Education (Build 16299 or later).
  - Hyper-V and Containers Windows features must be enabled.
  - The following hardware prerequisites are required to successfully run Client
Hyper-V on Windows 10:

     - 64 bit processor with [Second Level Address Translation (SLAT)](http://en.wikipedia.org/wiki/Second_Level_Address_Translation)
     - 8GB system RAM
    - BIOS-level hardware virtualization support must be enabled in the
    BIOS settings.

### Software requirements

Default installation of Bravetools runs on top of the [LXD](https://documentation.ubuntu.com/lxd/en/latest/) daemon, which is not supported natively on Windows.

To run effectively on Windows, Bravetool requires [multipass](multipass.run). Multipass uses Windows' native Hyper-V technology to provision fast and lightweight Ubuntu virtual machines, which are seamlessly used by Bravetools behind the scenes.

## Install Bravetools

### Stable release
1. Download the [latest stable release](https://github.com/bravetools/bravetools/releases) for your host platform
2. Add the `brave` binary to your `$PATH`.
3. Run `brave init` to get going.

### Development release
Bleeding edge release of Bravetools can be installed by:

```bash
git clone https://github.com/bravetools/bravetools
cd bravetools
go build -ldflags=“-s -X github.com/bravetools/bravetools/shared.braveVersion=VERSION” -o brave.exe
```

Where VERSION reflects the latest stable release of Bravetools e.g `shared.braveVersion=1.56`