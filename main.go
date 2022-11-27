





package main

import (
	//	"context"
	//	"fmt"
	//	"os"
	//	"os/signal"
	//	"syscall"
	//
	//	"go.mau.fi/whatsmeow"
	//	"go.mau.fi/whatsmeow/store/sqlstore"
	//	"go.mau.fi/whatsmeow/types/events"
	//	waLog "go.mau.fi/whatsmeow/util/log"
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func main() {
	array := [2]string{"Naute", "me"}
	fmt.Println(array)

	// Declaring the main application
	app := tview.NewApplication()

	// Root
	box := tview.NewFlex().SetDirection(tview.FlexRow)
	box.SetBorder(true).SetTitle("Muffin")

	// Input field
	text := tview.NewInputField().SetLabelWidth(0)
	text.SetBorder(true)
	text.SetFieldBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
	text.SetDoneFunc(func(key tcell.Key) {
		box.AddItem(tview.NewTextView().SetText(text.GetText()), 0, 1, false)
		text.SetText("")
	})

	// Users
	left := tview.NewFlex().SetDirection(tview.FlexRow)
	left.SetBorder(true).SetTitle("Users")
	for i := 0; i < len(array); i++ {
		left.AddItem(tview.NewTextView().SetText(array[i]), 0, 1, false)
	}

	// Messages
	right := tview.NewFlex().SetDirection(tview.FlexRow)
	right.AddItem(box, 0, 15, false)
	right.AddItem(text, 0, 1, true)

	body := tview.NewFlex().SetDirection(tview.FlexColumn)
	body.AddItem(left, 0, 1, false).AddItem(right, 0, 3, true)

	if err := app.SetRoot(body, true).Run(); err != nil {
		panic(err)
	}
}
