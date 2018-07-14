package axiom

import (
	dogma "github.com/antihax/go-dogma"
	"github.com/antihax/goesi/esi"
)

// GetAttributesFromKillmail takes an ESI killmail and returns the attributes
func (c *Axiom) GetAttributesFromKillmail(km *esi.GetKillmailsKillmailIdKillmailHashOk) error {
	// setup dogma context
	ctx, err := dogma.NewContext()
	if err != nil {
		return err
	}
	// Destroy the context when we are done
	defer ctx.Destroy()

	// set the context ship from the killmail
	if err := ctx.SetShip(uint32(km.Victim.ShipTypeId)); err != nil {
		return err
	}

	// Store modules and charges
	modules := make(map[int32]uint32)
	charges := make(map[int32]uint32)

	// loop all items on the lost ship, adding drones, and first pass to find modules and charges
	for _, item := range km.Victim.Items {
		if isFitted(item.Flag) {
			typeID := uint32(item.ItemTypeId)
			catID := ctx.GetCategory(typeID)
			if catID == 8 { // Charge
				charges[item.Flag] = typeID
			} else if catID == 18 { // Drone
				q := uint32(item.QuantityDestroyed + item.QuantityDropped)
				err := ctx.AddDrone(typeID, q)
				if err != nil {
					return err
				}
			} else {
				modules[item.Flag] = typeID
			}
		}
	}

	// Second pass with modules and loaded charges
	for i := range modules {
		if _, ok := charges[i]; ok {
			_, err := ctx.AddModuleAndCharge(modules[i], charges[i])
			if err != nil {
				return err
			}
		} else {
			_, err := ctx.AddModule(modules[i])
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func isFitted(location int32) bool {
	for _, n := range []int32{
		87, // Drone Bay

		11, // LoSlot0
		12, // LoSlot1
		13, // LoSlot2
		14, // LoSlot3
		15, // LoSlot4
		16, // LoSlot5
		17, // LoSlot6
		18, // LoSlot7

		19, // MedSlot0
		20, // MedSlot1
		21, // MedSlot2
		22, // MedSlot3
		23, // MedSlot4
		24, // MedSlot5
		25, // MedSlot6
		26, // MedSlot7

		27, // HiSlot0
		28, // HiSlot1
		29, // HiSlot2
		30, // HiSlot3
		31, // HiSlot4
		32, // HiSlot6
		33, // HiSlot7
		34, // HiSlot8

		92, // RigSlot0
		93, // RigSlot1
		94, // RigSlot2
		95, // RigSlot3
		96, // RigSlot4
		97, // RigSlot5
		98, // RigSlot6
		99, // RigSlot7

		125, // SubSystemSlot0
		126, // SubSystemSlot0
		127, // SubSystemSlot0
		128, // SubSystemSlot0
		129, // SubSystemSlot0
		130, // SubSystemSlot0
		131, // SubSystemSlot0
		132, // SubSystemSlot0

		159, // FighterTube0
		160, // FighterTube1
		161, // FighterTube2
		162, // FighterTube3
		163, // FighterTube4

	} {

		if n == location {
			return true
		}
	}
	return false
}
