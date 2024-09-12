package explorer

import (
	"bytes"
	"image/jpeg"
	"image/png"
	"os"
	"strings"

	"github.com/WinterSunset95/WhatsGo/debug"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func setupExplorerLists(parentDir *tview.List, currentDir *tview.List) {
	/////////////////////////////////////////////////////////////////////////
	//// Lets start the main explorer:									 ////
	//// 1. Get a list of files and directories in the parent directory  ////
	//// 2. Get a list of files and directories in the current directory ////
	//// 3. Get the first file in the current directory					 ////
	/////////////////////////////////////////////////////////////////////////

	// Parent Directory
	parentDirectoryList, err := os.ReadDir("../")
	if err != nil {
		return
	}
	for _, file := range parentDirectoryList {
		parentDir.AddItem(file.Name(), "", 0, nil)
	}
	// We need to put the selection on the current directory
	fullDirPath, _ := os.Getwd()
	fullDirPathSplit := strings.Split(fullDirPath, "/")
	_ = fullDirPathSplit
	debug.WhatsGoPrint(fullDirPathSplit[0])
	//currentDirName := fullDirPathSplit[len(fullDirPathSplit)]
	//directoryIndex := parentDir.FindItems(currentDirName, "", false, true)
	//parentDir.SetCurrentItem(directoryIndex[0])

	// Current Directory
	currentDirectoryList, err := os.ReadDir("./")
	if err != nil {
		return
	}
	for _, file := range currentDirectoryList {
		currentDir.AddItem(file.Name(), "", 0, nil)
	}
}

func loadAndSetImage(fileName string, textView *tview.TextView, imageView *tview.Image, previewPane *tview.Flex) {
		// If the file is an image
		// Load the image as bytes
		imageBytes, err := os.ReadFile(fileName)
		if err != nil {
			textView.SetText("Error reading image file")
			previewPane.AddItem(textView, 0, 1, false)
			return
		}
		// Decode the image
		graphics, err := jpeg.Decode(bytes.NewReader(imageBytes))
		if err != nil {
			graphics, err = png.Decode(bytes.NewReader(imageBytes))
			if err != nil {
				textView.SetText("Error decoding image")
				previewPane.AddItem(textView, 0, 1, false)
				return
			}
		}
		imageView.SetImage(graphics)
		previewPane.AddItem(imageView, 0, 1, false)
}

func setupPreviewPane(currentDir *tview.List, previewPane *tview.Flex, textView *tview.TextView, imageView *tview.Image, listView *tview.List) {
	// Clear the preview pane
	previewPane.Clear()

	itemIndex := currentDir.GetCurrentItem()
	fileName, _ := currentDir.GetItemText(itemIndex)

	// Get file info
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		textView.SetText("Error getting file type: " + fileName + " " + err.Error());
		previewPane.AddItem(textView, 0, 1, false)
		return
	}

	if fileInfo.IsDir() {
		// If the file is a directory, show the contents of the directory
		listView.Clear()
		directoryContents, err := os.ReadDir(fileName)
		if err != nil {
			return
		}
		for _, file := range directoryContents {
			listView.AddItem(file.Name(), "", 0, nil)
		}
		previewPane.AddItem(listView, 0, 1, false)
		return
	}

	if strings.HasSuffix(fileName, ".png") || strings.HasSuffix(fileName, ".jpg") || strings.HasSuffix(fileName, ".jpeg") {
		go loadAndSetImage(fileName, textView, imageView, previewPane)
	} else {
		// If the file is not an image, show the contents of the file
		file, err := os.ReadFile(fileName)
		if err != nil {
			return
		}
		textView.SetText(string(file))
		previewPane.AddItem(textView, 0, 1, false)
	}
}

func ExplorerApp(parentApp *tview.Application) (string) {
	// Import the views from explorer.go
	app, body, parentDir, currentDir, previewPane, textView, imageView, listView := drawExplorer()

	//////////////////////////////////////////////////
	//// Current directory. Should not be changed ////
	//////////////////////////////////////////////////
	baseDirectory, err := os.Getwd()
	if err != nil {
		return err.Error()
	}

	//////////////////////////////////////////////////////////////
	//// Declare a placeholder variable to hold the file path ////
	//////////////////////////////////////////////////////////////
	var filePath string = baseDirectory

	/////////////////////////////////////////
	//// Set the root of the application ////
	/////////////////////////////////////////
	setupExplorerLists(parentDir, currentDir)

	///////////////////////////////////////////
	//// Handle input events on everything ////
	///////////////////////////////////////////
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			app.Stop()
		}
		return event
	})
	body.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return event
	})
	parentDir.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return event
	})
	currentDir.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Change to the parent directory
		if event.Key() == tcell.KeyLeft {
			os.Chdir("../")
			parentDir.Clear()
			currentDir.Clear()
			setupExplorerLists(parentDir, currentDir)
			return event
		}
		// Change to the child directory, if it is a directory. If not, do nothing
		if event.Key() == tcell.KeyRight {
			selectedItem := currentDir.GetCurrentItem()
			selectedItemText, _ := currentDir.GetItemText(selectedItem)
			fileInfo, err := os.Stat(selectedItemText)
			if err != nil {
				return event
			}
			if fileInfo.IsDir() {
				os.Chdir(selectedItemText)
				parentDir.Clear()
				currentDir.Clear()
				setupExplorerLists(parentDir, currentDir)
			}
		}
		return event
	})
	currentDir.SetChangedFunc(func(i int, s1, s2 string, r rune) {
		setupPreviewPane(currentDir, previewPane, textView, imageView, listView)
	})
	currentDir.SetSelectedFunc(func(i int, s1, s2 string, r rune) {
		currentWorkingDirectory, _ := os.Getwd()
		filePath = currentWorkingDirectory + "/" + s1
		app.Stop()
	})
	previewPane.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return event
	})
	textView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return event
	})
	imageView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return event
	})
	listView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return event
	})

	parentApp.Suspend(func() {
		app.Run()
	})

	os.Chdir(baseDirectory)
	return filePath

}
