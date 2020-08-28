---
layout: default
title: brave init
parent: CLI
nav_order: 1
---

# brave init

Create a new Bravetools host

```
brave init
```

## Description

`brave init` is a utility that initialises a new Bravetools host on a local or remote instance. Typically, it is the first thing that's run when Bravetools is installed.

On Windows and Mac machines, initialisation requires [multipass](multipass.run), which is used to install an isolated LXD blackened. `brave init` conveniently configures your multipass resources, such as RAM, CPU, and storage to provide an optimal configuration.

On Linux hosts, bravetools is intialised natively and is restricted by the default size of the desired storage pool. Storage pool can be [configured](../brave_configure) to generate the most appropriate setting.

### Configuring Bravetools from command line

The simplest way to initialise and configure Bravetools host is to run

```bash
brave init
```

This operation will initialise networking with LXD bridge, create a new profile, and add a 10GB ZFS storage pool. These settings can then be [configured and customised](../brave_configure) further.

If you'd like a little bit more flexibility, you can invoke these options through command line:

```bash
brave init --backend lxd --storage 10GB --network 10.0.0.1
```

### Configuring Bravetools from file

If you have an existing configuration file, you can pass it to `brave init` via the `--config` option. For example, consider a configuration file `config.yml`:

```yaml
name: brave
trust: brave
profile: brave
storage:
  type: zfs
  name: brave-20200721144807
  size: 10GB
network:
  bridge: 10.0.0.1
backendsettings:
  type: lxd
  resources:
    name: ""
    os: ""
    cpu: ""
    ram: ""
    hd: ""
    ip: 0.0.0.0
status: inactive
```

This Bravetools host can be set up as:

```
brave init --config config.yml
```

## Options

```
  -b, --backend string   Backend type (multipass or lxd) [OPTIONAL]
      --config string    Path to the host configuration file [OPTIONAL]
  -h, --help             help for init
  -m, --memory string    Host memory size [OPTIONAL]
  -n, --network string   Host network IP range [OPTIONAL]
  -s, --storage string   Host storage size [OPTIONAL]
```

## See Also

* [brave](brave.md)	 - A complete System Container management platform

