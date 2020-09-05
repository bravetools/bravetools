---
layout: default
title: Use Cases
parent: Intro
nav_order: 2
---

In this section we highlight *some* uses cases for Bravetools, predominantly drawn from our own experiences and development pipelines. As Bravetools evolves, we're really excited to see additional applications where it can add value and benefit!


# Dev/Prod Parity
Bravetools was conceived to keep development, staging, and production environments as similar as possible. The key advantage for us were:

* Making the time gap small - our developers write code and deploy it within hours or even just minutes later.
* Making the personnel gap small - as a small company, our development team is also very small 😊. Developers who wrote code are also closely involved in deploying it and watching its behavior in production.
* No more "Works on my machine!" gap - when development and production environments are as similar as possible, tracking down software bugs becomes actually manageable.

# Edge/Cloud Interchange
Many applications that are being developed internally at Bering run both on edge devices and on large cloud-based compute infrastructures. We maintain only one device-agnostic version of our code base and using bravetools, [deploy](../../docs/cli/brave_deploy) software images to appropriate hardware.

# Pain-free dependency control
Whether it's managing multi-user Data Science environments or deploying legacy code, working with system-level dependencies is painful. Since Bravetools creates consistent images from [structured configuration files](../../docs/bravefile), we can explicitly control which dependencies are bundled with our code. Each image derived form the same configuration file is guaranteed to work on all supported machines.
