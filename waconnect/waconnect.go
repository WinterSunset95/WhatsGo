package waconnect

import (
	"context"
	"fmt"
	"os"

	/////////////////////////////////
	//// Do NOT remove this line ////
	whatsgotypes "github.com/WinterSunset95/WhatsGo/WhatsGoTypes"
	_ "github.com/mattn/go-sqlite3"
	/////////////////////////////////

	"github.com/mdp/qrterminal"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
)

var WAClient *whatsmeow.Client
var CurrentChat types.JID
var WhatsGoDatabase whatsgotypes.Database


func WAConnect(whatsGoDb string) (*whatsmeow.Client, error) {
	dbLog := waLog.Stdout("Database", "DEBUG", true);
	container, err := sqlstore.New(context.Background(),"sqlite3", "file:" + whatsGoDb + "?_foreign_keys=on", dbLog)
	if err != nil {
		return nil, err
	}
	deviceStore, err := container.GetFirstDevice(context.Background())
	if err != nil {
		panic(err)
	}
	WAClient = whatsmeow.NewClient(deviceStore, waLog.Noop)
	client := WAClient
	if client.Store.ID == nil {
		qrChan, err := client.GetQRChannel(context.Background())
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		err = client.Connect()
		if err != nil {
			return nil, err
		}
		for evt := range qrChan {
			fmt.Println("QR event: ", evt.Event)
			if evt.Event == "code" {
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
			} else {
				fmt.Println("Login event: ", evt.Event)
			}
		}
	} else {
		fmt.Println("Logged in: ", client.Store.ID);
		err = client.Connect()
		if err != nil {
			return nil, err
		}
	}
	return client, nil
}

