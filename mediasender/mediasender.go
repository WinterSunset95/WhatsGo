package mediasender

import (
	"bytes"
	"context"
	"image/jpeg"
	"image/png"
	"os"
	"strings"

	whatsgotypes "github.com/WinterSunset95/WhatsGo/WhatsGoTypes"
	"github.com/WinterSunset95/WhatsGo/debug"
	"github.com/WinterSunset95/WhatsGo/helpers"
	"github.com/WinterSunset95/WhatsGo/waconnect"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
)

func loadFile(filePathWithType string) ([]byte, string) {
	filePath := strings.Split(filePathWithType, ":")[1]
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		debug.WhatsGoPrint("Error reading file(mediasender.go): " + err.Error())
	}
	return fileBytes, filePath
}

func MediaSender(parentApp *tview.Application, currentChat types.JID,  filePathWithType string, database whatsgotypes.Database, messageList *tview.List) {

	client := waconnect.WAClient

	app, body, mediaTitleInput, previewPane, documentPreview, imagePreview, videoPreview := mediaSenderUi()
	_ = app
	_ = body
	_ = mediaTitleInput
	_ = previewPane
	_ = documentPreview
	_ = imagePreview
	_ = videoPreview

	//////////////////////////////////////////////////////////////////
	//// filePathWithType is of the format FileType:/path/to/file ////
	//////////////////////////////////////////////////////////////////
	fileBytes, _ := loadFile(filePathWithType)
	debug.WhatsGoPrint("MediaSender() recieved the following filePathWithType: " + filePathWithType)
	var uploadResponse whatsmeow.UploadResponse
	var err error
	previewPane.Clear()
	if strings.HasPrefix(filePathWithType, "Document:") {
		uploadResponse, err = client.Upload(context.Background(), fileBytes, whatsmeow.MediaDocument)
		documentPreview.SetText(string(fileBytes))
		previewPane.AddItem(documentPreview, 0, 1, false)
		if err != nil {
			debug.WhatsGoPrint("Error uploading document(mediasender.go): " + err.Error())
		}
	} else if strings.HasPrefix(filePathWithType, "Video:") {
		uploadResponse, err = client.Upload(context.Background(), fileBytes, whatsmeow.MediaVideo)
		documentPreview.SetText(string(fileBytes))
		previewPane.AddItem(documentPreview, 0, 1, false)
		if err != nil {
			debug.WhatsGoPrint("Error uploading document(mediasender.go): " + err.Error())
		}
	} else if strings.HasPrefix(filePathWithType, "Photo:") {
		uploadResponse, err = client.Upload(context.Background(), fileBytes, whatsmeow.MediaImage)
		if err != nil {
			debug.WhatsGoPrint("Error uploading document(mediasender.go): " + err.Error())
		}
		graphics, err := jpeg.Decode(bytes.NewReader(fileBytes))
		if err != nil {
			graphics, err = png.Decode(bytes.NewReader(fileBytes))
			if err != nil {
				debug.WhatsGoPrint("Error decoding image(mediasender.go): " + err.Error())
			}
		}
		imagePreview.SetImage(graphics)
		previewPane.AddItem(imagePreview, 0, 1, false)
	} else if strings.HasPrefix(filePathWithType, "Sticker:") {
		// Do nothing
	} else {
		// Do nothing
	}

	//////////////////////
	//// Handle Input ////
	//////////////////////
	mediaTitleInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			app.Stop()
			return event
		}
		if event.Key() == tcell.KeyCtrlSpace {
			helpers.SendMediaMessage(filePathWithType, uploadResponse, mediaTitleInput, messageList)
			app.Stop()
			return event
		}

		return event
	})

	parentApp.Suspend(func() {
		app.Run()
	})

}
