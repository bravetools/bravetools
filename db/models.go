package db

// BraveUnit type to store unit data in DB
type BraveUnit struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	UID  string `json:"uid"`
	Date string `json:"date"`
	Data []byte `json:"unitData"`
}

// UnitData Brave unit metadata
type UnitData struct {
	IP    string `json:"ip"`
	Image string `json:"image"`
	CPU   int    `json:"cpu"`
	RAM   string `json:"ram"`
}

// Unit Brave unit object
type Unit struct {
	ID   int64
	Name string
	UID  string
	Date string
	Data UnitData
}
