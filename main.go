package main

import (
	"github.com/WinterSunset95/WhatsGo/base"
	"github.com/WinterSunset95/WhatsGo/helpers"
)

func main() {
	////////////////////////////////////
	//// Run the base module		////
	//// It is the master control	////
	////////////////////////////////////
	helpers.SetupHelpers()
	base.WhatsGoBase()
}
