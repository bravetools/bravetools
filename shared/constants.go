package shared

const BraveHome = "/.bravetools"
const BraveCertStore = BraveHome + "/certs"
const BraveServerCertStore = BraveHome + "/servercerts"

// PlatformConfig ..
const PlatformConfig = BraveHome + "/config.yml"

// ImageStore ..
const ImageStore = BraveHome + "/images/"

// Bravetools local remote name
const BravetoolsRemote = "local"

// BraveRemoteStore is path to remotes dir
const BraveRemoteStore = BraveHome + "/remotes"

// BraveClientKey ..
const BraveClientKey = BraveCertStore + "/client.key"

// BraveClientCert ..
const BraveClientCert = BraveCertStore + "/client.crt"

// SnapLXC lxc command path in Snap
const SnapLXC = "/snap/bin/lxc"

// BraveDB path to Bravetools database
const BraveDB = BraveHome + "/bravetools.db"

// DefaultUnitCpuLimit - used if not specified
const DefaultUnitCpuLimit = "2"

// DefaultUnitRamLimit - used if not specified
const DefaultUnitRamLimit = "2GB"
