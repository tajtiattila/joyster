package parser

import (
	"fmt"
	"testing"
)

var testsrc = []byte(`
block input [gamepad: 0]
block output [vjoy: 1]

set Update=1m TapDelay=200m KeepPushed=250m

# axes

conn output.x [if [or [not fight] [not rolltoyaw]] ls.x 0]
conn output.y ls.y
conn output.z [add [if rolltoyaw ls.x 0] [if [not fight] ta 0]]

conn output.rx [if [not headlooktoggle] rs.x 0]
conn output.ry [if [not headlooktoggle] rs.y 0]
conn output.u headlook.x
conn output.v headlook.y

# buttons

conn output.1 [or abtn ta.break]
conn output.2 bbtn
conn output.3 xbtn
conn output.4 ybtn.1

conn output.5 abtn.2
conn output.6 bbtn.2
conn output.7 xbtn.2
conn output.8 ybtn.2

conn output.9 shift0
conn output.10 shift1

conn output.11 [if plane1 input.buttona off]
conn output.12 [if plane1 input.buttonb off]
conn output.13 [if plane1 input.buttonx off]
conn output.14 [if plane1 input.buttony off]

conn output.15 [if plane2 input.buttona off]
conn output.16 [if plane2 input.buttonb off]
conn output.17 [if plane2 input.buttonx off]
conn output.18 [if plane2 input.buttony off]

conn output.19 [and fight input.ltrigger]
conn output.20 [and fight input.rtrigger]

conn output.21 [if plane3 input.buttona off]
conn output.22 [if plane3 input.buttonb off]
conn output.23 [if plane3 input.buttonx off]
conn output.24 [if plane3 input.buttony off]

# hats

# targeting

# Up: target ahead
# Down: highest threat
# Left/Right: prev/next subsystem
conn output.hat1 [if plane0 input.dpad centre]

# LS + D-Pad: targeting extra

# Up/Down: prev/next hostile
# Left/Right: prev/next ship
conn output.hat2 [if plane1 input.dpad centre]

# RS + D-Pad: toggles (hardpoint, landing gear, flight assist, cargo scoop)

# Up: hardpoints
# Down: landing gear
# Left: flight assist
# Right: cargo scoop
conn output.hat3 [if plane2 input.dpad centre]

# LS + RS + dpad: misc. cycles, stealth

# Up: increase radar range
# Left: cycle fire group next
# Right: silent running
# Down: deploy heatsink
conn output.hat4 [if plane3 input.dpad centre]

# logic
port shift0 input.lbumper
port shift1 input.rbumper

block plane0 [and [not shift0] [not shift1]]
block plane1 [and shift0 [not shift1]]
block plane2 [and [not shift0] shift1]
block plane3 [and shift0 shift1]

block ta [triggeraxis input.lt input.rt: AxisThreshold=0.15 BreakThreshold=0.05 Exp=1.5]

block abtn [multibutton [if plane0 input.buttona off]: 2]

block bbtn [multibutton [if plane0 input.buttonb off]: 2]

block xbtn [multibutton [if plane0 input.buttonx off]: 2]

block ybtn [multibutton [if plane0 input.buttony off]: 2]

block rolltoyaw [toggle input.lthumb]
block fight [toggle input.back]
block headlooktoggle [toggle input.rthumb]

# left stick
block ls0 { x y
	$[deadzone: 0.05]
	$[multiply: 1.25]
	$[curvature: 0.2]
}
conn ls0.x input.lx
conn ls0.y input.ly

block ls1 { x y
	$[dampen: 0.2]
	$[smooth: 0.3]
}
conn ls1.x ls0.x
conn ls1.y ls0.y

block ls [stick [absmin ls0.x ls1.x] [absmin ls0.x ls1.x]]

# right stick
block rs { x y
	[circulardeadzone: 0.05]
	$[multiply: 1.25]
	$[curvature: 0.2]
}
conn rs.x input.rx
conn rs.y input.ry

block headlook [headlook: MovePerSec=2.0 AutoCenterDist=0.2 AutoCenterAccel=0.001 JumpToCenterAccel=0.1]
conn headlook.x [if headlooktoggle rs.x 0]
conn headlook.y [if headlooktoggle rs.y 0]
conn headlook.enable headlooktoggle
`)

func init() {
}

func TestParser(t *testing.T) {
	p := newparser(newtestnamespace())
	if err := p.parse(testsrc); err != nil {
		t.Error(err)
	}
	t.Log("parse ok")
	ctx := p.context
	if err := sort(ctx); err != nil {
		t.Error(err)
	}
	//t.Logf("%#v", p)
}

type testnamespace struct {
	m map[string]Type
}

func (m *testnamespace) GetType(n string) (Type, error) {
	if t, ok := m.m[n]; ok {
		return t, nil
	}
	return nil, fmt.Errorf("unknown type '%s'", n)
}

func (m *testnamespace) add(k Type, names ...string) {
	if m.m == nil {
		m.m = make(map[string]Type)
	}
	for _, n := range names {
		if _, ok := m.m[n]; ok {
			panic("duplicate: " + n)
		}
		m.m[n] = k
	}
}

func newtestnamespace() *testnamespace {
	m := &testnamespace{make(map[string]Type)}
	m.add(kind(bi(""), bo("")), "not", "toggle")
	m.add(kind(bi("1"), bi("2"), bo("")), "and", "or", "xor")
	m.add(&ifblkkind{*kind(bi("cond"), ai("then"), ai("else"))}, "if")

	m.add(kind(si("1"), si("2"), so("")),
		"add", "sub", "mul", "div", "mod", "pow", "min", "max", "absmin", "absmax")
	m.add(kind(si("1"), si("2"), bo("")), "eq", "ne", "lt", "gt", "le", "ge")
	m.add(kind(bi(""), bo(""), bo("1"), bo("2")).arg("NumTaps", "TapDelay", "KeepPushed"), "multibutton")

	m.add(kind(si(""), so("")).arg("Value"), "offset")
	m.add(kind(si(""), so("")).arg("Threshold"), "deadzone")
	m.add(kind(si(""), so("")).arg("Factor"), "multiply")
	m.add(kind(si(""), so("")).arg("Factor"), "curvature")
	m.add(kind(si(""), so("")).arg("Value"), "truncate")
	m.add(kind(si(""), so("")).arg("Value"), "dampen")
	m.add(kind(si(""), so("")).arg("Time"), "smooth")
	m.add(kind(si(""), so("")).arg("Speed").optarg("Rebound", "QuickCenter"), "incremental")

	m.add(kind(si("x"), si("y"), so("x"), so("y")), "stick")
	m.add(kind(si("x"), si("y"), so("x"), so("y")).arg("Factor"), "circlesquare")
	m.add(kind(si("x"), si("y"), so("x"), so("y")).arg("Threshold"), "circulardeadzone")

	m.add(kind(si("left"), si("right"), so(""), bo("break")).
		arg("AxisThreshold", "BreakThreshold", "Exp"), "triggeraxis")
	m.add(kind(si("x"), si("y"), bi("enable"), so("x"), so("y")).
		arg("MovePerSec", "AutoCenterDist", "AutoCenterAccel", "JumpToCenterAccel"), "headlook")

	var vk []testio
	for _, n := range []string{"x", "y", "z", "rx", "ry", "rz", "u", "v"} {
		vk = append(vk, si(n))
	}
	for i := 1; i <= 32; i++ {
		vk = append(vk, bi(fmt.Sprint(i)))
	}
	for i := 1; i <= 4; i++ {
		vk = append(vk, hi(fmt.Sprint("hat", i)))
	}
	m.add(kind(vk...).inopt().arg("device"), "vjoy")

	gk := []testio{ho("dpad")}
	for _, n := range []string{"lx", "ly", "rx", "ry", "lt", "rt"} {
		gk = append(gk, so(n))
	}
	for _, n := range []string{"buttona", "buttonb", "buttonx", "buttony",
		"ltrigger", "rtrigger",
		"lbumper", "rbumper", "lthumb", "rthumb", "back", "start"} {
		gk = append(gk, bo(n))
	}
	m.add(kind(gk...).arg("device"), "gamepad")
	return m
}

type testblkkind struct {
	inames   PortMap
	onames   PortMap
	optinput bool
	args     []karg
}

func kind(vio ...testio) *testblkkind {
	t := new(testblkkind)
	for _, io := range vio {
		if io.input {
			t.inames = append(t.inames, Port{io.name, io.typ})
		} else {
			t.onames = append(t.onames, Port{io.name, io.typ})
		}
	}
	return t
}

func (k *testblkkind) Input() PortMap         { return k.inames }
func (k *testblkkind) Output(PortMap) PortMap { return k.onames }
func (k *testblkkind) MustHaveInput() bool    { return !k.optinput }

func (k *testblkkind) Validate(p Param, c NamedParam) error {
	pr := p.Reader(c)
	for _, a := range k.args {
		if a.opt {
			pr.OptArg(a.name, 0)
		} else {
			pr.Arg(a.name)
		}
	}
	return pr.Err()
}

func (k *testblkkind) inopt() *testblkkind {
	nk := new(testblkkind)
	*nk = *k
	nk.optinput = true
	return nk
}

func (k *testblkkind) arg(v ...string) *testblkkind    { return k.xargs(false, v...) }
func (k *testblkkind) optarg(v ...string) *testblkkind { return k.xargs(true, v...) }

func (k *testblkkind) xargs(opt bool, v ...string) *testblkkind {
	nk := new(testblkkind)
	*nk = *k
	nk.args = make([]karg, len(k.args))
	copy(nk.args, k.args)
	for _, n := range v {
		nk.args = append(nk.args, karg{n, opt})
	}
	return nk
}

type karg struct {
	name string
	opt  bool
}

type ifblkkind struct {
	testblkkind
}

func (k *ifblkkind) Output(im PortMap) PortMap {
	if im == nil {
		im = k.Input()
	}
	th, el := im.Port("then"), im.Port("else")
	if th == el {
		return PortMap{Port{"", th}}
	}
	return PortMap{Port{"", Invalid}}
}

type testsig struct {
}

type testio struct {
	name  string
	input bool
	typ   PortType
}

func bo(name string) testio { return testio{name, false, Bool} }
func so(name string) testio { return testio{name, false, Scalar} }
func ho(name string) testio { return testio{name, false, Hat} }
func ai(name string) testio { return testio{name, true, Any} }
func bi(name string) testio { return testio{name, true, Bool} }
func si(name string) testio { return testio{name, true, Scalar} }
func hi(name string) testio { return testio{name, true, Hat} }
