---
layout: default
title: brave info
parent: CLI
nav_order: 1
---

# brave info

Display workspace information

```
brave info
```

## Description

Display workspace information. Disk and memory statistics reflects system utilisation by deployed Units. For example

```bash
brave info

NAME               	STATE  	IPV4         	DISK               	MEMORY         	CPU 
brave-ThinkPad-P43s	Running	192.168.1.112	269.22MB of 96.74GB	8.1GB of 40.7GB	8  	
```

In the above output, Units have reserved 269MB of disk space and 8.1GB of RAM.

## Options

```
  -h, --help    help for info
      --short   Returns host IP address
```

## See Also

* [brave](brave.md)	 - A complete System Container management platform

