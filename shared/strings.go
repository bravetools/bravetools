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
  - source: .
    target: /root/

service:
  name: example-container
  ports:
    - 8888:8888
  resources:
    ram: 2GB
    cpu: 2
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
1. lxc profile delete brave
2. lxc storage delete brave-[timestamp]
3. lxc network delete {$USER}br0
`

// REMOVEMP ..
const REMOVEMP = `
Bravetools already initialised.

To delete Bravetools:

1. rm -r $HOME/.bravetools
2. multipass delete brave
3. multipass purge"
`
