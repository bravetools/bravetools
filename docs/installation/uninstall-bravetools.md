---
layout: default
title: Uninstall Bravetools
parent: Install Bravetools
nav_order: 5
description: "Instructions to uninstall Bravetools"
---

# Uninstall Bravetools on MacOS/Windows with Multipass host

```bash
multipass delete bravetools; multipass purge
rm -r ~/.bravetools
```

# Uninstall Bravetools on Linux systems
Artefacts created by Bravetools can be uninstalled using LXD:

```bash
lxc profile delete bravetools-$USER
lxc storage delete bravetools-$USER
lxc network delete bravetoolsbr0
```

Finally, remove Bravetools images, databases, and certificates:

```bash
rm -r ~/.bravetools
```

