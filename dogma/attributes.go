package dogma

/*
 #include <dogma.h>
 #include <dogma-extra.h>

 // Dealing with unions
 double get_cap_duration(dogma_simple_capacitor_t *c) {
   if (c->stable) {
     return c->stable_fraction;
   }
   return c->depletion_time;
 }

 dogma_location_t module_location(dogma_key_t slot) {
   return (dogma_location_t){ .type = DOGMA_LOC_Module, .module_index = slot };
 }

 dogma_location_t drone_location(dogma_typeid_t typeid) {
   return (dogma_location_t){ .type = DOGMA_LOC_Drone, .drone_typeid = typeid };
 }

*/
import "C"
import (
	"errors"
	"time"
)

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

	Hull   Surface `json:",omitempty"`
	Armor  Surface `json:",omitempty"`
	Shield Surface `json:",omitempty"`

	MinEHP int64 `json:",omitempty"`
	MaxEHP int64 `json:",omitempty"`
	AvgEHP int64 `json:",omitempty"`

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

	Modules   map[uint8]Module `json:",omitempty"`
	Drones    []Drone          `json:",omitempty"`
	MaxDrones float64          `json:",omitempty"`
}

// GetAttributes gets all the fit attributes
func (c *Context) GetAttributes() (*Attributes, error) {
	var (
		att Attributes
		err error
	)
	att.Modules = make(map[uint8]Module)

	att.ShipID = int32(c.shipID)

	if att.PGRemaining, err = c.PowerLeft(); err != nil {
		return nil, err
	}

	if att.CPURemaining, err = c.CPULeft(); err != nil {
		return nil, err
	}

	if err := c.fillTankAttributes(&att); err != nil {
		return nil, err
	}

	if err := c.fillShipAttributes(&att); err != nil {
		return nil, err
	}

	if err := c.fillMWDAffectedAttributes(&att.WithMWD); err != nil {
		return nil, err
	}

	if err := c.fillModuleAttributes(&att); err != nil {
		return nil, err
	}

	if err := c.fillDroneAttributes(&att); err != nil {
		return nil, err
	}

	c.DeactivateMWD()
	if err := c.fillMWDAffectedAttributes(&att.WithoutMWD); err != nil {
		return nil, err
	}

	return &att, nil
}

func (c *Context) fillCapacitorAttributes(att *MWDAttributes) error {
	var (
		cap  *C.dogma_simple_capacitor_t
		size C.size_t
	)
	// Get the capacitor information
	if r := C.dogma_get_capacitor_all(c.ctx, true, &cap, &size); r != 0 {
		return errors.New("capacitor failure")
	}

	// Should only have one capacitor on the ship
	if size != 1 {
		return errors.New("wrong number of capacitors")
	}

	// Get the capacitor attributes
	att.Capacitor.Capacity = float64(cap.capacity)
	att.Capacitor.Stable = bool(cap.stable)
	if cap.stable {
		att.Capacitor.Fraction = float64(C.get_cap_duration(cap))
	} else {
		att.Capacitor.Duration = time.Duration(C.get_cap_duration(cap) * 1000000)
	}
	// Release memory used for the list
	C.dogma_free_capacitor_list(cap)
	return nil
}

func (c *Context) fillDroneAttributes(att *Attributes) error {
	for _, drone := range c.drones {
		typeID := C.dogma_typeid_t(drone.typeID)
		i := uint(0)
		var effect C.dogma_effectid_t

		for {
			if r := C.dogma_get_nth_type_effect_with_attributes(typeID, C.uint(i), &effect); r != 0 {
				break
			}
			i++

			var duration, tracking, discharge, optimal, falloff, chance C.double
			C.dogma_get_location_effect_attributes(c.ctx,
				C.drone_location(typeID), effect,
				&duration, &tracking, &discharge, &optimal, &falloff, &chance,
			)
			m := Drone{
				Duration:  float64(duration),
				Tracking:  float64(tracking),
				Discharge: float64(discharge),
				Optimal:   float64(optimal),
				Falloff:   float64(falloff),
				Chance:    float64(chance),
				TypeID:    int32(drone.typeID),
				Quantity:  int32(drone.quantity),
			}
			var err error
			if m.DroneBandwith, err = c.GetDroneAttribute(1272, typeID); err != nil {
				return err
			}
			if m.Damage.Explosive, err = c.GetDroneAttribute(116, typeID); err != nil {
				return err
			}
			if m.Damage.EM, err = c.GetDroneAttribute(114, typeID); err != nil {
				return err
			}
			if m.Damage.Thermal, err = c.GetDroneAttribute(118, typeID); err != nil {
				return err
			}
			if m.Damage.Kinetic, err = c.GetDroneAttribute(117, typeID); err != nil {
				return err
			}

			if m.DamageMultiplier, err = c.GetDroneAttribute(64, typeID); err != nil {
				return err
			}
			if m.DamageMultiplier == 0 {
				m.DamageMultiplier = 1
			}
			att.Drones = append(att.Drones, m)
		}
	}
	return nil
}

func (c *Context) fillModuleAttributes(att *Attributes) error {
	for _, mod := range c.mods {
		typeID := C.dogma_typeid_t(mod.typeID)
		i := uint(0)
		var effect C.dogma_effectid_t

		for {
			if r := C.dogma_get_nth_type_effect_with_attributes(typeID, C.uint(i), &effect); r != 0 {
				break
			}
			i++
			var hasit C.bool
			if C.dogma_type_has_effect(typeID, StateActive, effect, &hasit); !hasit {
				continue
			}

			var duration, tracking, discharge, optimal, falloff, chance C.double
			C.dogma_get_location_effect_attributes(c.ctx,
				C.module_location(mod.idx), effect,
				&duration, &tracking, &discharge, &optimal, &falloff, &chance,
			)

			m := Module{
				Duration:  float64(duration),
				Tracking:  float64(tracking),
				Discharge: float64(discharge),
				Optimal:   float64(optimal),
				Falloff:   float64(falloff),
				Chance:    float64(chance),
				TypeID:    int32(mod.typeID),
				ChargeID:  int32(mod.chargeID),
			}
			var err error
			if duration > 1e-10 && mod.chargeID > 0 {
				// Missile Specific
				if effect == EffectUseMissiles {
					if m.FlightTime, err = c.GetChargeAttribute(281, mod.idx); err != nil {
						return err
					}
					if m.MaxVelocity, err = c.GetChargeAttribute(37, mod.idx); err != nil {
						return err
					}
					if m.DamageMultiplier, err = c.GetModuleAttribute(212, mod.idx); err != nil {
						return err
					}
				}

				// Bomb Specific
				if effect == EffectEMPWave {
					m.DamageMultiplier = 1
				}
				if m.DamageMultiplier == 0 {
					if m.DamageMultiplier, err = c.GetModuleAttribute(64, mod.idx); err != nil {
						return err
					}
				}

				if m.Damage.Explosive, err = c.GetChargeAttribute(116, mod.idx); err != nil {
					return err
				}
				if m.Damage.EM, err = c.GetChargeAttribute(114, mod.idx); err != nil {
					return err
				}
				if m.Damage.Thermal, err = c.GetChargeAttribute(118, mod.idx); err != nil {
					return err
				}
				if m.Damage.Kinetic, err = c.GetChargeAttribute(117, mod.idx); err != nil {
					return err
				}

				m.AlphaDamage = m.DamageMultiplier * (m.Damage.Kinetic + m.Damage.EM + m.Damage.Explosive + m.Damage.Thermal)
				m.DamagePerSecond = m.AlphaDamage / (m.Duration / 1000)
			}

			switch effect {
			case EffectRemoteHullRepairFalloff:
				if m.RemoteStructureRepairAmount, err = c.GetModuleAttribute(83, mod.idx); err != nil {
					return err
				}
			case EffectRemoteArmorRepairFalloff:
				if m.RemoteArmorRepairAmount, err = c.GetModuleAttribute(84, mod.idx); err != nil {
					return err
				}
			case EffectAncillaryRemoteArmorRepairer:
				if m.RemoteArmorRepairAmount, err = c.GetModuleAttribute(84, mod.idx); err != nil {
					return err
				}
			case EffectRemoteShieldTransferFalloff:
				if m.RemoteShieldTransferAmount, err = c.GetModuleAttribute(68, mod.idx); err != nil {
					return err
				}
			case EffectAncillaryRemoteShieldBooster:
				if m.RemoteShieldTransferAmount, err = c.GetModuleAttribute(68, mod.idx); err != nil {
					return err
				}
			case EffectRemoteEnergyTransfer:
				if m.RemoteEnergyTransferAmount, err = c.GetModuleAttribute(90, mod.idx); err != nil {
					return err
				}
			case EffectEnergyNeutralizerFalloff:
				if m.NeutralizerAmount, err = c.GetModuleAttribute(97, mod.idx); err != nil {
					return err
				}
			case EffectEnergyNosferatuFalloff:
				if m.NosferatuAmount, err = c.GetModuleAttribute(90, mod.idx); err != nil {
					return err
				}

			case EffectArmorRepair:
				if m.ArmorRepair, err = c.GetModuleAttribute(84, mod.idx); err != nil {
					return err
				}
			case EffectFueledArmorRepair:
				if m.ArmorRepair, err = c.GetModuleAttribute(84, mod.idx); err != nil {
					return err
				}
			case EffectShieldBoosting:
				if m.ShieldRepair, err = c.GetModuleAttribute(68, mod.idx); err != nil {
					return err
				}
			case EffectFueledShieldBoosting:
				if m.ShieldRepair, err = c.GetModuleAttribute(68, mod.idx); err != nil {
					return err
				}

			case EffectStructureRepair:
				if m.StructureRepair, err = c.GetModuleAttribute(83, mod.idx); err != nil {
					return err
				}
			}

			att.Modules[uint8(mod.idx)] = m
		}
	}

	return nil
}

func (c *Context) fillShipAttributes(att *Attributes) error {
	var v C.double

	attributes := map[string]C.ushort{
		"warpSpeed":      600,
		"maxDrones":      283,
		"droneBandwidth": 1271,
		"maxTargetRange": 76,

		"agility": 70,

		"scanGravimetricStrength":   211,
		"scanLadarStrength":         209,
		"scanMagnetometricStrength": 210,
		"scanRadarStrength":         208,

		"scanResolution":     564,
		"shieldRechargeRate": 479,
	}

	for k, id := range attributes {
		if r := C.dogma_get_ship_attribute(c.ctx, id, &v); r != 0 {
			return errors.New("Could not get attributes")
		}
		switch k {
		case "warpSpeed":
			att.WarpSpeed = float64(v)

		case "maxDrones":
			att.MaxDrones = float64(v)
		case "droneBandwidth":
			att.DroneBandwith = float64(v)
		case "maxTargetRange":
			att.MaxTargetingRange = float64(v)

		case "agility":
			att.Agility = float64(v)

		case "scanGravimetricStrength":
			att.GravStrenth = float64(v)
		case "scanLadarStrength":
			att.LadarStrength = float64(v)
		case "scanMagnetometricStrength":
			att.MagStrength = float64(v)
		case "scanRadarStrength":
			att.RadarStrength = float64(v)

		case "scanResolution":
			att.ScanResolution = float64(v)
		case "shieldRechargeRate":
			att.ShieldRechargeRate = float64(v)

		}
	}

	return nil
}

func (c *Context) fillTankAttributes(att *Attributes) error {
	var v C.double
	resonances := map[string]C.ushort{
		"ArmorEm":        267,
		"ArmorExplosive": 268,
		"ArmorKinetic":   269,
		"ArmorThermal":   270,

		"ShieldEm":        271,
		"ShieldExplosive": 272,
		"ShieldKinetic":   273,
		"ShieldThermal":   274,

		"HullEm":        113,
		"HullExplosive": 111,
		"HullKinetic":   109,
		"HullThermal":   110,

		"ShieldCapacity": 263,
		"ArmorHp":        265,
		"Hp":             9,
	}
	for k, id := range resonances {
		if r := C.dogma_get_ship_attribute(c.ctx, id, &v); r != 0 {
			return errors.New("Could not get attributes")
		}
		switch k {
		case "ArmorEm":
			att.Armor.Resonance.EM = float64(v)
		case "ArmorExplosive":
			att.Armor.Resonance.Explosive = float64(v)
		case "ArmorKinetic":
			att.Armor.Resonance.Kinetic = float64(v)
		case "ArmorThermal":
			att.Armor.Resonance.Thermal = float64(v)

		case "ShieldEm":
			att.Shield.Resonance.EM = float64(v)
		case "ShieldExplosive":
			att.Shield.Resonance.Explosive = float64(v)
		case "ShieldKinetic":
			att.Shield.Resonance.Kinetic = float64(v)
		case "ShieldThermal":
			att.Shield.Resonance.Thermal = float64(v)

		case "HullEm":
			att.Hull.Resonance.EM = float64(v)
		case "HullExplosive":
			att.Hull.Resonance.Explosive = float64(v)
		case "HullKinetic":
			att.Hull.Resonance.Kinetic = float64(v)
		case "HullThermal":
			att.Hull.Resonance.Thermal = float64(v)

		case "ShieldCapacity":
			att.Shield.Hp = float64(v)
		case "ArmorHp":
			att.Armor.Hp = float64(v)
		case "Hp":
			att.Hull.Hp = float64(v)

		}
	}

	att.Shield.Resonance.Avg = avg([]float64{
		att.Shield.Resonance.EM,
		att.Shield.Resonance.Explosive,
		att.Shield.Resonance.Thermal,
		att.Shield.Resonance.Kinetic,
	})

	att.Armor.Resonance.Avg = avg([]float64{
		att.Armor.Resonance.EM,
		att.Armor.Resonance.Explosive,
		att.Armor.Resonance.Thermal,
		att.Armor.Resonance.Kinetic,
	})

	att.Hull.Resonance.Avg = avg([]float64{
		att.Hull.Resonance.EM,
		att.Hull.Resonance.Explosive,
		att.Hull.Resonance.Thermal,
		att.Hull.Resonance.Kinetic,
	})

	att.Shield.Resonance.Min = min([]float64{
		att.Shield.Resonance.EM,
		att.Shield.Resonance.Explosive,
		att.Shield.Resonance.Thermal,
		att.Shield.Resonance.Kinetic,
	})

	att.Armor.Resonance.Min = min([]float64{
		att.Armor.Resonance.EM,
		att.Armor.Resonance.Explosive,
		att.Armor.Resonance.Thermal,
		att.Armor.Resonance.Kinetic,
	})

	att.Hull.Resonance.Min = min([]float64{
		att.Hull.Resonance.EM,
		att.Hull.Resonance.Explosive,
		att.Hull.Resonance.Thermal,
		att.Hull.Resonance.Kinetic,
	})
	att.Shield.Resonance.Max = max([]float64{
		att.Shield.Resonance.EM,
		att.Shield.Resonance.Explosive,
		att.Shield.Resonance.Thermal,
		att.Shield.Resonance.Kinetic,
	})

	att.Armor.Resonance.Max = max([]float64{
		att.Armor.Resonance.EM,
		att.Armor.Resonance.Explosive,
		att.Armor.Resonance.Thermal,
		att.Armor.Resonance.Kinetic,
	})

	att.Hull.Resonance.Max = max([]float64{
		att.Hull.Resonance.EM,
		att.Hull.Resonance.Explosive,
		att.Hull.Resonance.Thermal,
		att.Hull.Resonance.Kinetic,
	})

	att.MinEHP =
		int64((att.Hull.Hp / att.Hull.Resonance.Max) +
			(att.Armor.Hp / att.Armor.Resonance.Max) +
			(att.Shield.Hp / att.Shield.Resonance.Max))

	att.MaxEHP =
		int64((att.Hull.Hp / att.Hull.Resonance.Min) +
			(att.Armor.Hp / att.Armor.Resonance.Min) +
			(att.Shield.Hp / att.Shield.Resonance.Min))

	att.AvgEHP =
		int64((att.Hull.Hp / att.Hull.Resonance.Avg) +
			(att.Armor.Hp / att.Armor.Resonance.Avg) +
			(att.Shield.Hp / att.Shield.Resonance.Avg))

	return nil
}
func (c *Context) fillMWDAffectedAttributes(att *MWDAttributes) error {
	var v C.double
	attributes := map[string]C.ushort{
		"MaxVelocity":     37,
		"SignatureRadius": 552,
	}
	for k, id := range attributes {
		if r := C.dogma_get_ship_attribute(c.ctx, id, &v); r != 0 {
			return errors.New("Could not get attributes")
		}
		switch k {
		case "MaxVelocity":
			att.MaxVelocity = float64(v)
		case "SignatureRadius":
			att.SignatureRadius = float64(v)
		}
	}

	return c.fillCapacitorAttributes(att)
}
