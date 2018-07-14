package axiom

import (
	"errors"
	"fmt"

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
	fmt.Println(km.Victim.ShipTypeId)

	// Store modules and charges
	modules := make(map[int32]uint32)
	charges := make(map[int32]uint32)

	// loop all items on the lost ship, adding drones, and first pass to find modules and charges
	for _, item := range km.Victim.Items {
		if dogma.IsFitted(item.Flag) {
			if dogma.IsAbyssal(item.ItemTypeId) {
				return errors.New("Cannot fit Abyssal modules")
			}
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

	c.GetAttributesFromContext(ctx)

	return nil
}

func (c *Axiom) GetAttributesFromContext(ctx *dogma.Context) error {

	/*cpu, err := ctx.CPULeft()
	if err != nil {
		return err
	}
	fmt.Println("CPU: ", cpu)

	pg, err := ctx.PowerLeft()
	if err != nil {
		return err
	}
	fmt.Println("Pg: ", pg)*/

	ctx.GetShipAttributes()
	return nil
}
