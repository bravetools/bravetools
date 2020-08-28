---
layout: default
title: brave mount
parent: CLI
nav_order: 1
---

# brave mount

Mount a directory to a Unit

```
brave mount [UNIT:]<source> UNIT:<target>
```

## Description

When you use `brave mount` a file or directory on the source machine or unit is mounted into a target unit. On the unit filesystem, a new directory is created and Bravetools manages that directory's content.

## Examples

### Mounting a host directory  to a unit:

To make a local directory accessible on a live unit, execute:

```bash
brave base alpine/edge/amd64
brave deploy brave-base-alpine-edge-1.0 --name alpine-edge

brave mount /PATH/TO/LOCAL/DIR alpine-edge:/PATH/TO/REMOTE/DIR
```

Local directory is now accessible at your preferred location on the live unit.

>**NOTE** It's important to ensure that file permissions are relaxed and set to 777 in order for the unit to see host directories and files.

### Sharing directories across units

`brave mount` also supports mounting directories between one or more units:

```bash
brave base alpine/edge/amd64
brave deploy brave-base-alpine-edge-1.0 --name alpine-edge-A
brave deploy brave-base-alpine-edge-1.0 --name alpine-edge-B

brave mount alpine-edge-A:/PATH/TO/LOCAL/DIR alpine-edge-B:/PATH/TO/REMOTE/DIR
```

In this case a directory on the `alpine-edge-A` units is exposed to `alpine-edge-B`.

## Options

```
  -h, --help   help for mount
```

## See Also

* [brave](brave.md)	 - A complete System Container management platform

