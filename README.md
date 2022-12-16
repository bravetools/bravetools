[![Gitter](https://badges.gitter.im/bravetools/community.svg)](https://gitter.im/bravetools/community?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge) [![Go Report Card](https://goreportcard.com/badge/github.com/bravetools/bravetools)](https://goreportcard.com/report/github.com/bravetools/bravetools)

![](https://github.com/bravetools/bravetools/blob/master/docs/assets/cli-bravetools-demo.gif)

# Bravetools
Bravetools is an end-to-end tool for creating and managing applications and environments using [System Containers](https://ubuntu.com/server/docs/containers-lxc). It uses a single source configuration to make it easy to build, deploy, and scale machine images.

Bravetools runs on Linux, MacOS, and Windows.

# Features

* [Build, version control, and share reproducible application and environment images](https://bravetools.github.io/bravetools/docs/bravefile/).
* [Compose multi-container systems](https://bravetools.github.io/bravetools/docs/compose/) in a simple and declarative way.
* [Deploy your systems](https://bravetools.github.io/bravetools/docs/cli/brave_deploy/) locally or remotely.

And [many more](https://bravetools.github.io/bravetools/intro/use_cases/).


# Installation

Prerequisites:

* Mac/Windows: [Multipass](https://multipass.run)
* Linux:
  - [LXD](https://linuxcontainers.org/lxd/getting-started-cli/)
  - Ensure your user belongs to the `lxd` group: `sudo usermod --append --groups lxd $USER`
  - You may also need `zfsutils`: `sudo apt install zfsutils-linux`

1. Download the [latest stable release](https://github.com/bravetools/bravetools/releases) for your host platform and add it to your `$PATH`.

2. Run `brave init` to get started.

# Installing from source

## Linux/MacOS
```bash
git clone https://github.com/bravetools/bravetools
cd bravetools
make [ubuntu]/[darwin]
```

## Windows
```bash
git clone https://github.com/bravetools/bravetools
cd bravetools
go build -ldflags=“-s -X github.com/bravetools/bravetools/shared.braveVersion=VERSION” -o brave.exe
```


# Command Reference

```
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
  remote      Manage remotes
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