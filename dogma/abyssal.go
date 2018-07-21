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
