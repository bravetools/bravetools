package shared

// REINIT ..
const REINIT = `
Bravetools home directory, settings, and images will be deleted.
You will need to run "brave init" to initialize a new Bravetools exvironment.
Do you want to continue?(Yes/no): 
`

// REMOVELIN ..
const REMOVELIN = `
Use LXC CLI to check if there is existing Bravetools storage and profile.
					
1. lxc profile delete brave
2. lxc storage delete brave_[timestamp]
3. lxc network delete bravebr0
`

// REMOVEMP ..
const REMOVEMP = `
Bravetools already initiated.

To delete Bravetools:

1. Delete directory .bravetools in home
2. Run: 
- "multipass delete brave"
- "multipass purge"

`
