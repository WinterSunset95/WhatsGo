package whatsgotypes

import (
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
) 


type MessageData struct {
	Info types.MessageInfo;
	Message waE2E.Message;
};

type Database map[types.JID][]MessageData;

