package main

import (
	"encoding/json"
	"io/fs"
	"os"
	"os/exec"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
)

func viewImage(content string) {
	// Firstly lets check if it is an image
	// Image messages will always start with image- and end with .jpg
	fehExec := exec.Command("feh", "./media/" + content)
	fehExec.Start()
}

func putContactsOnList(contacts map[types.JID]types.ContactInfo, list *tview.List) {
	list.Clear()
	for jid, contact := range contacts {
		list.AddItem(contact.FullName, jid.String(), 0, nil)
	}
}

func putMessagesToList(cli *whatsmeow.Client, database Database, jid types.JID, list *tview.List) {
	currList := database[jid];

	if len(database[jid]) > 100 {
		currList = database[jid][len(database[jid])-100:];
	}

	list.Clear()
	for _, messageData := range currList {
		pushName := messageData.Info.PushName;
		if len(pushName) > 20 {
			pushName = pushName[:20] + "...";
		}
		mainText := pushName + ": " + messageData.Info.Timestamp.String();
		if messageData.Message.Conversation != nil {
			// Normal messages I think
			list.AddItem(mainText, *messageData.Message.Conversation, 0, nil);
		} else if messageData.Message.ExtendedTextMessage != nil{
			// Maybe message replies
			list.AddItem(mainText, *messageData.Message.ExtendedTextMessage.Text, 0, nil);
		} else if messageData.Info.Type == "media" {
			// Media files handler
			// Name the file
			fileName := "image-" + messageData.Info.ID + ".jpeg";
			// Check if folder exists
			_, folderErr := os.Stat("./media")
			if folderErr != nil {
				// Make new folder
				os.Mkdir("media", fs.ModePerm)
			}
			// Check if the file already exists
			_, fileErr := os.Stat("./media/" + fileName)
			if fileErr != nil {
				// File does not exist
				// Download the image
				imageByte, err := cli.Download(messageData.Message.GetImageMessage())
				if (err != nil) {
					list.AddItem(mainText, "Error downloading image", 0, nil);
					continue;
				}

				// save the bytearray to a file
				os.WriteFile("./media/" + fileName, imageByte, 0644)
			}

			// Add the image to the list
			list.AddItem(mainText, fileName, 0, nil);

		} else {
			list.AddItem(mainText, "Unknown message type", 0, nil);
		}
	}
}

func scrollToBottom(list *tview.List) {
		endKeyEvent := tcell.NewEventKey(tcell.KeyEnd, 0, tcell.ModNone)
		list.InputHandler()(endKeyEvent, nil);
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

