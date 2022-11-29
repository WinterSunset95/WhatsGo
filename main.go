



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

type UserMessage struct {
	name string
	text string
}

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
	var database = make(map[types.JID][]UserMessage)
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
	for k, v := range users {
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

	// Filter Input
	filter_input := tview.NewInputField().SetLabel("Search: ")
	filter_input.SetBorder(true)
	filter_input.SetFieldBackgroundColor(tview.Styles.PrimitiveBackgroundColor)

	// Input field
	text := tview.NewInputField().SetLabelWidth(0)
	text.SetBorder(true)
	text.SetFieldBackgroundColor(tview.Styles.PrimitiveBackgroundColor)

	// Left side of screen - Contacts, Groups, Filter input
	left := tview.NewFlex().SetDirection(tview.FlexRow)
	left.AddItem(list, 0, 20, false)
	left.AddItem(filter_input, 0, 1, true)

	// Right side of screen - Messages, Input
	right := tview.NewFlex().SetDirection(tview.FlexRow)
	right.AddItem(box, 0, 15, false)
	right.AddItem(text, 0, 1, false)

	body := tview.NewFlex().SetDirection(tview.FlexColumn)
	body.AddItem(left, 0, 1, true).AddItem(right, 0, 3, false)

	// When contact is selected
	new_select := func(jid string) {
		rec, ok := parseJID(jid)
		recipient = rec
		if ok {
			box.SetTitle("Connected: " + jid)
		}
		db_check, d_ok := database[recipient]
		name_check, n_ok := name_map[recipient]
		// check if user is already in db
		if !d_ok && !n_ok && db_check == nil && name_check.Found == false {
			database[recipient] = []UserMessage{}
			name_map[recipient], err = cli.Store.Contacts.GetContact(recipient)
		}
		box.Clear()
	}
	// Filtering
	filter := func(text string) {
		filtered = make(map[types.JID]types.ContactInfo)
		list.Clear()
		usr_row = 1
		for k, v := range users {
			if strings.Contains(v.PushName, text) || strings.Contains(v.FullName, text) || strings.Contains(k.User, text) {
				filtered[k] = v
			}
		}
		for k, v := range filtered {
			if v.PushName != "" {
				list.SetCell(usr_row, 0, tview.NewTableCell(v.PushName))
				list.SetCell(usr_row, 3, tview.NewTableCell(k.User))
				list.SetCell(usr_row, 4, tview.NewTableCell(k.String()))
				usr_row++
			}
		}
	}

	// handlers
	msg := ""
	handler := func(rawEvt interface{}) {
		switch evt := rawEvt.(type) {
			case *events.Message:
				if evt.Info.Sender == recipient {
					var incoming_msg string
					img := evt.Message.GetImageMessage()
					sticker := evt.Message.GetStickerMessage()
					if img != nil {
						// if the incoming message is an image
						incoming_msg = "Image Message"
					} else if sticker != nil {
						// if sticker
						incoming_msg = "Sticker Message"
					} else {
						incoming_msg = evt.Message.GetConversation()
					}
					database[recipient] = append(database[recipient], UserMessage{name_map[recipient].PushName, incoming_msg})
					for i, s := range database[recipient] {
						box.SetCell(i, 0, tview.NewTableCell(s.name + ": "))
						box.SetCell(i, 1, tview.NewTableCell(s.text))
					}
					app.Draw()
				} else if evt.Info.Sender != recipient {
					// if the message is not from the selected user
					// then we need to add it to the database
					// and update the list
					// Add user if not in db
					if val, ok := database[evt.Info.Sender]; ok {
						database[evt.Info.Sender] = append(val, UserMessage{name_map[recipient].PushName, evt.Message.GetConversation()})
					} else {
						database[evt.Info.Sender] = []UserMessage{{name_map[recipient].PushName, evt.Message.GetConversation()}}
						name_map[evt.Info.Sender], err = cli.Store.Contacts.GetContact(evt.Info.Sender)
					}
				} else if evt.Info.MessageSource.IsFromMe {
					database[recipient] = append(database[recipient], UserMessage{"me", evt.Message.GetConversation()})
					for i, s := range database[recipient] {
						box.SetCell(i, 0, tview.NewTableCell(s.name +	": "))
						box.SetCell(i, 1, tview.NewTableCell(s.text))
					}
					app.Draw()
				}
			case *events.Receipt:
				for i, s := range database[recipient] {
					box.SetCell(i, 0, tview.NewTableCell(s.name + ": "))
					box.SetCell(i, 1, tview.NewTableCell(s.text))
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
		if event.Key() == tcell.KeyTab {
			app.SetFocus(box)
		} else if event.Key() == tcell.KeyEnter {
			app.SetFocus(text)
			row, col := list.GetSelection()
			list.GetCell(row, col).SetTextColor(tcell.ColorGreen)
			new_select(list.GetCell(row, 4).Text)
		}
		return event
	})

	box.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			app.SetFocus(text)
		}
		return event
	})

	filter_input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		filter(filter_input.GetText())
		if event.Key() == tcell.KeyEnter || event.Key() == tcell.KeyTab {
			app.SetFocus(list)
		}
		return event
	})

	text.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter && text.GetText() != "" {
			msg = text.GetText()
			text.SetText("")
			database[recipient] = append(database[recipient], UserMessage{"Me", msg})
			// u/darkhz told me I should use a goroutine for this.. No idea what that is...
			cli.SendMessage(context.Background()	, recipient, "", &waProto.Message{Conversation: proto.String(msg)})
			msg = ""
		} else if event.Key() == tcell.KeyTab {
			app.SetFocus(filter_input)
		}
		return event
	})

	if err := app.SetRoot(body, true).Run(); err != nil {
		panic(err)
	}

}
