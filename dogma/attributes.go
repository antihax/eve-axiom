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

	"github.com/antihax/eve-axiom/attributes"
	"github.com/bradfitz/slice"
)

// GetAttributes gets all the fit attributes
func (c *Context) GetAttributes() (*attributes.Attributes, error) {
	var (
		att attributes.Attributes
		err error
	)

	att.Attributes = make(map[string]float64)
	// Fix MWD/AB states
	c.ActivateAllModules()

	att.Modules = make(map[uint8]*attributes.Module)

	att.ShipID = int32(c.shipID)

	if att.PGRemaining, err = c.PowerLeft(); err != nil {
		return nil, err
	}

	if att.CPURemaining, err = c.CPULeft(); err != nil {
		return nil, err
	}

	// Fill basic ship attributes
	if err := c.fillShipAttributes(&att); err != nil {
		return nil, err
	}

	if err := c.fillTankAttributes(&att); err != nil {
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

	if err := c.optimalDroneConfiguration(&att); err != nil {
		return nil, err
	}

	c.DeactivateMWD()
	if err := c.fillMWDAffectedAttributes(&att.WithoutMWD); err != nil {
		return nil, err
	}

	if err := c.fillAllShipAttributes(&att); err != nil {
		return nil, err
	}

	att.TotalAlpha = att.DroneAlpha + att.ModuleAlpha
	att.TotalDPS = att.DroneDPS + att.ModuleDPS

	if att.Structure.Resonance.Max > 0 && att.Armor.Resonance.Max > 0 && att.Shield.Resonance.Max > 0 {
		att.MinEHP =
			int64((att.Structure.Hp / att.Structure.Resonance.Max) +
				(att.Armor.Hp / att.Armor.Resonance.Max) +
				(att.Shield.Hp / att.Shield.Resonance.Max))

		att.MaxEHP =
			int64((att.Structure.Hp / att.Structure.Resonance.Min) +
				(att.Armor.Hp / att.Armor.Resonance.Min) +
				(att.Shield.Hp / att.Shield.Resonance.Min))

		att.AvgEHP =
			int64((att.Structure.Hp / att.Structure.Resonance.Avg) +
				(att.Armor.Hp / att.Armor.Resonance.Avg) +
				(att.Shield.Hp / att.Shield.Resonance.Avg))

		att.MinRPS =
			(att.StructureRepairPerSecond / att.Structure.Resonance.Max) +
				(att.ArmorRepairPerSecond / att.Armor.Resonance.Max) +
				(att.ShieldRepairPerSecond / att.Shield.Resonance.Max)

		att.MaxRPS =
			(att.StructureRepairPerSecond / att.Structure.Resonance.Min) +
				(att.ArmorRepairPerSecond / att.Armor.Resonance.Min) +
				(att.ShieldRepairPerSecond / att.Shield.Resonance.Min)

		att.AvgRPS =
			(att.StructureRepairPerSecond / att.Structure.Resonance.Avg) +
				(att.ArmorRepairPerSecond / att.Armor.Resonance.Avg) +
				(att.ShieldRepairPerSecond / att.Shield.Resonance.Avg)
	}
	return &att, nil
}

func (c *Context) fillAllShipAttributes(att *attributes.Attributes) error {
	var v C.double

	// Get all known attributes
	for _, k := range typeAttributeMap[int32(c.shipID)] {
		if r := C.dogma_get_ship_attribute(c.ctx, C.ushort(k), &v); r == 0 {
			if v != 0 {
				att.Attributes[attributeMap[k]] = float64(v)
			}
		}
	}

	return nil
}

func (c *Context) fillCapacitorAttributes(att *attributes.MWDAttributes) error {
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

func (c *Context) fillDroneAttributes(att *attributes.Attributes) error {
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
			m := attributes.Drone{
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

			m.AlphaDamage = m.DamageMultiplier * (m.Damage.Kinetic + m.Damage.EM + m.Damage.Explosive + m.Damage.Thermal)
			m.DamagePerSecond = m.AlphaDamage / (m.Duration / 1000)

			att.Drones = append(att.Drones, m)
		}
	}

	return nil
}

func (c *Context) optimalDroneConfiguration(att *attributes.Attributes) error {
	// Early out if there is no dronebay
	if att.DroneBandwith < 1 {
		return nil
	}

	// Save our list of available drones
	type dcs struct {
		dps       float64
		alpha     float64
		bandwidth float64
		used      bool
	}
	dc := []dcs{}
	for _, d := range att.Drones {
		if d.DamagePerSecond > 0 {
			for i := int32(0); i < d.Quantity; i++ {
				dc = append(dc, dcs{
					dps:       d.DamagePerSecond,
					alpha:     d.AlphaDamage,
					bandwidth: d.DroneBandwith,
				})
			}
		}
	}

	// Sort the list lowest bandwidth first
	slice.Sort(dc[:], func(i, j int) bool {
		return dc[i].bandwidth < dc[j].bandwidth
	})

	// Determine optimal DPS
	droneSlot := [5]*dcs{}
	for i := range droneSlot {
		droneSlot[i] = &dcs{}
	}
	availableBandwith := att.DroneBandwith
	rounds := 2
	for rounds > 0 {
		for d := range dc {
			for i := range droneSlot {
				// Is this is a better unused drone and we have room?
				if !dc[d].used && dc[d].dps > droneSlot[i].dps &&
					availableBandwith+(droneSlot[i].bandwidth-dc[d].bandwidth) >= 0 {
					// Recalculate Bandwidth
					availableBandwith += droneSlot[i].bandwidth - dc[d].bandwidth
					// swap the drone in the slot
					droneSlot[i].used = false
					dc[d].used = true
					droneSlot[i] = &dc[d]
					break // don't fill with the same drone
				}
			}
		}
		rounds--
	}

	var dps, alpha float64
	for _, d := range droneSlot {
		dps += d.dps
		alpha += d.alpha
	}

	att.DroneDPS = dps
	att.DroneAlpha = alpha
	return nil
}

func (c *Context) fillModuleAttributes(att *attributes.Attributes) error {
	for _, mod := range c.mods {
		typeID := C.dogma_typeid_t(mod.typeID)
		i := uint(0)
		var effect C.dogma_effectid_t

		m := &attributes.Module{
			Attributes: make(map[string]float64),
			TypeID:     int32(mod.typeID),
			ChargeID:   int32(mod.chargeID),
			Location:   mod.location,
		}
		att.Modules[uint8(mod.idx)] = m
		var err error
		// Get all known attributes
		for _, k := range typeAttributeMap[int32(mod.typeID)] {
			var v float64
			if v, err = c.GetModuleAttribute(uint16(k), mod.idx); err == nil {
				if v != 0 {
					m.Attributes[attributeMap[k]] = v
				}
			}
		}
		for _, k := range typeAttributeMap[int32(mod.chargeID)] {
			var v float64
			if v, err = c.GetChargeAttribute(uint16(k), mod.idx); err == nil {
				if v != 0 {
					m.Attributes[attributeMap[k]] = v
				}
			}
		}
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
			m.Duration = float64(duration)
			m.Tracking = float64(tracking)
			m.Discharge = float64(discharge)
			m.Optimal = float64(optimal)
			m.Falloff = float64(falloff)
			m.Chance = float64(chance)

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

		}
	}

	for _, d := range att.Modules {
		att.ModuleDPS += d.DamagePerSecond
		att.ModuleAlpha += d.AlphaDamage
		if d.Duration > 0 {
			att.RemoteArmorRepairPerSecond += d.RemoteArmorRepairAmount / (d.Duration / 1000)
			att.RemoteShieldTransferPerSecond += d.RemoteShieldTransferAmount / (d.Duration / 1000)
			att.RemoteStructureRepairPerSecond += d.RemoteStructureRepairAmount / (d.Duration / 1000)
			att.RemoteEnergyTransferPerSecond += d.RemoteEnergyTransferAmount / (d.Duration / 1000)
			att.ArmorRepairPerSecond += d.ArmorRepair / (d.Duration / 1000)
			att.ShieldRepairPerSecond += d.ShieldRepair / (d.Duration / 1000)
			att.StructureRepairPerSecond += d.StructureRepair / (d.Duration / 1000)
			att.EnergyNeutralizerPerSecond += (d.NeutralizerAmount + d.NosferatuAmount) / (d.Duration / 1000)
		}
	}

	return nil
}

func (c *Context) fillShipAttributes(att *attributes.Attributes) error {

	attributes := map[string]uint16{
		"warpSpeed":      600,
		"maxDrones":      283,
		"droneBandwidth": 1271,
		"maxTargetRange": 76,

		"agility": 70,

		"power": 11,
		"cpu":   48,

		"scanGravimetricStrength":   211,
		"scanLadarStrength":         209,
		"scanMagnetometricStrength": 210,
		"scanRadarStrength":         208,

		"scanResolution":     564,
		"shieldRechargeRate": 479,
	}

	for k, id := range attributes {
		var (
			v   float64
			err error
		)
		if v, err = c.GetShipAttribute(id); err != nil {
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

		case "power":
			att.PGTotal = float64(v)
		case "cpu":
			att.CPUTotal = float64(v)

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

func (c *Context) fillTankAttributes(att *attributes.Attributes) error {
	resonances := map[string]uint16{
		"ArmorEm":        267,
		"ArmorExplosive": 268,
		"ArmorKinetic":   269,
		"ArmorThermal":   270,

		"ShieldEm":        271,
		"ShieldExplosive": 272,
		"ShieldKinetic":   273,
		"ShieldThermal":   274,

		"StructureEm":        113,
		"StructureExplosive": 111,
		"StructureKinetic":   109,
		"StructureThermal":   110,

		"ShieldCapacity": 263,
		"ArmorHp":        265,
		"Hp":             9,
	}
	for k, id := range resonances {
		var (
			v   float64
			err error
		)
		if v, err = c.GetShipAttribute(id); err != nil {
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

		case "StructureEm":
			att.Structure.Resonance.EM = float64(v)
		case "StructureExplosive":
			att.Structure.Resonance.Explosive = float64(v)
		case "StructureKinetic":
			att.Structure.Resonance.Kinetic = float64(v)
		case "StructureThermal":
			att.Structure.Resonance.Thermal = float64(v)

		case "ShieldCapacity":
			att.Shield.Hp = float64(v)
		case "ArmorHp":
			att.Armor.Hp = float64(v)
		case "Hp":
			att.Structure.Hp = float64(v)

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

	att.Structure.Resonance.Avg = avg([]float64{
		att.Structure.Resonance.EM,
		att.Structure.Resonance.Explosive,
		att.Structure.Resonance.Thermal,
		att.Structure.Resonance.Kinetic,
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

	att.Structure.Resonance.Min = min([]float64{
		att.Structure.Resonance.EM,
		att.Structure.Resonance.Explosive,
		att.Structure.Resonance.Thermal,
		att.Structure.Resonance.Kinetic,
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

	att.Structure.Resonance.Max = max([]float64{
		att.Structure.Resonance.EM,
		att.Structure.Resonance.Explosive,
		att.Structure.Resonance.Thermal,
		att.Structure.Resonance.Kinetic,
	})

	return nil
}
func (c *Context) fillMWDAffectedAttributes(att *attributes.MWDAttributes) error {
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
