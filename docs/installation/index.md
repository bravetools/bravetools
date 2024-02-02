---
layout: default
title: Install Bravetools
description: Lists the installation methods
has_children: true
nav_order: 2
---

## Supported platforms

Bravetools is available on a variety of Linux platforms, macOS, and Windows as a static binary.
The following table lists minimal requirements on each supported platform to use bravetools with a backend
running on your local machine:

| Platform      			 | Hardware          | Software 	  |
|:---------------------------|:------------------|:---------------|
| [Ubuntu 64-bit](ubuntu.md) | 1 GB Memory		 | [LXD >=3.0.3](https://documentation.ubuntu.com/lxd/en/latest/) |
| [macOS Mojave](macos.md)	 | 8 GB Memory		 | [Multipass](https://multipass.run/) |
| [Windows 10](windows.md)	 | 8 GB Memory 		 | [Multipass](https://multipass.run/) |

If using a [remote backend](../docs/init.md#remote-backends) the above requirements do not apply - the bravetools binary 
alone should be enough.

The easiest way to install Bravetools is to download the [latest stable release](https://github.com/bravetools/bravetools/releases) for your host platform and add it to your `$PATH`.