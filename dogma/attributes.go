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
	"math"
	"unicode"

	"github.com/antihax/eve-axiom/attributes"
	"github.com/bradfitz/slice"
)

// GetAttributes gets all the fit attributes
func (c *Context) GetAttributes() (*attributes.Attributes, error) {

	att := &attributes.Attributes{}

	att.TypeID = int32(c.shipID)
	att.Ship = make(map[string]float64)
	att.Modules = make(map[uint8]map[string]float64)
	a := att.Ship

	// Fix MWD/AB states
	c.ActivateAllModules()

	if pg, err := c.PowerLeft(); err != nil {
		return nil, err
	} else {
		a["powerRemaining"] = pg
	}

	if cpu, err := c.CPULeft(); err != nil {
		return nil, err
	} else {
		a["cpuRemaining"] = cpu
	}

	if err := c.fillAllShipAttributes(a); err != nil {
		return nil, err
	}
	if c.mwd > 0 {
		if err := c.fillMWDAffectedAttributes(a, "MWD"); err != nil {
			return nil, err
		}
	}
	c.fillTankAttributes(a)

	if err := c.fillModuleAttributes(att.Modules); err != nil {
		return nil, err
	}

	if err := c.fillDroneAttributes(att.Drones); err != nil {
		return nil, err
	}

	if err := c.optimalDroneConfiguration(att); err != nil {
		return nil, err
	}

	if err := c.sumModuleAttributes(att); err != nil {
		return nil, err
	}

	c.DeactivateMWD()
	if err := c.fillMWDAffectedAttributes(a, ""); err != nil {
		return nil, err
	}

	totalAlpha := a["droneAlphaDamage"] + a["moduleAlphaDamage"]
	totalDPS := a["droneDPS"] + a["moduleDPS"]
	if totalDPS > 0 {
		a["totalAlphaDamage"] = totalAlpha
		a["totalDPS"] = totalDPS
	}

	return att, nil
}

func (c *Context) fillAllShipAttributes(a map[string]float64) error {
	// Get all known attributes
	for _, k := range typeAttributeMap[int32(c.shipID)] {
		var v C.double
		if r := C.dogma_get_ship_attribute(c.ctx, C.ushort(k), &v); r == 0 {
			if v != 0 {
				a[attributeMap[k]] = float64(v)
			}
		}
	}
	return nil
}

func (c *Context) fillDroneAttributes(droneList []map[string]float64) error {
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

			m := make(map[string]float64)
			m["typeID"] = float64(drone.typeID)
			m["quantity"] = float64(drone.quantity)
			if duration > 0 {
				m["duration"] = float64(duration)
			}
			if tracking > 0 {
				m["trackingSpeed"] = float64(tracking)
			}
			if discharge > 0 {
				m["capacitorNeed"] = float64(discharge)
			}
			if optimal > 0 {
				m["maxRange"] = float64(optimal)
			}
			if tracking > 0 {
				m["falloff"] = float64(falloff)
			}
			if chance > 0 {
				m["chance"] = float64(chance)
			}

			// Get all known attributes
			var err error
			for _, k := range typeAttributeMap[int32(drone.typeID)] {
				var v float64

				if v, err = c.GetDroneAttribute(uint16(k), typeID); err == nil {
					if v != 0 {
						m[attributeMap[k]] = v
					}
				}
			}

			if effectIsAssistanceMap[int32(effect)] {
				m["isAssistance"] = 1
			}
			if sumDamage(m) > 0 {
				if m["damageMultiplier"] == 0 {
					m["damageMultiplier"] = 1
				}
				m["alphaDamage"] = m["damageMultiplier"] * sumDamage(m)
				m["dps"] = m["alphaDamage"] / (m["duration"] / 1000)
			}
			droneList = append(droneList, m)
		}
	}

	return nil
}

func sumDamage(m map[string]float64) float64 {
	return m["explosiveDamage"] +
		m["emDamage"] +
		m["kineticDamage"] +
		m["thermalDamage"]
}

func (c *Context) sumModuleAttributes(s *attributes.Attributes) error {
	sumValues := map[string]string{
		"dps":         "moduleDPS",
		"alphaDamage": "moduleAlphaDamage",
	}

	minimumValues := map[string]string{
		"speedFactor": "stasisWebifierStrength",
	}

	sumPositiveValues := map[string]string{
		"warpScrambleStrength": "totalWarpScrambleStrength",
	}

	sumDurations := map[string]string{
		"armorDamageAmount":       "armorDamageAmountPerSecond",
		"structureDamageAmount":   "structureDamageAmountPerSecond",
		"shieldBonus":             "shieldBonusAmountPerSecond",
		"powerTransferAmount":     "powerTransferAmountPerSecond",
		"energyNeutralizerAmount": "energyNeutralizerAmountPerSecond",
	}

	for _, m := range s.Modules {
		for k, v := range sumValues {
			if m[k] != 0 {
				s.Ship[v] += m[k]
			}
		}
		for k, v := range sumPositiveValues {
			if m[k] > 0 {
				s.Ship[v] += m[k]
			}
		}

		for k, v := range minimumValues {
			if m[k] < 0 && m[k] < s.Ship[v] {
				s.Ship[v] = m[k]
			}
		}

		for k, v := range sumDurations {
			if m[k] != 0 && m["duration"] > 0 {
				if m["isAssistance"] > 0 {
					s.Ship["remote"+UcFirst(v)] += m[k] / (m["duration"] / 1000)
				} else {
					s.Ship[v] += m[k] / (m["duration"] / 1000)
				}
			}
		}
	}
	return nil
}
func UcFirst(str string) string {
	for i, v := range str {
		return string(unicode.ToUpper(v)) + str[i+1:]
	}
	return ""
}

func (c *Context) optimalDroneConfiguration(s *attributes.Attributes) error {
	// Early out if there is no dronebay
	if s.Ship["droneBandwidth"] < 1 {
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
	for _, d := range s.Drones {
		if d["dps"] > 0 {
			for i := float64(0); i < d["quantity"]; i++ {
				dc = append(dc, dcs{
					dps:       d["dps"],
					alpha:     d["alphaDamage"],
					bandwidth: d["droneBandwidth"],
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
	availableBandwith := s.Ship["droneBandwidth"]
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

	s.Ship["droneDPS"] = dps
	s.Ship["droneAlphaDamage"] = alpha
	return nil
}

func (c *Context) fillModuleAttributes(moduleList map[uint8]map[string]float64) error {
	for _, mod := range c.mods {
		typeID := C.dogma_typeid_t(mod.typeID)
		i := uint(0)
		m := make(map[string]float64)

		m["typeID"] = float64(mod.typeID)
		if mod.chargeID > 0 {
			m["chargeTypeID"] = float64(mod.chargeID)
		}
		m["location"] = float64(mod.location)

		moduleList[uint8(mod.idx)] = m
		var err error
		// Get all known attributes
		for _, k := range typeAttributeMap[int32(mod.typeID)] {
			var v float64
			if v, err = c.GetModuleAttribute(uint16(k), mod.idx); err == nil {
				if v != 0 {
					m[attributeMap[k]] = v
				}
			}
		}
		// And the charge attributes
		for _, k := range typeAttributeMap[int32(mod.chargeID)] {
			var v float64
			if v, err = c.GetChargeAttribute(uint16(k), mod.idx); err == nil {
				if v != 0 {
					m[attributeMap[k]] = v
				}
			}
		}
		for {
			var effect C.dogma_effectid_t
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

			if duration > 0 {
				m["duration"] = float64(duration)
			}
			if tracking > 0 {
				m["trackingSpeed"] = float64(tracking)
			}
			if discharge > 0 {
				m["capacitorNeed"] = float64(discharge)
			}
			if optimal > 0 {
				m["maxRange"] = float64(optimal)
			}
			if tracking > 0 {
				m["falloff"] = float64(falloff)
			}
			if chance > 0 {
				m["chance"] = float64(chance)
			}
			if effectIsAssistanceMap[int32(effect)] {
				m["isAssistance"] = 1
			}
		}

		if sumDamage(m) > 0 {
			if m["damageMultiplier"] == 0 {
				m["damageMultiplier"] = 1
			}
			m["alphaDamage"] = m["damageMultiplier"] * sumDamage(m)
			m["dps"] = m["alphaDamage"] / (m["duration"] / 1000)
		}
	}

	return nil
}

func (c *Context) fillTankAttributes(a map[string]float64) {

	// Early out if it isn't a proper ship
	if a["shieldEmDamageResonance"] == 0 || a["armorEmDamageResonance"] == 0 ||
		a["hullEmDamageResonance"] == 0 {
		return
	}

	a["shieldAvgDamageResonance"] = avg([]float64{
		a["shieldEmDamageResonance"],
		a["shieldExplosiveDamageResonance"],
		a["shieldKineticDamageResonance"],
		a["shieldThermalDamageResonance"],
	})

	a["armorAvgDamageResonance"] = avg([]float64{
		a["armorEmDamageResonance"],
		a["armorExplosiveDamageResonance"],
		a["armorKineticDamageResonance"],
		a["armorThermalDamageResonance"],
	})

	a["hullAvgDamageResonance"] = avg([]float64{
		a["hullEmDamageResonance"],
		a["hullExplosiveDamageResonance"],
		a["hullKineticDamageResonance"],
		a["hullThermalDamageResonance"],
	})

	a["shieldMinDamageResonance"] = min([]float64{
		a["shieldEmDamageResonance"],
		a["shieldExplosiveDamageResonance"],
		a["shieldKineticDamageResonance"],
		a["shieldThermalDamageResonance"],
	})

	a["armorMinDamageResonance"] = min([]float64{
		a["armorEmDamageResonance"],
		a["armorExplosiveDamageResonance"],
		a["armorKineticDamageResonance"],
		a["armorThermalDamageResonance"],
	})

	a["hullMinDamageResonance"] = min([]float64{
		a["hullEmDamageResonance"],
		a["hullExplosiveDamageResonance"],
		a["hullKineticDamageResonance"],
		a["hullThermalDamageResonance"],
	})

	a["shieldMaxDamageResonance"] = max([]float64{
		a["shieldEmDamageResonance"],
		a["shieldExplosiveDamageResonance"],
		a["shieldKineticDamageResonance"],
		a["shieldThermalDamageResonance"],
	})

	a["armorMaxDamageResonance"] = max([]float64{
		a["armorEmDamageResonance"],
		a["armorExplosiveDamageResonance"],
		a["armorKineticDamageResonance"],
		a["armorThermalDamageResonance"],
	})

	a["hullMaxDamageResonance"] = max([]float64{
		a["hullEmDamageResonance"],
		a["hullExplosiveDamageResonance"],
		a["hullKineticDamageResonance"],
		a["hullThermalDamageResonance"],
	})

	a["minEHP"] =
		(a["hp"] / a["hullMaxDamageResonance"]) +
			(a["armorHp"] / a["armorMaxDamageResonance"]) +
			(a["shieldCapacity"] / a["shieldMaxDamageResonance"])

	a["maxEHP"] =
		(a["hp"] / a["hullMinDamageResonance"]) +
			(a["armorHp"] / a["armorMinDamageResonance"]) +
			(a["shieldCapacity"] / a["shieldMinDamageResonance"])

	a["avgEHP"] =
		(a["hp"] / a["hullAvgDamageResonance"]) +
			(a["armorHp"] / a["armorAvgDamageResonance"]) +
			(a["shieldCapacity"] / a["shieldAvgDamageResonance"])

	if a["shieldRechargeRate"] > 0 {
		rate := a["shieldRechargeRate"] / 1000
		capacity := a["shieldCapacity"]
		peak := 10 / rate * math.Sqrt(0.25) * (1 - math.Sqrt(0.25)) * capacity

		a["minRPS"] = peak / a["shieldMaxDamageResonance"]
		a["maxRPS"] = peak / a["shieldMinDamageResonance"]
		a["avgRPS"] = peak / a["shieldAvgDamageResonance"]
	}

	if a["structureDamageAmountPerSecond"] > 0 || a["armorDamageAmountPerSecond"] > 0 || a["shieldBonusAmountPerSecond"] > 0 {
		a["minRPS"] +=
			(a["shieldBonusAmountPerSecond"] / a["hullMaxDamageResonance"]) +
				(a["armorDamageAmountPerSecond"] / a["armorMaxDamageResonance"]) +
				(a["shieldBonusAmountPerSecond"] / a["shieldMaxDamageResonance"])

		a["maxRPS"] +=
			(a["structureDamageAmountPerSecond"] / a["hullMinDamageResonance"]) +
				(a["armorDamageAmountPerSecond"] / a["armorMinDamageResonance"]) +
				(a["shieldBonusAmountPerSecond"] / a["shieldMinDamageResonance"])

		a["avgRPS"] +=
			(a["structureDamageAmountPerSecond"] / a["hullAvgDamageResonance"]) +
				(a["armorDamageAmountPerSecond"] / a["armorAvgDamageResonance"]) +
				(a["shieldBonusAmountPerSecond"] / a["shieldAvgDamageResonance"])
	}

}
func (c *Context) fillMWDAffectedAttributes(a map[string]float64, postfix string) error {
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
			a["maxVelocity"+postfix] = float64(v)
		case "SignatureRadius":
			a["signatureRadius"+postfix] = float64(v)
		}
	}
	return c.fillCapacitorAttributes(a, postfix)
}

func (c *Context) fillCapacitorAttributes(a map[string]float64, postfix string) error {
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
	a["capacitorCapacity"] = float64(cap.capacity)
	if cap.stable {
		a["capacitorStable"+postfix] = 1
		a["capacitorFraction"+postfix] = float64(C.get_cap_duration(cap))
	} else {
		a["capacitorDuration"+postfix] = float64(C.get_cap_duration(cap))
	}

	// Release memory used for the list
	C.dogma_free_capacitor_list(cap)
	return nil
}
