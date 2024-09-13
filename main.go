package main

import (
	"github.com/WinterSunset95/WhatsGo/base"
	"github.com/WinterSunset95/WhatsGo/explorer"
	"github.com/WinterSunset95/WhatsGo/helpers"
	"github.com/WinterSunset95/WhatsGo/mediasender"
	"github.com/WinterSunset95/WhatsGo/ui"
)

func main() {
	////////////////////////////////////
	//// Run the base module		////
	//// It is the master control	////
	////////////////////////////////////
	helpers.SetupHelpers()
	ui.UIInitialize()
	explorer.ExInitialize()
	mediasender.MsInitialize()
	base.WhatsGoBase()
}
