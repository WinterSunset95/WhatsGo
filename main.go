// the top of my laptop screen is broken
// i'm leaving this empty space here so I can see code at the top
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	//"strconv"
	"strings"
	//"time"

	"github.com/gdamore/tcell/v2"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal"
	"github.com/rivo/tview"
	"go.mau.fi/whatsmeow"

	//"go.mau.fi/whatsmeow/appstate"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
)

var log waLog.Logger

type UserMessage struct {
	name string
	text string
}

func WAConnect() (*whatsmeow.Client, error) {
	store.DeviceProps.RequireFullSync = proto.Bool(true)
	fmt.Println(store.DeviceProps.Os)
	fmt.Println(store.DeviceProps.GetRequireFullSync())
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
	// If user selects a message to reply to, this will be set to the message ID
	msgReplyTo := &events.Message{}
	// map holding the JID with an array of messages
	var newDb = make(map[types.JID][]events.Message)
	// First read db.json file to see if there is any data
	content, err := os.ReadFile("./db.json")
	if err != nil {
		fmt.Println("Error reading db.json file")
	}
	json.Unmarshal(content, &newDb)
	// map holding the JID with the username
	var name_map = make(map[types.JID]types.ContactInfo)

	var recipient types.JID

	// putting my test number for now
	cli, err := WAConnect()
	if err != nil {
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
	// Initial render of the contacts
	// Will get overwritten when searching
	for k, v := range users {
		if v.PushName != "" {
			list.SetCell(usr_row, 0, tview.NewTableCell(v.FullName))
			list.SetCell(usr_row, 1, tview.NewTableCell(v.PushName))
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

	// Modal view
	modal := tview.NewModal().SetText("Go To Page")
	modal.AddButtons([]string{"Logs", "Settings", "Logout"})

	// Header
	notifs := tview.NewTextView().SetText("Notifications")
	notifs.SetBorder(true)
	commands := tview.NewTextView().SetText("Commands")
	commands.SetBorder(true)
	pagesBtn := tview.NewTextView()
	pagesBtn.SetText("Pages").SetBorder(true)
	header := tview.NewFlex().SetDirection(tview.FlexColumn)
	header.AddItem(notifs, 0, 10, false)
	header.AddItem(commands, 0, 10, false)
	header.AddItem(pagesBtn, 0, 1, false)

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
	left.AddItem(list, 0, 15, false)
	left.AddItem(filter_input, 0, 1, true)

	// Right side of screen - Messages, Input
	right := tview.NewFlex().SetDirection(tview.FlexRow)
	right.AddItem(header, 0, 1, false)
	right.AddItem(box, 0, 10, false)
	right.AddItem(text, 0, 1, false)

	// Error logs
	logText := tview.NewTextArea()
	logText.SetBorder(true)
	logText.SetBorderColor(tcell.ColorRed).SetTitle("Error logs for debugging")
	logs := tview.NewFlex().SetDirection(tview.FlexRow)
	logs.AddItem(header, 0, 1, false)
	logs.AddItem(logText, 0, 15, true)

	// The main body
	body := tview.NewFlex().SetDirection(tview.FlexColumn)
	body.AddItem(left, 0, 1, true).AddItem(right, 0, 3, false)

	pages := tview.NewPages()
	pages.AddPage("WhatsGo", body, true, true)
	pages.AddPage("Logs", logs, true, false)

	// Export database to json file
	exportDb := func() {
		jsonType, err := json.MarshalIndent(newDb, "", "	")
		if err != nil {
			logText.SetText("Error marshalling json: "+err.Error(), true)
		} else {
			os.WriteFile("db.json", jsonType, 0644)
		}
	}

	// When contact is selected
	new_select := func(jid string) {
		rec, ok := parseJID(jid)
		recipient = rec
		if ok {}
		db_check, d_ok := newDb[recipient]
		name_check, n_ok := name_map[recipient]
		// check if user is already in db
		if !d_ok && !n_ok && db_check == nil && name_check.Found == false {
			newDb[recipient] = []events.Message{}
			name_map[recipient], err = cli.Store.Contacts.GetContact(recipient)
		} else {
			name, _ := cli.Store.Contacts.GetContact(rec)
			box.SetTitle(name.PushName)
		}
		box.Clear()
	}

	// Filtering
	filter := func(text string) {
		list.Clear()
		usr_row = 1
		filtered = make(map[types.JID]types.ContactInfo)
		filtered_groups := make(map[types.JID]types.GroupInfo)
		// Filtering contacts
		for k, v := range users {
			if strings.Contains(v.PushName, text) || strings.Contains(v.FullName, text) || strings.Contains(k.User, text) {
				filtered[k] = v
			}
		}
		// Filtering groups
		for _, v := range groups {
			if strings.Contains(v.GroupName.Name, text) {
				filtered_groups[v.JID] = *v
			}
		}
		// Render filtered contacts
		for k, v := range filtered {
			if v.PushName != "" {
				list.SetCell(usr_row, 0, tview.NewTableCell(v.PushName))
				list.SetCell(usr_row, 1, tview.NewTableCell(v.FullName))
				list.SetCell(usr_row, 3, tview.NewTableCell(k.User))
				list.SetCell(usr_row, 4, tview.NewTableCell(k.String()))
				usr_row++
			}
		}
		// Render filtered groups
		for k, v := range filtered_groups {
			list.SetCell(usr_row, 0, tview.NewTableCell(v.GroupName.Name))
			list.SetCell(usr_row, 3, tview.NewTableCell(k.User))
			list.SetCell(usr_row, 4, tview.NewTableCell(v.JID.String()))
			usr_row++
		}
	}

	setCell := func(r int, str string) {
		box.SetCell(r, 1, tview.NewTableCell(str).SetMaxWidth(300).SetExpansion(1))
	}

	render_messages := func() {
		toRender := make(map[types.JID][]events.Message)
		for i := 50; i > 0; i-- {
			if len(newDb[recipient]) > i {
				toRender[recipient] = append(toRender[recipient], newDb[recipient][len(newDb[recipient])-i])
			}
		}
		for i, s := range toRender[recipient] {
			// Check if message is from me
			if s.Info.MessageSource.IsFromMe {
				box.SetCell(i, 0, tview.NewTableCell("Me" + ": "))
			} else {
				box.SetCell(i, 0, tview.NewTableCell(s.Info.PushName + ": "))
			}
			// Check message type and act accordingly
			if s.Info.MediaType == "" {
				// Check if message is a reply
				field := s.Message.ExtendedTextMessage
				if field != nil {
					// Message is replying to a text message
					if field.ContextInfo.QuotedMessage.Conversation != nil {
						setCell(i, ("\"" + *field.ContextInfo.QuotedMessage.Conversation + "\"" + " <- " + *field.Text))
					} else {
					// Message is replying to a media message
						setCell(i, ("\"" + "UNKNOWN TYPE" + "\"" + " <- " + *field.Text))
					}
				} else {
				// Message is a normal text message
					setCell(i, s.Message.GetConversation())
				}
			} else if s.Info.MediaType == "image" {
				// Message is an image
				result_message := "IMAGE MESSAGE"
				img := s.Message.GetImageMessage()
				_, err := cli.Download(img)
				if err != nil {
					result_message = "IMAGE LOAD ERROR"
				}
				setCell(i, result_message)
			} else if s.Info.MediaType == "sticker" {
				// Message is a sticker
				result_message := "STICKER MESSAGE"
				img := s.Message.GetStickerMessage()
				_, err := cli.Download(img)
				if err != nil {
					result_message = "STICKER LOAD ERROR"
				}
				setCell(i, result_message)
			} else {
				// Message is an unknown type
				logText.SetText("Unknown message type: "+s.Info.MediaType, true)
			}
			// Put the message id in column 100
			box.SetCell(i, 100, tview.NewTableCell(s.Info.ID))
		}
	}

	msg := ""
	// handlers
	handler := func(rawEvt interface{}) {
		switch evt := rawEvt.(type) {
			// History Sync event
			case *events.HistorySync:
				logs.SetBorder(false)
			// Incoming message
			case *events.Message:
				if val, ok := newDb[evt.Info.Chat]; ok {
					newDb[evt.Info.Chat] = append(val, *evt)
				} else {
					newDb[evt.Info.Chat] = []events.Message{*evt}
				}
				render_messages()
				exportDb()
				app.Draw()
			// Contact update
			case *events.Receipt:
				render_messages()
				exportDb()
				if evt.Type == events.ReceiptTypeDelivered {
					notifs.SetText("Delivered to " + evt.Chat.User)
				} else if evt.Type == events.ReceiptTypeRead {
					notifs.SetText("Read by " + evt.Chat.User)
				} else if evt.MessageSource.IsFromMe {
				}
				app.Draw()
		}
	}

	cli.AddEventHandler(handler)

	// Input handlers
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			app.SetFocus(pagesBtn)
			commands.SetText("Ctrl+Underscore, Enter")
		} else if event.Key() == tcell.KeyEnter {
			app.SetFocus(text)
			commands.SetText("Ctrl+Underscore, Enter, Tab")
			row, col := list.GetSelection()
			list.GetCell(row, col).SetTextColor(tcell.ColorGreen)
			new_select(list.GetCell(row, 4).Text)
			render_messages()
		}
		return event
	})

	pagesBtn.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			app.SetFocus(box)
			commands.SetText("Ctrl+Underscore, Tab, Enter")
		} else if event.Key() == tcell.KeyEnter {
			commands.SetText("Ctrl+Underscore, Enter")
			body.AddItem(modal, 0, 0, false)
			app.SetFocus(modal)
		}
		return event
	})

	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Logs" {
			pages.ShowPage("Logs")
			commands.SetText("Esc")
			app.SetFocus(logText)
		} else {
			notifs.SetText("Page not implemented yet")
			app.SetFocus(text)
		}
		body.RemoveItem(modal)
	})

	box.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			app.SetFocus(text)
			commands.SetText("Ctrl+Underscore, Enter, Tab")
		} else if event.Key() == tcell.KeyEnter {
			// When selecting text u want to reply to
			row, _ := box.GetSelection()
			replyTo := box.GetCell(row, 100).Text
			notifs.SetText(box.GetCell(row, 100).Text)
			for i := range newDb[recipient] {
				if newDb[recipient][i].Info.ID == replyTo {
					msgReplyTo = &newDb[recipient][i]
				}
			}
			app.SetFocus(text)
			commands.SetText("Ctrl+Underscore, Enter, Tab")
		}
		return event
	})

	filter_input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		filter(filter_input.GetText())
		if event.Key() == tcell.KeyEnter || event.Key() == tcell.KeyTab {
			app.SetFocus(list)
			commands.SetText("Ctrl+Underscore, Enter, Tab")
		}
		return event
	})

// The following is code for sending a reply
//				if evt.Info.Sender == recipient {
//					newMsg := &waProto.Message{
//						ExtendedTextMessage: &waProto.ExtendedTextMessage{
//							Text: proto.String("A reply"),
//							ContextInfo: &waProto.ContextInfo{
//								StanzaId: proto.String(evt.Info.ID),
//								Participant: proto.String(evt.Info.Sender.String()),
//								QuotedMessage: evt.Message,
//							},
//						},
//					}
//					cli.SendMessage(context.Background(), recipient, "", newMsg)
//				}

	text.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter && text.GetText() != "" {
			msg = text.GetText()
			text.SetText("")
			// Build a new message
			newMsg := events.Message{}
			if msgReplyTo.Message != nil {
				newMsg.Message = &waProto.Message{
					ExtendedTextMessage: &waProto.ExtendedTextMessage{
						Text: proto.String(msg),
						ContextInfo: &waProto.ContextInfo{
							StanzaId: proto.String(msgReplyTo.Info.ID),
							Participant: proto.String(msgReplyTo.Info.Sender.String()),
							QuotedMessage: msgReplyTo.Message,
						},
					},
				}
				msgReplyTo = &events.Message{}
			} else {
				newMsg.Message = &waProto.Message{
					Conversation: proto.String(msg),
				}
			}
			newMsg.Info.MessageSource.IsFromMe = true
			newDb[recipient] = append(newDb[recipient], newMsg)

			// u/darkhz told me I should use a goroutine for this.. No idea what that is...
			cli.SendMessage(context.Background()	, recipient, "", newMsg.Message)
			msg = ""
		} else if event.Key() == tcell.KeyTab {
			app.SetFocus(filter_input)
			commands.SetText("Ctrl+Underscore, Enter, Tab")
		}
		return event
	})

	body.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlUnderscore {
			pages.ShowPage("Logs")
			commands.SetText("Esc")
		}
		return event
	})

	logs.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			pages.HidePage("Logs")
			commands.SetText("Ctrl+Underscore, Enter, Tab")
		}
		return event
	})

	if err := app.SetRoot(pages, true).Run(); err != nil {
		panic(err)
	}

}
