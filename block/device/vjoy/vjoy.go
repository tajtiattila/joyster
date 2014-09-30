package vjoy

import (
	"fmt"
	"github.com/tajtiattila/joyster/block"
	vj "github.com/tajtiattila/vjoy"
)

func init() {
	block.RegisterParam("vjoy", func(p block.Param) (block.Block, error) {
		return newVjoyBlock(uint(p.Arg("device")))
	})
}

var axes = []string{"x", "y", "z", "rx", "ry", "rz", "u", "v"}

type vjoyblk struct {
	dev *vj.Device

	axes    map[string]*axis
	buttons []*btn
	hats    []*hat
}

type axis struct {
	v *float64
	p *vj.Axis
}

type hat struct {
	v *int
	p vj.Hat
}

type btn struct {
	v *bool
	p *vj.Button
}

func newVjoyBlock(idev uint) (block.Block, error) {
	d, err := vj.Acquire(idev)
	if err != nil {
		return nil, err
	}
	blk := &vjoyblk{dev: d, axes: make(map[string]*axis)}

	zero := new(float64)
	blk.axes["x"] = &axis{zero, d.Axis(vj.AxisX)}
	blk.axes["y"] = &axis{zero, d.Axis(vj.AxisY)}
	blk.axes["z"] = &axis{zero, d.Axis(vj.AxisZ)}
	blk.axes["rx"] = &axis{zero, d.Axis(vj.AxisRX)}
	blk.axes["ry"] = &axis{zero, d.Axis(vj.AxisRY)}
	blk.axes["rz"] = &axis{zero, d.Axis(vj.AxisRZ)}
	blk.axes["u"] = &axis{zero, d.Axis(vj.Slider0)}
	blk.axes["v"] = &axis{zero, d.Axis(vj.Slider1)}

	off := new(bool)
	for i := 0; i < 32; i++ {
		blk.buttons = append(blk.buttons, &btn{off, d.Button(uint(i))})
	}

	centre := new(int)
	for i := 0; i < 4; i++ {
		blk.hats = append(blk.hats, &hat{centre, d.Hat(i)})
	}

	return blk, nil
}

func (v *vjoyblk) Input() block.InputMap {
	m := make(map[string]interface{})
	for n, a := range v.axes {
		m[n] = &(a.v)
	}
	for i, h := range v.hats {
		m[fmt.Sprint("hat", i+1)] = &(h.v)
	}
	for i, b := range v.buttons {
		m[fmt.Sprint(i+1)] = &(b.v)
	}
	return block.MapInput("vjoy", m)
}

func (v *vjoyblk) Output() block.OutputMap { return nil }

func (v *vjoyblk) Tick() {
	for _, a := range v.axes {
		a.p.Setf(float32(*a.v))
	}
	for _, h := range v.hats {
		h.p.SetDiscrete(hatmap[*h.v&block.HatMask])
	}
	for _, b := range v.buttons {
		b.p.Set(*b.v)
	}
	v.dev.Update()
}

var hatmap []vj.HatState

func init() {
	hatmap := make([]vj.HatState, block.HatMax)
	for i := 0; i < block.HatMax; i++ {
		var hs vj.HatState
		switch {
		case (i & block.HatNorth) != 0:
			hs = vj.HatN
		case (i & block.HatSouth) != 0:
			hs = vj.HatS
		case (i & block.HatEast) != 0:
			hs = vj.HatE
		case (i & block.HatWest) != 0:
			hs = vj.HatW
		default:
			hs = vj.HatOff
		}
		hatmap[i] = hs
	}
}
