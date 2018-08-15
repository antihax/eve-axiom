package attributes

import "time"

// DamageProfile is the four types of damage
type DamageProfile struct {
	EM        float64 `json:",omitempty"`
	Thermal   float64 `json:",omitempty"`
	Kinetic   float64 `json:",omitempty"`
	Explosive float64 `json:",omitempty"`
	Avg       float64 `json:",omitempty"`
	Min       float64 `json:",omitempty"`
	Max       float64 `json:",omitempty"`
}

type Surface struct {
	Resonance DamageProfile `json:",omitempty"`
	Hp        float64       `json:",omitempty"`
}

type MWDAttributes struct {
	Capacitor Capacitor `json:",omitempty"`

	MaxVelocity     float64 `json:",omitempty"`
	SignatureRadius float64 `json:",omitempty"`
}

type Module struct {
	// Basic Attributes
	Duration, Tracking, Discharge, Optimal, Falloff, Chance float64 `json:",omitempty"`
	DamageMultiplier, AlphaDamage, DamagePerSecond          float64 `json:",omitempty"`

	// Missile Attributes
	FlightTime, MaxVelocity float64 `json:",omitempty"`

	// Module location
	Location int32

	// Damage Data
	Damage DamageProfile `json:",omitempty"`

	TypeID   int32 `json:",omitempty"`
	ChargeID int32 `json:",omitempty"`

	RemoteStructureRepairAmount float64 `json:",omitempty"`
	RemoteArmorRepairAmount     float64 `json:",omitempty"`
	RemoteShieldTransferAmount  float64 `json:",omitempty"`
	RemoteEnergyTransferAmount  float64 `json:",omitempty"`
	NeutralizerAmount           float64 `json:",omitempty"`
	NosferatuAmount             float64 `json:",omitempty"`

	ArmorRepair     float64 `json:",omitempty"`
	ShieldRepair    float64 `json:",omitempty"`
	StructureRepair float64 `json:",omitempty"`
}

type Drone struct {
	// Basic Attributes
	Duration, Tracking, Discharge, Optimal, Falloff, Chance float64 `json:",omitempty"`
	DamageMultiplier, AlphaDamage, DamagePerSecond          float64 `json:",omitempty"`

	DroneBandwith float64 `json:",omitempty"`
	// Damage Data
	Damage DamageProfile `json:",omitempty"`

	TypeID   int32 `json:",omitempty"`
	Quantity int32 `json:",omitempty"`
}

type Capacitor struct {
	Stable   bool
	Fraction float64       `json:",omitempty"`
	Duration time.Duration `json:",omitempty"`
	Capacity float64       `json:",omitempty"`
}

// Attributes for all the fit attributes
type Attributes struct {
	ShipID     int32         `json:",omitempty"`
	WithMWD    MWDAttributes `json:",omitempty"`
	WithoutMWD MWDAttributes `json:",omitempty"`

	Structure Surface `json:",omitempty"`
	Armor     Surface `json:",omitempty"`
	Shield    Surface `json:",omitempty"`

	MinEHP int64 `json:",omitempty"`
	MaxEHP int64 `json:",omitempty"`
	AvgEHP int64 `json:",omitempty"`

	MinRPS float64 `json:",omitempty"`
	MaxRPS float64 `json:",omitempty"`
	AvgRPS float64 `json:",omitempty"`

	Agility        float64 `json:",omitempty"`
	ScanResolution float64 `json:",omitempty"`

	WarpSpeed         float64 `json:",omitempty"`
	DroneBandwith     float64 `json:",omitempty"`
	MaxTargetingRange float64 `json:",omitempty"`

	GravStrenth   float64 `json:",omitempty"`
	LadarStrength float64 `json:",omitempty"`
	MagStrength   float64 `json:",omitempty"`
	RadarStrength float64 `json:",omitempty"`

	ShieldRechargeRate float64 `json:",omitempty"`

	CPURemaining float64 `json:",omitempty"`
	PGRemaining  float64 `json:",omitempty"`

	CPUTotal float64 `json:",omitempty"`
	PGTotal  float64 `json:",omitempty"`

	Modules   map[uint8]Module `json:",omitempty"`
	Drones    []Drone          `json:",omitempty"`
	MaxDrones float64          `json:",omitempty"`

	// Totals
	DroneDPS, DroneAlpha, ModuleDPS, ModuleAlpha, TotalDPS, TotalAlpha    float64 `json:",omitempty"`
	RemoteArmorRepairPerSecond, RemoteShieldTransferPerSecond             float64 `json:",omitempty"`
	RemoteStructureRepairPerSecond, RemoteEnergyTransferPerSecond         float64 `json:",omitempty"`
	ArmorRepairPerSecond, ShieldRepairPerSecond, StructureRepairPerSecond float64 `json:",omitempty"`
	EnergyNeutralizerPerSecond                                            float64 `json:",omitempty"`
}
