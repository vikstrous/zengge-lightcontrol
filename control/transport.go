package control

type Transport interface {
	SendReceive(data []byte, responseSize int) (response []byte, err error)
	Remote() bool
	Close()
}
