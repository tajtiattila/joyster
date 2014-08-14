
Joyster
=======

XBOX compatible gamepad remapper using [vJoy][] and XInput.

Requirements
------------

[vJoy]: http://vjoystick.sourceforge.net

[vJoy][]: http://vjoystick.sourceforge.net
XInput: should be available on any modern Windows system

Tested only with vJoy 2.0.4 on Windows 7 64-bit.

Usage
-----

`joyster`

Run with default configuration.

`joyster -save config.json`

Write default config to specified file and exit.

`joyster -cfg config.json`

Run using specified config file.

Thumbsticks
-----------

Left thumb -> Axis X/Y
Right thumb -> Axis RX/RY

Deadzones and response curves can be specified independently for all four
axes using config.ThumbLX/LY/RX/RY:
    Min: Deadzone setting, all values under this value will yield no movement
    Max: Values above this will be set to full movement
    Pow: Smoothness factor to allow fine tuned control

Furthermore config.ThumbCircle can be specified to be able to use the full
axes of the virtual joystick when the thumbstick is moved diagonally.

Triggers -> Axis Z/RZ (only when config.Left/RightTrigger.Axis is enabled)

Buttons
-------

Input buttons are recognised according to the following list:

     1     DPAD_UP
     2     DPAD_DOWN
     3     DPAD_LEFT
     4     DPAD_RIGHT
     5     START
     6     BACK
     7     LEFT_THUMB
     8     RIGHT_THUMB
     9     LEFT_SHOULDER
    10     RIGHT_SHOULDER
    11     LEFT_TRIGGER
    12     RIGHT_TRIGGER
    13     BUTTON_A
    14     BUTTON_B
    15     BUTTON_X
    16     BUTTON_Y
    17     LEFT_TRIGGER pull
    18     LEFT_TRIGGER touch
    19     RIGHT_TRIGGER pull
    20     RIGHT_TRIGGER touch

They are mapped according to elements in config.Buttons to output buttons.

A maximum number of 128 output buttons are supported by [vJoy][], but most
programs can recognise only 32 of them.

Triggers may have two virtual buttons assigned to them based on how far they
were pulled in. If any of the tresholds are greater than one, then the
corresponding virtual button will never be activated. TouchThreshold should be smaller
than PullThreshold, otherwise it will never be activated.

POV Hats
--------

POV Hats are not supported by [vJoy][] yet, so the D-Pad will be mapped to normal buttons.

Configuration
-------------

Values in the configuration are integers (button indexes), boolean values (for flags) and
floating point values usually in the range (0..1).

`config.UpdateMicros` update frequency in microseconds

`config.TapDelayMicros` is the maximum time between taps for "double-tap" buttons

`config.KeepPushedMicros` specifies how long should a double-tapped button kept pressed

`config.ThumbCircle` modifies the logic of reading thumbstick positions. Normally when the
thumbstick is moved diagonally, the distinct X and Y axes cannot reach their extreme values.
With this flag enabled, the thumbstick position inside the unit circle are mapped
to values reaching the corners of the unit square in the output.

--------------------------------------------------------------------------------

`config.ThumbLX`
`config.ThumbLY`
`config.ThumbRX`
`config.ThumbRY`

Settings for thumb stick axes. These are maps of following values:

`Min` sets the deadzone. Input values values under this value will yield zero output.
Valid values are between 0 and 1.

`Max` sets the upper bound. Input Values above this will yield maximal output value.
Valid values are between 0 and 1.

`Pow` is the smoothness factor. The normalized input value between Min and Max will
be raised to this power. Larger values yield fine control easier at the expense of
reduced precision and responsiveness for large output values. Useful values are
between 1 and 5.

--------------------------------------------------------------------------------

`config.LeftTrigger`
`config.RightTrigger`

Settings for triggers. These are maps with the following values:

`PullTreshold` sets the threshold, above which the corresponding "pull" output
button will be set.

`TouchTreshold` sets the threshold, above which the corresponding "touch"
output button will be set.

`Axis` enables the trigger into an axis.

`TouchTreshold` should be smaller than `PullTreshold` to be of any use. If set to an equal
or larger value than `PullThreshold` then the "touch" button in the output will never fire.

--------------------------------------------------------------------------------

`config.Buttons`

Array of buttons that will appear in the output.

`Output` is the output button index, ideally in the range 1-32. If the output button
is not configured for the [vJoy][] device, this button setting will have no effect.

`Sources` is an array of input buttons, they must be held together to have the button
triggered. Normally only a single input is specified here, but more than one may be used
if an input button is used as a shift button. 

If any input button index appears in multiple `Source` blocks, they will be subjected
to special handling. Eg. with

`"Buttons":[
	{"Output":1, "Sources": [13]},
	{"Output":2, "Sources": [9, 13]},
]`

Output 1 will be fired only when input 13 is pressed, but 9 is not.

`Double` is a second output for this button, when the input(s) are double tapped.
The output will be triggered if the input button is pressed twice in quick succession.
The normal output will be disabled while the double output is set. This is useful for
continuous outputs, eg. normal: increase throttle, double: set throttle to 100%.

virtual joystick
----------------

The output virtual joystick has the following controls:

Left thumb: Joy_XAxis and Joy_YAxis
Right thumb: Joy_RXAxis and Joy_RYAxis

Buttons: Joy_1, Joy_2, ... Joy_32

Example
-------

An example configuration for Elite: Dangerous with the controls below
can be found in the misc/ed directory.

### Flight controls

#### Triggers

Fire Primary/Secondary

#### Shoulders

used as shift (LS and RS)

#### D-Pad: targeting

* Up: target ahead
* Down: highest thread
* Left/Right: prev/next subsystem

#### LS + D-Pad: targeting extra

* Up/Down: prev/next hostile
* Left/Right: prev/next ship

#### RS + D-Pad: toggles (hardpoint, landing gear, flight assist, cargo scoop)

* Up: hardpoints
* Down: landing gear
* Left: flight assist
* Right: cargo scoop

#### LS + RS + dpad: misc. toggles sensor range, lights

* Up/Down: sensor range
* Right: ship spotlight

#### Buttons: thrust

* X: increase throttle (double: set to 100%)
* A: reduce throttle (double: set to 0%)
* B: engine boost
* Y: throttle to 50% (double: set to -100%)

#### LS + Buttons: thrust extra

* A: thrust backward
* X: thrust forward
* B: frameshift

#### RS + Buttons: power distribution

* X: increase SYS
* Y: increase ENG
* B: increase WEP
* A: balance power

#### Left Thumb

* Axes: Pitch and Yaw (Galaxy map: Translate X and Y)
* Button: yaw to roll toggle (Galaxy map: Z/Y toggle)

#### Right Thumb

* Axes: Lateral and vertical thrust (Galaxy map: Rotate X and Y)
* Button: headlook toggle

----------

### UI Focus

* D-Pad: move selection
* Triggers: prev/next page
* A: select

### Galaxy map

#### Left Thumb

* Axes: Translate X and Y
* Button: Z/Y toggle

#### Right Thumb

* Axes: Rotate X and Y

#### Other

* Triggers: prev/next page
* A: select

### Notes

Buttons that support double click should be shifted within joyster,
otherwise a double click could be triggered even if a shift button is pressed.

Outputs:

* 1-20: same as inputs (see "Buttons" above)
* 21-24: buttons with LS
* 25-28: buttons with RS
* 29-32: buttons double tap

Config:

		  "Buttons": [
			  { "Output":1,  "Inputs":[1]  },
			  { "Output":2,  "Inputs":[2]  },
			  { "Output":3,  "Inputs":[3]  },
			  { "Output":4,  "Inputs":[4]  },
			  { "Output":5,  "Inputs":[5]  },
			  { "Output":6,  "Inputs":[6]  },
			  { "Output":7,  "Inputs":[7]  },
			  { "Output":8,  "Inputs":[8]  },
			  { "Output":9,  "Inputs":[9]  },
			  { "Output":10, "Inputs":[10] },
			  { "Output":11, "Inputs":[11] },
			  { "Output":12, "Inputs":[12] },
			  { "Output":13, "Inputs":[13], "Double":29 },
			  { "Output":14, "Inputs":[14], "Double":30 },
			  { "Output":15, "Inputs":[15], "Double":31 },
			  { "Output":16, "Inputs":[16], "Double":32 },
			  { "Output":17, "Inputs":[17] },
			  { "Output":18, "Inputs":[18] },
			  { "Output":19, "Inputs":[19] },
			  { "Output":20, "Inputs":[20] },
			  { "Output":21, "Inputs":[9, 13] },
			  { "Output":22, "Inputs":[9, 14] },
			  { "Output":23, "Inputs":[9, 15] },
			  { "Output":24, "Inputs":[9, 16] },
			  { "Output":25, "Inputs":[10, 13] },
			  { "Output":26, "Inputs":[10, 14] },
			  { "Output":27, "Inputs":[10, 15] },
			  { "Output":28, "Inputs":[10, 16] }
		  ]

