---
layout: default
title: Initialising Bravetools
parent: Docs
nav_order: 1
description: "Bravetools initialisation."
---

# Introduction

Bravetools uses [LXD](https://documentation.ubuntu.com/lxd/en/latest/) to build, deploy, and manage System Containers. This means that on most Linux-like systems it can run natively, provided that LXD is installed. On MacOS/Windows, Bravetools uses a lightweight [Multipass VM](https://multipass.run) to take care of the Linux kernel features required by LXC containers. Alternatively, a [remote LXD instance can be set as a backend](#remote-backends).

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

>> **Note** that on Linux machines and remote backends, `backensettings` option is ignored.

A file-based configuration allows you to explicitly specify profile, storage, and network names, as well as allocate more appropriate hardware resources to the LXD backend. The configuration file can be used to initialise Bravetools as:

```bash
brave init --config config.yaml
```

# Remote backends

Bravetools can also use a remote LXD instance as a backend, allowing for seamless build and deployment of images
on that remote machine. If using a remote backend, users on Windows and Mac will not require Multipass to be installed locally
to use bravetools - instead they can contact an LXD instance running on a separate machine over TCP/IP. Alternatively, other Virtual Machine implementations such as VirtualBox could be set up manualy and used as a backend.

To do this, run `brave init --remote`. This will skip the setup of the local LXD server and multipass VM.
Then add a remote backend with `brave remote add local ...` providing the configuration options matching your
LXD instance as detailed [here](remotes.md).

For example:
```sh
# Initialize bravetools for first time
brave init --remote

# Add an existing remote LXD instance as a backend
brave remote add local https://10.0.0.10:8443 \
--profile bravetools-profile \
--storage bravetools-storage \
--network bravetoolsbr0 \
--password bravetools-password
```
