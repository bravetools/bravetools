---
layout: default
title: brave build
parent: CLI
nav_order: 1
---

# brave build

Build an image from a Bravefile

```
brave build
```

## Description

The `brave build` command builds an image by systematically following instructions specified in a [Bravefile](../../bravefile). The output of the command is an image tar ball stored in the `$HOME/.bravetools` directory and accessible through the `brave images` command.

By default, `brave build` looks for a Bravefile in the current working directory, however an absolute path to a Bravefile can also be passed using the `-p, --path` option.

## Examples

Consider the following Bravefile that describes Ubuntu Bionic with a miniconda3 distribution:

```yaml
base:
  image: ubuntu/bionic/amd64
  location: public
packages:
  manager: apt
  system:
  - wget
run:
- command: wget
  args:
  - https://repo.anaconda.com/miniconda/Miniconda3-latest-Linux-x86_64.sh
- command: bash
  args:
  - Miniconda3-latest-Linux-x86_64.sh
  - -b
  - -p
  - $HOME/miniconda
service:
  name: ubuntu-bionic-miniconda3
  version: "1.0"
  ip: ""
  ports: []
  resources:
    ram: 4GB
    cpu: 2
    gpu: false
```

To build this image execute within the same directory:

```bash
brave build
```

## Options

```
  -h, --help          help for build
  -p, --path string   Absolute path to Bravefile [OPTIONAL]
```

## See Also

* [brave](brave.md)	 - A complete System Container management platform

