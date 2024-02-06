package shared

// Bravefile template
const BravefileTemplate = `image: example-image/v1.0

base:
  image: alpine/3.16

packages:
  manager: apk
  system:
    - curl

run: 
  - command: echo
    args:
      - hello world

copy:
  - source: ./Bravefile
    target: /root/

service:
  name: example-container
  ports:
    - 8888:8888
  resources:
    ram: 2GB
    cpu: 2
    disk: 10GB
`

// REINIT ..
const REINIT = `
Bravetools home directory, settings, and images will be deleted.
You will need to run "brave init" to initialize a new Bravetools environment.
Do you want to continue?(Yes/no): 
`

// REMOVELIN ..
const REMOVELIN = `
Looks like you're initialising an already configured Bravetools environment.

You can manually cleanup your Bravetools LXD configuration using LXD CLI:				
1. lxc profile delete bravetools-$USER
2. lxc storage delete bravetools-$USER
3. lxc network delete bravetoolsbr0
`

// REMOVEMP ..
const REMOVEMP = `
Bravetools already initialised.

To delete Bravetools:

1. rm -r $HOME/.bravetools
2. multipass delete bravetools
3. multipass purge"
`
