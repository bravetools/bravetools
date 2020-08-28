---
layout: default
title: brave base
parent: CLI
nav_order: 1
---

# brave base

Build a base unit

```
brave base NAME
```

## Description

Base images are the building blocks of every environment created by Bravetools. Base images can be created quickly on the command line by either pulling from the [linuxcontainers](https://images.linuxcontainers.org) repository, or by specifying a GitHub repository.

Base images are subsequently modified by instructions inside a [Bravefile](../../bravefile).

## Examples

### Base images as clean Linux distributions

Many flavours of Linux can be used as base images. For a full list refer to [linuxcontainers](https://images.linuxcontainers.org). The typical argument naming convention is ``Distribution/Release/Architecture``.

```bash
# Create an Alpine Edge base image
brave base alpine/edge/amd64

# Create Ubuntu 18.04 LTS base image
brave base ubuntu/bionic/amd64
```

### Base images from GitHub repositories

Sometimes it may be desirable to build a base image using a Bravefile stored inside a GitHub repository. This can be especially useful if your image is highly customised and builds require portability. Bellow is an example of an Ubuntu Bionic base with Python3 built using a custom Bravefile located at https://github.com/beringresearch/bravefiles/tree/master/ubuntu/ubuntu-bionic-py3.

GitHub builds follow the convention ``github.com/username/repository/directory/subdirectory``.

```bash
brave base github.com/beringresearch/bravefiles/ubuntu/ubuntu-bionic-py3
```

## Options

```
  -h, --help   help for base
```

## See Also

* [brave](brave.md)	 - A complete System Container management platform

