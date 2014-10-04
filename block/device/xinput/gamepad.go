package xinput

import (
	"github.com/tajtiattila/joyster/block"
	xi "github.com/tajtiattila/xinput"
)

func init() {
	block.RegisterParam("gamepad", func(p block.Param) (block.Block, error) {
		return &gamepad{dev: uint(p.OptArg("device", 0))}, nil
	})
}

type gamepad struct {
	dev uint
	xs  xi.State

	a, b, x, y         bool
	start, back        bool
	lbumper, rbumper   bool
	lthumb, rthumb     bool
	ltrigger, rtrigger bool

	lt, rt float64
	lx, ly float64
	rx, ry float64

	dpad int
}

func (p *gamepad) Input() block.InputMap { return nil }
func (p *gamepad) Validate() error       { return nil }

func (p *gamepad) Output() block.OutputMap {
	return block.MapOutput("Device",
		pt("a", &p.a),
		pt("b", &p.b),
		pt("x", &p.x),
		pt("y", &p.y),
		pt("start", &p.start),
		pt("back", &p.back),
		pt("ltrigger", &p.ltrigger),
		pt("rtrigger", &p.rtrigger),
		pt("lbumper", &p.lbumper),
		pt("rbumper", &p.rbumper),
		pt("lthumb", &p.lthumb),
		pt("rthumb", &p.rthumb),
		pt("lx", &p.lx),
		pt("ly", &p.ly),
		pt("rx", &p.rx),
		pt("ry", &p.ry),
		pt("lt", &p.lt),
		pt("rt", &p.rt),
		pt("dpad", &p.dpad),
	)
}

func (p *gamepad) Tick() {
	last := p.xs.PacketNumber
	xi.GetState(p.dev, &p.xs)
	if last == p.xs.PacketNumber {
		return // state unchanged
	}

	xpad := &p.xs.Gamepad

	p.a = (xpad.Buttons & xi.BUTTON_A) != 0
	p.b = (xpad.Buttons & xi.BUTTON_B) != 0
	p.x = (xpad.Buttons & xi.BUTTON_X) != 0
	p.y = (xpad.Buttons & xi.BUTTON_Y) != 0

	p.start = (xpad.Buttons & xi.START) != 0
	p.back = (xpad.Buttons & xi.BACK) != 0

	p.lbumper = (xpad.Buttons & xi.LEFT_SHOULDER) != 0
	p.rbumper = (xpad.Buttons & xi.RIGHT_SHOULDER) != 0

	p.lthumb = (xpad.Buttons & xi.LEFT_THUMB) != 0
	p.rthumb = (xpad.Buttons & xi.RIGHT_THUMB) != 0

	p.ltrigger = xpad.LeftTrigger != 0
	p.rtrigger = xpad.RightTrigger != 0

	d := int(xpad.Buttons & (xi.DPAD_UP | xi.DPAD_DOWN | xi.DPAD_LEFT | xi.DPAD_RIGHT))
	p.dpad = dpadmap[d]

	p.lt = uint8scalar(xpad.LeftTrigger)
	p.rt = uint8scalar(xpad.RightTrigger)

	p.lx = int16scalar(xpad.ThumbLX)
	p.ly = int16scalar(xpad.ThumbLY)
	p.rx = int16scalar(xpad.ThumbRX)
	p.ry = int16scalar(xpad.ThumbRY)
}

func uint8scalar(v uint8) float64 {
	return float64(v) / 255
}

func int16scalar(v int16) float64 {
	if v < 0 {
		return float64(v) / 0x8000
	} else {
		return float64(v) / 0x7fff
	}
	return 0 // not reached
}

func pt(n string, v interface{}) block.MapDecl { return block.MapDecl{n, v} }

var dpadmap []int

func init() {
	dpadmap = make([]int, 16)
	for i := 0; i < 16; i++ {
		b := uint16(i)
		v := 0
		if (b & xi.DPAD_UP) != 0 {
			v += block.HatNorth
		}
		if (b & xi.DPAD_DOWN) != 0 {
			v += block.HatSouth
		}
		if (b & xi.DPAD_RIGHT) != 0 {
			v += block.HatEast
		}
		if (b & xi.DPAD_LEFT) != 0 {
			v += block.HatWest
		}
		dpadmap[i] = v
	}
}
