package main

import (
	"heat-transfer/gui"
)

func main() {
	gui.StartGUILoop()

	<-make(chan struct{})
}
