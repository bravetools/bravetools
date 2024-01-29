---
layout: default
title: Initialising Bravetools
parent: Docs
nav_order: 1
description: "Bravetools initialisation."
---

# Introduction

Bravetools uses [LXD](https://documentation.ubuntu.com/lxd/en/latest/) to build, deploy, and manage System Containers. This means that on most Linux-like systems it can run natively, provided that LXD is installed. On MacOS/Windows, Bravetools uses a lightweight [Multipass VM](https://multipass.run) to take care of the Linux kernel features required by LXC containers.

Bravetools relies on the following LXD features:

* Network bridge
* Storage device
* User profile
* Remotes

# Default Configuration

The easiest way to configure Bravetools is to accept default parameters and simply run

```bash
brave init
```

This will create:

* Network bridge - bravetoolsbr0, randomly-assigned private IP from a range of available IP values.
* Storage device - bravetools-${USER}, ZFS file system
* User profile - bravetools-${USER}
* Remotes - local

# Configuring Bravetools using a `yaml` file

It is also possible to configure Bravetools by passing a configuration file `config.yaml`, such as:


```yaml
name: brave
trust: brave
profile: brave
storage:
  type: zfs
  name: brave-20220923163002
  size: 98GB
network:
  name: bravebr0
  ip: 10.0.0.1
backendsettings:
  type: multipass
  resources:
    name: brave
    os: bionic
    cpu: "2"
    ram: 8GB
    hd: 100GB
    ip: 192.168.64.62
status: inactive
remote: local
```

>> **Note** that on Linux machines, `backensettings` option is ignored.

A file-based configuration allows you to explicitly specify profile, storage, and network names, as well as allocate more appropriate hardware resources to the LXD backend. The configuration file can be used to initialise Bravetools as:

```bash
brave init --config config.yaml
```