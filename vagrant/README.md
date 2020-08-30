**Note:** Vagrant file configured to use `$HOME/workspace` as a `$GOPATH`. Replace `workspace` with your value in Vagrantfile.


To develop or test **Bravetools** with Vagrant:

1. Start Vagrant VM and ssh into running machine

```
$ vagrant up
$ vagrant ssh
```

2. Uninstall current version of LXD/LXC (Bravetools uses LXD from Canonical Snapstore. It will be installed during `brave init`)

```
$ sudo apt remove lxd
$ sudo apt autoremove
```

3. Build **Bravetools** for Ubuntu

```
$ cd $HOME/workspace/src/github.com/bravetools/bravetools
$ make ubuntu
```

4. Install **Bravetools** 

```
$ brave init
```