---
layout: default
description: Instructions for installing Bravetools on Windows
keywords: requirements, installation, windows, install, uninstall, upgrade, update
title: Install Bravetools on Windows
parent: Install Bravetools
nav_order: 1
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

Default installation of Bravetools runs on top of the [LXD](https://linuxcontainers.org/lxd/introduction/) daemon, which is not supported natively on Windows.

To run effectively on Windows, Bravetool requires [multipass](multipass.run). Multipass uses Windows' native Hyper-V technolgy to provision fast and lightweight Ubuntu virtual machines, which are seamlessly used by Bravetools behind the scenes.

## Install Bravetools

Latest stable release of Bravetools can be installed by:

```bash
git clone https://github.com/beringresearch/bravetools
cd bravetools
make ubuntu
```