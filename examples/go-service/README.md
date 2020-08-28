## How ro run a go-service example

1. Install Bravetools and verify it is up and running:

```bash
brave info
```

2. Build base unit based on Ubuntu Buonic:

```bash
$ brave base ubuntu/bionic
```

3. Build go-service image. Run in `go-service` project root:

```bash
$ brave build
```

This command creates a **go-service-1.0** image. Verify that image stored in the local image store:

```bash
$ brave images
```

4. Deploy a unit:

``` bash
5. brave deploy
```

Verify that unit `go-service` is up and running

```bash
$ brave units
```

Unit is a web service running on port 3000. Type `http://[HOST_IP]:3000/bravetools` in web browser. This should display a web page with a siple content.

URL http://[HOST_IP]:3000/bravetools shows **Hi there, I love bravetools!**

7. Deleting unit:

Run in terminal:

```bash
$ brave remove go-service
```
