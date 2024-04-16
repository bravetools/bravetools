---
layout: default
title: Writing a Bravefile
parent: Docs
nav_order: 2
description: "Bravetools builds images automatically by reading instructions from a Bravefile."
---

# Writing a Bravefile
{: .no_toc }

Bravetools creates reproducible environments by following instructions logged in a ``Bravefile``. A ``Bravefile`` adheres to a strict structure, which ensures reproducibility and makes it easy to manage complex system structures and operations.

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## A Simple Bravefile

A ``Bravefile`` follows YAML format and at a high-level looks like this:

```yaml
image: alpine-python3/1.0
base:
  image: alpine/edge
  location: public
packages:
  manager: apk
  system:
  - python3
  - python3-pip
run:
- command: ls
  args:
  - -a
service:
  name: python3
  ip: ""
  resources:
    ram: "1GB"
    cpu: 4
    gpu: "no"
    disk: "10GB"
```

Running `brave build` followed by `brave deploy` on a ``Bravefile`` above, will pull a blank Alpine Edge system image for your CPU architecture, install python3, and make the container available on your network with 1GB of RAM and 4 CPUs. Image itself, will be sotred as `alpine-python3` and can be viewed by running `brave images`.

```bash
IMAGE         	VERSION	ARCH 	CREATED 	SIZE	HASH
alpine-python3	latest 	arm64	just now	20MB	e40460891f90e73ceb17f9952919a571
```

## Key Components

The minimal structural unit of a Bravefile is an **Entry**. ``Bravefile`` supports five entry types - image, base, system, copy, run, and service.

### image
`image` refers to the target image to be built using the instractions in your ``Bravefile``. General syntax is [NAME]/[VERSION]/[ARCH]. If [ARCH] is not specified, Bravetools will automatically determine your host's CPU architecture and build an appropriate image.

```yaml
image: alpine-python3/1.0
```

### base
Describes base requirements for your image, such as base image and location of the image file.

```yaml
base:
  image: alpine/edge/amd64
  location: public
```

Three types of image locations are supported:

1. ``public`` - specifies that images are to be pulled from the configured "public" image remote. The default setting pulls from the [ubuntu cloud](https://cloud-images.ubuntu.com/releases) repository.
2. ``local`` - images stored locally. Naming follows the convention ``NAME/VERSION/ARCH``.
3. ``github`` - images that can be built and imported on the fly from Bravefiles stored inside GitHub directories. Naming convention is ``username/repository/directory``. Bravetools will search for a Bravefiles inside the ``/directory`` location.

In cases where Bravefiles are ingested from GitHub, a local copy of the resulting image will be kept. The local image copy will be re-used next time you run ``brave build``.

If the location field is not present, bravetools will resolve the image location itself. Local images will be checked first, then public LXD images. Image names starting with "github.com/" will be imported from GitHub.

### system
Describes system packages to be installed through a specified package manager. Supported package managers are ``apt`` and ``apk``.

```yaml
packages:
  manager: apk
  system:
  - python3
```

### copy
This is a specialised entity designed for file and directory transfers between hosts and Brave Images. The Entity supports multiple **Blocks**. Each **Block** contains a source and a target. Optionally, action specifies additional actions to perform once the file or directory has been copied to the image. All actions are executed on an image during build.

To copy a single file, include:
```yaml
copy:
  - source: configuration/init.sh
    target: /root/
    action: |-
      chmod +x init.sh
```

To copy a directory:
```yaml
copy:
  - source: configuration
    target: /root/configuration
```

### run
Executes commands on the Brave image during build time. This **Entity** supports multiple Blocks and a diverse range of syntax. In its simplest embodiment, run Entity supports command, followed by an argument string. For example,

```yaml
run:
- command: ls
  args:
  - -a
```

### service
Controls image properties, such as name, version, and run-time configuration. It is also possible to specify  post-deployment operations, such as ``copy`` and ``run``.

```yaml
service:
  image: alpine-python3/1.0   # If not specified at top of file
  name: python3
  profile: brave              # Optional, defaults to your local LXD profile
  network: lxdbr0             # Optional, defaults to your local LXD network bridge
  storage: brave-deploy-disk  # Optional, defaults to your local LXD storage device
  ip: ""
  ports: []
  postdeploy:
    run:
    - command: echo
      args: "Hello World"
    copy:
    - source: /file/or/directory
      target: /file/or/directory
      action: chmod 0700 /file/or/directory
  resources:
    ram: "4GB"
    cpu: 4
    gpu: "no"
    disk: "10GB"
```

If you're deploying to a remote Bravetools host, you can append `<remote>:` to the `name` field. Note that you have to ensure that `profile` and `network` options are set and reflect the set up of your remote LXD instance.

## Brave Configuration Language (BCL)

BCL is a simplified configuration script for Bravetools Images. It is json-based and supports arbitrary TAB and SPACE placements, as well as comments. BCL can be installed through a [github repository](https://github.com/beringresearch/bcl)