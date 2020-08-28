---
layout: default
title: brave deploy
parent: CLI
nav_order: 1
---

# brave deploy

Deploy Unit from image

```
brave deploy IMAGE
```

## Description

Units are our shorthand for **System Containers** with some additional elements thrown in. System Containers are similar to Virtual Machines, since they share host's kernel and provide user space isolation. However, they bring additional [efficiencies](https://linuxcontainers.org/lxd/introduction/#whats-lxd). Additionally, System Containers support installation of various libraries and services, which makes them well-suited to running both dedicated tasks and environments. It is this combination that we lovingly refer to as a **Unit**.

Units are launched from pre-built images and always run as background processes. There are several useful command line arguments to speed up deployments, for example see [networking](#networking). More advanced unit configurations are expressed in a Bravefile using the [service](#launching-a-unit-using-bravefile) section.

## Examples

### Launching a unit from command line

Let's create an Alpine Edge base image:

```bash
brave base alpine/edge/amd64
```

You can verify that the image was successfully generated:

```bash
brave images

IMAGE                           CREATED         SIZE            HASH                             
brave-base-alpine-edge-1.0      just now        4.9 MB          3533785ddf47507cfd9e49c4454f2b15
```

This base image was automatically assigned a name and hash that can verify its integrity in the future should you wish to reuse it elsewhere. We can now launch this image as a System Container. Technically, since this is a bare-bones Alpine OS, it's not yet a unit!

You can launch this unit using:

```bash
brave deploy brave-base-alpine-edge-1.0 --name alpine-edge

Importing brave-base-alpine-edge-1.0.tar.gz
Unit launched:  alpine-edge
Service started:  alpine-edge
```

To verify that the launch was successful, list all running units:

```bash
brave units

NAME            STATUS  IPV4            DEVICES           
alpine-edge     Running 10.0.0.205      eth0(nic):bridged
```

This unit is now accessible on 10.0.0.205.

### Launching a unit using Bravefile

Advanced unit configurations beyond IP address and forwarded port should be passed through a Bravefile. This encourages full reproducibility of complex environments. Unit configurations are controlled through the **service** section at the very end of a Bravefile:

```yaml
service:
  image: brave-base-alpine-edge-1.0
  name: alpine-edge
  version: 1.0
  ip: 10.0.0.10
  ports:
    - 8888:8988
  resources:
    ram: 4GB
    cpu: 1
```

`brave deploy` will by default preferentially parse a Bravefile **service** section to configure a unit. As you can see, you can have very tight control over the networking parameters as well as hardware resources.


### Networking

Note that by default, Bravetools assigns a random static IP address to you launched unit. Sometimes it's desirable to create a pre-defined IP address and set up port forwarding between the unit and host. This can be easily achieved through the command line arguments:

```bash
brave deploy brave-base-alpine-edge-1.0 --name alpine-edge -ip 10.0.0.10 --port 8888:8888
```

The command will launch the unit, assign it a static ip address 10.0.0.10, and forward unit port 8888 to local host port 8888. Port forwarding follows UNIT:HOST convention.

## Options

```
      --config string   Path to Unit configuration file [OPTIONAL]
  -h, --help            help for deploy
  -i, --ip string       IPv4 address (e.g., 10.0.0.20) [OPTIONAL]
  -n, --name string     Assign name to deployed Unit
  -p, --port string     Publish Unit port to host [OPTIONAL]
```

## See Also

* [brave](brave.md)	 - A complete System Container management platform

