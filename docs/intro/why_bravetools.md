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

**No overheads of a VM**. Bravetools runs on [LXD](https://linuxcontainers.org/lxd/introduction/). LXD uses Linux containers to offer a user experience similar to virtual machines, but without the expensive overhead. You can run either single images on a local machines or scale to thousands of compute nodes.

**Focus on code not infrastructure**. Maintaining and configuring infrastructure is difficult! With any application built and deployed using Bravetools infrastructure and environment have to be configured just once. Developers can spend more time on creating and improving software and less time on managing production environments.

## Why not Docker?

Docker is the industry-standard way for building **Application Containers**. Surprising as it may seem, Bravetools was not designed to compete with Docker! In fact, you can [run Docker containers inside Bravetools Units](../docs/../docker). The main reasons we decided to use System Containers for our infrastructure and core are:

* **Docker containers were being treated like VMs**. Our team was pushing Vim, python, RStudio, R, libraries, and Shiny webservers into the same Docker container and treating it like a lightweight VM. Docker containers are simply not designed to run multiple processes and our approach was an anti-pattern.

* **Valuable time was being spent on devops and not research**. As our container practices improved, the number of Docker containers required to run our research pipelines increased drastically. We found that we were spending more time on managing containers instead of doing iterative research.

* **Persistent data storage and data sharing is complicated**. By design, all files created inside a Docker container are stored on a writable container layer. This means that the data doesn’t persist when that container no longer exists, and it can be difficult to get the data out of the container if another process needs it. Furthermore, a container’s writable layer is tightly coupled to the host machine where the container is running - you can’t easily move the data somewhere else. Docker volumes are a good workaround, but are difficult to share and reproduce across collaborating teams.

## Let's ...
[get started](../../installation)!