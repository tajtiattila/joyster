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
	savecfg  = flag.String("save", "", "save default configfile into file specified and exit")
	loadcfg  = flag.String("cfg", "", "configfile to use")
	quiet    = flag.Bool("quiet", false, "don't print info at startup")
	prtver   = flag.Bool("version", false, "print version and exit")
	webgui   = flag.Bool("web", false, "enable web gui")
	addr     = flag.String("addr", ":7489", "web gui address")  // "JY"
	sharedir = flag.String("share", "share", "share directory") // "JY"
	Version  = "development"
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

	if *webgui {
		WebGUI(*addr, *sharedir)
	}

	var xs xinput.State
	t := &ticker{config: cfg, gamepad: &xs.Gamepad, dev: d}
	for {
		xinput.GetState(0, &xs)
		t.update()
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

	lastb uint32

	view viewaccumulatelogic

	dev *vjoy.Device

	st Status
}

func (t *ticker) update() {
	t.updateAxes()
	t.updateButtons()
	t.updateFlight()

	t.updateDevice()
	t.updateListeners()
}

func (t *ticker) updateDevice() {
	d, o := t.dev, &t.st.O
	d.Axis(vjoy.AxisX).Setf(o.X)
	d.Axis(vjoy.AxisY).Setf(o.Y)
	d.Axis(vjoy.AxisZ).Setf(o.Z)
	d.Axis(vjoy.AxisRX).Setf(o.RX)
	d.Axis(vjoy.AxisRY).Setf(o.RY)
	d.Axis(vjoy.AxisRZ).Setf(o.RZ)
	d.Axis(vjoy.Slider0).Setf(o.U)
	d.Axis(vjoy.Slider1).Setf(o.V)
	d.Update()
}

func (t *ticker) updateListeners() {
	DefaultStatusDispatcher.Update(&t.st)
}

func (t *ticker) updateFlight() {
	b := uint32(t.gamepad.Buttons)
	if t.config.RollToYaw {
		if (((b ^ t.lastb) & b) & uint32(xinput.LEFT_THUMB)) != 0 {
			t.st.RollToYaw = !t.st.RollToYaw
			t.st.TriggerYaw = false
		}
	}
	if t.config.HeadLook != nil {
		if (((b ^ t.lastb) & b) & uint32(xinput.RIGHT_THUMB)) != 0 {
			t.st.HeadLook = !t.st.HeadLook
		}
	}
	if t.config.TriggerAxis != nil {
		if (((b ^ t.lastb) & b) & t.config.TriggerAxis.imask) != 0 {
			t.st.TriggerYaw = !t.st.TriggerYaw
			if t.st.TriggerYaw {
				t.st.RollToYaw = false
			}
		}
	}
	t.lastb = b
}

func (t *ticker) updateAxes() {
	var lt, rt thumbStick
	lt.set(t.gamepad.ThumbLX, t.gamepad.ThumbLY)
	rt.set(t.gamepad.ThumbRX, t.gamepad.ThumbRY)
	t.st.I.LX, t.st.I.LY = lt.values32()
	t.st.I.RX, t.st.I.RY = rt.values32()
	if t.config.ThumbCircle {
		lt.circularize()
		rt.circularize()
	}

	lx := float32(axismap(t.config.ThumbLX, lt.xv, lt.xs))
	ly := float32(axismap(t.config.ThumbLY, lt.yv, lt.ys))
	rx := float32(axismap(t.config.ThumbRX, rt.xv, rt.xs))
	ry := float32(axismap(t.config.ThumbRY, rt.yv, rt.ys))
	o := &t.st.O
	if t.st.RollToYaw {
		o.X = 0
		o.Y = ly
		o.Z = lx
	} else {
		o.X = lx
		o.Y = ly
		o.Z = 0
	}
	if t.st.HeadLook {
		o.RX, o.RY = 0, 0
		o.U, o.V = t.view.update(t.config, rx, ry)
	} else {
		o.RX, o.RY = rx, ry
		o.U, o.V = t.view.jumpToOrigin(t.config)
	}

	ltr := float32(t.gamepad.LeftTrigger) / 255
	rtr := float32(t.gamepad.RightTrigger) / 255
	t.st.I.LTrigger, t.st.I.RTrigger = ltr, rtr
	if t.config.LeftTrigger.Axis {
		t.dev.Axis(vjoy.AxisZ).Setuf(ltr)
	}
	if t.config.RightTrigger.Axis {
		t.dev.Axis(vjoy.AxisRZ).Setuf(rtr)
	}
}

func (t *ticker) updateButtons() {
	btns := uint32(t.gamepad.Buttons)
	lv, rv := uint16(t.gamepad.LeftTrigger), uint16(t.gamepad.RightTrigger)
	if t.st.TriggerYaw {
		tac := t.config.TriggerAxis
		lx, rx := lv > tac.breakthreshold, rv > tac.breakthreshold
		if lx == rx {
			if lx {
				btns |= tac.breakmask
			}
		} else {
			lf := triggermap(lv, tac.axisthreshold, tac.Pow)
			rf := triggermap(rv, tac.axisthreshold, tac.Pow)
			t.st.O.Z += rf - lf
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
	t.st.I.Buttons = uint64(btns)
	for _, bc := range t.config.Buttons {
		bc.handler.Handle(t, bc, (btns&bc.fmask) == bc.imask)
	}
}

func (t *ticker) SetButton(idx uint, value bool) {
	t.dev.Button(idx).Set(value)
	if value {
		t.st.O.Buttons |= 1 << idx
	} else {
		t.st.O.Buttons &= ^(1 << idx)
	}
}

func abort(a ...interface{}) {
	fmt.Println(a...)
	os.Exit(1)
}

func autoloadconfig(fn string) <-chan *Config {
	ch := make(chan *Config)
	if fi, err := os.Stat(fn); err == nil {
		t := fi.ModTime()
		go func() {
			for {
				if fi, err := os.Stat(fn); err == nil && fi.ModTime().After(t) {
					t = fi.ModTime()
					cfg := new(Config)
					if err := cfg.Load(fn); err == nil {
						ch <- cfg
						fmt.Println("new config loaded")
					} else {
						fmt.Println("config error:", err)
					}
				}
				time.Sleep(time.Second)
			}
		}()
	} else {
		panic("autoloadcfg " + fn + err.Error())
	}
	return ch
}
