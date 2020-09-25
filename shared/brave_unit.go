package shared

// BraveUnit ..
type BraveUnit struct {
	Name    string
	Status  string
	Address string
	Devices []Device
}

// Device ..
type Device struct {
	Name string
	Info string
}
