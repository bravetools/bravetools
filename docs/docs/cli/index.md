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
brave - tools for working with the Brave Platform

Usage:
  brave [command]

Available Commands:
  base        Build a base unit
  build       Build an image from a Bravefile
  configure   Configure local host parameters such as storage
  deploy      Deploy Unit from image
  help        Help about any command
  images      List images
  import      Import an image into local Bravetools image repository
  info        Display workspace information
  init        Create a new Bravetools host
  mount       Mount directory to a Unit
  publish     Publish deployed Unit as image
  remote      Add a remote connection to a Bravetools host
  remove      Remove Unit or Image
  start       Start Unit
  stop        Stop Unit
  umount      Unmount <disk> from UNIT
  units       List Units
  version     Show current bravetools version

Flags:
  -h, --help   help for brave

Use "brave [command] --help" for more information about a command.
```