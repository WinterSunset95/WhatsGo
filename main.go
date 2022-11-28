




package main

import (
	"os"
	"context"
	"fmt"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal"
	//"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
	"go.mau.fi/whatsmeow/types"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"
	"go.mau.fi/whatsmeow/types/events"
)

var log waLog.Logger
var globalMsg []string

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

func parseJID(arg string) (types.JID, bool) {
	if arg[0] == '+' {
		arg = arg[1:]
	}
	if !strings.ContainsRune(arg, '@') {
		return types.NewJID(arg, types.DefaultUserServer), true
	} else {
		recipient, err := types.ParseJID(arg)
		if err != nil {
			log.Errorf("Invalid JID %s: %v", arg, err)
			return recipient, false
		} else if recipient.User == "" {
			log.Errorf("Invalid JID %s: no server specified", arg)
			return recipient, false
		}
		return recipient, true
	}
}

func main() {
	jid := "+916009019522@s.whatsapp.net"
	// putting my test number for now
	recipient, ok := parseJID(jid)

	if ok {
		fmt.Println("Ok")
	}
	cli, err := WAConnect()
	if err != nil {
		fmt.Println(err)
		return
	}


	groups, err := cli.GetJoinedGroups()
	users, err := cli.Store.Contacts.GetAllContacts()

	// Declaring the main application
	app := tview.NewApplication()

	// Messages container
	box := tview.NewTable().SetSelectable(true, false)
	box.SetBorder(true).SetTitle("Messages")
	count := 1
	box.SetCell(0, 0, tview.NewTableCell("Start"))

	// Users
	usr_row := 1
	left := tview.NewTable().SetSelectable(true, false)
	left.GetMouseCapture()
	left.SetBorder(true).SetTitle("Users")
	left.SetCell(0, 0, tview.NewTableCell("Connected"))
	for k, v := range users {
		if v.PushName != "" {
			left.SetCell(usr_row, 0, tview.NewTableCell(v.PushName))
			left.SetCell(usr_row, 1, tview.NewTableCell(k.User))
			usr_row++
		}
	}
	for i := 0; i < len(groups); i++ {
		left.SetCell(usr_row, 0, tview.NewTableCell(groups[i].Name))
		usr_row++
	}

	// Input field
	text := tview.NewInputField().SetLabelWidth(0)
	text.SetBorder(true)
	text.SetFieldBackgroundColor(tview.Styles.PrimitiveBackgroundColor)

	// Messages
	right := tview.NewFlex().SetDirection(tview.FlexRow)
	right.AddItem(box, 0, 15, false)
	right.AddItem(text, 0, 1, true)

	body := tview.NewFlex().SetDirection(tview.FlexColumn)
	body.AddItem(left, 0, 1, false).AddItem(right, 0, 3, true)
	
	// handlers
	handler := func(rawEvt interface{}) {
		switch evt := rawEvt.(type) {
			case *events.Message:
				global := evt.Message.GetConversation()
				box.SetCell(count, 0, tview.NewTableCell(global))
				count++
			case *events.Receipt:
				if evt.Type == events.ReceiptTypeDelivered {
					box.SetCell(count, 0, tview.NewTableCell(text.GetText()))
					text.SetText("")
					box.SetTitle("Delivered")
					count++
				} else if evt.Type == events.ReceiptTypeRead {
					box.SetTitle("Read")
				}
		}
	}
	cli.AddEventHandler(handler)
	time.Sleep(0 * time.Millisecond)

	// Input captures
	left.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == rune(tcell.KeyTab) {
			app.SetFocus(box)
		}
		return event
	})
	box.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == rune(tcell.KeyTab) {
			app.SetFocus(text)
		}
		return event
	})
	text.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == rune(tcell.KeyEnter) {
			msg := text.GetText()
			cli.SendMessage(context.Background()	, recipient, "", &waProto.Message{Conversation: proto.String(msg)})
		} else if event.Rune() == rune(tcell.KeyTab) {
			app.SetFocus(left)
		}
		return event
	})

	if err := app.SetRoot(body, true).Run(); err != nil {
		panic(err)
	}

}
