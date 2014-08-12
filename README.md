
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

Buttons are mapped according to the following list:

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
    11     (unused)
    12     (unused)
    13     BUTTON_A
    14     BUTTON_B
    15     BUTTON_X
    16     BUTTON_Y
    17     LEFT_TRIGGER pull
    18     RIGHT_TRIGGER pull
    19     LEFT_TRIGGER touch
    20     RIGHT_TRIGGER touch

When config.ShiftButtonIndex is nonzero, the following buttons are also used.
They will be triggered then the specified button is pressed together with the
shift button. Note that buttons in the other plane will be released whenever
the state of the shift button changes.

    21     DPAD_UP
    22     DPAD_DOWN
    23     DPAD_LEFT
    24     DPAD_RIGHT
    25     START
    26     BACK
    27     LEFT_THUMB
    28     RIGHT_THUMB
    29     LEFT_SHOULDER
    30     RIGHT_SHOULDER
    31     (unused)
    32     (unused)
    33     BUTTON_A
    34     BUTTON_B
    35     BUTTON_X
    36     BUTTON_Y
    37     LEFT_TRIGGER pull
    38     RIGHT_TRIGGER pull
    39     LEFT_TRIGGER touch
    40     RIGHT_TRIGGER touch

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

`config.ShiftButtonIndex` specifies the index for the shift button. Useful values are
between 0 and 20. Specify 0 to turn the shift button feature off.

`config.ShiftButtonHide` hides the state of the shift button in the vJoy device.

`config.DontShift` is a list of integers for buttons that should not be shifted.

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

