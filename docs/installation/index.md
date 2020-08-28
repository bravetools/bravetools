---
layout: default
title: Install Bravetools
description: Lists the installation methods
has_children: true
nav_order: 2
---

## Supported platforms

Bravetools is available on a variety of Linux platforms, macOS, and Windows 10 as a static binary.
The following table lists minimal requirements on each supported platform:

| Platform      			 | Hardware          | Software 	  |
|:---------------------------|:------------------|:---------------|
| [Ubuntu 64-bit](ubuntu.md) | 4 GB Memory		 | [LXD 4.3](https://linuxcontainers.org/lxd/introduction/) |
| [macOS Mojave](macos.md)	 | 8 GB Memory		 | [Multipass](https://multipass.run/) |
| [Windows 10](windows.md)	 | 8 GB Memory 		 | [Multipass](https://multipass.run/) |

## Release channels

Bravetools has three types of update channels - **stable**, **test**, and **nightly**.

### Stable

Stable releases are made from a release branch diverged from the master branch. All further patch releases are performed from that branch.

### Test

Test release is rolled out from a branch created from a master branch when milestones outlined in the Project Plan have achieved feature-complete status.

### Nightly

Nightly builds are automatically generated once per day from the master branch. These builds allow for testing from the latest code on the master branch.