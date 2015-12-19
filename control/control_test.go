package control

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"testing"
)

// one time server
func Server(t *testing.T, sock net.Listener, channel chan []byte) {
	conn, err := sock.Accept()
	if err != nil {
		close(channel)
		return
	}

	buf, err := ioutil.ReadAll(conn)
	if err != nil {
		close(channel)
		return
	}
	channel <- buf
	close(channel)
}

func setupServer(t *testing.T) (net.Listener, chan []byte) {
	sock, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatalf("Failed to create test server %s.", err)
	}
	channel := make(chan []byte)

	return sock, channel
}

func setupClient(t *testing.T, sock net.Listener) *Controller {
	port := sock.Addr().(*net.TCPAddr).Port
	transport, err := NewLocalTransport(fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Fatalf("Failed to connect to test server %s.", err)
	}
	return &Controller{transport}
}

func TestConnectAndSendMessage(t *testing.T) {
	sock, channel := setupServer(t)
	go Server(t, sock, channel)
	conn := setupClient(t, sock)

	conn.SendHex("beef", 0)
	conn.Close()
	res := <-channel
	expected := []byte{0xbe, 0xef, False, byte((0xbe + 0xef + uint(False)) % 0x100)}
	if !bytes.Equal(res, expected) {
		t.Fatalf("Did not receive the correct message. Expected %x, got %x", expected, res)
	}
}

func TestConnectAndSetColor(t *testing.T) {
	sock, channel := setupServer(t)
	go Server(t, sock, channel)
	conn := setupClient(t, sock)

	conn.SetColor(Color{0xff, 0x00, 0x00, 0, false})
	conn.Close()
	res := <-channel
	expected := []byte{0x31, 0xff, 0x00, 0x00, 0x00, 0xf0, 0x0f, (0x31 + 0xff + 0x00 + 0x00 + 0x00 + 0xf0 + 0x0f) % 0x100}
	if !bytes.Equal(res, expected) {
		t.Fatalf("Did not receive the correct message. Expected %x, got %x", expected, res)
	}
}

func TestColorParsing(t *testing.T) {
	tests := []struct {
		Input  string
		Output Color
	}{
		{
			Input: "f00",
			Output: Color{
				R:    0xff,
				G:    0,
				B:    0,
				W:    0,
				UseW: false,
			},
		}, {
			Input: "#f00",
			Output: Color{
				R:    0xff,
				G:    0,
				B:    0,
				W:    0,
				UseW: false,
			},
		}, {
			Input: "ff0000",
			Output: Color{
				R:    0xff,
				G:    0,
				B:    0,
				W:    0,
				UseW: false,
			},
		}, {
			Input: "#ff0000",
			Output: Color{
				R:    0xff,
				G:    0,
				B:    0,
				W:    0,
				UseW: false,
			},
		}, {
			Input: "red",
			Output: Color{
				R:    0xff,
				G:    0,
				B:    0,
				W:    0,
				UseW: false,
			},
		},
	}
	for _, test := range tests {
		if actual := ParseColorString(test.Input); actual == nil || *actual != test.Output {
			if actual == nil {
				t.Errorf("ParseColorString failed; expected %v, actual %v", test.Output, actual)
			} else {
				t.Errorf("ParseColorString failed; expected %v, actual %v", test.Output, *actual)
			}
		}
	}
}
