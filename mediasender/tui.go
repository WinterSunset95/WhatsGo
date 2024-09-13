package mediasender

import "github.com/rivo/tview"

var MsApp *tview.Application
var MsBody *tview.Flex
var MsMediaTitleInput *tview.InputField
var MsPreviewPane *tview.Flex
var MsDocumentPreview *tview.TextView
var MsImagePreview *tview.Image
var MsVideoPreview *tview.TextView

func MsInitialize() {
	MsApp = tview.NewApplication()

	MsBody = tview.NewFlex().SetDirection(tview.FlexRow)

	MsPreviewPane = tview.NewFlex().SetDirection(tview.FlexRow)
	MsPreviewPane.SetBorder(true).SetTitle("Preview Pane")

	MsMediaTitleInput = tview.NewInputField()
	MsMediaTitleInput.SetBorder(true).SetTitle("Caption")

	MsDocumentPreview = tview.NewTextView()
	MsDocumentPreview.SetBorder(true).SetTitle("Document Preview")

	MsImagePreview = tview.NewImage()
	MsImagePreview.SetBorder(true).SetTitle("Image Preview")

	MsVideoPreview = tview.NewTextView()
	MsVideoPreview.SetBorder(true).SetTitle("Video Preview")

	MsBody.AddItem(MsPreviewPane, 0, 10, false)
	MsBody.AddItem(MsMediaTitleInput, 0, 1, false)

	MsApp.SetRoot(MsBody, true).SetFocus(MsMediaTitleInput)
}
