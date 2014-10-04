
Joyster
=======

XBOX compatible gamepad remapper using [vJoy][] and XInput.

Features
--------

* Set axis deadzones
* Remap axes to allow better fine grained control (see "Pow")
* Triggers as rudder and break (toggle using user defined button)
* Roll-to-yaw mode (toggle using left thumb)
* Custom auto centering right stick headlook mode (toggle using right thumb)
* Shift state support
* Double tapped buttons
* Website to display internal states (eg. on your phone while playing)

All features are optional, they may be turned on and off individually.

Requirements
------------

[vJoy]: http://vjoystick.sourceforge.net

[vJoy][]: http://vjoystick.sourceforge.net
XInput: should be available on any modern Windows system

Tested only with vJoy 2.0.4 on Windows 7 64-bit.

Operation
---------

The operation principle of joyster is based on logic blocks connected to each other.
Blocks can have any number of inputs and outputs. Some block types have parameters
to fine tune their behavior. The config file defines what block
types are created, and how are they connected to each other. This set of blocks is
called the profile, which is run `Update` times per second.

The inputs and outputs are called ports. Each port has a type, which can be:

* Numeric ports are for axes. Uses floating point numbers in the range -1..1. Specific ports,
  such as the triggers on the gamepad may provide only non-negative values.
* Boolean ports represend buttons and logical flags. They can be either on or off.
* Hat ports represent directional pads or hats. Internally they are 4-bit numbers, so all possible
  combinations of the four components (north, east, south, west) are supported.

Configuration
-------------

The config file is a set of block, port and connection specifications.

Lines starting with a '#' are comments.

`block` creates a named block to be referred from other block and connection specifications.

`port` creates a new name for an existing input or output port.

`conn` connects an input port to an output port.

`set` sets global configuration parameters, and defines defaults for blocks.

Blocks and single-port inputs can be defined using:

	[blocktype input1 input2 .... : parameters]
	block blockname [blocktype input1 input2 .... : parameters]

Furthermore, a chained group of blocks can be defined using curly braces. A
block created this way will have input and output ports matching portnames. The
input of the group is the first, the output is the last block.

	block groupname { portnames
		[block1 ...]
		[block2 ...]
	}

Ports of blocks are referred to using a period between the block and port names. Ports of
blocks having a single input or output have no name, so they are referred to simply
using the block name with no period and port name following.

Ports and connection are defined using the following syntax:

	port portname blockname.portname
	conn portname sourceblockname.portname
	conn blockname.port sourceblockname.portname

Examples:

	port shift gamepad.lthumb
	conn vjoy.x gamepad.rx
	conn vjoy.1 [and gamepad.1 gamepad.lbumper]

Parameters are always floating point values. Values support some SI prefixes. Block parameters
are specified after a colon, and can be either a named (`Param1=0.5 Param2=1`) or positional
argument list (`0.5 1`).

The config file is parsed according to the following syntax pseudo-specification.

	stmt =
		'block' name blockspec
		'port' name block ['.' spec]
		'conn' portspec portspec
		'set' namedarglist
	arglist = posarglist | namedarglist
	posarglist =
		value [' ' value]*
	value =
		digit* [
	namedarglist =
		name '=' value [' ' value]*
	portspecs = portspec [' ' portspec]*
	portspec =
		name
		name.selector
	blockspec =
		plugvalue
		portspec
		newblockspec
		{ newblockspec [newblockspec]* }
	newblockspec =
		'[' blocktype [portspecs] [':' arglist] ']'
		'$' '[' blocktype [':' arglist] ']'
		'{' inputnames [blockspec]* '}'
	plugvalue =
		on | off | true | false
		digits* ['.' digits*] ['k' | 'm' | 'u' | 'μ' | 'n']
		centre | north | south | east | west

Block types
-----------

A list of supported block types follows.

Some blocks have parameters. Currently only floating point values are supported. The actual unit of parameter values
typically fall into one of the following categories:

* relative axis value
* time in seconds (fractions are supported)
* axis value per second
* boolean flag (nonzero meaning true)

# Device blocks

Device blocks typically have only either inputs or outputs.

## vjoy

`vjoy` creates a [vJoy] output block. The single optional `Device` parameter specifies
which one will be used. The default is 1 (first vJoy device).

* Axes: `x` `y` `z` `rx` `ry` `rz` `u` `v`
* Buttons: `1` .. `32`
* Hats: `hat1` .. `hat4`

The `vjoy` block supports 4-way discrete hats only, so only one of the input components will be used. Inputs
representing diagonals will yield their vertical (north or south) component.

## gamepad

`gamepad` creates an XBOX gamepad input block. The single optional `Device` parameter specifies
which one will be used. The default is 0 (first controller).

* Axes: `lx` `ly` `rx` `ry` `lt` `rt`
* Buttons: `a` `b` `x` `y` `ltrigger` `rtrigger` `lbumper` `rbumper` `lthumb` `rthumb` `start` `back`
* Hats: `dpad`

Triggers are provided both as buttons (`ltrigger` and `rtrigger`) and also as axes (`lt` and `rt`).

# logic blocks

Logic blocks process input. They may or may not have an internal state. Pure functional blocks like `and`
process their input directly to yield the output, while a block with state like `toggle` output value based on
the internal state which may be adjusted with input.

# Simple blocks

Simple blocks have only input and output, but no parameters.

| Name | Input | Output | Category |
|------|-------|--------|----------|
| `not` |  bool |   bool  | logical not |
| `and` `or` `xor`|  2x bool|   bool | logical operators |
| `lt` `gt` `le` `ge` |  2x axis | bool | comparison |
| `eq` `ne` |  2x axis | bool | equality (see also `xeq` and `xne`) |
| `add` `sub` `mul` `div` `mod` `pow` | 2x axis | axis | math |
| `min` `max` | 2x axis | axis | select smaller/larger input |
| `absmin` `absmax` | 2x axis | axis | select input with smaller/larger abs. value |
| `if` | bool, 2x axis | axis | select input based on bool condition |
| `hatadd` `hatsub` `hatxor` | 2x hat | hat | combine/subtract/flip hat values |

# Blocks with parameters

`xeq` and `xne` are similar to `eq` and `ne`, but uses the `Range` parameter to decide
it the values are close enough, rather than using exact comparison that may be inaccurate
because of floating point rounding errors.

# Blocks with state

## toggle

`toggle` outputs a fixed bool value that is toggled when the input signal is changed from `off` to `on`.

## headlook

`headlook` is for incremental head look behavior with optional snap to centre.

	block headlook [headlook: MovePerSec=2.0 AutoCenterDist=0.2 AutoCenterAccel=0.001 JumpToCenterAccel=0.1]
	conn headlook.x [if headlooktoggle rs.x 0]
	conn headlook.y [if headlooktoggle rs.y 0]
	conn headlook.enable headlooktoggle

## doublebutton

`doublebutton` defines a block that can be "double clicked". It provides two
outputs, one unnamed representing the input itself, and `double` which is set
if two successive inputs came within `TapDelay` seconds. `double` will be set
to `on` for `KeepPushed` seconds. Only one of the outputs will ever be set to
`on`, that is, a "double click" forces the unnamed output to be `off` for the
time being itself is pushed.

## multibutton

`multibutton` is a push counter. It counts how many times a button was pressed
with at most `TapDelay` seconds between button pushes, and sets the
corresponding numbered output to `on` for `KeepPushed` seconds. Only one of the
outputs will be set, meaning two quick pushes will yield an output on `2` only,
but not on `1`.

This means `multibutton` has to wait `TapDelay` seconds after the last button
press to know which output must be set, therefore a `multibutton` with two
outputs is differen than `doublebutton` in this regard.

# Special blocks

## pedals

`pedals` takes two axis values and turn them into a single combined axis. In addition to that a boolean
flag will be set if both inputs are in use. The axis output is always zero if break is `on`. Parameters:

| Parameter | Description |
|-----------|-------------|
| AxisThreshold | only use input above this treshold for axis output |
| BreakThreshold | only use input above this treshold for break output |
| Exp | exponent to use on axis output (default is 1 - linear) |

Example assuming `z` to be the rudder, and button `1` to be the break:

	block triggerYaw [pedal gamepad.lt gamepad.rt: AxisThreshold=0.15 BreakThreshold=0.05 Exp=1.5]
	conn output.z triggerYaw
	conn output.1 triggerYaw.break

## hatelem and makehat

`makehat` combines four bool inputs into a hat.

`hatelem` decomposes a hat into four distinct bool outputs.

The four boolean inputs are named `n`, `s`, `e` and `w` after the cardinal directions.

Example:

	block dpaddir [hatelem gamepad.dpad]
	port dpadup    dpaddir.n
	port dpaddown  dpaddir.s
	port dpadleft  dpaddir.w
	port dpadright dpaddir.e

	block buttonshat [makehat]
	conn buttonshat.n gamepad.y
	conn buttonshat.s gamepad.a
	conn buttonshat.w gamepad.x
	conn buttonshat.e gamepad.b

## combo

`combo` multiplexes a single hat input to create four hat outputs. An output is triggered
if any of the hats are pressed twice within `TapDelay` seconds. The first input selects
one of the outputs `n`, `s`, `e` and `w`, the second tells what value it should be set for
`KeepPushed` seconds. The unnamed output will be set for `KeepPushed` seconds if only
one press happens within `TapDelay` seconds. `combo` will output cardinal directions only,
and requires that the hat is released before moving on.

# Stick filters

Stick filters have exactly two inputs and two outputs. Both inputs and outputs
are named `x` and `y`. A stick filter operates on their values as if they were vectors.

## circlesquare

`circlesquare` converts the `x` and `y` axis positions in a circle into vectors
on a square. This lets gamepad stick vectors constrained to be within the range
less than or equal to one around the center to reach the extreme corner
coordinates. That is, an input of x,y=√2,√2 yields x,y=1,1. The optional
parameter `Factor` can be used to reduce the effect of this block: 0 yields
output is same as input, 1 (default) yields full effect.

Using this block is not recommended, `multiply` on the two axes tipically
porduce more intuitive behaviour.

## circulardeadzone

`circulardeadzone` are used for inputs representing movement. Vectors around
the centre closer than `Threshold` will report zero. Output for vectors outside
the circle defined by `Threshold` will have the same direction as the input,
with magnitude reduced by `Threshold`.

# Axis filters

## deadzone

`deadzone` reduces the absolute input value with a constant `Threshold`, and
outputs either zero if the reduced value is negative, or the reduced value with
the sign of the input. The output is multiplied by 1/(1-`Threshold`) so that
input of 1 or -1 yields the same output.

## offset

`offset` adds the constant parameter `Value` to its input.

## multiply

`multiply` multiplies the input with constant parameter `Factor`.

## curvature

`curvature` applies the power function with exponent 2 ** `Factor` to the
input, making the input near the center less responsible, thus more precise
(`Factor`: 0 - linear, positive: exponential).  The output will have the same
sign as the input.

## truncate

`truncate` inputs with magniture above `Value` to be equal to `Value`. The sign
of the input is preserved.

## dampen

`dampen` constraints the input to have a maximum change `Value` per second. In
other words its output chases it inputs with a maximum speed of `Value`.

## smooth

`smooth` accumulates input over the specified amount of `Time`, and yields the
average value.

## incremental

`incremental` implements a logic where the input is used to adjust an otherwise
fixed internal value.  The output is adjusted by input multiplied by `Speed`
per second.  If `Rebound` is nonzero, then the output will converge to zero by
`Rebound` per second, if input is zero. If `Quickcenter` is nonzero, then an
input with opposite sign compared to that of the internal value will set the
internal value back to zero immediately.

Example config
--------------

	block input [gamepad]
	block output [vjoy]

	set Update=1000 TapDelay=200m KeepPushed=250m

	# set up bumper as shift
	port shift input.lbumper
	block plane0 [and [not shift]]
	block plane1 [and shift]

	block abtn [doublebutton [if plane0 input.buttona off]]

	# 11-14: unshifted buttons
	conn output.11 abtn
	conn output.12 [if plane0 input.buttonb off]
	conn output.13 [if plane0 input.buttonx off]
	conn output.14 [if plane0 input.buttony off]

	# 15-18: shifted buttons
	conn output.15 [if plane1 input.buttona off]
	conn output.16 [if plane1 input.buttonb off]
	conn output.17 [if plane1 input.buttonx off]
	conn output.18 [if plane1 input.buttony off]

	# 20: fired when A is pressed twice
	conn output.20 abtn.double



