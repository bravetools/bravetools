---
layout: default
title: Brave Compose
parent: Docs
nav_order: 4
description: "Bravetools builds images automatically by reading instructions from a Bravefile."
---

# Brave Compose

While a `Bravefile` defines configuration options for building and deploying a single service, often a complex application is backed by multiple services working together in concert. For example, you may have an API that contacts multiple backend services such as a database or authentication server.

`brave compose` can help manage the building and deploying of many services, allowing you to treat a complex system made of many components as a single entity, defined in a single `brave-compose.yaml` file. Thanks to the close integration with `Bravefile`, it's easy to take existing standalone services you've defined and combine them into a single system. 

## Compose command

When you run `brave compose` with no arguments, bravetools will use the `brave-compose.yaml` file in the current directory - pass a directory to the command to use the compose file in that directory.

```bash
brave compose path/to/dir
``` 

The directory containing the compose file will become the root directory for the ensuing build/deploy. This means that you can (and should) use relative paths in the compose file to make the project more portable.


## Compose file

The `brave-compose.yaml` file defines a set of services to build/deploy. A basic compose file consists of a map of service names with deploy configurations - the name of the service in the composefile will be the name of the deployed unit, while deploy config can come from a `Bravefile` or can be defined in the compose file.

For example, deploying the service below will result in a unit named "example-service" being deployed. The deploy configuration will be loaded from the provided bravefile - it's also possible to define the deployment configuration inline (see below).

```yaml
services:
  example-service:
    bravefile: ./path/to/bravefile
```

## Service configuration

### Bravefile
If a path to a `Bravefile` is provided for a service, the deployment configuration will be loaded from the file. This is useful if your standalone service's default configurations are applicable to the system you are composing, since you can reuse the settings with minimal fuss.

### Build

If a `Bravefile` path is supplied, it is also possible to use `brave compose` to build the service by setting the "build" field to "true". The image will be built according to the build instructions in the `Bravefile` before deployment if it does not already exist. If the image already exists, that image will be reused and the build will be skipped.

```yaml
services:
  example-service:
    bravefile: ./path/to/bravefile
    build: true
```

The directory containing the `Bravefile` will become the context from which the build/deploy of a service is executed by default. This means that resources referenced by relative paths in the `Bravefile` work seamlessly with compose.

If for some reason a different build/deploy context is required, you may specify an alternate path in the "context" field of the service.

```yaml
services:
  example-service:
    bravefile: ./path/to/bravefile
    build: true
    context: ./path/to/context/dir
```


### Inline configuration
It is possible to deploy a service without a `Bravefile` by specifying the deploy configuration for the service within the compose file. Any field from the "service" section of the Bravefile can be used to configure a service in the compose file.

```yaml
services:
  example-service:
    image: example-image-name
    version: 1.0
    ip: 10.0.0.20
    docker: yes
    ports:
      - 5000:5000
    resources:
      ram: 500MB
      cpu: 1
      gpu: yes
    postdeploy:
      run:
        - command: echo
          args:
            - "hello world"
      copy:
        - source: /example/host/file
          target: /example/unit/path
```

As you can see, it can get quite verbose compared with the version that loaded the `Bravefile`. However, it may be beneficial to have all the deployment configuration in one place.

### Bravefile defaults, selective overwriting

But what if there are just a few problematic settings in the `Bravefile` that don't work for the system you're setting up with `compose`? Instead of copying the "service" section of the Bravefile into the compose file and editing it, you can load the default config from the `Bravefile` and overwrite what you need in the compose file.

For example, if the base `Bravefile` of "example-service" defines an IP of 10.0.0.10 but that IP clashes with another service we are composing, you can overwrite the IP address in the compose file with 10.0.0.20. All other settings will be loaded from the `Bravefile` as normal.

```yaml
services:
  example-service:
    bravefile: ./path/to/bravefile
    ip: 10.0.0.20
```

### Dependencies between services

Services will often depend on each other to properly function. For example, a server may require a database to be reachable to work. In cases like these we can define dependencies between services in the "depends_on" field of the compose file. Bravetools will build/deploy the services so that each service is deployed after the services on which it depends.

In the following example, the api server depends on both auth and log - therefore it will be deployed last. auth also depends on log, so log must be deployed first and auth second.

```yaml
# Expected order: log -> auth -> api
services:
  api:
    bravefile: ./api/Bravefile
    depends_on:
      - auth
      - log
  auth:
    bravefile: ./auth/Bravefile
    depends_on:
      - log
  log:
    bravefile: ./log/Bravefile
```

### Reusing base images

Often, images will have some overlap in their environments, sharing the same base distribution and the majority of installed packages. You can think of it as a superclass and subclasses, with specialized subclass services inheriting from the same base superclass. This scenario is perfect for incremental builds, where certain images are created and then reused and specialized by other services.

The "base" field, used in tandem with the "depends_on" field, is a way to express this scenario declaratively in the compose file. Services marked as "base" are built only to be used by their dependents - they are not deployed themselves. The dependent services can then use this image as a base during build as a starting point.

"base" images are by default transient - they exist only during the build to facilitate the building of other services. Therefore by default they are deleted after the end of the completion of the compose operation to avoid using disk space. In the event that you want to keep a "base" image around afterwards, you can set the "build" flag for the service to "true".

In the example below, the "base" service is built first. Every other service imports it locally as their "base" image from within their Bravefile as part of the build. Finally, after the other services are deployed, the "base" image is deleted (to avoid this, uncomment the "build: true" line)

```yaml
# Build order: base -> log -> auth -> api
services:
  base:
    bravefile: ./base/Bravefile
    #build: true
    base: true
  api:
    bravefile: ./api/Bravefile
    build: true
    depends_on:
      - base 
      - auth
      - log
  auth:
    bravefile: ./auth/Bravefile
    build: true
    depends_on:
      - base
      - log
  log:
    bravefile: ./log/Bravefile
    build: true
    depends_on:
      - base
```
