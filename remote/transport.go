package remote

import (
	"encoding/hex"
	"fmt"
	"net/url"
)

type RemoteTransport struct {
	Controller
	MAC string
}

func NewRemoteTransport(controller *Controller, mac string) *RemoteTransport {
	r := RemoteTransport{
		Controller: *controller,
		MAC:        mac,
	}
	return &r
}

func (e *RemoteTransport) Remote() bool {
	return true
}

func (e *RemoteTransport) Close() {
}

func (r *RemoteTransport) SendReceive(data []byte, responseLength int) ([]byte, error) {
	res, err := r.SendCommand("DataCommand", &url.Values{
		"Data":          {hex.EncodeToString(data)},
		"ResponseCount": {fmt.Sprintf("%d", responseLength)},
		"MacAddress":    {r.MAC},
	})
	if err != nil {
		return nil, err
	}
	return hex.DecodeString(res.(string))
}
