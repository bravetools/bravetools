---
layout: default
title: Reproducible Bioinformatics
parent: Examples
nav_order: 1
description: "Using Bravetools to create and ship reproducible environments"
---


# Reproducibile Environments in Bioinformatics

1. TOC
{:toc}

## Motivation

Reproducible research is a key element of modern science as well as a hallmark for any industrial application. However, the raw data, tools, libraries, and workflows may not be enough to guarantee reproducibility. Indeed, different releases of the same tools and/or of the system libraries might lead to unwanted issues when projects are shared between teams or across the wider research community.

Our research team works extensively with Docker to build, maintain and scale a number of bioinformatics and machine learning pipelines. However, we found several issues when using Docker in our _research environments_.

## Why Docker didn't work for us

* **Docker containers were being treated like VMs**. Our team was pushing Vim, python, RStudio, R, libraries, and Shiny webservers into the same Docker container and treating it like a lightweight WM. Docker containers are simply not designed to run multiple processes and our approach was an anti-pattern.

* **Valuable time was being spent on devops and not research**. As our container practices improved, the number of Docker containers required to run our research pipelines increased drastically. We found that we were spending more time on managing containers instead of doing iterative research.

* **Persistent data storage and data sharing is complicated**. By design, all files created inside a Docker container are stored on a writable container layer. This means that the data doesn’t persist when that container no longer exists, and it can be difficult to get the data out of the container if another process needs it. Furthermore, a container’s writable layer is tightly coupled to the host machine where the container is running - you can’t easily move the data somewhere else. Docker volumes are a good workaround, but are difficult to share and reproduce across collaborating teams.

## Our solution

We found that our issues could be largely addressed through System Containers. Our researchers could once more focus on experiments and not devops, whilst ensuring efficiency, scalability, security, and reproducibility of all pipelines. Bellow we provide an overview of our typical Bioinformatics research environment and how Bravetools can be used to build, scale, and publish all dependencies.

> **NOTE**: System Containers are not a replacement for Docker. They are a replacement for VMs. Docker can be easily deployed inside LXD containers, a feature we take advantage of in production.

In this example, we assume that Bravetools [has been installed](../../installation). We will demonstrate how to:

1. Build an Ubuntu 18.04 image with the latest R and RStudio Server
2. Deploy the image as a standalone research environment
3. Publish the research environment as a single file that can then be shared

### Build RStudio Server Image

Let's get started by grabbing the Bravefile that scripts our full environment:

```bash
> wget https://raw.githubusercontent.com/beringresearch/bravefiles/master/ubuntu/ubuntu-bionic-rstudio-server/Bravefile
```

This Bravefile will install R and RStudio Server with all of their dependencies into an Ubuntu 18.04 image. To build the image run:

```bash
> brave build

Creating ubuntu-bionic-rstudio-server
Unit launched:  ubuntu-bionic-rstudio-server ubuntu/bionic/amd64
[ubuntu-bionic-rstudio-server] RUN:  [apt update]

WARNING: apt does not have a stable CLI interface. Use with caution in scripts.

Hit:1 http://archive.ubuntu.com/ubuntu bionic InRelease
Get:2 http://archive.ubuntu.com/ubuntu bionic-updates InRelease [88.7 kB]
Get:3 http://security.ubuntu.com/ubuntu bionic-security InRelease [88.7 kB
...
```

This operation will take a little bit of time, depending on the speed of your Internet connection. When the build completes, you can verify that the image is now stored in your local repository.

```bash
> brave images

IMAGE                           	CREATED   	SIZE 	HASH                            
ubuntu-bionic-rstudio-server-1.0	just now  	709MB	26a46ed26c074d3fa29d2a6fec7dcdfe
```

### Deploy the RStudio Server image as a service

Before deployment, you may wish to change some hardware allocations, e.g. RAM and CPU resources. Edit the Bravefile and modify its `resources` section to fit your specifications:

```yaml
resources:
    ram: 4GB
    cpu: 4
    gpu: "no"
```

Once you're happy with hardware allocations, deploy the image as a unit:

``` bash
> brave deploy

Importing ubuntu-bionic-rstudio-server-1.0.tar.gz
Unit launched:  ubuntu-bionic-rstudio-server
Service started:  ubuntu-bionic-rstudio-server
...
```

During deployment, Bravetools will set you up with a new user `rstudio` and the corresponding password `password`. Now we can check the status of the running unit:

``` bash
> brave units

NAME                        	STATUS 	IPV4      	DEVICES
ubuntu-bionic-rstudio-server	Running	10.0.0.23 	ubuntu-bionic-rstudio-serverproxy-8787:8787
                            	       	          	eth0(nic):bridged
```

A note on networking. Bravetools tries to simplify the networking process by allocating static or dynamic IP address to your unit and by automating port forwarding between the host and the unit. These parameters can be modified in the `service` section of the Bravefile:

```yaml
service:
  ip: ""
  ports:
  - 8787:8787
```

A blank IP field will result in an ephemeral IP being generated for this unit. Since a default RStudio runs on port 8787, we will leave this unchanged.

You can now navigate to 10.0.0.23:8787 and login using `rstudio`/`password` credentials.

> **NOTE** If you're using multipass (e.g. on a Mac or Windows machine), RStudio service will be accessible from HOSTIP:8787, where HOSTIP is the IPV4 address of your multipass host, easily obtained by running `brave info`.

As a side node, installing an OpenSSH server inside this environment, will allow others to connect to it over ssh, facilitating collaborative research.

### Publishing the environment

Now that the RStudio environment is deployed, you can work inside the unit as though you're on a remote server or inside a virtual machine. Data and packages can be added as needed.

> **NOTE** If you do install additional packages, do update your Bravefile, to ensure that de novo builds are consistent with your existing environment.

Once you're happy with your environment and are ready to migrate it to a more powerful infrastructure, share it with your team, or make it available as part of your publication, you need to simply run:

```bash
> brave publish ubuntu-bionic-rstudio-server
```

This command will create a file `ubuntu-bionic-rstudio-server-TIMESTAMP.tar.gz` in the current working directory. The file is our RStudio image with all the data and dependencies, which can be shared together with the Bravefile as needed.

To restore the original environment, simply import the image archive into Bravetools and deploy it:

```bash
> brave import ubuntu-bionic-rstudio-server-TIMESTAMP.tar.gz
> brave deploy
```

## Conclusions

System Containers offer a lightweight solution to reproducible environment management for Bioinformatics research. After we adopted this approach for our environment and experiment management we noticed that:

1. Researchers spend their time designing and running experiments, rather than focusing on devops and
2. Shipping OS, dependencies, and data as a single unit drastically improves reproducibility.