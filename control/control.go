package control

import "encoding/hex"

type Controller struct {
	Transport Transport
}

func ScanLANAndConnect() *Controller {
	// TODO
	return nil
}

func (c *Controller) SendBytesRaw(data []byte, responseSize int) ([]byte, error) {
	return c.Transport.SendReceive(data, responseSize)
}

func (c *Controller) SendBytes(data []byte, responseSize int) ([]byte, error) {
	if c.Transport.Remote() {
		data = append(data, True)
	} else {
		data = append(data, False)
	}
	sum := Checksum(data)
	data = append(data, sum)
	return c.SendBytesRaw(data, responseSize)
}

func (c *Controller) SendHex(str string, responseSize int) ([]byte, error) {
	data, err := hex.DecodeString(str)
	if err != nil {
		return nil, err
	}
	return c.SendBytes(data, responseSize)
}

func (c *Controller) SetPower(on bool) error {
	_, err := c.SendBytes(FormatSetPower(on), 0)
	return err
}

func (c *Controller) SetColor(color Color) error {
	_, err := c.SendBytes(FormatSetColor(color), 0)
	return err
}

func (c *Controller) GetState() (*State, error) {
	data := []byte{CommandGetState, CommandGetState2, CommandGetState3}
	sum := Checksum(data)
	data = append(data, sum)

	data, err := c.SendBytesRaw(data, 14)
	if err != nil {
		return nil, err
	}
	state := ParseState(data)
	return &state, nil
}

func (c *Controller) GetTime() (*Time, error) {
	data, err := c.SendBytes(FormatGetTime(), 12)
	if err != nil {
		return nil, err
	}
	var time Time
	time.Parse(data)
	return &time, nil
}

func (c *Controller) GetTimers() (*TimerList, error) {
	data, err := c.SendBytes(FormatGetTimers(), 88)
	if err != nil {
		return nil, err
	}
	var timers TimerList
	timers.Parse(data)
	return &timers, nil
}

func (c *Controller) Close() {
	c.Transport.Close()
}
