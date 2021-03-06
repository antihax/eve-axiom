package axiom

import (
	"github.com/antihax/eve-axiom/attributes"
	"github.com/antihax/eve-axiom/dogma"
	"github.com/antihax/goesi/esi"
)

func (c *Axiom) getAttributesFromKillmail(km *esi.GetKillmailsKillmailIdKillmailHashOk) (*attributes.Attributes, error) {
	// setup dogma context
	ctx, err := dogma.NewContext()
	if err != nil {
		return nil, err
	}
	// Destroy the context when we are done
	defer ctx.Destroy()

	// set the context ship from the killmail
	if err := ctx.SetShip(uint32(km.Victim.ShipTypeId)); err != nil {
		return nil, err
	}

	// Store modules and charges
	modules := make(map[int32]uint32)
	charges := make(map[int32]uint32)

	// loop all items on the lost ship, adding drones, and first pass to find modules and charges
	for _, item := range km.Victim.Items {
		if dogma.IsFitted(item.Flag) {

			// Swap any abyssal mods to t2
			item.ItemTypeId = dogma.SwapAbyssal(item.ItemTypeId)

			if !dogma.IsCyno(item.ItemTypeId) {
				typeID := uint32(item.ItemTypeId)
				catID := ctx.GetCategory(typeID)
				if catID == 8 { // Charge
					charges[item.Flag] = typeID
				} else if catID == 18 { // Drone
					q := uint8(item.QuantityDestroyed + item.QuantityDropped)
					err := ctx.AddDrone(typeID, q)
					if err != nil {
						return nil, err
					}
				} else {
					modules[item.Flag] = typeID
				}
			}
		}
	}

	// Second pass with modules and loaded charges
	for i := range modules {
		if _, ok := charges[i]; ok {
			_, err := ctx.AddModuleAndCharge(modules[i], charges[i], i)
			if err != nil {
				return nil, err
			}
		} else {
			_, err := ctx.AddModule(modules[i], i)
			if err != nil {
				return nil, err
			}
		}
	}

	return ctx.GetAttributes()
}
