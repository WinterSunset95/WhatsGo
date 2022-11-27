package main

import (
	//	"context"
	//	"fmt"
	"os"
	//"os/signal"
	//	"syscall"
	//

	"context"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal"
	//"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func WAConnect() (*whatsmeow.Client, error) {
	container, err := sqlstore.New("sqlite3", "file:wapp.db?_foreign_keys=on", waLog.Noop)
	if err != nil {
		return nil, err
	}
	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		panic(err)
	}
	client := whatsmeow.NewClient(deviceStore, waLog.Noop)
	if client.Store.ID == nil {
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			return nil, err
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
			} else {
				fmt.Println("Login event: ", evt.Event)
			}
		}
	} else {
		err = client.Connect()
		if err != nil {
			return nil, err
		}
	}
	return client, nil
}

func main() {
	cli, err := WAConnect()
	if err != nil {
		fmt.Println(err)
		return
	}
	var conn_status string
	if cli.IsConnected() {
		conn_status = "connected"
	} else {
		conn_status = "hehe"
	}

	array := [2]string{"Naute", "me"}

	// Declaring the main application
	app := tview.NewApplication()

	message := tview.NewTextView()
	message.SetText(conn_status)

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
	left.AddItem(message, 0, 1, false)
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
