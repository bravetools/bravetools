---
layout: default
title: Quickstart
parent: Intro
nav_order: 3
---

# Quickstart

This overview will demonstrate how to quickly create and deploy a Bravetools image on your system.

{: .no_toc }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

## Install Bravetools

Begin by [installing the latest binary](../../installation) of Bravetools. When Bravetools is installed for the first time, your host will need to be configured to interact with LXD either directly (Linux) or via Multipass (Mac, Windows). This is achieved through a single command:

```bash
brave init
```

## Define an Image

Every image is configured through specifications described in a Bravefile. Let's create a simple configuration for an Alpine Edge image with a Python3 installation:

```bash
touch Bravefile
```

Populate the empty Bravefile with this simple yaml config:

```yaml
image: alpine-python3/1.0
base:
  image: alpine/edge
  location: public
packages:
  manager: apk
  system:
    - python3-dev
    - python3
service:
  name: alpine-edge-python3
  image: alpine-python3/1.0
  resources:
    ram: "4GB"
    cpu: 2
```

## Build the Image
Now build the image:

``` bash
$ brave build
```

You can verify that the image was created by listing all images available on your system:

```bash
$ brave images
```

The output should look something like this:

```bash
IMAGE                  	VERSION	ARCH 	CREATED   	SIZE 	HASH
alpine-python3         	1.0    	arm64	just now  	47MB 	6c3199081322e7006bc2b3b631e6cf9a
```

Note that Bravetools will build an image for your host CPU architecture. In case of this example, this is ARM64. You can configure a [bravetools remote](../../docs/remotes) to take advantge of multi-arch build capabilities. 

## Deploy the Image

Since Bravetools uses a single configuration file for both building and deploying your image, all you need to deploy this Alpine image is already specified in the `service` section of you Bravefile. Deploy this image just by running:

```bash
$ brave deploy
```

Verify that the image is deployed:

```bash
$ brave units

NAME               	STATUS 	IPV4          	VOLUMES                                           	PORTS
alpine-edge-python3 Running	10.137.211.85 	root:->/                                          	
```

To delete the live image, run

```bash
$ brave remove alpine-edge-python3
```

That's it! You have now configured and deployed your very first Bravetools image. Check out the [docs](../../docs) to dive deeper.