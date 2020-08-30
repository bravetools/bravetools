---
layout: default
title: Why use Bravetools
parent: Intro
nav_order: 1
---

# Why use Bravetools

Configurable system images have a lot advantages, but their use has been limited. In our own development practice, we found that there were either no existing tools to automate the full lifecycle of a System Container or they had a steep learning curve.

Bravetools addresses these limitations. It has a simple and lightweight command line interface underpinned by a fixed set of modern development principles. It encourages structure and reproducibility in your configuration, helping you to focus on maintaining your application code rather than your environment.

## Advantages of using Bravetools

**Improved Stability**. All software and configurations are installed into your images at build-time. Once your image is launched and tested, you can be confident that any environment launched from that image will function properly.

**No overheads of a VM**. Bravetools runs on [LXD](https://linuxcontainers.org/#LXD). LXD uses Linux containers to offer a user experience similar to virtual machines, but without the expensive overhead. You can run either single images on a local machines or scale to thousands of compute nodes.

**Focus on code not infrastructure**. Maintaining and configuring infrastructure is difficult! With any application built and deployed using Bravetools infrastructure and environment have to be configured just once. Developers can spend more time on creating and improving software and less time on managing production environments.

Let's [get started](../../installation)!