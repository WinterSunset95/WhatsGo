package main

import (
	"context"

	"encoding/json"
	"fmt"
	"os"

	//"strconv"
	//"strings"
	//"time"

	"github.com/gdamore/tcell/v2"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal"
	"google.golang.org/protobuf/proto"

	//"github.com/rivo/tview"
	"go.mau.fi/whatsmeow"

	//"go.mau.fi/whatsmeow/appstate"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"

	//"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"

	//"go.mau.fi/whatsmeow/types"
	//"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	//"google.golang.org/protobuf/proto"
)

var log waLog.Logger

type MessageData struct {
	Info types.MessageInfo;
	Message waProto.Message;
};

type Database map[types.JID][]MessageData;

func WAConnect() (*whatsmeow.Client, error) {
	dbLog := waLog.Stdout("Database", "DEBUG", true);
	container, err := sqlstore.New("sqlite3", "file:wapp.db?_foreign_keys=on", dbLog)
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
		fmt.Println(client.Store.ID);
		err = client.Connect()
		if err != nil {
			return nil, err
		}
	}
	return client, nil
}

func main() {
	currentChat := types.JID{};
	database := make(Database);

	oldDatabase, err := os.ReadFile("./db.json")
	if err == nil {
		err = json.Unmarshal(oldDatabase, &database);
	} else {
		fmt.Println("No database found");
	}

	cli, err := WAConnect()
	if err != nil {
		fmt.Println("Error with connection");
		return
	}

	////////////////////////////////////////
	/// Constant that must not change //////
	////////////////////////////////////////
	fullListOfContacts, err := cli.Store.Contacts.GetAllContacts();
	if err != nil {
		fmt.Println("Error getting contacts")
		return
	}
	fullListOfGroups, err := cli.GetJoinedGroups();
	if err != nil {
		fmt.Println("Error getting groups")
		return
	}
	////////////////////////////////////////

	contacts := listOfContacts("", fullListOfContacts, fullListOfGroups);
	app, contactsList, messageList, searchInput, messageInputField, debugPage, pages := drawApp();
	putContactsOnList(contacts, contactsList)

	////////////////////////////////////////
	///// Lets handle some input here //////
	////////////////////////////////////////
	// Here is where we get the message
	messageInputField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyESC {
			pages.SendToFront("Debug")
			app.SetFocus(debugPage)
		} else if event.Key() == tcell.KeyTAB {
			app.SetFocus(searchInput);
		} else if event.Key() == tcell.KeyEnter {
			text := messageInputField.GetText();

			messageInfo := types.MessageSource{
				Chat: currentChat,
				Sender: *cli.Store.ID,
				IsFromMe: true,
			}

			messageData := MessageData{
				Info: types.MessageInfo{
					MessageSource: messageInfo,
					PushName: cli.Store.PushName,
				},
				Message: waProto.Message{Conversation: proto.String(text)}}
			textToSend := &waProto.Message{
				Conversation: proto.String(text),
			}

			cli.SendMessage(context.Background(), currentChat, textToSend)
			database[currentChat] = append(database[currentChat], messageData);
			pushToDatabase(database)
			putMessagesToList(database, currentChat, messageList);
			messageInputField.SetText("");
		}

		return event;
	})

	// The search input. Pretty straightforward
	searchInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTAB || event.Key() == tcell.KeyEnter {
			app.SetFocus(contactsList)
		} 

		text := searchInput.GetText();
		contacts = listOfContacts(text, fullListOfContacts, fullListOfGroups);
		putContactsOnList(contacts, contactsList);

		return event
	})
	
	// The contacts list. Also straightforward
	contactsList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTAB {
			app.SetFocus(messageList)
		}

		return event
	})
	contactsList.SetSelectedFunc(func(index int, userName string, userJid string, shortcut rune) {
		converted, _ := types.ParseJID(userJid);
		currentChat = converted;
		putMessagesToList(database, currentChat, messageList);

		searchInput.SetText("");
		contacts = listOfContacts("", fullListOfContacts, fullListOfGroups);
		putContactsOnList(contacts, contactsList);

		messageList.SetTitle(" " + userName + " ");
		app.SetFocus(messageInputField);
	})


	// Next is the message list.
	messageList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTAB {
			app.SetFocus(messageInputField)
		}

		return event
	})
	messageList.SetSelectedFunc(func(index int, userName string, content string, shortcut rune) {
		viewImage(content, debugPage)
	})

	// This one can double as both the debug page and a multi-line input for sending long messages
	debugPage.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlSpace {
			pages.SendToFront("Home")
			app.SetFocus(messageInputField)
		
		}

		return event;
	})



	////////////////////////////////////////
	///// We need to handle the events /////
	////////////////////////////////////////
	cli.AddEventHandler(func(event interface{}) {
		switch evt := event.(type) {
			case *events.HistorySync:
				debugPage.SetText(evt.Data.Conversations[0].Messages[0].Message.String(), true);
				break

			case *events.Message:
				jid, _ := types.ParseJID("status@broadcast");
				if evt.Info.Chat == jid {
					break
				}

				info := evt.Info;
				message := evt.Message;
				messageData := MessageData{Info: info, Message: *message};
				chatId := evt.Info.Chat;

				database[chatId] = append(database[chatId], messageData);
				pushToDatabase(database)
				if chatId == currentChat {
					putMessagesToList(database, currentChat, messageList);
				}

				break

			default:
				_ = evt
				break

		}
	})


	// Turn everything into a box and run the app
	contactsList.SetBorder(true).SetTitle("Contacts");
	messageList.SetBorder(true).SetTitle("Messages");
	debugPage.SetBorder(true).SetTitle("Debug");
	app.Run();
}
