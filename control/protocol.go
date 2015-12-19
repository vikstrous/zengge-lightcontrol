package control

import "time"

type Color struct {
	R    uint8
	G    uint8
	B    uint8
	W    uint8
	UseW bool
}

type State struct {
	DeviceType    uint8
	IsOn          bool
	LedVersionNum uint8
	Mode          uint8
	Slowness      uint8
	Color         Color
}

const (
	ModeColor    = 97
	ModeMusic    = 98
	ModeCustom   = 35
	ModePreset1  = 37
	ModePreset2  = 38
	ModePreset3  = 39
	ModePreset4  = 40
	ModePreset5  = 41
	ModePreset6  = 42
	ModePreset7  = 43
	ModePreset8  = 44
	ModePreset9  = 45
	ModePreset10 = 46
	ModePreset11 = 47
	ModePreset12 = 48
	ModePreset13 = 49
	ModePreset14 = 50
	ModePreset15 = 51
	ModePreset16 = 52
	ModePreset17 = 53
	ModePreset18 = 54
	ModePreset19 = 55
	ModePreset20 = 56
)

func ModeName(mode uint8) string {
	switch mode {
	case ModeColor:
		return "Color"
	case ModeMusic:
		return "Music"
	case ModeCustom:
		return "Custom"
	case ModePreset1:
		return "Preset1"
	case ModePreset2:
		return "Preset2"
	case ModePreset3:
		return "Preset3"
	case ModePreset4:
		return "Preset4"
	case ModePreset5:
		return "Preset5"
	case ModePreset6:
		return "Preset6"
	case ModePreset7:
		return "Preset7"
	case ModePreset8:
		return "Preset8"
	case ModePreset9:
		return "Preset9"
	case ModePreset10:
		return "Preset10"
	case ModePreset11:
		return "Preset11"
	case ModePreset12:
		return "Preset12"
	case ModePreset13:
		return "Preset13"
	case ModePreset14:
		return "Preset14"
	case ModePreset15:
		return "Preset15"
	case ModePreset16:
		return "Preset16"
	case ModePreset17:
		return "Preset17"
	case ModePreset18:
		return "Preset18"
	case ModePreset19:
		return "Preset19"
	case ModePreset20:
		return "Preset20"
	default:
		return "Unknown"
	}
}

func (c Color) Format() []byte {
	ignoreW := True
	if c.UseW {
		ignoreW = False
	}
	return []byte{c.R, c.G, c.B, c.W, ignoreW}
}

func (c *Color) Parse(data []byte, ignoreW byte) {
	c.R = data[0]
	c.G = data[1]
	c.B = data[2]
	c.W = data[3]
	if ignoreW == False {
		c.UseW = true
	} else {
		c.UseW = false
	}
}

type Time struct {
	Time time.Time
}

func (t *Time) Format() []byte {
	b := []byte{}
	b = append(b, byte(t.Time.Year()-2000))
	b = append(b, byte(t.Time.Month()))
	b = append(b, byte(t.Time.Day()))
	b = append(b, byte(t.Time.Hour()))
	b = append(b, byte(t.Time.Minute()))
	b = append(b, byte(t.Time.Second()))
	return b
}

func (t *Time) RawParse(data []byte) {
	t.Time = time.Date(2000+int(data[0]), time.Month(data[1]), int(data[2]), int(data[3]), int(data[4]), int(data[5]), 0, time.Local)
}

func (t *Time) Parse(data []byte) {
	// byte 0 is ignored (0f?)
	// byte 1 is always 17 (0x11)
	// byte 2 is ignored ???
	// bytes 9-10 ignored (0x0300)
	// byte 11 is the checksum
	t.RawParse(data[3:])
}

type TimerList struct {
	Timers []Timer
}
type Timer struct {
	Enabled  bool
	Mode     uint8
	PowerOn  bool
	Weekdays []bool
	Data     [4]byte
	Time     Time
}

func (tl *TimerList) Format() []byte {
	b := []byte{}
	for _, t := range tl.Timers {
		b = append(b, t.Format()...)
	}
	return b
}

func (tl *TimerList) Parse(data []byte) {
	tl.Timers = []Timer{}
	// data[1] == 34
	// data[87] is checksum
	// there are exactly 6 timers
	// first timer at [2]
	for i := 0; i < 6; i++ {
		var timer Timer
		timer.Parse(data[2+i*14 : 2+(i+1)*14])
		tl.Timers = append(tl.Timers, timer)
	}
}

func (t *Timer) Format() []byte {
	b := []byte{}
	if t.Enabled {
		b = append(b, True)
	} else {
		b = append(b, False)
	}
	b = append(b, t.Time.Format()...)
	weekdays := byte(0)
	for i, on := range t.Weekdays {
		if on {
			weekdays |= 1 << uint(i)
		}
	}
	b = append(b, weekdays)
	b = append(b, t.Mode)
	b = append(b, t.Data[:]...)
	if t.PowerOn {
		b = append(b, True)
	} else {
		b = append(b, False)
	}
	return b
}

func (t *Timer) Parse(data []byte) {
	if data[0] == True {
		t.Enabled = true
	} else {
		t.Enabled = false
	}
	t.Time.RawParse(data[1:7])

	t.Weekdays = []bool{}
	for i := uint(0); i < 7; i++ {
		t.Weekdays = append(t.Weekdays, data[7]&(1<<i) > 0)
	}
	t.Mode = data[8]
	copy(t.Data[:], data[9:13])
	if data[13] == True {
		t.PowerOn = true
	} else {
		t.PowerOn = false
	}
}

const (
	CommandSetTime       = uint8(0x10)
	CommandSetTime2      = uint8(0x14)
	CommandGetTime       = uint8(0x11)
	CommandGetTime2      = uint8(0x1A)
	CommandGetTime3      = uint8(0x1B)
	CommandGetTimers     = uint8(0x22)
	CommandGetTimers2    = uint8(0x2A)
	CommandGetTimers3    = uint8(0x2B)
	CommandSetColor      = uint8(0x31)
	CommandSetMusicColor = uint8(0x41)
	CommandSetMode       = uint8(0x61)
	CommandSetPower      = uint8(0x71)
	CommandGetState      = uint8(0x81)
	CommandGetState2     = uint8(0x8a)
	CommandGetState3     = uint8(0x8b)
)

const (
	True  = uint8(0xf0)
	False = uint8(0x0f)
	On    = uint8(0x23)
	Off   = uint8(0x24)
)

func FormatSetPower(on bool) []byte {
	if on {
		return []byte{CommandSetPower, On}
	} else {
		return []byte{CommandSetPower, Off}
	}
}

func FormatSetColor(c Color) []byte {
	buff := []byte{CommandSetColor}
	buff = append(buff, c.Format()...)
	return buff
}

func FormatGetTime() []byte {
	return []byte{CommandGetTime, CommandGetTime2, CommandGetTime3}
}
func FormatGetTimers() []byte {
	return []byte{CommandGetTimers, CommandGetTimers2, CommandGetTimers3}
}

func FormatSetMode(mode, speed uint8) []byte {
	// TODO
	return nil
}
func ParseState(data []byte) State {
	state := State{}
	// data[0] is always 0x81
	// device type is always 68
	state.DeviceType = data[1]
	state.IsOn = data[2] == On
	state.Mode = data[3]
	// data[4] is always 33
	state.Slowness = data[5]
	state.Color.Parse(data[6:10], data[12])
	state.LedVersionNum = data[10] // always 4
	// data[11] is always 0
	return state
}
