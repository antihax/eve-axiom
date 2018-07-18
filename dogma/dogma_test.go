package dogma

import "testing"
import "github.com/stretchr/testify/assert"

func BenchmarkRifter(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ctx, err := NewContext()
		assert.Nil(b, err)
		ctx.SetShip(587)              // Rifter
		i, err := ctx.AddModule(3831) // Medium Shield Extender II
		assert.Nil(b, err)
		ctx.GetShipAttribute(263)
		ctx.GetModuleAttribute(50, i)
		ctx.Destroy()
	}
}

func BenchmarkRifterReuseCtx(b *testing.B) {
	ctx, err := NewContext()
	assert.Nil(b, err)
	ctx.SetShip(587) // Rifter
	for n := 0; n < b.N; n++ {
		i, err := ctx.AddModule(3831) // Medium Shield Extender II
		assert.Nil(b, err)
		ctx.GetShipAttribute(263)
		ctx.GetModuleAttribute(50, i)
		ctx.Strip()
	}
}

func TestDogmaContext_PowerLeft(t *testing.T) {
	ctx, err := NewContext()
	assert.Nil(t, err)
	ctx.SetShip(587) // Rifter
	p, err := ctx.PowerLeft()
	assert.Nil(t, err)
	assert.EqualValues(t, 51.25, p)
	ctx.AddModule(3831) // Medium Shield Extender II
	p, err = ctx.PowerLeft()
	assert.Nil(t, err)
	assert.EqualValues(t, 28.75, p)
}

func TestDogmaContext_Validate(t *testing.T) {
	ctx, err := NewContext()
	assert.Nil(t, err)
	ctx.SetShip(587) // Rifter
	powerLeft, cpuLeft, err := ctx.Validate()
	assert.Nil(t, err)
	assert.EqualValues(t, 51.25, powerLeft)
	assert.EqualValues(t, 162.5, cpuLeft)
}

func TestDogmaContext_GetChargeAttributes(t *testing.T) {
	ctx, err := NewContext()
	assert.Nil(t, err)
	ctx.SetShip(587)               // Rifter
	i, err := ctx.AddModule(10631) // Rocket Launcher II
	assert.Nil(t, err)
	ctx.AddCharge(2514, i) // Inferno Rocket
	//DamageTypeEm:        dmgMultiplier * f.ReadChargeAttribute(114, idx),
	//DamageTypeKinetic:   dmgMultiplier * f.ReadChargeAttribute(117, idx),
	//DamageTypeExplosive: dmgMultiplier * f.ReadChargeAttribute(116, idx),
	//DamageTypeThermal:   dmgMultiplier * f.ReadChargeAttribute(118, idx),
	attrs, err := ctx.GetChargeAttributes([]uint16{114, 116, 117, 118}, i)
	assert.Nil(t, err)
	assert.EqualValues(t, 0.0, attrs[0])
	assert.EqualValues(t, 0.0, attrs[1])
	assert.EqualValues(t, 0.0, attrs[2])
	assert.EqualValues(t, 45.375, attrs[3])
}
