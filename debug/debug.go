package debug

import (
	"os"
)

func WhatsGoPrint(text string) {
	////////////////////////////////////////////////
	//// Print 'text' to a file called logs.txt ////
	////////////////////////////////////////////////
	baseDirectory, _ := os.UserHomeDir()
	whatsGoHome := baseDirectory + "/.whatsgo"
	file, err := os.OpenFile(whatsGoHome + "/logs/logs.txt", os.O_APPEND | os.O_WRONLY | os.O_CREATE , 0600)
	if err != nil {
		panic(err)
	}
	if _, err := file.WriteString(text); err != nil {
		panic (err)
	}
}

