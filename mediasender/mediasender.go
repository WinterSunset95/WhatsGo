package mediasender

import (
	"bytes"
	"context"
	"encoding/json"
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

func uploadFile(filePathWithType string, fileBytes []byte) (*whatsmeow.UploadResponse, error) {
	client := waconnect.WAClient

	if strings.HasPrefix(filePathWithType, "Document:") {
		uploadResponse, err := client.Upload(context.Background(), fileBytes, whatsmeow.MediaDocument)
		MsDocumentPreview.SetText(string(fileBytes))
		MsPreviewPane.AddItem(MsDocumentPreview, 0, 1, false)
		if err != nil {
			debug.WhatsGoPrint("Error uploading document(mediasender.go): " + err.Error())
			return &uploadResponse, err
		}
		MsDocumentPreview.SetText(string(fileBytes))
		MsPreviewPane.Clear()
		MsPreviewPane.AddItem(MsDocumentPreview, 0, 1, false)
		return &uploadResponse, nil
	} else if strings.HasPrefix(filePathWithType, "Video:") {
		uploadResponse, err := client.Upload(context.Background(), fileBytes, whatsmeow.MediaVideo)
		MsDocumentPreview.SetText(string(fileBytes))
		MsPreviewPane.AddItem(MsDocumentPreview, 0, 1, false)
		if err != nil {
			debug.WhatsGoPrint("Error uploading document(mediasender.go): " + err.Error())
			return &uploadResponse, err
		}
		MsDocumentPreview.SetText("Video previews are not supported yet")
		MsPreviewPane.Clear()
		MsPreviewPane.AddItem(MsDocumentPreview, 0, 1, false)
		return &uploadResponse, nil
	} else if strings.HasPrefix(filePathWithType, "Photo:") {
		uploadResponse, err := client.Upload(context.Background(), fileBytes, whatsmeow.MediaImage)
		uploadResponseJson, err := json.MarshalIndent(&uploadResponse, "", "    ")
		debug.WhatsGoPrint("Upload response(inside the uploadFile function): " + string(uploadResponseJson))
		if err != nil {
			debug.WhatsGoPrint("Error uploading document(mediasender.go): " + err.Error())
			return &uploadResponse, err
		}
		graphics, err := jpeg.Decode(bytes.NewReader(fileBytes))
		if err != nil {
			graphics, err = png.Decode(bytes.NewReader(fileBytes))
			if err != nil {
				debug.WhatsGoPrint("Error decoding image(mediasender.go): " + err.Error())
				return &uploadResponse, err
			}
		}
		MsImagePreview.SetImage(graphics)
		MsPreviewPane.Clear()
		MsPreviewPane.AddItem(MsImagePreview, 0, 1, false)
		return &uploadResponse, nil
	} else if strings.HasPrefix(filePathWithType, "Sticker:") {
		// Do nothing
	} else {
		// Do nothing
	}
	return &whatsmeow.UploadResponse{}, nil
}

func MediaSender(parentApp *tview.Application, currentChat types.JID,  filePathWithType string, database whatsgotypes.Database, messageList *tview.List) {
	//////////////////////////////////////////////////////////////////
	//// filePathWithType is of the format FileType:/path/to/file ////
	//////////////////////////////////////////////////////////////////
	filePath := strings.Split(filePathWithType, ":")[1]
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		debug.WhatsGoPrint("Error reading file(mediasender.go): " + err.Error())
	}

	uploadResponse, err := uploadFile(filePathWithType, fileBytes)
	if err != nil {
		debug.WhatsGoPrint("Error uploading file(mediasender.go): " + err.Error())
	}
	uploadResponseJson, err := json.MarshalIndent(&uploadResponse, "", "    ")
	debug.WhatsGoPrint("\nUpload response: " + string(uploadResponseJson))
	debug.WhatsGoPrint("" + uploadResponse.ObjectID)

	//////////////////////
	//// Handle Input ////
	//////////////////////
	MsMediaTitleInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			MsApp.Stop()
		}
		if event.Key() == tcell.KeyEnter {
			helpers.SendMediaMessage(MsApp, filePathWithType, &fileBytes, uploadResponse, MsMediaTitleInput, messageList)
		}
		return event
	})

	// Initialize UI
	parentApp.Suspend(func() {
		MsApp.Run()
	})
}
