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

	ThumbCircle bool

	ThumbLX      *AxisConfig
	ThumbLY      *AxisConfig
	ThumbRX      *AxisConfig
	ThumbRY      *AxisConfig
	LeftTrigger  *TriggerConfig
	RightTrigger *TriggerConfig
	Buttons      []*ButtonConfig
}

func NewConfig() *Config {
	var btn buttonConfigBuilder
	c := &Config{
		ThumbCircle:      true,
		ThumbLX:          NewAxisConfig(),
		ThumbLY:          NewAxisConfig(),
		ThumbRX:          NewAxisConfig(),
		ThumbRY:          NewAxisConfig(),
		LeftTrigger:      NewTriggerConfig(),
		RightTrigger:     NewTriggerConfig(),
		UpdateMicros:     1000,   // 1ms
		TapDelayMicros:   300000, // 0.3s
		KeepPushedMicros: 500000, // 0.05s
		Buttons: []*ButtonConfig{
			btn.NewSimple(1, 13),
			btn.NewSimple(2, 14),
			btn.NewSimple(3, 15),
			btn.NewSimple(4, 16),
			btn.NewSimple(5, 9, 13),
			btn.NewMulti(6, 13, 2),
			//btn.NewMulti(7, 14, 2),
		},
	}
	c.update()
	return c
}

func (c *Config) Load(fn string) error {
	f, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(c)
	c.update()
	return err
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

func (c *Config) update() {
	if c.UpdateMicros == 0 {
		c.UpdateMicros = 1000
	}

	c.LeftTrigger.update()
	c.RightTrigger.update()

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

type AxisConfig struct {
	Min float64 // 0 will be reported under this value (0..1)
	Max float64 // 1 will be reported above this value (0..1)
	Pow float64 // smoothen small movements in input: raise to this power (1..∞)
}

func NewAxisConfig() *AxisConfig {
	return &AxisConfig{Min: 0.1, Max: 0.95, Pow: 2.0}
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
	if t.pull < t.touch {
		t.touch = t.pull
	}
}

type ButtonConfig struct {
	Output  int    // output
	Inputs  []int  // these sources must be pressed together to trigger
	Multi   uint   // press this many times
	imask   uint32 // input mask created from Inputs
	fmask   uint32 // filtering mask when button is used in multiple configs
	handler ButtonHandler
}

type buttonConfigBuilder struct {
	output int
}

func (b *buttonConfigBuilder) NewSimple(output int, inputs ...int) *ButtonConfig {
	b.output++
	return &ButtonConfig{Output: b.output, Inputs: inputs}
}

func (b *buttonConfigBuilder) NewMulti(output, input int, multi uint) *ButtonConfig {
	b.output++
	return &ButtonConfig{Output: b.output, Inputs: []int{input}, Multi: multi}
}

func (b *ButtonConfig) update(c *Config) {
	var m uint32
	for _, v := range b.Inputs {
		if 0 < v && v <= 16 {
			m |= 1 << uint(v-1)
		}
	}
	b.imask = m
	if b.Multi != 0 {
		b.handler = &MultiButton{
			taptick: c.tapDelayTicks(),
			needtap: b.Multi,
			pushlen: c.keepPushedTicks(),
		}
	} else {
		b.handler = &SimpleButton{}
	}
}
