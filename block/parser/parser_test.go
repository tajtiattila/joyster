package parser

import (
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

conn output.rx [if [not headlooktoggle] rs.x]
conn output.ry [if [not headlooktoggle] rs.y]
conn output.u headlook.x
conn output.v headlook.y

# buttons

conn output.1 [or abtn ta.break]
conn output.2 bbtn
conn output.3 xbtn
conn output.4 ybtn.1

conn output.5 abtn.double
conn output.6 bbtn.double
conn output.7 xbtn.double
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

conn output.21 [if plane3 input.buttona off]
conn output.22 [if plane3 input.buttonb off]
conn output.23 [if plane3 input.buttonx off]
conn output.24 [if plane3 input.buttony off]

# hats

# targeting

# Up: target ahead
# Down: highest threat
# Left/Right: prev/next subsystem
conn output.hat1 [if plane0 input.dpad hatoff]

# LS + D-Pad: targeting extra

# Up/Down: prev/next hostile
# Left/Right: prev/next ship
conn output.hat2 [if plane1 input.dpad hatoff]

# RS + D-Pad: toggles (hardpoint, landing gear, flight assist, cargo scoop)

# Up: hardpoints
# Down: landing gear
# Left: flight assist
# Right: cargo scoop
conn output.hat3 [if plane2 input.dpad hatoff]

# LS + RS + dpad: misc. cycles, stealth

# Up: increase radar range
# Left: cycle fire group next
# Right: silent running
# Down: deploy heatsink
conn output.hat4 [if plane3 input.dpad hatoff]

# logic
block shift0 input.lbumper
block shift1 input.rbumper

block plane0 [and [not shift0] [not shift1]]
block plane1 [and shift0 [not shift1]]
block plane2 [and [not shift0] shift1]
block plane3 [and shift0 shift1]

block ta [triggeraxis input.ltrigger input.rtrigger]

block abtn [multibutton [if plane0 input.buttona off]: 2]

block bbtn [multibutton [if plane0 input.buttonb off]: 2]

block xbtn [multibutton [if plane0 input.buttonx off]: 2]

block ybtn [tapmultibutton [if plane0 input.buttony off]: 2]

block rolltoyaw [toggle input.lthumb]
block fight [toggle input.back]
block headlooktoggle [toggle input.rthumb]

# left stick
block ls0 {
	[xy deadzone: 0.05]
	[xy multiply: 1.25]
	[xy curvature: 0.2]
}
conn ls0.x input.lx
conn ls0.y input.ly

block ls1 {
	[xy dampen: 0.2]
	[xy smooth: 0.3]
}
conn ls1.x ls0.x
conn ls1.y ls0.y

block ls [stick [absmin ls0.x ls1.x] [absmin ls0.x ls1.x]]

# right stick
block rs {
	[circulardeadzone: 0.05]
	[xy multiply: 1.25]
	[xy curvature: 0.2]
}
conn rs.x input.rx
conn rs.y input.ry

block headlook [headlook: MovePerSec=2.0 AutoCenterDist=0.2 AutoCenterAccel=0.001 JumpToCenterAccel=0.1]
conn headlook.x [if headlooktoggle rs.x 0]
conn headlook.y [if headlooktoggle rs.y 0]
`)

func TestParser(t *testing.T) {
	p := newparser(testsrc)
	if err := p.parse(); err != nil {
		t.Error(err)
	}
	t.Logf("%#v", p)
}
