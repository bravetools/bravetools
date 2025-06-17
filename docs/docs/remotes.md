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

The default remote is called `local` and it is automatically added on first use of Bravetools.

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


## Configurable base image servers

### Default base image server
Canonical has deployed their own images server that hosts images from other distributions (including alpine, centOS, Debian etc...). This server is available at https://images.lxd.canonical.com. The one downside to this image repository is that it does not ship Ubuntu server images, only Ubuntu desktop images. If you need Ubuntu server images they are available at a different repository: https://cloud-images.ubuntu.com/releases.

You can configure bravetools to use Ubuntu server images by editing or adding the config option 'public_image_remote' in ~/.bravetools/config.yml:
```
public_image_remote: https://cloud-images.ubuntu.com/minimal/releases/
```

Alternatively you can follow the instructions below to add an "ubuntu" remote and use that remote when specifying your base image in your Bravefile.

### Adding additional image servers
It's possible to add remote image server as bravetools remotes and explicitly select which remote to use to retrieve the base image in the Bravefile instead of adjusting the default in bravetools config. This can be useful if you want to use a remote by default except for certain images - for example, using https://images.lxd.canonical.com for most images but https://cloud-images.ubuntu.com/releases for cloud Ubuntu server images.

For example, you could add a remote named "ubuntu" for Ubuntu server cloud images (although this remote is created for you during `brave init`):
```sh
brave remote add --protocol simplestreams --public ubuntu https://cloud-images.ubuntu.com/releases/                     
```

Then in the Bravefile specify that this remote is to be used to retrieve the base image:
```yaml
image: example-image/v1.0

base:
  image: ubuntu:20.04

service:
  name: example-container
```


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


## Remote image builds

By default, Bravetools uses a `local` remote for an image build. On Mac/Windows, this is a Multipass VM, whilst on Linux host this is your local LXD server. Sometimes, it may be desirable to use a remote LXD server to cary out Image builds. For example, if your remote has a different CPU architecture (arm64 vs x86) or has more allocated resources.

To use a remote LXD server to build an image use the `--remote` flag of the `brave build` command to select the remote. For example, the following command will use the remote named "utm_x86-64" to build the image and store the result in the local machine's image store.

```sh
brave build --remote utm_x86-64
```

Note that if you build an image on a remote with a different CPU architecture than your current machine you will not be able to launch that image on your local machine's LXD server. However you can deploy that image to any remote LXD server running on the same CPU architecture (x86_64 for the above example).
