
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

`joyster -defcfg config.json`

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

`config.TapDelayMicros` is the maximum time between taps for `Multi` buttons

`config.KeepPushedMicros` specifies how long should a `Multi` button kept pressed

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
output button will be set. Currently unused.

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

`Multi` is used for multiple taps. Value of 2 means double-tap. The output will
be triggered if the input button is pressed this many times in quick succession.

virtual joystick
----------------

The output virtual joystick has the following controls:

Left thumb: Joy_XAxis and Joy_YAxis
Right thumb: Joy_RXAxis and Joy_RYAxis

Buttons: Joy_1, Joy_2, ... Joy_32

Example
-------

Example configuration for Elite: Dangerous follows.

### Left Thumb

* Axes: Pitch and Yaw (Galaxy map: Translate X and Y)
* Button: yaw to roll toggle (Galaxy map: Z/Y toggle)

### Right Thumb

* Axes: Lateral and vertical thrust (Galaxy map: Rotate X and Y)
* Button: headlook toggle

### D-Pad

* Unshifted: targeting (target ahead, prev/next subsystem, highest treat)
* Shifted: targeting (prev hostile, prev/next target, next hostile)
* Extra (both shoulders): power (increase SYS/ENG/WEP and reset)
* UI focus: move select

### Buttons

* A: increase throttle (double tap: throttle to 100%, shifted: throttle to -100%)
* X: decrease throttle (double tap: throttle to 0%)
* B: engine boost (shifted: frameshift drive)
* Y: throttle to 50% (shifted: throttle to -100%)
* Back: toggle hardpoints (shifted: silent running, extra: landing gear)

#### UI focus

* A: select
* B: back

#### Landing

* B: thrust back
* Y: thrust forward

### Triggers

* Left: secondary fire
* Right: primary fire

#### UI Focus

* Left: prev page
* Right: next page

### Shoulders

* Left: SHIFT
* Right: double: flight assist toggle

### Notes

Map normal buttons except shift so that they yield the same value in
the output as in input.

Outputs:

* 1-8: same as input
* 9: left shoulder double tap
* 10-16: same as input
* 17-20: DPad shifted
* 21: (unused)
* 22: Back shifted
* 23-24: A, X double tap
* 25-28: DPad special shift
* 29-32: A, B, X, Y shifted

		"Buttons":[
			{"Output":1, "Inputs":[1]},
			{"Output":2, "Inputs":[2]},
			{"Output":3, "Inputs":[3]},
			{"Output":4, "Inputs":[4]},
			{"Output":5, "Inputs":[5]},
			{"Output":6, "Inputs":[6]},
			{"Output":7, "Inputs":[7]},
			{"Output":8, "Inputs":[8]},
			{"Output":9, "Inputs":[9], "Multi":2},
			{"Output":10, "Inputs":[10], "Multi":2},
			{"Output":11, "Inputs":[11]},
			{"Output":12, "Inputs":[12]},
			{"Output":13, "Inputs":[13]},
			{"Output":14, "Inputs":[14]},
			{"Output":15, "Inputs":[15]},
			{"Output":16, "Inputs":[16]},
			// D-Pad shifted
			{"Output":17, "Inputs":[9, 1]},
			{"Output":18, "Inputs":[9, 2]},
			{"Output":19, "Inputs":[9, 3]},
			{"Output":20, "Inputs":[9, 4]},
			// Back extra
			{"Output":21, "Inputs":[9, 10, 6]},
			// Back shifted
			{"Output":22, "Inputs":[9, 6]},
			// A, X double tap
			{"Output":23, "Inputs":[13], "Multi":2},
			{"Output":24, "Inputs":[15], "Multi":2},
			// D-Pad extra
			{"Output":17, "Inputs":[9, 10, 1]},
			{"Output":18, "Inputs":[9, 10, 2]},
			{"Output":19, "Inputs":[9, 10, 3]},
			{"Output":20, "Inputs":[9, 10, 4]},
			// A, B, X, Y shifted
			{"Output":29, "Inputs":[9, 13]},
			{"Output":30, "Inputs":[9, 14]},
			{"Output":31, "Inputs":[9, 15]},
			{"Output":32, "Inputs":[9, 16]}
		]
