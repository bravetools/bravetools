package shared

// BraveUnit ..
type BraveUnit struct {
	Name    string
	Status  string
	Address string
	Disk    []DiskDevice
	Proxy   []ProxyDevice
	NIC     NicDevice
}

// DiskDevice ..
type DiskDevice struct {
	Name   string
	Path   string
	Source string
}

// ProxyDevice ..
type ProxyDevice struct {
	Name      string
	ConnectIP string
	ListenIP  string
}

// NicDevice ..
type NicDevice struct {
	Name    string
	Type    string
	NicType string
	Parent  string
	IP      string
}

// BraveProfile ..
type BraveProfile struct {
	Name       string
	Storage    string
	Bridge     string
	LxdVersion string
}
