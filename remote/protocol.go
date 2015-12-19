package remote

import (
	"fmt"
)

type Result struct {
	OK      bool        `json:"OK"`
	ErrCode int         `json:"err_code"`
	ErrMsg  string      `json:"err_msg"`
	Result  interface{} `json:"Result"`
}

func (e *Result) Error() string {
	return fmt.Sprintf("Zengge error: #%d %s", e.ErrCode, e.ErrMsg)
}

// this is used as a json object
type Device struct {
	DeviceType    int
	LedVersionNum int
	ModuleID      string
	MacAddress    string
	TimeZoneID    *interface{}
	DSToffset     int
	RawOffset     int
	IsOnline      bool
}

// this is used as a json object
type UserDevice struct {
	DeviceID string
	DevType  string
	DevName  string
	Spec     string
}

// this is used as a json object
type Module struct {
	DeviceType    int
	LedVersionNum int
	ModuleID      string
	MacAddress    string
	IsOnline      bool
}
