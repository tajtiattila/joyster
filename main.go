package main

import (
	"flag"
	"fmt"
	"github.com/tajtiattila/vjoy"
	"github.com/tajtiattila/xinput"
	"os"
	"time"
)

var (
	savecfg = flag.String("save", "", "save default configfile into file specified and exit")
	loadcfg = flag.String("cfg", "", "configfile to use")
	quiet   = flag.Bool("quiet", false, "don't print info at startup")
	prtver  = flag.Bool("version", false, "print version and exit")
	Version = "development"
)

const (
	LTriggerPull  = 1 << 10
	RTriggerPull  = 1 << 11
	LTriggerTouch = 1 << 16
	RTriggerTouch = 1 << 17
)

func main() {
	cfg := NewConfig()
	flag.Parse()

	if flag.NArg() != 0 {
		abort("Positional arguments not supported")
	}

	var cfgch <-chan *Config
	exit := false
	if *prtver {
		fmt.Println(Version)
		exit = true
	}
	if *savecfg != "" {
		err := cfg.Save(*savecfg)
		if err != nil {
			abort(err)
		}
		exit = true
	}
	if exit {
		return
	}

	if *loadcfg != "" {
		err := cfg.Load(*loadcfg)
		if err != nil {
			abort(err)
		}
		cfgch = autoloadconfig(*loadcfg)
	} else {
		fn := "joyster.cfg"
		if err := cfg.Load(fn); err != nil {
			// no error, but reset to default
			cfg = NewConfig()
			fmt.Println(err)
		} else {
			cfgch = autoloadconfig(fn)
		}
	}

	if !*quiet {
		fmt.Println("joyster version:", Version)
		fmt.Println("vJoy version:", vjoy.Version())
		fmt.Println("  Product:       ", vjoy.ProductString())
		fmt.Println("  Manufacturer:  ", vjoy.ManufacturerString())
		fmt.Println("  Serial number: ", vjoy.SerialNumberString())
	}
	d, err := vjoy.Acquire(1)
	if err != nil {
		abort(err)
	}
	defer d.Relinquish()

	var xs xinput.State
	t := &ticker{config: cfg, gamepad: &xs.Gamepad}
	for {
		xinput.GetState(0, &xs)
		t.update(d)
		time.Sleep(time.Duration(cfg.UpdateMicros) * time.Microsecond)
		select {
		case cfg := <-cfgch:
			t.config = cfg
		default:
		}
	}
}

type ticker struct {
	config  *Config
	gamepad *xinput.Gamepad

	x, y, z, rx, ry, rz, u, v float32

	lastb uint32

	rolltoyaw  bool
	triggeryaw bool
	headlook   bool

	view viewaccumulatelogic
}

func (t *ticker) update(d *vjoy.Device) {
	t.updateAxes(d)
	t.updateButtons(d)
	t.updateFlight()

	t.updateDevice(d)
}

func (t *ticker) updateDevice(d *vjoy.Device) {
	d.Axis(vjoy.AxisX).Setf(t.x)
	d.Axis(vjoy.AxisY).Setf(t.y)
	d.Axis(vjoy.AxisZ).Setf(t.z)
	d.Axis(vjoy.AxisRX).Setf(t.rx)
	d.Axis(vjoy.AxisRY).Setf(t.ry)
	d.Axis(vjoy.AxisRZ).Setf(t.rz)
	d.Axis(vjoy.Slider0).Setf(t.u)
	d.Axis(vjoy.Slider1).Setf(t.v)
	d.Update()
}

func (t *ticker) updateFlight() {
	b := uint32(t.gamepad.Buttons)
	if t.config.RollToYaw {
		if (((b ^ t.lastb) & b) & uint32(xinput.LEFT_THUMB)) != 0 {
			t.rolltoyaw = !t.rolltoyaw
			t.triggeryaw = false
		}
	}
	if t.config.HeadLook != nil {
		if (((b ^ t.lastb) & b) & uint32(xinput.RIGHT_THUMB)) != 0 {
			t.headlook = !t.headlook
		}
	}
	if t.config.TriggerAxis != nil {
		if (((b ^ t.lastb) & b) & t.config.TriggerAxis.imask) != 0 {
			t.triggeryaw = !t.triggeryaw
			if t.triggeryaw {
				t.rolltoyaw = false
			}
		}
	}
	t.lastb = b
}

func (t *ticker) updateAxes(d *vjoy.Device) {
	var lt, rt thumbStick
	lt.set(t.gamepad.ThumbLX, t.gamepad.ThumbLY)
	rt.set(t.gamepad.ThumbRX, t.gamepad.ThumbRY)
	if t.config.ThumbCircle {
		lt.circularize()
		rt.circularize()
	}

	lx := float32(axismap(t.config.ThumbLX, lt.xv, lt.xs))
	ly := float32(axismap(t.config.ThumbLY, lt.yv, lt.ys))
	rx := float32(axismap(t.config.ThumbRX, rt.xv, rt.xs))
	ry := float32(axismap(t.config.ThumbRY, rt.yv, rt.ys))
	if t.rolltoyaw {
		t.x = 0
		t.y = ly
		t.z = lx
	} else {
		t.x = lx
		t.y = ly
		t.z = 0
	}
	if t.headlook {
		t.rx, t.ry = 0, 0
		t.u, t.v = t.view.update(t.config, rx, ry)
	} else {
		t.rx, t.ry = rx, ry
		t.u, t.v = t.view.jumpToOrigin(t.config)
	}

	if t.config.LeftTrigger.Axis {
		d.Axis(vjoy.AxisZ).Setuf(float32(t.gamepad.LeftTrigger) / 255)
	}
	if t.config.RightTrigger.Axis {
		d.Axis(vjoy.AxisRZ).Setuf(float32(t.gamepad.RightTrigger) / 255)
	}
}

func (t *ticker) updateButtons(d *vjoy.Device) {
	btns := uint32(t.gamepad.Buttons)
	lv, rv := uint16(t.gamepad.LeftTrigger), uint16(t.gamepad.RightTrigger)
	if t.triggeryaw {
		tac := t.config.TriggerAxis
		lx, rx := lv > tac.breakthreshold, rv > tac.breakthreshold
		if lx == rx {
			if lx {
				btns |= tac.breakmask
			}
		} else {
			lf := triggermap(lv, tac.axisthreshold, tac.Pow)
			rf := triggermap(rv, tac.axisthreshold, tac.Pow)
			t.z += rf - lf
		}
	} else {
		if t.config.LeftTrigger.touch <= lv {
			if t.config.LeftTrigger.pull <= lv {
				btns |= LTriggerPull
			} else {
				btns |= LTriggerTouch
			}
		}
		if t.config.RightTrigger.touch <= rv {
			if t.config.RightTrigger.pull <= rv {
				btns |= RTriggerPull
			} else {
				btns |= RTriggerTouch
			}
		}
	}
	for _, bc := range t.config.Buttons {
		bc.handler.Handle(d, bc, (btns&bc.fmask) == bc.imask)
	}
}

func abort(a ...interface{}) {
	fmt.Println(a...)
	os.Exit(1)
}

func autoloadconfig(fn string) <-chan *Config {
	ch := make(chan *Config)
	if fi, err := os.Stat(fn); err != nil {
		t := fi.ModTime()
		go func() {
			for {
				if fi, err := os.Stat(fn); err != nil && fi.ModTime().After(t) {
					t = fi.ModTime()
					cfg := new(Config)
					if err := cfg.Load(fn); err == nil {
						ch <- cfg
					}
				}
				time.Sleep(time.Second)
			}
		}()
	}
	return ch
}
