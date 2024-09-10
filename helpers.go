package main

import (
	"encoding/json"
	"io/fs"
	"os"
	"os/exec"
	"strings"
	"time"
	"context"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"

	"google.golang.org/protobuf/proto"
	waProto "go.mau.fi/whatsmeow/binary/proto"
)

func viewImage(content string) {
	// Image messages will always start with image- and end with .jpg, they are in the media folder
	// Sticker messages will always start with sticker- and end with .webp, they are in the sticker folder
	// Video messages will always start with video- and end with .mp4, they are in the media folder
	// If it is not any of these, then it is an unknown message type
	
	if strings.HasPrefix(content, "image-") && strings.HasSuffix(content, ".jpeg") {
		fehExec := exec.Command("feh", "./media/" + content)
		fehExec.Start()
	} else if strings.HasPrefix(content, "sticker-") && strings.HasSuffix(content, ".webp") {
		fehExec := exec.Command("feh", "./sticker/" + content)
		fehExec.Start()
	} else if strings.HasPrefix(content, "video-") && strings.HasSuffix(content, ".mp4") {
		mpvExec := exec.Command("mpv", "./media/" + content)
		mpvExec.Start()
	} else {
		// Unknown message type
		// Do nothing
	}
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
			var fileName string;
			var folderName string;
			if messageData.Message.ImageMessage != nil {
				fileName = "image-" + messageData.Info.ID + ".jpeg";
				folderName = "media";
			} else if messageData.Message.StickerMessage != nil {
				fileName = "sticker-" + messageData.Info.ID + ".webp";
				folderName = "sticker";
			} else if messageData.Message.VideoMessage != nil {
				fileName = "video-" + messageData.Info.ID + ".mp4";
				folderName = "media";
			} else {
				fileName = "unknown-" + messageData.Info.ID + ".unknown";
				folderName = "unknown";
			}
			// Check if folder exists
			_, folderErr := os.Stat(folderName)
			if folderErr != nil {
				// Make new folder
				os.Mkdir(folderName, fs.ModePerm)
			}
			// Check if the file already exists
			fullPath := "./" + folderName + "/" + fileName;
			_, fileErr := os.Stat(fullPath)
			if fileErr != nil {
				//////////////////////////////////////
				// File does not exist? Download it //
				//////////////////////////////////////
				go backgroundDownloader(cli, list, fullPath, mainText, messageData)
			}

			// Add the image to the list
			list.AddItem(mainText, fileName, 0, nil);
		} else {
			list.AddItem(mainText, "Unknown message type", 0, nil);
		}
	}
}

func backgroundDownloader(cli *whatsmeow.Client, list *tview.List, fullPath string, mainText string, messageData MessageData) {
	/////////////////////////////////////////////
	////// Handle Image, Sticker and Video //////
	/////////////////////////////////////////////
	if messageData.Message.ImageMessage != nil {
		imageByte, err := cli.Download(messageData.Message.GetImageMessage())
		if (err != nil) {
			list.AddItem(mainText, "Error downloading image", 0, nil);
		}
		// save the bytearray to a file
		os.WriteFile(fullPath, imageByte, 0644)
	} else if messageData.Message.StickerMessage != nil {
		stickerByte, err := cli.Download(messageData.Message.GetStickerMessage())
		if (err != nil) {
			list.AddItem(mainText, "Error downloading sticker", 0, nil);
		}
		// save the bytearray to a file
		os.WriteFile(fullPath, stickerByte, 0644)
	} else if messageData.Message.VideoMessage != nil {
		videoByte, err := cli.Download(messageData.Message.GetVideoMessage())
		if (err != nil) {
			list.AddItem(mainText, "Error downloading video", 0, nil);
		}
		// save the bytearray to a file
		os.WriteFile(fullPath, videoByte, 0644)
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

func sendTextMessage(cli *whatsmeow.Client, currentChat types.JID, text string, database Database, messageList *tview.List) {
			messageInfo := types.MessageSource{
				Chat: currentChat,
				Sender: *cli.Store.ID,
				IsFromMe: true,
			}

			currentTime := time.Now();
			messageData := MessageData{
				Info: types.MessageInfo{
					MessageSource: messageInfo,
					PushName: cli.Store.PushName,
					Timestamp: currentTime,
					Type: "text",
				},
				Message: waProto.Message{Conversation: proto.String(text)}}
			textToSend := &waProto.Message{
				Conversation: proto.String(text),
			}

			cli.SendMessage(context.Background(), currentChat, textToSend)
			database[currentChat] = append(database[currentChat], messageData);
			pushToDatabase(database)
			putMessagesToList(cli, database, currentChat, messageList);
}
