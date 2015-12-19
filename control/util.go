package control

import (
	"encoding/hex"
	"fmt"
	"regexp"
)

func ColorToStr(c Color) string {
	if c.UseW {
		return fmt.Sprintf("%d%", c.W)
	} else {
		return fmt.Sprintf("#%x", []byte{c.R, c.G, c.B})
	}
}

func HexStrToColor(str string) *Color {
	rgb, err := hex.DecodeString(str)
	if err != nil {
		return nil
	}
	if len(rgb) < 3 {
		return nil
	}
	color := Color{rgb[0], rgb[1], rgb[2], 0, false}
	return &color
}

func ParseColorString(name string) *Color {
	if matched, _ := regexp.Match("#?[0-9a-f]{6}", []byte(name)); matched {
		if name[0] == '#' {
			name = name[1:]
		}
		return HexStrToColor(name)
	} else if matched, _ := regexp.Match("#?[0-9a-f]{3}", []byte(name)); matched {
		if name[0] == '#' {
			name = name[1:]
		}
		expandedName := fmt.Sprintf("%c%c%c%c%c%c", name[0], name[0], name[1], name[1], name[2], name[2])
		return HexStrToColor(expandedName)
	} else if color := ColorToHex[name]; color != "" {
		return HexStrToColor(color)
	} else {
		return nil
	}
}

func Checksum(data []byte) byte {
	sum := byte(0)
	for _, b := range data {
		sum += b
	}
	return sum
}
