package shared

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
