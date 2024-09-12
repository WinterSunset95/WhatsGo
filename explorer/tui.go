package explorer

import "github.com/rivo/tview"

func drawExplorer() (*tview.Application, *tview.Flex, *tview.List, *tview.List, *tview.Flex, *tview.TextView, *tview.Image, *tview.List) {

	app := tview.NewApplication()

	body := tview.NewFlex().SetDirection(tview.FlexColumn)
	body.SetBorder(true).SetTitle("Explorer")

	//////////////////////////////
	// Parent directory section //
	//////////////////////////////
	parentDir := tview.NewList().ShowSecondaryText(false)
	parentDir.SetBorder(true).SetTitle("Parent Directory")

	///////////////////////////////
	// Current directory section //
	///////////////////////////////
	currentDir := tview.NewList().ShowSecondaryText(false)
	currentDir.SetBorder(true).SetTitle("Current Directory")

	//////////////////
	// Preview pane //
	//////////////////
	previewPane := tview.NewFlex().SetDirection(tview.FlexRow)
	previewPane.SetBorder(true).SetTitle("Preview Pane")

	////////////////////////////////////////////////////////
	//// The preview pane can have one of the following ////
	//// 1. A text view, for viewing text files			////
	//// 2. A image view, for viewing image files		////
	//// 3. A list, for viewing directories				////
	////////////////////////////////////////////////////////
	textView := tview.NewTextView()
	textView.SetBorder(true).SetTitle("Text Preview")
	imageView := tview.NewImage()
	imageView.SetBorder(true).SetTitle("Image Preview")
	listView := tview.NewList()
	listView.SetBorder(true).SetTitle("Directory Preview")

	/////////////////////////////////////////////
	// Add the three sections to the main body //
	/////////////////////////////////////////////
	body.AddItem(parentDir, 0, 1, false)
	body.AddItem(currentDir, 0, 1, false)
	body.AddItem(previewPane, 0, 1, false)

	app.SetRoot(body, true).SetFocus(currentDir)

	return app, body, parentDir, currentDir, previewPane, textView, imageView, listView;
	
}
