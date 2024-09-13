package base

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	whatsgotypes "github.com/WinterSunset95/WhatsGo/WhatsGoTypes"
	"github.com/WinterSunset95/WhatsGo/explorer"
	"github.com/WinterSunset95/WhatsGo/helpers"
	"github.com/WinterSunset95/WhatsGo/mediasender"
	"github.com/WinterSunset95/WhatsGo/ui"
	"github.com/WinterSunset95/WhatsGo/waconnect"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

func WhatsGoBase() {
	////////////////////////////////////////////////////
	//// The main folder for whatsgo is ~/.whatsgo/ ////
	////////////////////////////////////////////////////
	oldDatabase, err := os.ReadFile(helpers.WhatsGoDbJson)
	if err == nil {
		err = json.Unmarshal(oldDatabase, &waconnect.WhatsGoDatabase);
	} else {
		fmt.Println("No waconnect.WhatsGoDatabase found");
	}

	///////////////////////////////////////
	//// Connect with the whatsapp cli ////
	///////////////////////////////////////
	cli, err := waconnect.WAConnect(helpers.WhatsGoDb)
	// After the above code is executed once
	// waconnect.WAClient should be available for use... i think
	if err != nil {
		fmt.Println("Error with connection: " + err.Error());
		return
	}

	////////////////////////////////////////
	//// Constants that must not change ////
	////////////////////////////////////////
	//var fullListOfContacts map[types.JID]types.ContactInfo
	fullListOfContacts, err := cli.Store.Contacts.GetAllContacts();
	if err != nil {
		fmt.Println("Error getting contacts main.go")
	}
	fullListOfGroups, err := cli.GetJoinedGroups();
	if err != nil {
		fmt.Println("Error getting groups main.go")
	}
	////////////////////////////////////////

	contacts := listOfContacts("", fullListOfContacts, fullListOfGroups);
	///////////////////////////
	//// Initialize the UI ////
	///////////////////////////
	ui.UIInitialize()
	// After this, the UI elements should be available for use
	app := ui.UIApp
	body := ui.UIBody
	contactsList := ui.UIContactsList
	messageList := ui.UIMessageList
	searchInput := ui.UISearchInput
	messageInputField := ui.UIMessageInputField
	debugPage := ui.UIDebugPage
	pages := ui.UIPages
	notificationsBox := ui.UINotificationsBox
	miscActions := ui.UIHelpBox
	modalSelector := ui.UIModalSelector

	_ = miscActions;
	_ = body;
	_ = modalSelector;
	helpers.PutContactsOnList(contacts, contactsList)

	////////////////////////////////////////
	//// An array of the input sections ////
	////////////////////////////////////////
	sectionsArray := []tview.Primitive{searchInput, contactsList, messageList, messageInputField};
	sectionsArrayIndex := 0;
	app.SetFocus(sectionsArray[sectionsArrayIndex])

	////////////////////////////////////////
	///// Lets handle some input here //////
	////////////////////////////////////////
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Escape key to open up the menu
		if event.Key() == tcell.KeyESC {
			modalSelector.SetText("Where would you like to go? ")
			modalSelector.ClearButtons()
			modalSelector.AddButtons([]string{"Home", "Help", "Multi Line Message", "Document", "Photo", "Video", "Exit"})
			modalSelector.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				if buttonLabel == "Home" {
					ui.UIPages.SendToFront("Home")
				} else if buttonLabel == "Help" {
					// Do nothing
				} else if buttonLabel == "Multi Line Message" {
					ui.UIPages.SendToFront("Debug")
				} else if buttonLabel == "Document" {
					filePath := explorer.ExplorerApp(ui.UIApp)
					ui.UIDebugPage.SetText(filePath, true)
					ui.UIPages.SendToFront("Home")
					mediasender.MediaSender(ui.UIApp, waconnect.CurrentChat, "Document:" + filePath, waconnect.WhatsGoDatabase, messageList)
				} else if buttonLabel == "Photo" {
					filePath := explorer.ExplorerApp(ui.UIApp)
					ui.UIDebugPage.SetText(filePath, true)
					ui.UIPages.SendToFront("Home")
					mediasender.MediaSender(ui.UIApp, waconnect.CurrentChat, "Photo:" + filePath, waconnect.WhatsGoDatabase, messageList)
				} else if buttonLabel == "Video" {
					filePath := explorer.ExplorerApp(ui.UIApp)
					ui.UIDebugPage.SetText(filePath, true)
					ui.UIPages.SendToFront("Home")
					mediasender.MediaSender(ui.UIApp, waconnect.CurrentChat, "Video:" + filePath, waconnect.WhatsGoDatabase, messageList)
				} else if buttonLabel == "Exit" {
					ui.UIApp.Stop()
				} else {
					ui.UIPages.SendToFront("Modal")
				}
			})
			pages.SendToFront("Modal")
			return event
		}

		// Cycle through the sections
		if event.Key() == tcell.KeyTAB {
			if sectionsArrayIndex == len(sectionsArray) - 1 {
				sectionsArrayIndex = 0;
			} else {
				sectionsArrayIndex++;
			}
			app.SetFocus(sectionsArray[sectionsArrayIndex]);
			return event
		}

		return event
	})

	// Here is where we handle inputs on the message input field
	messageInputField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() != tcell.KeyEnter {
			// For anything that is not enter
			helpers.ScrollToBottom(messageList)
			return event;
		}

		// The user pressed enter
		// Send a text message
		text := messageInputField.GetText();
		helpers.SendTextMessage(cli, waconnect.CurrentChat, text, waconnect.WhatsGoDatabase, messageList);
		messageInputField.SetText("");
		return event
	})

	// This one can double as both the debug page and a multi-line input for sending long messages
	debugPage.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlSpace {
			// Send the message
			text := debugPage.GetText();
			helpers.SendTextMessage(cli, waconnect.CurrentChat, text, waconnect.WhatsGoDatabase, messageList);

			pages.SendToFront("Home")
			app.SetFocus(messageInputField)
		}

		if event.Key() == tcell.KeyESC {
			pages.SendToFront("Home")
			app.SetFocus(sectionsArray[sectionsArrayIndex])
		}

		helpers.ScrollToBottom(messageList)
		return event;
	})

	// The search input. Pretty straightforward
	searchInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			sectionsArrayIndex = 1;
			app.SetFocus(sectionsArray[sectionsArrayIndex])
			return event
		}

		text := searchInput.GetText();
		contacts = listOfContacts(text, fullListOfContacts, fullListOfGroups);
		helpers.PutContactsOnList(contacts, contactsList);

		return event
	})
	
	// The contacts list. Also straightforward
	contactsList.SetSelectedFunc(func(index int, userName string, userJid string, shortcut rune) {
		converted, _ := types.ParseJID(userJid);
		waconnect.CurrentChat = converted;
		helpers.PutMessagesToList(cli, waconnect.WhatsGoDatabase, waconnect.CurrentChat, messageList);
		searchInput.SetText("");
		contacts = listOfContacts("", fullListOfContacts, fullListOfGroups);
		helpers.PutContactsOnList(contacts, contactsList);
		messageList.SetTitle(" " + userName + " ");
		helpers.ScrollToBottom(messageList)
		sectionsArrayIndex = 3;
		app.SetFocus(sectionsArray[sectionsArrayIndex])
	})

	// Next is the message list.
	messageList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return event
	})
	messageList.SetSelectedFunc(func(index int, userName string, content string, shortcut rune) {
		helpers.ViewImage(content)
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
				// We don't want to handle the status messages
				jid, _ := types.ParseJID("status@broadcast");
				if evt.Info.Chat == jid {
					break
				}

				// Notify for new messages
				userName := evt.Info.PushName;
				notificationsBox.SetText(userName + " Sent a message");

				// Prepare the message data
				// We need to add the message to the waconnect.WhatsGoDatabase
				info := evt.Info;
				message := evt.Message;
				messageData := whatsgotypes.MessageData{Info: info, Message: *message};
				chatId := evt.Info.Chat;
				waconnect.WhatsGoDatabase[chatId] = append(waconnect.WhatsGoDatabase[chatId], messageData);
				helpers.PushToDatabase(waconnect.WhatsGoDatabase)
				if chatId == waconnect.CurrentChat {
					helpers.PutMessagesToList(cli, waconnect.WhatsGoDatabase, waconnect.CurrentChat, messageList);
				}

				break

			case *events.Receipt:
				// Get the jid
				userJid := evt.Chat;
				// Get the name by searching through the contacts map
				userName := "Unknown";
				if val, ok := contacts[userJid]; ok {
					userName = val.FullName;
				}

				// Get the type of the event
				// sender, Delivered, TypeRead
				evtType := evt.Type.GoString();
				if strings.Contains(evtType, "sender") {
					evtType = "Sent";
					notificationsBox.SetText("Sent to " + userName);
				} else if strings.Contains(evtType, "Delivered") {
					evtType = "Delivered";
					notificationsBox.SetText("Delivered to " + userName);
				} else if strings.Contains(evtType, "Read") {
					evtType = "Read";
					notificationsBox.SetText("Read by " + userName);
				}
				if userJid == waconnect.CurrentChat {
					messageList.SetTitle(userName + "(" + evtType + ")");
				}
				break;

			default:
				_ = evt
				break
		}

		app.Draw();
	})


	// Turn everything into a box and run the app
	contactsList.SetBorder(true).SetTitle("Contacts");
	messageList.SetBorder(true).SetTitle("Messages");
	debugPage.SetBorder(true).SetTitle("Debug");
	app.Run();
}
