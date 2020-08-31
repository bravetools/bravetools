# Bravetools
Bravetools is an end-to-end System Container management platform. Bravetools makes it easy to configure, build, and deploy reproducible and isolated environments either on single machines or large clusters.

## Quickstart

To get started using Bravetools, download a platform-specific binary and add it to your PATH variable:

| Operating System | Binary | Version |
|------------------|--------|---------|
| Ubuntu           | [download](https://github.com/bravetools/bravetools/releases/download/1.53/brave-release-1.53-ubuntu) | release-1.53 |
| macOS            | [download](https://github.com/bravetools/bravetools/releases/download/1.53/brave-release-1.53-darwin) | release-1.53 |
| Windows 8/10     | [download](https://github.com/bravetools/bravetools/releases/download/1.53/brave-release-1.53-win.exe)  | release-1.53 |

> **NOTE:** Bravetools can be built from source on any platform that supports Go.

## Using Bravetools

To learn more about using Bravetools, please refer to our [Bravetools Documentation](https://beringresearch.github.io/bravetools/).

## Install from Source

### Ubuntu

**Minimum Requirements**
* Operating System
  * Ubuntu 18.04 (64-bit)
* Hardware
  * 4GB of Memory
* Software
  * [Go](https://golang.org/)
  * [LXD 4.3](https://linuxcontainers.org/lxd/getting-started-cli/)

> **NOTE**: LXD up to 3.0.x were published as non-snap versions. Bravetools will not work with these distributions. The user is encouraged to use snap-LXD before continuing with installation.

```bash
git clone https://github.com/bravetools/bravetools
cd bravetools
make ubuntu
brave init
```

### Mac OS

**Minimum Requirements**
* Operating System
  * MacOS Mojave (64-bit)
* Hardware
  * 4GB of Memory
* Software
  * [Go](https://golang.org/)
  * [Multipass](https://multipass.run/)

```bash
git clone https://github.com/bravetools/bravetools
cd bravetools
make darwin
brave init
```

### Vagrant

1. Start Vagrant VM:

```bash
cd vagrant
vagrant up
vagrant ssh

// execute inside Vagrant VM
cd $HOME/workspace/src/github.com/bravetools/bravetools
make ubuntu
brave init
```

### Update Bravetools

To update existing installation of Bravetools for your platform:

```bash
git clone https://github.com/bravetools/bravetools
cd bravetools
make [darwin][ubuntu]
```

## Build Documentation

Follow installation instructions for [Jekyll](https://jekyllrb.com/) on your platform.
To serve documentation locally run:

```bash
cd docs
bundle exec jekyll serve --trace
```

and point your browser to http://127.0.0.1:4000/bravetools/.


## Command Reference

```
Usage:
  brave [command]

Available Commands:
  base        Build a base unit
  build       Build an image from a Bravefile
  configure   Configure local host parameters such as storage
  deploy      Deploy Unit from image
  help        Help about any command
  images      List images
  import      Import a tarball into local Bravetools image repository
  info        Display workspace information
  init        Create a new Bravetools host
  mount       Mount a directory to a Unit
  publish     Publish deployed Unit as image
  remote      Add a remote connection to a Bravetools host via an IP address
  remove      Remove a Unit or an Image
  start       Start Unit
  stop        Stop Unit
  umount      Unmount <disk> from UNIT
  units       List Units
  version     Show current bravetools version

Flags:
  -h, --help   help for brave
```