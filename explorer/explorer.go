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

func setupExplorerLists() {
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
		ExParentDir.AddItem(file.Name(), "", 0, nil)
	}
	// We need to put the selection on the current directory
	fullDirPath, _ := os.Getwd()
	fullDirPathSplit := strings.Split(fullDirPath, "/")
	_ = fullDirPathSplit
	debug.WhatsGoPrint(fullDirPathSplit[0])
	//currentDirName := fullDirPathSplit[len(fullDirPathSplit)]
	//directoryIndex := ExParentDir.FindItems(currentDirName, "", false, true)
	//ExParentDir.SetCurrentItem(directoryIndex[0])

	// Current Directory
	currentDirectoryList, err := os.ReadDir("./")
	if err != nil {
		return
	}
	for _, file := range currentDirectoryList {
		ExCurrentDir.AddItem(file.Name(), "", 0, nil)
	}
}

func loadAndSetImage(fileName string) {
		// If the file is an image
		// Load the image as bytes
		imageBytes, err := os.ReadFile(fileName)
		if err != nil {
			ExTextView.SetText("Error reading image file")
			ExPreviewPane.AddItem(ExTextView, 0, 1, false)
			return
		}
		// Decode the image
		graphics, err := jpeg.Decode(bytes.NewReader(imageBytes))
		if err != nil {
			graphics, err = png.Decode(bytes.NewReader(imageBytes))
			if err != nil {
				ExTextView.SetText("Error decoding image")
				ExPreviewPane.AddItem(ExTextView, 0, 1, false)
				return
			}
		}
		ExImageView.SetImage(graphics)
		ExPreviewPane.AddItem(ExImageView, 0, 1, false)
		ExApp.Draw()
}

func setupPreviewPane() {
	// Clear the preview pane
	ExPreviewPane.Clear()

	itemIndex := ExCurrentDir.GetCurrentItem()
	fileName, _ := ExCurrentDir.GetItemText(itemIndex)

	// Get file info
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		ExTextView.SetText("Error getting file type: " + fileName + " " + err.Error());
		ExPreviewPane.AddItem(ExTextView, 0, 1, false)
		return
	}

	if fileInfo.IsDir() {
		// If the file is a directory, show the contents of the directory
		ExListView.Clear()
		directoryContents, err := os.ReadDir(fileName)
		if err != nil {
			return
		}
		for _, file := range directoryContents {
			ExListView.AddItem(file.Name(), "", 0, nil)
		}
		ExPreviewPane.AddItem(ExListView, 0, 1, false)
		return
	}

	if strings.HasSuffix(fileName, ".png") || strings.HasSuffix(fileName, ".jpg") || strings.HasSuffix(fileName, ".jpeg") {
		go loadAndSetImage(fileName)
	} else {
		// If the file is not an image, show the contents of the file
		file, err := os.ReadFile(fileName)
		if err != nil {
			return
		}
		ExTextView.SetText(string(file))
		ExPreviewPane.AddItem(ExTextView, 0, 1, false)
	}
}

func ExplorerApp(parentApp *tview.Application) (string) {
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
	setupExplorerLists()

	///////////////////////////////////////////
	//// Handle input events on everything ////
	///////////////////////////////////////////
	ExApp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			ExApp.Stop()
		}
		return event
	})
	ExBody.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return event
	})
	ExParentDir.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return event
	})
	ExCurrentDir.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Change to the parent directory
		if event.Key() == tcell.KeyLeft {
			os.Chdir("../")
			ExParentDir.Clear()
			ExCurrentDir.Clear()
			setupExplorerLists()
			return event
		}
		// Change to the child directory, if it is a directory. If not, do nothing
		if event.Key() == tcell.KeyRight {
			selectedItem := ExCurrentDir.GetCurrentItem()
			selectedItemText, _ := ExCurrentDir.GetItemText(selectedItem)
			fileInfo, err := os.Stat(selectedItemText)
			if err != nil {
				return event
			}
			if fileInfo.IsDir() {
				os.Chdir(selectedItemText)
				ExParentDir.Clear()
				ExCurrentDir.Clear()
				setupExplorerLists()
			}
		}
		return event
	})
	ExCurrentDir.SetChangedFunc(func(i int, s1, s2 string, r rune) {
		setupPreviewPane()
	})
	ExCurrentDir.SetSelectedFunc(func(i int, s1, s2 string, r rune) {
		currentWorkingDirectory, _ := os.Getwd()
		filePath = currentWorkingDirectory + "/" + s1
		ExApp.Stop()
	})

	parentApp.Suspend(func() {
		ExApp.Run()
	})

	os.Chdir(baseDirectory)

	return filePath
}
