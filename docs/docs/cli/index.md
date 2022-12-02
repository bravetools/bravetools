---
layout: default
title: CLI
description: Command Line Interface
parent: Docs
has_children: true
nav_order: 1
---

## Command Line Interface

To list available commands, either run `brave` without parameters, or execute `brave help`.

```bash
A complete System Container management platform

Usage:
  brave [command]

Available Commands:
  base        Pull a base image from LXD Image Server or public Github Bravefile
  build       Build an image from a Bravefile
  compose     Compose a system from a set of images
  configure   Configure local host parameters
  deploy      Deploy Unit from image
  help        Help about any command
  images      List images
  import      Import LXD image tarballs into local Bravetools image repository
  info        Display workspace information
  init        Create a new Bravetools host
  mount       Mount a directory to a Unit
  publish     Publish deployed Units as images
  remove      Remove Units or Images
  start       Start Units
  stop        Stop Units
  template    Generate a template Bravefile
  umount      Unmount <disk> from UNIT
  units       List Units
  version     Show current bravetools version

Flags:
  -h, --help   help for brave

Use "brave [command] --help" for more information about a command.
```