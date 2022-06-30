# brave compose example

Brave compose is a bravetools command that manages the build and deployment of systems consisting of multiple units. By integrating with Bravefiles complex deployment scenarios can be expressed in just a few lines.

In this example we will build and deploy a small system of three python web services that interact with each other.


## Compose file

The minimal compose file that we will use in this example is at "brave-compose.yml" and is shown below:

```yaml
services:
  api:
    bravefile: ./api/Bravefile
    build: true
    depends_on:
      - auth
      - log
  auth:
    bravefile: ./auth/Bravefile
    build: true
  log:
    bravefile: ./log/Bravefile
    build: true
```

The compose file defines a set of services by name, with paths to the respective Bravefiles. There's also an optional build flag, and a field "depends_on" that determines the order the services will be spun up - in this case, because the `api` depends on `auth` and `log` it will be deployed last.

This file is very short - this is possible because the instructions for building the image and the defaults for deploying the image are drawn from the Bravefiles for each service, keeping this higher level view clear of the small details.

All deployment variables from the "Service" field of the Bravefile can be specified per service in the compose file - these will take precedence over the values in the Bravefile. For example, if we wanted to change the CPU and RAM allocation of the `log` service we can just specify the new values here.

```yaml
  log:
    bravefile: ./log/Bravefile
    build: true
    resources:
        cpu: 2
        ram: 1GB
```

## Build and deploy

To build and deploy this system, run the command `brave compose` in the root directory of this example (python-multi-service).

After this command completes, the whole system should be built and deployed.
You can verify that the units are running by running `brave units`.

The api server should be reachable at reachable on port 5000 of the bravetools host. To verify your host IP run `brave info` and check the IPV4 field.

When visiting `http://HOST_IP:5000` you should see a message confirming that a test user has been authenticated and that the request has been logged at a certain date time. These messages come from the `auth` and `log` services. To prove this, we can shutdown the log service using `brave stop log`. Now an error should appear as the `api` service can no longer contact the `log` service.

## Conclusion

This example has shown how easy it is to build and deploy multi-unit systems using `brave compose` and a simple "brave-compose.yml" file.

To clean up the units and images after you are done use the `brave remove` command.
