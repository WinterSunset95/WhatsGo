package waconnect

import (
	"context"
	"fmt"
	"os"

	/////////////////////////////////
	//// Do NOT remove this line ////
	_ "github.com/mattn/go-sqlite3"
	/////////////////////////////////
	
	"github.com/mdp/qrterminal"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

var WAClient *whatsmeow.Client

func WAConnect(whatsGoDb string) (*whatsmeow.Client, error) {
	dbLog := waLog.Stdout("Database", "DEBUG", true);
	container, err := sqlstore.New("sqlite3", "file:" + whatsGoDb + "?_foreign_keys=on", dbLog)
	if err != nil {
		return nil, err
	}
	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		panic(err)
	}
	WAClient = whatsmeow.NewClient(deviceStore, waLog.Noop)
	client := WAClient
	if client.Store.ID == nil {
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			return nil, err
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
			} else {
				fmt.Println("Login event: ", evt.Event)
			}
		}
	} else {
		fmt.Println(client.Store.ID);
		err = client.Connect()
		if err != nil {
			return nil, err
		}
	}
	return client, nil
}

