---
layout: default
title: brave publish
parent: CLI
nav_order: 1
---

# brave publish

Publish deployed Unit as image

```
brave publish NAME
```

## Description

Bravetools creates stateful units that can be easily migrated between hosts without risk of data loss. This is accomplished through the `brave publish` command, which creates a compressed tar ball image of a live unit inside the current working directory. The image can later be imported using [`brave import`](../brave_import).

## Options

```
  -h, --help   help for publish
```

## See Also

* [brave](brave.md)	 - A complete System Container management platform

