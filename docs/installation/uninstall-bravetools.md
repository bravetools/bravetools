---
layout: default
title: Uninstall Bravetools
parent: Install Bravetools
nav_order: 5
description: "Instructions to uninstall Bravetools"
---

# Uninstall Bravetools on MacOS/Windows with Multipass host

```bash
multipass delete $USER; multipass purge
rm -r ~/.bravetools
```

# Uninstall Bravetools on Linux systems
Artefacts created by Bravetools can be uninstalled using LXD:

```bash
lxc profile delete $USER
lxc storage delete $USER-[TIMESTAMP] # You will need to get specific storage name using lxc storage list
lxc network delete $USER"br0"
```

Finally, remove Bravetools images, databases, and certificates:

```bash
rm -r ~/.bravetools
```

