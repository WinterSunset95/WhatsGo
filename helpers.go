package main

import (
	"encoding/json"
	"os"

	"github.com/rivo/tview"
	_ "go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
)

func putContactsOnList(contacts map[types.JID]types.ContactInfo, list *tview.List) {
	list.Clear()
	for jid, contact := range contacts {
		list.AddItem(contact.FullName, jid.String(), 0, nil)
	}
}

func putMessagesToList(database Database, jid types.JID, list *tview.List) {
	currList := database[jid];

	if len(database[jid]) > 15 {
		currList = database[jid][len(database[jid])-15:];
	}

	list.Clear()
	for _, messageData := range currList {
		mainText := messageData.Info.PushName + ": " + messageData.Info.Timestamp.String();
		if messageData.Message.Conversation != nil {
			list.AddItem(mainText, *messageData.Message.Conversation, 0, nil);
		} else if messageData.Message.ExtendedTextMessage != nil{
			list.AddItem(mainText, *messageData.Message.ExtendedTextMessage.Text, 0, nil);
		} else {
			list.AddItem(mainText, "Unknown message type", 0, nil);
		}
	}
}

func pushToDatabase(db Database) {
	// TODO: print 'db' to a json file
	jsonConvert, err := json.MarshalIndent(db, "", "    ");
	if err != nil {
		panic(err);
	}
	os.WriteFile("db.json", jsonConvert, 0644);

	return
}

