package remote

import (
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

type Controller struct {
	Client     http.Client
	ControlURL string
	Secret     []byte
	AppKey     string
	DevID      string
	Version    string
	System     string
}

func NewController(controlURL, secret, devID string) *Controller {
	jar, _ := cookiejar.New(nil)
	c := Controller{
		Client:     http.Client{Jar: jar},
		ControlURL: controlURL,
		Secret:     []byte(secret),
		DevID:      devID,
	}
	return &c
}

// returns signature, timestamp
func (c *Controller) Signature() (string, string, error) {
	// use a static timestamp because the server doesn't check
	timestamp := "timestamp"
	str := c.DevID + c.AppKey + timestamp
	cyphertext, err := AESCBC([]byte(str), c.Secret)
	if err != nil {
		return "", "", err
	}
	return hex.EncodeToString(cyphertext), timestamp, nil
}

func (c *Controller) SendCommand(cmd string, data *url.Values) (interface{}, error) {
	reqBody := &strings.Reader{}
	if data != nil {
		reqBody = strings.NewReader(data.Encode())
	}
	req, err := http.NewRequest("POST", c.ControlURL, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Add("zg-app-cmd", cmd)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	zerr := Result{}
	err = json.Unmarshal(body, &zerr)
	if err != nil {
		return nil, err
	}
	if !zerr.OK {
		return nil, &zerr
	}
	return zerr.Result, nil
}

func (c *Controller) Login() error {
	signature, timestamp, err := c.Signature()
	if err != nil {
		return err
	}
	_, err = c.SendCommand("Login", &url.Values{
		"AppKey":    {c.AppKey},
		"DevID":     {c.DevID},
		"AppVer":    {c.AppKey},
		"AppSys":    {c.System},
		"Timestamp": {timestamp},
		"CheckCode": {signature},
	})
	return err
}

func (c *Controller) RegisterDevice(mac string) error {
	udev := UserDevice{
		DeviceID: c.DevID,
		DevType:  c.System,
		// TODO: wire these up
		DevName: "A0001",
		Spec:    "A0001",
	}
	udevBytes, err := json.Marshal(&udev)
	if err != nil {
		return err
	}
	module := Module{
		// TODO: get these from the bulb itself (state command and management handshake)
		DeviceType:    68,
		LedVersionNum: 4,
		ModuleID:      "HF-LPB100-ZJ200",
		MacAddress:    mac,
		// XXX: does this matter?
		IsOnline: false,
	}
	moduleBytes, err := json.Marshal(&module)
	if err != nil {
		return err
	}
	signature, timestamp, err := c.Signature()
	if err != nil {
		return err
	}
	_, err = c.SendCommand("AuthorizationDevice", &url.Values{
		"DevID":      {c.DevID},
		"AppSys":     {c.System},
		"AppKey":     {c.AppKey},
		"AppVer":     {c.Version},
		"Timestamp":  {timestamp},
		"CheckCode":  {signature},
		"UserDevice": {string(udevBytes)},
		"Module":     {string(moduleBytes)},
	})
	return err
}

func (c *Controller) DeregisterDevice(mac string) error {
	signature, timestamp, err := c.Signature()
	if err != nil {
		return err
	}
	_, err = c.SendCommand("RemoveAuthorization", &url.Values{
		"AppKey":     {c.AppKey},
		"DeviceID":   {c.DevID},
		"Timestamp":  {timestamp},
		"CheckCode":  {signature},
		"MacAddress": {mac},
	})
	return err
}

func (c *Controller) GetOwners(mac string) ([]UserDevice, error) {
	info, err := c.SendCommand("GetAuthUserDevice", &url.Values{
		"MacAddress": {mac},
	})
	if err != nil {
		return nil, err
	}
	devices := []UserDevice{}
	err = json.Unmarshal([]byte(info.(string)), &devices)
	if err != nil {
		return nil, err
	}
	return devices, err
}

func (c *Controller) GetDevices() ([]Device, error) {
	res, err := c.SendCommand("OnlineDevices", nil)
	if err != nil {
		return nil, err
	}
	devices := []Device{}
	err = json.Unmarshal([]byte(res.(string)), &devices)
	if err != nil {
		return nil, err
	}
	return devices, nil
}
