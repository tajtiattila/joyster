package vjoy

import (
	"fmt"
	"github.com/tajtiattila/joyster/block"
	vj "github.com/tajtiattila/vjoy"
)

func init() {
	block.RegisterParam("vjoy", func(p block.Param) (block.Block, error) {
		if p == block.ProtoParam {
			return new(vjoyproto), nil
		}
		return newVjoyBlock(int(p.OptArg("Device", 1)))
	})
}

type devnode struct {
	device   *vj.Device
	usecount int
}

var devices []devnode

func openvjoy(idev int) (*vj.Device, error) {
	if idev < len(devices) {
		node := &devices[idev]
		if node.device != nil {
			node.usecount++
			return node.device, nil
		}
	}
	dev, err := vj.Acquire(uint(idev))
	if err != nil {
		return nil, err
	}
	if len(devices) <= idev {
		n := make([]devnode, idev+1, 3*idev+1)
		copy(n, devices)
		devices = n
	}
	devices[idev] = devnode{dev, 1}
	return dev, nil
}

func closevjoy(idev int) error {
	if idev < len(devices) {
		node := &devices[idev]
		if node.usecount != 0 && node.device != nil {
			node.usecount--
			if node.usecount == 0 {
				node.device.Relinquish()
				node.device = nil
			}
			return nil
		}
	}
	return fmt.Errorf("device %d not open", idev)
}

var axes = []string{"x", "y", "z", "rx", "ry", "rz", "u", "v"}

type vjoyblk struct {
	idev int
	dev  *vj.Device

	axes    []*axis
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

func newVjoyBlock(idev int) (block.Block, error) {
	d, err := openvjoy(idev)
	if err != nil {
		return nil, err
	}
	blk := &vjoyblk{idev: idev, dev: d}

	zero := new(float64)
	blk.axes = []*axis{
		{zero, d.Axis(vj.AxisX)},
		{zero, d.Axis(vj.AxisY)},
		{zero, d.Axis(vj.AxisZ)},
		{zero, d.Axis(vj.AxisRX)},
		{zero, d.Axis(vj.AxisRY)},
		{zero, d.Axis(vj.AxisRZ)},
		{zero, d.Axis(vj.Slider0)},
		{zero, d.Axis(vj.Slider1)},
	}

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
	var decl []block.MapDecl
	for i, a := range v.axes {
		decl = append(decl, pt(axes[i], &a.v))
	}
	for i, h := range v.hats {
		decl = append(decl, pt(fmt.Sprint("hat", i+1), &h.v))
	}
	for i, b := range v.buttons {
		decl = append(decl, pt(fmt.Sprint(i+1), &b.v))
	}
	return block.MapInput("vjoy", decl...)
}

func (v *vjoyblk) Output() block.OutputMap { return nil }
func (v *vjoyblk) Validate() error         { return nil }
func (v *vjoyblk) Close() error {
	return closevjoy(v.idev)
}

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

type vjoyproto struct{}

func (v *vjoyproto) Output() block.OutputMap { return nil }
func (v *vjoyproto) Validate() error         { return nil }
func (v *vjoyproto) Input() block.InputMap {
	fp, hp, bp := new(float64), new(int), new(bool)
	var decl []block.MapDecl
	for _, n := range axes {
		decl = append(decl, pt(n, &fp))
	}
	for i := 0; i < 4; i++ {
		decl = append(decl, pt(fmt.Sprint("hat", i+1), &hp))
	}
	for i := 0; i < 32; i++ {
		decl = append(decl, pt(fmt.Sprint(i+1), &bp))
	}
	return block.MapInput("vjoyproto", decl...)
}

var hatmap []vj.HatState

func init() {
	hatmap = make([]vj.HatState, block.HatMax)
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

func pt(n string, v interface{}) block.MapDecl { return block.MapDecl{n, v} }
