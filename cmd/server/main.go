package main

import (
	"fmt"
	"net/http"

	"github.com/vikstrous/zengge-lightcontrol/control"
	"github.com/vikstrous/zengge-lightcontrol/local"
)

func handlerOn(w http.ResponseWriter, r *http.Request) {
	transport, err := local.NewTransport("home.viktorstanchev.com:5577")
	if err != nil {
		fmt.Printf("Failed to connect. %s", err)
	}
	controller := &control.Controller{transport}
	err = controller.SetPower(true)
	if err != nil {
		fmt.Printf("Failed to set power. %s", err)
	}
	w.Write([]byte("on"))
}

func handlerOff(w http.ResponseWriter, r *http.Request) {
	transport, err := local.NewTransport("home.viktorstanchev.com:5577")
	if err != nil {
		fmt.Printf("Failed to connect. %s", err)
	}
	controller := &control.Controller{transport}
	err = controller.SetPower(false)
	if err != nil {
		fmt.Printf("Failed to set power. %s", err)
	}
	w.Write([]byte("off"))
}

func handlerState(w http.ResponseWriter, r *http.Request) {
	transport, err := local.NewTransport("home.viktorstanchev.com:5577")
	if err != nil {
		fmt.Printf("Failed to connect. %s", err)
	}
	controller := &control.Controller{transport}
	state, err := controller.GetState()
	if err != nil {
		fmt.Printf("Failed to get state. %s", err)
	}
	if !state.IsOn {
		w.Write([]byte(fmt.Sprintf("Off")))
	} else {
		w.Write([]byte(control.ColorToStr(state.Color)))
	}
}

func main() {
	http.HandleFunc("/on", handlerOn)
	http.HandleFunc("/off", handlerOff)
	http.HandleFunc("/state", handlerState)
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.ListenAndServe(":8099", nil)
}
