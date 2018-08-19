package attributes

// Attributes for all the fit attributes
type Attributes struct {
	Ship    map[string]float64           `json:",omitempty"`
	Modules map[uint8]map[string]float64 `json:",omitempty"`
	Drones  []map[string]float64         `json:",omitempty"`
	TypeID  int32
}
