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
