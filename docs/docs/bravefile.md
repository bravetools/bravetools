---
layout: default
title: Writing a Bravefile
parent: Docs
nav_order: 3
description: "Bravetools builds images automatically by reading instructions from a Bravefile."
---

# Anatomy of a Bravefile

Bravetools creates reproducible environments by following instructions logged in a ``Bravefile``. A ``Bravefile`` adheres to a strict structure, which ensures reproducibility and makes it easy to manage complex system structures and operations.

## Key Components

The minimal structural unit of a Bravefile is an **Entry**. ``Bravefile`` supports five entry types - base, system, copy, run, and service.

### base
Describes base requirements for your image, such as base image and location of the image file.

```yaml
base:
  image: alpine/edge/amd64
  location: public
```

Three types of image locations are supported:

1. ``public`` - specifies that images are to be pulled from the [linuxcontainers](https://images.linuxcontainers.org) repository. Accepted image name syntax is ``Distribution/Release/Architecture``.
2. ``local`` - images stored locally. Naming follows the convention ``image-name-version``.
3. ``github`` - images that can be built and imported on the fly from Bravefiles stored inside GitHub directories. Naming convention is ``username/repository/directory``. Bravetools will search for a Bravefiles inside the ``/directory`` location.

In cases where Bravefiles are ingested from GitHub, a local copy of the resulting image will be kept. The local image copy will be re-used next time you run ``brave build``.

If the location field is not present, bravetools will resolve the image location itself. Local images will be checked first, then public LXD images. Image names starting with "github.com/" will be imported from GitHub.

### system
Describes system packages to be installed through a specified package manager. Supported package managers are ``apt`` and ``apk``.

```yaml
packages:
  manager: apk
  system:
  - bash
  - curl
  - openjdk8
  - gcc
  - g++
  - linux-headers
  - zip
  - python3-dev
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
- command: ln
  args:
  - -s
  - /usr/bin/python3
  - /usr/bin/python
```

### service
Controls image properties, such as name, version, and run-time configuration. It is also possible to specify  post-deployment operations, such as ``copy`` and ``run``.

```yaml
service:
  name: alpine-edge-bazel
  version: 0.27.1
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
```

## Brave Configuration Language (BCL)

BCL is a simplified configuration script for Bravetools Images. It is json-based and supports arbitrary TAB and SPACE placements, as well as comments. BCL can be installed through a [github repository](https://github.com/beringresearch/bcl)