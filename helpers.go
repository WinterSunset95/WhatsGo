package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image/jpeg"
	"io/ioutil"
	"net/http"
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
		} else if messageData.Message.StickerMessage != nil || messageData.Message.ImageMessage != nil {
			// Fetch the image, GET request
			var url string;
			if messageData.Message.ImageMessage != nil {
				url = *messageData.Message.ImageMessage.Url;
			} else if messageData.Message.StickerMessage != nil {
				url = *messageData.Message.StickerMessage.Url;
			}

			response, err := http.Get(url)
			if err != nil {
				// Error handling.. of course
				list.AddItem(mainText, "Error fetching sticker", 0, nil);
				continue;
			}
			defer response.Body.Close();

			// Read image data
			imageData, err := ioutil.ReadAll(response.Body);
			if err != nil {
				// Error handling.. again
				list.AddItem(mainText, "Error reading sticker", 0, nil);
				continue;
			}
			imageBase64 := base64.StdEncoding.EncodeToString(imageData);
			ioutil.WriteFile("sticker.jpeg", []byte(imageBase64), 0644)

			b, err := base64.StdEncoding.DecodeString(imageBase64)
			if err != nil {
				// Error handling.. again
				list.AddItem(mainText, "Error decoding sticker", 0, nil);
				continue;
			}

			image, err := jpeg.Decode(bytes.NewReader(b));
			if err != nil {
				// Error handling.. again
				list.AddItem(mainText, err.Error(), 0, nil);
				continue;
			}

			// Make a new image view
			imageView := tview.NewImage();
			imageView.SetImage(image)

			// Add the image to the list
			list.AddItem(mainText, "", 0, func() {imageView.SetImage(image)});

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

