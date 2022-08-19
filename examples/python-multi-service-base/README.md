# brave compose base example

The example builds on the [python-multi-service example](../python-multi-service/) to demonstrate the use of the "base" field to reuse the same base image when building multiple services. It's recommended to start with that example before reading this one.

In this scenario we will build and deploy a small system of three python web services that interact with each other. These three python web services mostly share the same environment - all are based on alpine, and all require Python and the same python packages.

To accelerate the build process and reduce resource utilization, we will first build a base image with the shared environment of the services. Each service will then import this image as a starting point for their build.

For this simple example these changes sped up the build on my machine by about 1 minute, reducing the time taken to run the `compose` command from 3.5 to 2.5 minutes. More complex scenarios with more installed packages or larger images will lead to larger time savings.


## Compose file

The compose file that we will use in this example is at "brave-compose.yaml" and is shown below:

```yaml
services:
  python-base:
    bravefile: ./base/Bravefile
    base: true
    #build: true
  api:
    bravefile: ./api/Bravefile
    build: true
    depends_on:
      - python-base 
      - auth
      - log
  auth:
    bravefile: ./auth/Bravefile
    build: true
    depends_on:
      - python-base
      - log
  log:
    bravefile: ./log/Bravefile
    build: true
    depends_on:
      - python-base
```

Focusing on just the build phase, the "python-base" service has its **base** field set to *true*, marking it as an image which other images will base themselves on. All other services include the "python-base" service in their **depends_on** fields and have **build** set to *true*. This ensures that the image will be available when building each service.

As a **base** image, the "python-base" image will not be deployed - it will just be used in the build phase of each dependent service. In addition, by default the "python-base" image is transient, and will automatically be cleaned up after a build. To disable the automatic cleanup, uncomment the "build: true" line.

The build scenario described above can be viewed as an inheritance hierarchy with several subclasses inheriting from a single abstract superclass. The image below visualizes this relationship.  

![build-dependecy-hierarchy](build-deps.svg)

Note that the Bravefiles need to be set up to reference the base image image. A snippet of the top of the `api` Bravefile shows this:

```yaml
# ./api/Bravefile
base:
  image: python-base-1.0
  location: local
```


## Build and deploy

To build and deploy this system, run the command `brave compose` in the root directory of this example (python-multi-service-base). The "python-base" image should be built first and then imported by each other service during their build phases.

After the command completes, the whole system should be built and deployed.
You can verify that the units are running with the command `brave units`. All services except for "python-base" should be running.

When checking the stored images with `brave images` we should see images corresponding to the `api`, `auth`, and `log` services, but none for `python-base`. The base image was just used to accelerate the build process by removing the need to re-download the same image and packages repeatedly for each service. 

To test that everything is working, try accessing the api. The api server should be reachable at reachable on port 5000 of the bravetools host. To verify your host IP run `brave info` and check the IPV4 field. When visiting `http://HOST_IP:5000` you should see a message confirming that a test user has been authenticated and that the request has been logged at a certain date time.


## Conclusion

This example has shown how to reuse build stages in multiple services using brave compose. This can improve build times and reduce resource utilization by avoiding repeated work and also can help logically break up build stages. 

Using this technique only required a small adjustment to the "brave-compose.yaml" file to express the dependencies and identify the base service. After that, it was simply a matter of running the `brave compose` command to build and deploy the system.

To clean up the units and images after you are done use the `brave remove` command.
