package local

import (
	"io"
	"net"
)

type Transport struct {
	Conn *net.TCPConn
}

func (lt *Transport) Receive(length int) ([]byte, error) {
	dst := make([]byte, length)
	_, err := io.ReadFull(lt.Conn, dst)
	if err != nil {
		return nil, err
	}
	return dst, nil
}

func (lt *Transport) SendReceive(data []byte, responseSize int) ([]byte, error) {
	_, err := lt.Conn.Write(data)
	if responseSize > 0 {
		data, err = lt.Receive(responseSize)
		if err != nil {
			return nil, err
		}
		return data, err
	}
	return nil, err
}

func (lc *Transport) Remote() bool {
	return false
}

func (lc *Transport) Close() {
	lc.Conn.Close()
}

func NewTransport(ipPort string) (*Transport, error) {
	lt := Transport{}
	conn, err := net.Dial("tcp", ipPort)
	if err != nil {
		return nil, err
	}
	lt.Conn = conn.(*net.TCPConn)
	return &lt, nil
}
