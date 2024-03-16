package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"go.mau.fi/whatsmeow/types"
)

func viewImage(content string, debugPage *tview.TextArea) {
	link := content
	link = strings.ReplaceAll(link, "Image: ", "");
	link = strings.ReplaceAll(link, "Sticker: ", "");
	link = strings.ReplaceAll(link, "\u0026", "&");

	response, err := http.Get(link)
	if err != nil {
		debugPage.SetText(err.Error(), true);
		return
	}

	defer response.Body.Close();

	imageData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		debugPage.SetText(err.Error(), true);
		return
	}

	ioutil.WriteFile("image.enc", []byte(imageData), 0644)

	fehExec := exec.Command("feh", "image.jpeg")
	fehExec.Start()

}

func putContactsOnList(contacts map[types.JID]types.ContactInfo, list *tview.List) {
	list.Clear()
	for jid, contact := range contacts {
		list.AddItem(contact.FullName, jid.String(), 0, nil)
	}
}

func putMessagesToList(database Database, jid types.JID, list *tview.List) {
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
			list.AddItem(mainText, *messageData.Message.Conversation, 0, nil);
		} else if messageData.Message.ExtendedTextMessage != nil{
			list.AddItem(mainText, *messageData.Message.ExtendedTextMessage.Text, 0, nil);
		} else if messageData.Info.Type == "media" {
			// Fetch the image, GET request
			var url string;
			if messageData.Info.MediaType == "image" {
				url = *messageData.Message.ImageMessage.Url;
				url = strings.ReplaceAll(url, "\u0026", "&");
				list.AddItem(mainText, "Image: " + url, 0, nil);
			} else if messageData.Info.MediaType == "sticker" {
				url = *messageData.Message.StickerMessage.Url + *messageData.Message.StickerMessage.DirectPath;
				list.AddItem(mainText, "Sticker: " + url, 0, nil);
			}

			//response, err := http.Get(url)
			//if err != nil {
			//	// Error handling.. of course
			//	list.AddItem(mainText, "Error fetching sticker", 0, nil);
			//	continue;
			//}
			//defer response.Body.Close();

			//// Read image data
			//imageData, err := ioutil.ReadAll(response.Body);
			//if err != nil {
			//	// Error handling.. again
			//	list.AddItem(mainText, "Error reading sticker", 0, nil);
			//	continue;
			//}
			//imageBase64 := base64.StdEncoding.EncodeToString(imageData);
			//ioutil.WriteFile("sticker.jpeg", []byte(imageBase64), 0644)

			//b, err := base64.StdEncoding.DecodeString(imageBase64)
			//if err != nil {
			//	// Error handling.. again
			//	list.AddItem(mainText, "Error decoding sticker", 0, nil);
			//	continue;
			//}

			//image, err := jpeg.Decode(bytes.NewReader(b));
			//if err != nil {
			//	// Error handling.. again
			//	list.AddItem(mainText, err.Error(), 0, nil);
			//	continue;
			//}

			//// Make a new image view
			//imageView := tview.NewImage();
			//imageView.SetImage(image)

			//// Add the image to the list
			//list.AddItem(mainText, "", 0, func() {imageView.SetImage(image)});

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

