---
layout: default
title: Working with Bravetools remotes
parent: Docs
nav_order: 7
description: "Bravetools enables local and remote builds and unit deployments"
---

# Bravetools Remotes
{: .no_toc }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Introduction

Bravetools supports building images and deploying units across LXD Remotes. Remotes are an LXD concept that refers to various LXD servers or clusters running at a particular URL.

Ability to manipulate containers across multiple enables features such as:

* cross-platform image builds
* remote unit deployment
* remote management of system containers

The default remote is called `local` and it is automatically added on first use of Bravetools. All remotes are stored in `~/.bravetools/remotes/` directory. Here's a basic configuration of a `local` remote, which follows a JSON structure:

```json
{
    "name": "local",
    "url": "https://192.168.64.60:8443",
    "protocol": "lxd",
    "public": false,
    "profile": "user",
    "network": "userbr0",
    "storage": "user-20220915121119"
}
```

## Adding a New Remote

### Setting up a Bravetools Remote
Any server that has been initialised using `lxd init` can be configured as a Bravetools remote. To confirm that your LXD server has been initialised, you should see the following outputs:

``` bash
lxc profile list
+---------+---------+
|  NAME   | USED BY |
+---------+---------+
| default | 0       |
+---------+---------+

lxc network list
+--------+----------+---------+-------------+---------+
|  NAME  |   TYPE   | MANAGED | DESCRIPTION | USED BY |
+--------+----------+---------+-------------+---------+
| eth0   | physical | NO      |             | 0       |
+--------+----------+---------+-------------+---------+
| lxdbr0 | bridge   | YES     |             | 1       |
+--------+----------+---------+-------------+---------+
```

Now expose your LXD server to the network and set a trust password:

```bash
lxc config set core.https_address $SERVER_IP_ADDRESS
lxc config set core.trust_password root
```

In the above example, the trust password ``root`` is used.

### Adding Remote to Bravetools

Finally, on your host machine, add this LXD server as a Bravetools remote:

```bash
brave remote add myremote https://$SERVER_IP_ADDRESS:8443 --password root
```

This step will use [built-in LXD authentication](https://documentation.ubuntu.com/lxd/en/latest/authentication/) to establish a connection between your host and an LXD Remote via LXD socket.

Confirm that the remote has been added:

```bash
brave remote list

myremote
local
```

### Configuring Bravetools Remotes

When a new remote is added, its configuration is stored in `~/.bravetools/remotes/$REMOTE_NAME.json`:

```json
{
    "name": "macpine",
    "url": "https://127.0.0.1:8443",
    "protocol": "lxd",
    "public": false,
    "profile": "default",
    "network": "lxdbr0",
    "storage": ""
}
```

`profile` and `network` parameters refer to LXD profile and bridge on your remote respectively. You may need to alter these values, depending on your remote set up and manually edit `profile` and `network` fields to reflect your remote LXD configuration.

## Configuring Bravetools to use Remotes for image builds

By default, Bravetools uses a `local` remote for an image build. On Mac/Windows, this is a Multipass VM, whilst on Linux host this is your local LXD server. Sometimes, it may be desirable to use a remote LXD server to cary out Image builds. For example, if your remote has a different CPU architecture (arm64 vs x86) or has more allocated resources.

Bravetools remote backend can be set in the global configuration file `~/.bravetools/config.yml` by setting the `remote` field to the name of one of your added remotes.

```yaml
name: user
trust: user
profile: user
storage:
  type: zfs
  name: user-20220915121119
  size: 98GB
network:
  name: userbr0
  ip: 10.57.220.1
backendsettings:
  type: multipass
  resources:
    name: user
    os: bionic
    cpu: "2"
    ram: 4GB
    hd: 100GB
    ip: 192.168.64.60
status: active
remote: local
```

Next time you run `brave build`, Bravetools will execute Bravefile instructions on the remote LXD server.

## Deploying Units to Remotes

To deploy an image to a specific remote, you can simply append the remote name to your target Unit name either on the command line or in the ``service`` section of your ``Bravefile``.

```
brave deploy brave-base-alpine-edge-1.0 --name myremote:test --port 1234:1234
```

## Manipulating Units on Remotes

Basic Bravetools commands such as `brave start`, `brave stop`, and `brave remove` can access both local and remote units. If remote name is not appended to the unit name, Bravetools will assume that the unit is running on a `local` remote. To interact with a unit on a remote LXD server, simply append <remote>: to the unit name:

```bash
brave start myremote:test
```