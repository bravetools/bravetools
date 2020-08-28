---
layout: default
title: brave import
parent: CLI
nav_order: 1
---

# brave import

Import a tarball into local Bravetools image repository

```
brave import NAME
```

## Description

Brave images are simply compressed file system tar balls. `brave import` makes an arbitrary image on your file system available to Bravetools by copying it to `$HOME/.bravetools` directory and generating a unique hash.

## Options

```
  -h, --help   help for import
```

## See Also

* [brave](brave.md)	 - A complete System Container management platform

