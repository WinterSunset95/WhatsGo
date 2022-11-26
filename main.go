





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

	"github.com/rivo/tview"
)

// golang declare an array
// https://www.geeksforgeeks.org/golang-declare-an-array/

func main() {
	array := [2]string{"Naute", "me"}
	fmt.Println(array)

	app := tview.NewApplication()

	textBox := tview.NewButton(array[0])

	box := tview.NewTextView()
	box.SetBorder(true).SetTitle("Muffin")

	text := tview.NewTextArea().SetPlaceholder("Enter text here.")
	text.SetBorder(true)

	left := tview.NewFlex()
	left.AddItem(textBox, 0, 1, false)
	left.SetBorder(true).SetTitle("Users")
	right := tview.NewFlex().SetDirection(tview.FlexRow)
	right.AddItem(box, 0, 15, false)
	right.AddItem(text, 0, 1, true)

	body := tview.NewFlex().SetDirection(tview.FlexColumn)
	body.AddItem(left, 0, 1, false).AddItem(right, 0, 3, true)

	if err := app.SetRoot(body, true).Run(); err != nil {
		panic(err)
	}
}
