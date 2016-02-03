package main

import (
	"fmt"
	"net/http"

	"github.com/vikstrous/zengge-lightcontrol/control"
	"github.com/vikstrous/zengge-lightcontrol/local"
)

var transport *local.Transport
var controller *control.Controller

func init() {
	var err error
	transport, err = local.NewTransport("home.viktorstanchev.com:5577")
	if err != nil {
		fmt.Printf("Failed to connect. %s", err)
		return
	}
	controller = &control.Controller{transport}
}

func handlerOn(w http.ResponseWriter, r *http.Request) {
	err := controller.SetPower(true)
	if err != nil {
		fmt.Printf("Failed to set power. %s", err)
		return
	}
	w.Write([]byte("on"))
}

func handlerOff(w http.ResponseWriter, r *http.Request) {
	err := controller.SetPower(false)
	if err != nil {
		fmt.Printf("Failed to set power. %s", err)
		return
	}
	w.Write([]byte("off"))
}

func handlerState(w http.ResponseWriter, r *http.Request) {
	state, err := controller.GetState()
	if err != nil {
		fmt.Printf("Failed to get state. %s", err)
		return
	}
	if !state.IsOn {
		w.Write([]byte(fmt.Sprintf("Off")))
	} else {
		w.Write([]byte(control.ColorToStr(state.Color)))
	}
}

func handlerColor(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	colorStr := r.FormValue("color")
	color := control.ParseColorString(colorStr)
	if color == nil {
		fmt.Printf("Failed to parse color.")
		return
	}

	err := controller.SetColor(*color)
	if err != nil {
		fmt.Printf("Failed to set color. %s", err)
		return
	}
	w.Write([]byte(fmt.Sprintf("done")))
}

func main() {
	http.HandleFunc("/on", handlerOn)
	http.HandleFunc("/off", handlerOff)
	http.HandleFunc("/state", handlerState)
	http.HandleFunc("/color", handlerColor)
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.ListenAndServe(":8099", nil)
}
