# Use `macpine` to spin up bravetools testing environment on Mac

```bash
alpine launch --image alpine_3.16.0_lxd --name bravetools --mount $PWD
alpine exec bravetools "ash /root/mnt/macpine/bootstrap.sh"
```

## Unit Testing

Run Unit tests directly on the VM

```bash
alpine ssh bravetools
cd mnt
go test -v ./...
```