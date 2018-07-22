package dogma

/*
 #cgo LDFLAGS: -ldogma -lJudy -lm
 #include <dogma.h>
 #include <dogma-extra.h>

 int get_charge_attributes_batch(dogma_context_t* ctx, dogma_key_t index, dogma_attributeid_t attributeid[],
                                       double out[], int size) {
   int i;
   for (i = 0; i < size; ++i) {
     double v;
     dogma_get_charge_attribute(ctx, index, attributeid[i], &v);
     out[i] = v;
   }
   return DOGMA_OK;
 }

 int get_power_left(dogma_context_t* ctx, double* out) {
   double power_output, power_left;
   dogma_get_ship_attribute(ctx, 11, &power_output);
   dogma_get_ship_attribute(ctx, 15, &power_left);
   *out = power_output - power_left;
   return DOGMA_OK;
 }

 int get_cpu_left(dogma_context_t* ctx, double* out) {
   double cpu_output, cpu_left;
   dogma_get_ship_attribute(ctx, 48, &cpu_output);
   dogma_get_ship_attribute(ctx, 49, &cpu_left);
   *out = cpu_output - cpu_left;
   return DOGMA_OK;
 }

 int validate(dogma_context_t* ctx, double* power, double *cpu) {
   get_power_left(ctx, power);
   get_cpu_left(ctx, cpu);
   return DOGMA_OK;
 }


*/
import "C"
import (
	"errors"
)

// bootstrap the dogma engine
func init() {
	if r := C.dogma_init(); r != 0 {
		panic("failed to initilize dogma engine")
	}
}

// Context is a single instance in the fitting engine
type Context struct {
	ctx    *C.dogma_context_t
	mods   []module
	drones []drone
	mwd    C.dogma_key_t
	ab     C.dogma_key_t
	shipID uint32
}

type module struct {
	idx      C.dogma_key_t
	typeID   uint32
	chargeID uint32
}

type drone struct {
	quantity uint8
	typeID   uint32
}

// NewContext returns a new context. This must be destroyed if successful.
func NewContext() (*Context, error) {
	c := &Context{}
	if r := C.dogma_init_context(&c.ctx); r != 0 {
		return nil, errors.New("failed to initialize")
	}
	return c, nil
}

// SetShip changes the context ship
func (c *Context) SetShip(t uint32) error {
	if r := C.dogma_set_ship(c.ctx, C.dogma_typeid_t(t)); r != 0 {
		return errors.New("failed to set ship")
	}
	c.shipID = t
	return nil
}

// AddModule to a ship
func (c *Context) AddModule(t uint32) (C.dogma_key_t, error) {
	var i C.dogma_key_t
	var hasit C.bool
	if r := C.dogma_add_module_s(c.ctx, C.dogma_typeid_t(t), &i, StateActive); r != 0 {
		return 0, errors.New("failed to add module")
	}
	c.mods = append(c.mods, module{typeID: t, idx: i})

	// Find any MWD and store
	if r := C.dogma_type_has_effect(C.dogma_typeid_t(t), StateActive, C.dogma_effectid_t(6730), &hasit); r != 0 {
		return 0, errors.New("failed to add module")
	}

	if bool(hasit) {
		c.mwd = i
	}

	// Find any MWD and store
	if r := C.dogma_type_has_effect(C.dogma_typeid_t(t), StateActive, C.dogma_effectid_t(6731), &hasit); r != 0 {
		return 0, errors.New("failed to add module")
	}

	if bool(hasit) {
		c.ab = i
	}

	return i, nil
}

// AddModuleAndCharge to a ship
func (c *Context) AddModuleAndCharge(t uint32, ch uint32) (C.dogma_key_t, error) {
	var i C.dogma_key_t
	if r := C.dogma_add_module_sc(c.ctx, C.dogma_typeid_t(t), &i, StateActive, C.dogma_typeid_t(ch)); r != 0 {
		return 0, errors.New("failed to add module with charge")
	}

	c.mods = append(c.mods, module{chargeID: ch, typeID: t, idx: i})
	return i, nil
}

// AddDrone to a ship
func (c *Context) AddDrone(t uint32, num uint8) error {
	if r := C.dogma_add_drone(c.ctx, C.dogma_typeid_t(t), C.uint(num)); r != 0 {
		return errors.New("failed to add module")
	}
	c.drones = append(c.drones, drone{typeID: t, quantity: num})
	return nil
}

// AddCharge to a module
func (c *Context) AddCharge(t uint32, idx C.dogma_key_t) error {
	if r := C.dogma_add_charge(c.ctx, idx, C.dogma_typeid_t(t)); r != 0 {
		return errors.New("failed to add module")
	}
	return nil
}

// GetCategory of a type
func (c *Context) GetCategory(t uint32) uint8 {
	return uint8(C.dogma_get_category_for_typeid(C.dogma_typeid_t(t)))
}

// GetShipAttribute of the current fit
func (c *Context) GetShipAttribute(t uint16) (float64, error) {
	var value C.double
	if r := C.dogma_get_ship_attribute(c.ctx, C.dogma_attributeid_t(t), &value); r != 0 {
		return 0, errors.New("failed to get attribute")
	}
	return float64(value), nil
}

// GetModuleAttribute of a fitted module on a ship
func (c *Context) GetModuleAttribute(t uint16, i C.dogma_key_t) (float64, error) {
	var value C.double
	if r := C.dogma_get_module_attribute(c.ctx, i, C.dogma_attributeid_t(t), &value); r != 0 {
		return 0, errors.New("failed get attribute")
	}
	return float64(value), nil
}

// GetChargeAttribute of a charge in a module, on a ship
func (c *Context) GetChargeAttribute(t uint16, i C.dogma_key_t) (float64, error) {
	var value C.double
	if r := C.dogma_get_charge_attribute(c.ctx, i, C.dogma_attributeid_t(t), &value); r != 0 {
		return 0, errors.New("failed to get attribute")
	}
	return float64(value), nil
}

// GetDroneAttribute of a fitted drone on a ship
func (c *Context) GetDroneAttribute(t uint16, i C.dogma_typeid_t) (float64, error) {
	var value C.double
	if r := C.dogma_get_drone_attribute(c.ctx, i, C.dogma_attributeid_t(t), &value); r != 0 {
		return 0, errors.New("failed get attribute")
	}
	return float64(value), nil
}

// GetChargeAttributes Get all the things of a charge, in a module, on a ship.
func (c *Context) GetChargeAttributes(attrIds []uint16, idx C.dogma_key_t) ([]float64, error) {
	size := len(attrIds)
	attrs := make([]C.dogma_attributeid_t, size)
	for i := 0; i < size; i++ {
		attrs[i] = C.dogma_attributeid_t(attrIds[i])
	}
	values := make([]C.double, size)
	if r := C.get_charge_attributes_batch(c.ctx, idx, &attrs[0], &values[0], C.int(size)); r != 0 {
		return nil, errors.New("failed to get attribute")
	}
	results := make([]float64, size)
	for i := 0; i < size; i++ {
		results[i] = float64(values[i])
	}
	return results, nil
}

// Destroy the context when finished
func (c *Context) Destroy() error {
	if r := C.dogma_free_context(c.ctx); r != 0 {
		return errors.New("failed to destroy context")
	}
	return nil
}

// Strip all the modules from a ship
func (c *Context) Strip() {
	for _, i := range c.mods {
		c.RemoveModule(i.idx)
	}
	c.mods = c.mods[:0]
}

// RemoveModule from a ship
func (c *Context) RemoveModule(idx C.dogma_key_t) error {
	if r := C.dogma_remove_module(c.ctx, idx); r != 0 {
		return errors.New("failed removing module")
	}
	return nil
}

// ActivateAllModules on a ship
func (c *Context) ActivateAllModules() error {
	for _, mod := range c.mods {
		if r := C.dogma_set_module_state(c.ctx, mod.idx, StateActive); r != 0 {
			return errors.New("failed activating module")
		}
	}

	// Turn off AB
	if c.ab > 0 {
		if r := C.dogma_set_module_state(c.ctx, c.ab, StateOnline); r != 0 {
			return errors.New("failed deactivating microwarp drive")
		}
	}

	return nil
}

// DeactivateMWD on a ship
func (c *Context) DeactivateMWD() error {
	// Turn off mwd
	if c.mwd > 0 {
		if r := C.dogma_set_module_state(c.ctx, c.mwd, StateOnline); r != 0 {
			return errors.New("failed deactivating microwarp drive")
		}
	}
	// Turn on AB
	if c.ab > 0 {
		if r := C.dogma_set_module_state(c.ctx, c.ab, StateActive); r != 0 {
			return errors.New("failed deactivating microwarp drive")
		}
	}
	return nil
}

// PowerLeft of the current fitting
func (c *Context) PowerLeft() (float64, error) {
	var value C.double
	if r := C.get_power_left(c.ctx, &value); r != 0 {
		return 0, errors.New("failed getting power")
	}
	return float64(value), nil
}

// CPULeft of the current fit
func (c *Context) CPULeft() (float64, error) {
	var value C.double
	if r := C.get_cpu_left(c.ctx, &value); r != 0 {
		return 0, errors.New("failed getting cpu")
	}
	return float64(value), nil
}

// Validate the current fit? Probably delete this
func (c *Context) Validate() (float64, float64, error) {
	var power, cpu C.double
	if r := C.validate(c.ctx, &power, &cpu); r != 0 {
		return 0, 0, errors.New("failed validating fit")
	}
	return float64(power), float64(cpu), nil
}
