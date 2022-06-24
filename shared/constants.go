package shared

// BRAVEFILE variable - minimal Bravefile template
const BRAVEFILE = `
base:
  image: <name>
  location: public

service:
  name: <service>
  version: 1.0
`

const BraveHome = "/.bravetools"
// PlatformConfig ..
const PlatformConfig = "/.bravetools/config.yml"

// ImageStore ..
const ImageStore = "/.bravetools/images/"

// BraveClientKey ..
const BraveClientKey = "/.bravetools/certs/client.key"

// BraveClientCert ..
const BraveClientCert = "/.bravetools/certs/client.crt"

// SnapLXC lxc command path in Snap
const SnapLXC = "/snap/bin/lxc"

// BraveDB path to Bravetools database
const BraveDB = "/.bravetools/bravetools.db"
