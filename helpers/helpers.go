package helpers

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"

	whatsgotypes "github.com/WinterSunset95/WhatsGo/WhatsGoTypes"
	"github.com/WinterSunset95/WhatsGo/debug"
	"github.com/WinterSunset95/WhatsGo/waconnect"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"

	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"
)

var log waLog.Logger
var UserHomeDir string
var WhatsGoDir string
var WhatsGoDb string
var WhatsGoDbJson string
var WhatsGoMediaDir string
var WhatsGoStickerDir string
var WhatsGoUnknownDir string
var WhatsGoLogsDir string

func SetupHelpers() {
	UserHomeDir, _ = os.UserHomeDir();
	WhatsGoDir = UserHomeDir + "/.whatsgo";
	WhatsGoDb = WhatsGoDir + "/wapp.db";
	WhatsGoDbJson = WhatsGoDir + "/db.json";
	WhatsGoMediaDir = WhatsGoDir + "/media/";
	WhatsGoStickerDir = WhatsGoDir + "/stickers/";
	WhatsGoUnknownDir = WhatsGoDir + "/unknown/";
	WhatsGoLogsDir = WhatsGoDir + "/logs/";
	_ = os.MkdirAll(WhatsGoDir, 0755);
	_ = os.MkdirAll(WhatsGoMediaDir, 0755);
	_ = os.MkdirAll(WhatsGoStickerDir, 0755);
	_ = os.MkdirAll(WhatsGoUnknownDir, 0755);
	_ = os.MkdirAll(WhatsGoLogsDir, 0755);

	waconnect.CurrentChat = types.JID{};
	waconnect.WhatsGoDatabase = make(whatsgotypes.Database);
}


func ViewImage(content string) {
	// Image messages will always start with image- and end with .jpg, they are in the media folder
	// Sticker messages will always start with sticker- and end with .webp, they are in the sticker folder
	// Video messages will always start with video- and end with .mp4, they are in the media folder
	// If it is not any of these, then it is an unknown message type
	
	if strings.HasPrefix(content, "image-") && strings.HasSuffix(content, ".jpeg") {
		fehExec := exec.Command("feh", WhatsGoMediaDir + content)
		fehExec.Start()
	} else if strings.HasPrefix(content, "sticker-") && strings.HasSuffix(content, ".webp") {
		fehExec := exec.Command("feh", WhatsGoStickerDir + content)
		fehExec.Start()
	} else if strings.HasPrefix(content, "video-") && strings.HasSuffix(content, ".mp4") {
		mpvExec := exec.Command("mpv", WhatsGoMediaDir + content)
		mpvExec.Start()
	} else {
		// Unknown message type
		// Do nothing
	}
}

func PutContactsOnList(contacts map[types.JID]types.ContactInfo, list *tview.List) {
	list.Clear()
	for jid, contact := range contacts {
		list.AddItem(contact.FullName, jid.String(), 0, nil)
	}
}

func getFullPathOfMedia(url string, mimeType string, isLocal bool) (string, string, string) {
	fileIdWithExtension := strings.Split(url, "/")[5];
	var fileId string
	if isLocal {
		fileId = "whatsgo"
	} else {
		fileId = strings.Split(fileIdWithExtension, ".")[0]
	}
	prefix := strings.Split(mimeType, "/")[0];
	suffix := strings.Split(mimeType, "/")[1];
	return prefix, fileId, suffix
}

func PutMessagesToList(cli *whatsmeow.Client, database whatsgotypes.Database, jid types.JID, list *tview.List) {
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
			var fullPath string;
			var isLocal bool = false;
			if messageData.Info.ID == "whatsgo" {
				isLocal = true;
			}
			if messageData.Message.ImageMessage != nil {
				prefix, imageId, suffix := getFullPathOfMedia(*messageData.Message.ImageMessage.URL, *messageData.Message.ImageMessage.Mimetype, isLocal)
				fileName = prefix + "-" + imageId + "." + suffix;
				fullPath = WhatsGoMediaDir + fileName
			} else if messageData.Message.StickerMessage != nil {
				_, stickerId, suffix := getFullPathOfMedia(*messageData.Message.StickerMessage.URL, *messageData.Message.StickerMessage.Mimetype, isLocal)
				fileName = "sticker" + "-" + stickerId + "." + suffix;
				fullPath = WhatsGoStickerDir + fileName
			} else if messageData.Message.VideoMessage != nil {
				prefix, videoId, suffix := getFullPathOfMedia(*messageData.Message.VideoMessage.URL, *messageData.Message.VideoMessage.Mimetype, isLocal)
				fileName = prefix + "-" + videoId + "." + suffix;
				fullPath = WhatsGoMediaDir + fileName
			} else if messageData.Message.DocumentMessage !=nil {
				prefix, docId, suffix := getFullPathOfMedia(*messageData.Message.DocumentMessage.URL, *messageData.Message.DocumentMessage.Mimetype, isLocal)
				fileName = prefix + "-" + docId + "." + suffix;
				fullPath = WhatsGoMediaDir + fileName
			} else {
				fileName = "unknown-" + messageData.Info.ID + ".unknown";
				fullPath = WhatsGoUnknownDir + fileName
			}

			// Check if the file already exists
			_, fileErr := os.Stat(fullPath)
			if fileErr != nil && messageData.Info.ID != "whatsgo" {
				//////////////////////////////////////
				// File does not exist? Download it //
				//////////////////////////////////////
				go BackgroundDownloader(cli, list, fullPath, mainText, messageData)
			}

			// Add the image to the list
			list.AddItem(mainText, fileName, 0, nil);
		} else {
			list.AddItem(mainText, "Unknown message type", 0, nil);
		}
	}
}

func BackgroundDownloader(cli *whatsmeow.Client, list *tview.List, fullPath string, mainText string, messageData whatsgotypes.MessageData) {
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
	} else if messageData.Info.MediaType == "document" {
		documentByte, err := cli.Download(messageData.Message.GetDocumentMessage())
		if (err != nil) {
			list.AddItem(mainText, "Error downloading document", 0, nil);
		}
		// save the bytearray to a file
		os.WriteFile(fullPath, documentByte, 0644)
	} else {
		// Do nothing
	}
}

func ScrollToBottom(list *tview.List) {
		endKeyEvent := tcell.NewEventKey(tcell.KeyEnd, 0, tcell.ModNone)
		list.InputHandler()(endKeyEvent, nil);
}

func PushToDatabase(db whatsgotypes.Database) {
	// TODO: print 'db' to a json file
	jsonConvert, err := json.MarshalIndent(db, "", "    ");
	if err != nil {
		panic(err);
	}
	os.WriteFile(WhatsGoDbJson, jsonConvert, 0644);

	return
}

func SendTextMessage(cli *whatsmeow.Client, currentChat types.JID, text string, database whatsgotypes.Database, messageList *tview.List) {
			messageInfo := types.MessageSource{
				Chat: currentChat,
				Sender: *cli.Store.ID,
				IsFromMe: true,
			}

			currentTime := time.Now();
			messageData := whatsgotypes.MessageData{
				Info: types.MessageInfo{
					MessageSource: messageInfo,
					PushName: cli.Store.PushName,
					Timestamp: currentTime,
					Type: "text",
				},
				Message: waProto.Message{Conversation: proto.String(text)}}
			textToSend := &waE2E.Message{
				Conversation: proto.String(text),
			}

			cli.SendMessage(context.Background(), currentChat, textToSend)
			database[currentChat] = append(database[currentChat], messageData);
			PushToDatabase(database)
			PutMessagesToList(cli, database, currentChat, messageList);
}

func SendMediaMessage(app *tview.Application, filePathWithType string, fileBytes *[]byte, uploadResponse *whatsmeow.UploadResponse, mediaTitleInput *tview.InputField, messageList *tview.List) {
	mimeType := mimetype.Detect(*fileBytes)

	client := waconnect.WAClient
	currentChat := waconnect.CurrentChat
	var messageToSend *waE2E.Message
	var mediaType string = ""
	// Here we send
	if strings.HasPrefix(filePathWithType, "Document:") {
		mediaType = "document"
		documentMessage := &waE2E.DocumentMessage{
			Caption: proto.String(mediaTitleInput.GetText()),
			Mimetype: proto.String(mimeType.String()),
			URL: &uploadResponse.URL,
			DirectPath: &uploadResponse.DirectPath,
			MediaKey: uploadResponse.MediaKey,
			FileSHA256: uploadResponse.FileSHA256,
			FileEncSHA256: uploadResponse.FileEncSHA256,
			FileLength: &uploadResponse.FileLength,
			FileName: &uploadResponse.Handle,
		}
		messageToSend = &waE2E.Message{
			DocumentMessage: documentMessage,
		}
	} else if strings.HasPrefix(filePathWithType, "Video:") {
		mediaType = "video"
		videoMessage := &waE2E.VideoMessage{
			Caption: proto.String(mediaTitleInput.GetText()),
			Mimetype: proto.String("video/mp4"),
			URL: &uploadResponse.URL,
			DirectPath: &uploadResponse.DirectPath,
			MediaKey: uploadResponse.MediaKey,
			FileSHA256: uploadResponse.FileSHA256,
			FileEncSHA256: uploadResponse.FileEncSHA256,
			FileLength: &uploadResponse.FileLength,
		}
		messageToSend = &waE2E.Message{
			VideoMessage: videoMessage,
		}
	} else if strings.HasPrefix(filePathWithType, "Photo:") {
		mediaType = "image"
		imageMessage := &waE2E.ImageMessage{
			Caption: proto.String(mediaTitleInput.GetText()),
			Mimetype: proto.String("image/jpeg"),

			URL: &uploadResponse.URL,
			DirectPath: &uploadResponse.DirectPath,
			MediaKey: uploadResponse.MediaKey,
			FileEncSHA256: uploadResponse.FileEncSHA256,
			FileSHA256: uploadResponse.FileSHA256,
			FileLength: &uploadResponse.FileLength,
		}
		messageToSend = &waE2E.Message{
			ImageMessage: imageMessage,
		}
	} else {
		// Do nothing
		debug.WhatsGoPrint("Unknown file type")
	}

	messageInfo := types.MessageSource{
		Chat: currentChat,
		Sender: *client.Store.ID,
		IsFromMe: true,
	}
	currentTime := time.Now();
	messageData := whatsgotypes.MessageData{
		Info: types.MessageInfo{
			MessageSource: messageInfo,
			ID: "whatsgo",
			PushName: client.Store.PushName,
			Timestamp: currentTime,
			Type: "media",
			MediaType: mediaType,
		},
		Message: *messageToSend,
	}
	resp, _ := client.SendMessage(context.Background(), currentChat, messageToSend)
	respJson, _ := json.MarshalIndent(resp, "", "    ")
	debug.WhatsGoPrint("\nSent message Response: " + string(respJson))
	//messageJson, _ := json.MarshalIndent(messageData, "", "    ")
	waconnect.WhatsGoDatabase[currentChat] = append(waconnect.WhatsGoDatabase[currentChat], messageData);
	PushToDatabase(waconnect.WhatsGoDatabase)
	PutMessagesToList(client, waconnect.WhatsGoDatabase, currentChat, messageList);
	app.Stop()
}
