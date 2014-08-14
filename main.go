package main

import (
	"flag"
	"fmt"
	"github.com/tajtiattila/vjoy"
	"github.com/tajtiattila/xinput"
	"math"
	"os"
	"time"
)

func main() {
	cfg := NewConfig()
	defcfg := flag.String("defcfg", "", "save default configfile into file specified")
	fn := flag.String("cfg", "", "configfile")
	flag.Parse()

	if *defcfg != "" {
		err := cfg.Save(*defcfg)
		if err != nil {
			abort(err)
		}
		return
	}

	if *fn != "" {
		err := cfg.Load(*fn)
		if err != nil {
			abort(err)
		}
	} else {
		if cfg.Load("joyster.cfg") != nil {
			// no error, but reset to default
			cfg = NewConfig()
		}
	}

	fmt.Println("vJoy version:", vjoy.Version())
	fmt.Println("  Product:       ", vjoy.ProductString())
	fmt.Println("  Manufacturer:  ", vjoy.ManufacturerString())
	fmt.Println("  Serial number: ", vjoy.SerialNumberString())
	d, err := vjoy.Acquire(1)
	if err != nil {
		abort(err)
	}
	defer d.Relinquish()

	var xs xinput.State
	for {
		xinput.GetState(0, &xs)
		update(cfg, &xs.Gamepad, d)
		time.Sleep(time.Duration(cfg.UpdateMicros) * time.Microsecond)
	}
}

func abort(err error) {
	fmt.Println(err)
	os.Exit(1)
}

func update(c *Config, gp *xinput.Gamepad, d *vjoy.Device) {
	updateAxes(c, gp, d)
	updateButtons(c, gp, d)
	d.Update()
}

func updateAxes(c *Config, gp *xinput.Gamepad, d *vjoy.Device) {
	var lt, rt thumbStick
	lt.set(gp.ThumbLX, gp.ThumbLY)
	rt.set(gp.ThumbRX, gp.ThumbRY)
	if c.ThumbCircle {
		lt.circularize()
		rt.circularize()
	}
	d.Axis(vjoy.AxisX).Setf(axismap(c.ThumbLX, lt.xv, lt.xs))
	d.Axis(vjoy.AxisY).Setf(axismap(c.ThumbLY, lt.yv, lt.ys))
	d.Axis(vjoy.AxisRX).Setf(axismap(c.ThumbRX, rt.xv, rt.xs))
	d.Axis(vjoy.AxisRY).Setf(axismap(c.ThumbRY, rt.yv, rt.ys))
	if c.LeftTrigger.Axis {
		d.Axis(vjoy.AxisZ).Setuf(float32(gp.LeftTrigger) / 255)
	}
	if c.RightTrigger.Axis {
		d.Axis(vjoy.AxisRZ).Setuf(float32(gp.RightTrigger) / 255)
	}
}

func updateButtons(c *Config, gp *xinput.Gamepad, d *vjoy.Device) {
	btns := uint32(gp.Buttons)
	if c.LeftTrigger.touch <= uint16(gp.LeftTrigger) {
		if c.LeftTrigger.pull <= uint16(gp.LeftTrigger) {
			btns |= 1 << 16
		} else {
			btns |= 1 << 17
		}
	}
	if c.RightTrigger.touch <= uint16(gp.RightTrigger) {
		if c.RightTrigger.pull <= uint16(gp.RightTrigger) {
			btns |= 1 << 18
		} else {
			btns |= 1 << 19
		}
	}
	for _, bc := range c.Buttons {
		bc.handler.Handle(d, bc, (btns&bc.fmask) == bc.imask)
	}
}

type thumbStick struct {
	xv float64
	xs float32
	yv float64
	ys float32
}

func (t *thumbStick) set(xi, yi int16) {
	t.xv, t.xs = float64FromInt16(xi)
	t.yv, t.ys = float64FromInt16(yi)
}

func (t *thumbStick) circularize() {
	if t.xv*t.xv+t.yv*t.yv < 1e-3 {
		return
	}

	var u float64
	if t.xv > t.yv {
		u = t.yv / t.xv // yu/xu = yv/xv
	} else {
		u = t.xv / t.yv // xu/yu = xv/yv
	}
	m := math.Sqrt(1 + u*u)
	t.xv *= m
	t.yv *= m
}

func axismap(c *AxisConfig, vabs float64, sign float32) float32 {
	if vabs <= c.Min {
		return 0
	}
	if c.Max <= vabs {
		return sign
	}
	vabs = (vabs - c.Min) / (c.Max - c.Min)
	return sign * float32(math.Pow(vabs, c.Pow))
}

func float64FromInt16(v int16) (abs float64, sign float32) {
	if v < 0 {
		return float64(-int(v)) / 0x8000, -1
	} else {
		return float64(v) / 0x7fff, 1
	}
	return 0, 0 // not reached
}
