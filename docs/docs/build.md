---
layout: default
title: Brave Build
parent: Docs
nav_order: 3
description: "Bravetools builds images automatically by reading instructions from a Bravefile."
---

# Building an Image
{: .no_toc }

A `Bravefile` defines a set of instructions and configuration options for building a single system container. Bravetools supports [multiple CPU architectures](https://documentation.ubuntu.com/lxd/en/latest/architectures/) and a [wide number of base Linux distributions](https://uk.lxd.images.canonical.com).

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Image Build Instructions

Lets consider a simple `Bravefile`:

```yaml
image: cowsay/1.0
base:
  image: alpine/edge
  location: public
packages:
  manager: apk
  system:
  - python3
  - py3-pip
run:
- command: sh
  args:
  - -c
  - python3 -m pip install cowsay
- command: sh
  args:
  - -c
  - python3 -c "import cowsay; cowsay.cow('Hello World!')"
service:
  name: cowsay
  docker: "no"
  resources:
    ram: 1GB
    cpu: 2
    gpu: "no"

```

This image will use Alpine Edge as its base and running `brave build` will install `python3` and `pip3` system packages using Alpine's `apk` package manager.

> NOTE: Bravetools supports [a large number of base Linux distributions](https://uk.lxd.images.canonical.com). These can be imported by setting the base `image` field to NAME/VERSION/ARCH. For example, `ubuntu/focal/arm64`.

## Image Versioning
Bravetools encourages incremental version control of each image. Image name and version can be specified in the `image` field of the `Bravefile`. For example, `cowsay/1.0`. If version is not specified, Bravetools will automatically add `untagged` label to the image.

Specific image version can be referenced in the `service` section during deployment.

## Specifying a Build Host
Bravetools can use either a local machine or a [preconfigured remote](remotes.md) to perform the build process. This can be useful if, for example, you require large computational resources to build your image or need an image with a non-host architecture.

To specify a remote to be used for your build, run:

```bash
brave build -r $REMOTE
```

Where `$REMOTE` is the name of a trusted Bravetools remote.

## Using a Local Image Store
Every image built by Bravetools can be used as a base for any subsequent image configurations. For example, you might have pre-built images containing the full python3 development environment, which can be re-used as bases for python3-dependent applications.

To use images in the local store, set the `location` field of the Bravefile to `local`.
```yaml
base:
  image: alpine/edge
  location: local
```

## Remote Image Storage
Upon build completion, every Bravetools image is stored locally in `~/.bravetools/images` directory as tar.gz files. This simplifies the process of sharing each image, which can then be imported using [`brave import`](cli/brave_import.md) command.

However, sometimes it can be desirable to also store an image on a remote LXD server, which acts as an [image repository](https://documentation.ubuntu.com/lxd/en/latest/reference/remote_image_servers/#remote-server-types). Bravetools enables this by specifying the remote name in the `image` field:

```yaml
image: qemu:cowsay/1.0
base:
  image: alpine/edge
  location: public
```

Upon build completion, the image will be pushed to a remote called `qemu` for later use and a local copy will be stored on the host's file system

To use an image stored on your remote, reference it in your `Bravefile` as:

```yaml
base:
  image: qemu:alpine/edge
  location: public
```

This will pull the pre-built image from the `qemu` remote and reuse it for downstream builds.