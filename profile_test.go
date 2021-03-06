package main

import (
	"github.com/tajtiattila/joyster/block"
	_ "github.com/tajtiattila/joyster/block/device/vjoy"
	_ "github.com/tajtiattila/joyster/block/device/xinput"
	_ "github.com/tajtiattila/joyster/block/logic"
	"testing"
)

var testsrc = `
block input [gamepad: 0]
block output [vjoy: 1]

set Update=1000 TapDelay=200m KeepPushed=250m

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
`

func TestLoad(t *testing.T) {
	_, err := block.Parse(testsrc)
	if err != nil {
		t.Error(err)
	}
}
