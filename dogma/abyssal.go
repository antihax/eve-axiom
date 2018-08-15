package dogma

func IsAbyssal(typeID int32) bool {
	for _, n := range []int32{
		47408, 47458, 47465, 47702, 47732, 47736, 47740, 47745, 47749, 47753,
		47757, 47769, 47773, 47777, 47781, 47785, 47789, 47793, 47800, 47804,
		47808, 47812, 47817, 47820, 47824, 47828, 47832, 47833, 47834, 47836,
		47838, 47840, 47842, 47844, 47846, 47876, 47877, 47878, 47879, 47880,
	} {

		if n == typeID {
			return true
		}
	}
	return false
}

// SwapAbyssal to a T2 module
func SwapAbyssal(typeID int32) int32 {
	switch typeID {
	case 47702:
		return 527 // Stasis Webifier II
	case 47769:
		return 1183 // Small Armor Repairer II
	case 47804:
		return 3831 // Medium Shield Extender II
	case 47408:
		return 12076 // 50MN Microwarpdrive II
	case 47757:
		return 12068 // 100MN Afterburner II
	case 47800:
		return 380 // Small Shield Extender II
	case 47753:
		return 12058 // 10MN Afterburner II
	case 47793:
		return 10842 // X-Large Shield Booster II
	case 47749:
		return 438 // 1MN Afterburner II
	case 47789:
		return 10842 // X-Large Shield Booster II
	case 47745:
		return 12084 // 500MN Microwarpdrive II
	case 47785:
		return 10850 // Medium Shield Booster II
	case 47832:
		return 12271 // Heavy Energy Neutralizer II
	case 47740:
		return 440 // 5MN Microwarpdrive II
	case 47781:
		return 400 // Small Shield Booster II
	case 47828:
		return 12267 // Medium Energy Neutralizer II
	case 47736:
		return 3244 // Warp Disruptor II
	case 47777:
		return 3540 // Large Armor Repairer II
	case 47824:
		return 13003 // Small Energy Neutralizer II
	case 47732:
		return 448 // Warp Scrambler II
	case 47773:
		return 3530 // Medium Armor Repairer II
	case 47808:
		return 3841 // Large Shield Extender II
	case 47812:
		return 20351 // Small Abyssal Armor Plates
	case 47817:
		return 20347 // Medium Abyssal Armor Plates
	case 47820:
		return 20353 // Large Abyssal Armor Plates
	case 48439:
		return 3504 // Large Cap Battery
	}
	return typeID
}

func IsCyno(typeID int32) bool {
	for _, n := range []int32{
		28646, 21096,
	} {

		if n == typeID {
			return true
		}
	}
	return false
}

func IsFitted(location int32) bool {
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
