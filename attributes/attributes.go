package attributes

// Attributes for all the fit attributes
type Attributes struct {
	Ship    map[string]float64           `json:"ship,omitempty"`
	Modules map[uint8]map[string]float64 `json:"modules,omitempty"`
	Drones  []map[string]float64         `json:"drones,omitempty"`
	Types   map[int32]string             `json:"types,omitempty"`
	TypeID  int32                        `json:"typeID,omitempty"`
}
