package explorer

import "github.com/rivo/tview"

var ExApp *tview.Application
var ExBody *tview.Flex
var ExParentDir *tview.List
var ExCurrentDir *tview.List
var ExPreviewPane *tview.Flex
var ExTextView *tview.TextView
var ExImageView *tview.Image
var ExListView *tview.List

func ExInitialize() {
	ExApp = tview.NewApplication()

	ExBody = tview.NewFlex().SetDirection(tview.FlexColumn)
	ExBody.SetBorder(true).SetTitle("Explorer")

	//////////////////////////////
	// Parent directory section //
	//////////////////////////////
	ExParentDir = tview.NewList().ShowSecondaryText(false)
	ExParentDir.SetBorder(true).SetTitle("Parent Directory")

	///////////////////////////////
	// Current directory section //
	///////////////////////////////
	ExCurrentDir = tview.NewList().ShowSecondaryText(false)
	ExCurrentDir.SetBorder(true).SetTitle("Current Directory")

	//////////////////
	// Preview pane //
	//////////////////
	ExPreviewPane = tview.NewFlex().SetDirection(tview.FlexRow)
	ExPreviewPane.SetBorder(true).SetTitle("Preview Pane")

	////////////////////////////////////////////////////////
	//// The preview pane can have one of the following ////
	//// 1. A text view, for viewing text files			////
	//// 2. A image view, for viewing image files		////
	//// 3. A list, for viewing directories				////
	////////////////////////////////////////////////////////
	ExTextView = tview.NewTextView()
	ExTextView.SetBorder(true).SetTitle("Text Preview")
	ExImageView = tview.NewImage()
	ExImageView.SetBorder(true).SetTitle("Image Preview")
	ExListView = tview.NewList()
	ExListView.SetBorder(true).SetTitle("Directory Preview")

	/////////////////////////////////////////////
	// Add the three sections to the main body //
	/////////////////////////////////////////////
	ExBody.AddItem(ExParentDir, 0, 1, false)
	ExBody.AddItem(ExCurrentDir, 0, 1, false)
	ExBody.AddItem(ExPreviewPane, 0, 1, false)

	ExApp.SetRoot(ExBody, true).SetFocus(ExCurrentDir)
}
