package main

import (
	"bytes"
	"encoding/json"
	"os"
)

type Config struct {
	UpdateMicros     uint
	TapDelayMicros   uint
	KeepPushedMicros uint

	RollToYaw   bool               // use left thumbstich to switch ThumbLX -> XAxis/ZAxis
	TriggerAxis *TriggerAxisConfig // convert triggers to yaw/break
	HeadLook    *HeadLookConfig

	LeftStick  []interface{}
	RightStick []interface{}

	LeftTrigger  *TriggerConfig
	RightTrigger *TriggerConfig
	Buttons      []*ButtonConfig

	leftStickLogic  StickLogic
	rightStickLogic StickLogic
}

func NewConfig() *Config {
	c := &Config{
		LeftStick:        DefaultStickFilter,
		RightStick:       DefaultStickFilter,
		LeftTrigger:      NewTriggerConfig(),
		RightTrigger:     NewTriggerConfig(),
		UpdateMicros:     1000,   // 1ms
		TapDelayMicros:   300000, // 0.3s
		KeepPushedMicros: 500000, // 0.05s
		Buttons: []*ButtonConfig{
			newDouble(1, 11, 13),
			newDouble(2, 12, 14),
			newSimple(3, 15),
			newSimple(4, 16),
			newSimple(5, 9, 13),
			newSimple(6, 6),
		},
		RollToYaw: true,
		TriggerAxis: &TriggerAxisConfig{
			Switch:         []uint{6},
			BreakButton:    13,
			AxisThreshold:  0.15,
			BreakThreshold: 0.05,
			Pow:            1.5,
		},
		HeadLook: &HeadLookConfig{
			MovePerSec:        2.0,
			AutoCenterDist:    0.2,
			AutoCenterAccel:   0.0001,
			JumpToCenterAccel: 0.1,
		},
	}
	if err := c.update(); err != nil {
		println(err.Error())
		panic("defaultConfigError")
	}
	return c
}

func (c *Config) Load(fn string) error {
	f, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer f.Close()
	if err = json.NewDecoder(f).Decode(c); err != nil {
		return err
	}
	return c.update()
}

func (c *Config) Save(fn string) error {
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err = json.Indent(&buf, data, "", "  "); err != nil {
		return err
	}
	buf.WriteByte('\n')
	f, err := os.Create(fn)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(buf.Bytes())
	return err
}

func (c *Config) update() error {
	if c.UpdateMicros == 0 {
		c.UpdateMicros = 1000
	}

	c.LeftTrigger.update()
	c.RightTrigger.update()

	var err error
	if c.leftStickLogic, err = c.stickLogic(c.LeftStick); err != nil {
		return err
	}
	if c.rightStickLogic, err = c.stickLogic(c.RightStick); err != nil {
		return err
	}

	for _, b := range c.Buttons {
		b.update(c)
	}

	for i, bi := range c.Buttons {
		bi.fmask = bi.imask
		for j, bj := range c.Buttons {
			if i == j {
				continue
			}
			if (bi.imask & bj.imask) != 0 {
				bi.fmask |= bj.imask
			}
		}
	}

	if c.TriggerAxis != nil {
		c.TriggerAxis.update()
	}
	if c.HeadLook != nil {
		c.HeadLook.update(c)
	}
	return nil
}

func (c *Config) stickLogic(fcv []interface{}) (StickLogic, error) {
	var fv []StickLogic
	for _, fc := range fcv {
		f, err := NewStickLogic(c, fc)
		if err != nil {
			return nil, err
		}
		fv = append(fv, f)
	}
	return StickFunc(func(sp *StickPos) {
		for _, f := range fv {
			sp.Apply(f)
		}
	}), nil
}

func (c *Config) tapDelayTicks() uint {
	n := c.TapDelayMicros / c.UpdateMicros
	if n == 0 {
		n = 1
	}
	return n
}

func (c *Config) keepPushedTicks() uint {
	n := c.KeepPushedMicros / c.UpdateMicros
	if n == 0 {
		n = 1
	}
	return n
}

var DefaultStickFilter = []interface{}{
	[]interface{}{"deadzone", 0.1},
	[]interface{}{"multiplier", 1.25},
	[]interface{}{"curvature", 1.1},
}

type TriggerConfig struct {
	TouchTreshold float64
	PullTreshold  float64
	Axis          bool
	touch         uint16
	pull          uint16
}

func NewTriggerConfig() *TriggerConfig {
	t := &TriggerConfig{
		TouchTreshold: 0.15,
		PullTreshold:  0.9,
		Axis:          false,
	}
	t.update()
	return t
}

func (t *TriggerConfig) update() {
	if t.TouchTreshold <= 1.0 {
		t.touch = uint16(t.TouchTreshold * 255)
	} else {
		t.touch = 1000
	}
	if t.PullTreshold <= 1.0 {
		t.pull = uint16(t.PullTreshold * 255)
	} else {
		t.pull = 1000
	}
	if t.pull == 0 {
		t.pull = 1
	}
	if t.touch == 0 {
		t.touch = 1
	}
	if t.pull < t.touch {
		t.touch = t.pull
	}
}

type ButtonConfig struct {
	Output  uint   // output
	Double  uint   // output for doubleclick
	Inputs  []uint // these sources must be pressed together to trigger
	imask   uint32 // input mask created from Inputs
	fmask   uint32 // filtering mask when button is used in multiple configs
	handler ButtonHandler
}

func newSimple(output uint, inputs ...uint) *ButtonConfig {
	return &ButtonConfig{Output: output, Inputs: inputs}
}

func newDouble(output, double, input uint) *ButtonConfig {
	return &ButtonConfig{Output: output, Double: double, Inputs: []uint{input}}
}

func (b *ButtonConfig) update(c *Config) {
	b.imask = inputmask(b.Inputs)
	if b.Double != 0 {
		b.handler = &MultiButton{
			taptick: c.tapDelayTicks(),
			needtap: 2,
			pushlen: c.keepPushedTicks(),
		}
	} else {
		b.handler = &SimpleButton{}
	}
}

type TriggerAxisConfig struct {
	Switch         []uint
	BreakButton    uint
	BreakThreshold float64
	AxisThreshold  float64
	Pow            float64

	breakthreshold uint16
	axisthreshold  uint16
	imask          uint32
	breakmask      uint32
}

func (t *TriggerAxisConfig) update() {
	t.imask = inputmask(t.Switch)
	t.axisthreshold = uint16(255 * t.AxisThreshold)
	t.breakthreshold = uint16(255 * t.BreakThreshold)
	if t.breakthreshold == 0 {
		t.breakthreshold = 1
	}
	if t.BreakButton != 0 {
		t.breakmask = 1 << (t.BreakButton - 1)
	} else {
		t.breakmask = 0
	}
}

type HeadLookConfig struct {
	MovePerSec        float64
	AutoCenterDist    float64
	AutoCenterAccel   float64
	JumpToCenterAccel float64

	movepertick float64
	acapertick  float64
	jumppertick float64
}

func (v *HeadLookConfig) update(c *Config) {
	tickpersec := 1e6 / float64(c.UpdateMicros)
	v.movepertick = v.MovePerSec / tickpersec
	v.acapertick = v.AutoCenterAccel / tickpersec
	v.jumppertick = v.JumpToCenterAccel / tickpersec
	if v.acapertick <= 0.0 {
		v.acapertick = 1
	}
	if v.jumppertick <= 0.0 {
		v.jumppertick = 1
	}
}

func inputmask(iv []uint) uint32 {
	var m uint32
	for _, v := range iv {
		if 0 < v && v < 32 {
			m |= 1 << uint(v-1)
		}
	}
	return m
}
