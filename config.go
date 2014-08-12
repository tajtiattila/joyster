package main

import (
	"bytes"
	"encoding/json"
	"os"
)

type Config struct {
	ShiftButtonIndex int
	ShiftButtonHide  bool
	DontShift        []uint32
	ThumbCircle      bool
	ThumbLX          *AxisConfig
	ThumbLY          *AxisConfig
	ThumbRX          *AxisConfig
	ThumbRY          *AxisConfig
	LeftTrigger      *TriggerConfig
	RightTrigger     *TriggerConfig

	shiftButtonMask uint32
	shiftableMask   uint32
}

func NewConfig() *Config {
	c := &Config{
		ShiftButtonIndex: 9,
		ShiftButtonHide:  true,
		DontShift:        []uint32{7, 8, 17, 18, 19, 20},
		ThumbCircle:      true,
		ThumbLX:          NewAxisConfig(),
		ThumbLY:          NewAxisConfig(),
		ThumbRX:          NewAxisConfig(),
		ThumbRY:          NewAxisConfig(),
		LeftTrigger:      NewTriggerConfig(),
		RightTrigger:     NewTriggerConfig(),
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
	if c.ShiftButtonIndex != 0 {
		c.shiftButtonMask = 1 << uint(c.ShiftButtonIndex-1)
	} else {
		c.shiftButtonMask = 0
	}

	var m uint32
	for _, v := range c.DontShift {
		if 0 < v && v <= 20 {
			m |= 1 << uint(v-1)
		}
	}
	c.shiftableMask = ^m

	c.LeftTrigger.update()
	c.RightTrigger.update()
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
