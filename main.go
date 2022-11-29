



// the top of my laptop screen is broken
// i'm leaving this empty space here so I can see code at the top
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
	// map holding the JID with an array of messages
	var database = make(map[types.JID][]string)
	// map holding the JID with the username
	var name_map = make(map[types.JID]types.ContactInfo)

	var recipient types.JID

	// putting my test number for now
	cli, err := WAConnect()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Getting all the groups and contacts
	groups, err := cli.GetJoinedGroups()
	users, err := cli.Store.Contacts.GetAllContacts()
	filtered := make(map[types.JID]types.ContactInfo)
	for k, v := range users {
		// Ignore this part please. I will come back to it later
		if strings.Contains(v.PushName, "") || strings.Contains(v.PushName, "") {
			filtered[k] = v
		}
	}

	// Declaring the main application
	app := tview.NewApplication()

	// Messages container
	box := tview.NewTable().SetSelectable(true, false)
	box.SetBorder(true).SetTitle("Messages")
	box.SetCell(0, 0, tview.NewTableCell("Start"))

	// Users
	usr_row := 1
	list := tview.NewTable().SetSelectable(true, false)
	list.GetMouseCapture()
	list.SetBorder(true).SetTitle("Users")
	list.SetCell(0, 0, tview.NewTableCell("Connected"))
	for k, v := range filtered {
		if v.PushName != "" {
			list.SetCell(usr_row, 0, tview.NewTableCell(v.PushName))
			list.SetCell(usr_row, 3, tview.NewTableCell(k.User))
			list.SetCell(usr_row, 4, tview.NewTableCell(k.String()))
			usr_row++
		}
	}
	for i := 0; i < len(groups); i++ {
		list.SetCell(usr_row, 0, tview.NewTableCell(groups[i].Name))
		list.SetCell(usr_row, 1, tview.NewTableCell(groups[i].JID.Server))
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
	body.AddItem(list, 0, 1, false).AddItem(right, 0, 3, true)

	// When contact is selected
	new_select := func(jid string) {
		rec, ok := parseJID(jid)
		recipient = rec
		if ok {
			fmt.Println("Ok")
		}
		db_check, d_ok := database[recipient]
		name_check, n_ok := name_map[recipient]
		// check if user is already in db
		if !d_ok && !n_ok && db_check == nil && name_check.Found == false {
			database[recipient] = []string{}
			name_map[recipient], err = cli.Store.Contacts.GetContact(recipient)
		}
		box.Clear()
	}

	// handlers
	msg := ""
	handler := func(rawEvt interface{}) {
		switch evt := rawEvt.(type) {
			case *events.Message:
				if evt.Info.Sender == recipient {
					global := evt.Message.GetConversation()
					database[recipient] = append(database[recipient], name_map[recipient].PushName + ": " + global)
					for i, s := range database[recipient] {
						box.SetCell(i, 0, tview.NewTableCell(s))
					}
					app.Draw()
				} else if evt.Info.Sender != recipient {
					// if the message is not from the selected user
					// then we need to add it to the database
					// and update the list
					// Add user if not in db
					if val, ok := database[evt.Info.Sender]; ok {
						database[evt.Info.Sender] = append(val, name_map[evt.Info.Sender].PushName + ":" + evt.Message.GetConversation())
					} else {
						database[evt.Info.Sender] = []string{evt.Message.GetConversation()}
						name_map[evt.Info.Sender], err = cli.Store.Contacts.GetContact(evt.Info.Sender)
					}
				} else if evt.Info.MessageSource.IsFromMe {
					global := evt.Message.GetConversation()
					database[recipient] = append(database[recipient], name_map[recipient].PushName + ": " + global)
					for i, s := range database[recipient] {
						box.SetCell(i, 0, tview.NewTableCell(s))
					}
					app.Draw()
				}
			case *events.Receipt:
				for i, s := range database[recipient] {
					box.SetCell(i, 0, tview.NewTableCell(s))
				}
				if evt.Type == events.ReceiptTypeDelivered {
					box.SetTitle("Delivered")
				} else if evt.Type == events.ReceiptTypeRead {
					box.SetTitle("Read")
				} else if evt.MessageSource.IsFromMe {
				}
				app.Draw()
			case *events.HistorySync:
				box.SetTitle("History Sync")
		}
	}

	cli.AddEventHandler(handler)
	// Ignore this too I might need it later
	time.Sleep(0 * time.Millisecond)

	// Input captures
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == rune(tcell.KeyTab) {
			app.SetFocus(box)
		} else if event.Rune() == rune(tcell.KeyEnter) {
			row, col := list.GetSelection()
			list.GetCell(row, col).SetTextColor(tcell.ColorGreen)
			new_select(list.GetCell(row, 4).Text)
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
		if event.Rune() == rune(tcell.KeyEnter) && text.GetText() != "" {
			msg = text.GetText()
			text.SetText("")
			database[recipient] = append(database[recipient], "Me: " + msg)
			// u/darkhz told me I should use a goroutine for this.. No idea what that is...
			cli.SendMessage(context.Background()	, recipient, "", &waProto.Message{Conversation: proto.String(msg)})
			msg = ""
		} else if event.Rune() == rune(tcell.KeyTab) {
			app.SetFocus(list)
		}
		return event
	})

	if err := app.SetRoot(body, true).Run(); err != nil {
		panic(err)
	}

}
