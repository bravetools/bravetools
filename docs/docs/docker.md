---
layout: default
title: Working with Docker
parent: Docs
nav_order: 5
description: "Bravetools enables seamless Docker integration."
---

# Bravetools and Docker

Docker is the industry-standard way for building **Application Containers**. Bravetools provides an intuitive interface to deploy Docker inside our System Container [units](../cli/brave_deploy), making it very easy to ship Docker applications inside secure and reproducible environments.

## Configuring a Bravetools Unit to run Docker

A unit can be configured to run Docker containers using a [Bravefile](../bravefile). This can be achieved simply by adding `docker: "yes"` to the **service** section:

```yaml
service:
  name: ubuntu-bionic-docker
  docker: "yes"
  version: "1.0"
  ip: ""
  ports: []
  resources:
    ram: 4GB
    cpu: 2
    gpu: "no"
```

and then running:

```bash
brave deploy
```

Bravetools will automatically configure this unit to run Docker! A complete Bravefile is available at our [Bravefiles repository](https://github.com/beringresearch/bravefiles/tree/master/ubuntu/ubuntu-bionic-docker)