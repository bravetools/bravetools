**Note:** Vagrant file configured to use `$HOME/workspace` as a `$GOPATH`. Replace `workspace` with your value in Vagrantfile.


To develop or test **Bravetools** with Vagrant:

1. Start Vagrant VM and ssh into running machine

```
$ vagrant up
$ vagrant ssh
```

2. Build **Bravetools** for Ubuntu

```
$ sudo apt remove lxd
$ sudo apt autoremove
$ sudo snap install lxd

$ cd $HOME/workspace/src/github.com/bravetools/bravetools
$ make ubuntu
```

4. Install **Bravetools** 

```
$ cd $HOME/workspace/src/github.com/bravetools/bravetools/vagrant
$ brave init --config config.yml
```