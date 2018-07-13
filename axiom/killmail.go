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

	// loop all items on the lost ship and add to the context
	for _, item := range km.Victim.Items {
		if isFitted(item.Flag) {
			// sum the quantity of items
			q := item.QuantityDestroyed + item.QuantityDropped
			for q {
				ctx.AddModule(item.ItemTypeId)
				q--
			}
		}
	}
}

func isFitted(location int32) bool {
	for _, n := range []int32{
		//87, // Drone Bay
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
	} {
		if n == location {
			return true
		}
	}
	return false
}
