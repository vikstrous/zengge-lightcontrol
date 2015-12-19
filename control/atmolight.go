package control

import (
	"fmt"
	"io"
	"os"
)

// To use this feature create a pair of pseudo ttys with
// `socat -d -d pty,raw,echo=0 pty,raw,echo=0`
// This creates a pipe that VLC can write to and we can read from. It's usually
// found at `/dev/pts/2` and `/dev/pts/3`. Feed one end to VLC and the other to
// this script. VLC will send the "average" color of the scene to us and we can set
// the color of the lightbulb based on it.
func (c *Controller) AtmolightDaemon(pty string) error {
	file, _ := os.Open(pty)
	defer file.Close()
	for {
		// We read exactly 19 bytes at a time continuously because
		// that's how the atmolight protocol works
		msg := make([]byte, 19)
		_, err := io.ReadFull(file, msg)
		if err != nil {
			return err
		}
		color := Color{msg[4], msg[5], msg[6], 0, false}
		fmt.Println(ColorToStr(color))
		c.SetColor(color)
	}
	return nil
}
