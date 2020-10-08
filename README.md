[![Gitter](https://badges.gitter.im/bravetools/community.svg)](https://gitter.im/bravetools/community?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge) [![Go Report Card](https://goreportcard.com/badge/github.com/bravetools/bravetools)](https://goreportcard.com/report/github.com/bravetools/bravetools)

# Bravetools
Bravetools is an end-to-end System Container management platform. Bravetools makes it easy to configure, build, and deploy reproducible environments either on single machines or large clusters.

## Why use Bravetools

Configurable system images have a lot advantages, but their use has been limited. In our own development practice, we found that there were either no existing tools to automate the full lifecycle of a System Container or they had a steep learning curve.

* **Improved Stability**. All software and configurations are installed into your images at build-time. Once your image is launched and tested, you can be confident that any environment launched from that image will function properly.

* **No overheads** of a VM. Bravetools runs on LXD. LXD uses Linux containers to offer a user experience similar to virtual machines, but without the expensive overhead. You can run either single images on a local machines or scale to thousands of compute nodes.

* **Focus on code not infrastructure**. Maintaining and configuring infrastructure is difficult! With any application built and deployed using Bravetools infrastructure and environment have to be configured just once. Developers can spend more time on creating and improving software and less time on managing production environments.

## Quick tour

Here's a toy example showing how to create a simple container configuration, add some useful packages to it, and deploy your image as a service.

Configuration instructions are stored in a [Bravefile](https://bravetools.github.io/bravetools/docs/bravefile/). Let's crate a simple Bravefile that uses Alpine Edge image and installs python3:

```bash
touch Bravefile
```

Now populate your Bravefile:

```yaml
base:
  image: alpine/edge/amd64
  location: public
packages:
  manager: apk
  system:
  - python3
service:
  image: alpine-example-1.0
  name: alpine-example
  docker: "no"
  version: "1.0"
  ip: ""
  ports: []
  resources:
    ram: 4GB
    cpu: 2
    gpu: "no"
```

To create an image from this configuration, run:

```bash
$ brave build

[alpine-example] IMPORT:  alpine/edge/amd64
[alpine-example] RUN:  [apk update]
fetch http://dl-cdn.alpinelinux.org/alpine/edge/main/x86_64/APKINDEX.tar.gz
...

OK: 56 MiB in 30 packages
Exporting image alpine-example
9691e2cf3a58abd4ca411e8085c3117a
```

Finally, we can deploy this image as a container:

```bash:
$ brave deploy

Importing alpine-example-1.0.tar.gz
```

Confirm that the service is up and running:

```bash
NAME            STATUS  IPV4              DISK  PROXY
alpine-example  Running eth0:10.0.0.117                                      
```

This is a very basic example - Bravertools makes it easy to create complex System Containers, abstracting configuration options such as [GPU support](https://bravetools.github.io/bravetools/docs/gpu-units/), [Docker integration](https://bravetools.github.io/bravetools/docs/docker/), and seamless port-forwarding, just to name a few. To learn more about using Bravetools, please refer to our [Bravetools Documentation](https://bravetools.github.io/bravetools/).

## Installing Bravetools

### Latest stable binary 

To get started using Bravetools:

1. Download a platform-specific binary, rename it to `brave`, and add it to your PATH variable:

| Operating System | Binary | Version |
|------------------|--------|---------|
| Ubuntu           | [download](https://github.com/bravetools/bravetools/releases/download/1.54/brave-release-1.54-ubuntu) | release-1.54 |
| macOS            | [download](https://github.com/bravetools/bravetools/releases/download/1.54/brave-release-1.54-darwin) | release-1.54 |
| Windows 8/10     | [download](https://github.com/bravetools/bravetools/releases/download/1.54/brave-release-1.54-win.exe)  | release-1.54 |

2. Add your user to `lxd group`:
```bash
sudo usermod --append --groups lxd $USER
```


### Install from source

Bravetools can be built from source on any platform that supports Go and LXD.

### Ubuntu

**Minimum Requirements**
* Operating System
  * Ubuntu 18.04 (64-bit)
* Hardware
  * 2GB of Memory
* Software
  * [Go](https://golang.org/)
  * [LXD >3.0.3](https://linuxcontainers.org/lxd/getting-started-cli/)

```bash
git clone https://github.com/bravetools/bravetools
cd bravetools
make ubuntu
```

Add your user to `lxd group`:
```bash
sudo usermod --append --groups lxd $USER
```

You may also need to install `zfsutils`:

```bash
sudo apt install zfsutils-linux
```

If this is your first time setting up Bravetools, run `brave init` to initialise the required profile, storage pool, and LXD bridge.

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
```

If this is your first time setting up Bravetools, run `brave init` to initialise the required profile, storage pool, and LXD bridge.


### Windows

**Minimum Requirements**
* Operating System
  * Windows 8 (64-bit)
* Hardware
  * 8GB of Memory
* Software
  * [Go](https://golang.org/)
  * [Multipass](https://multipass.run/)
  * BIOS-level hardware virtualization support must be enabled in the BIOS settings.

```bash
git clone https://github.com/beringresearch/bravetools
cd bravetools
go build -ldflags=“-s -X github.com/bravetools/bravetools/shared.braveVersion=VERSION” -o brave.exe
```

Where VERSION reflects the latest stable release of Bravetools e.g `shared.braveVersion=1.53`

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
  remove      Remove a Unit or an Image
  start       Start Unit
  stop        Stop Unit
  umount      Unmount <disk> from UNIT
  units       List Units
  version     Show current bravetools version

Flags:
  -h, --help   help for brave
```