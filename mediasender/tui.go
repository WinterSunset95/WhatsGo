package mediasender

import "github.com/rivo/tview"

func mediaSenderUi() (*tview.Application, *tview.Flex, *tview.InputField, *tview.Flex, *tview.TextView, *tview.Image, *tview.TextView) {
	app := tview.NewApplication()

	body := tview.NewFlex().SetDirection(tview.FlexRow)

	previewPane := tview.NewFlex().SetDirection(tview.FlexRow)
	previewPane.SetBorder(true).SetTitle("Preview Pane")

	mediaTitleInput := tview.NewInputField()
	mediaTitleInput.SetBorder(true).SetTitle("Caption")

	documentPreview := tview.NewTextView()
	documentPreview.SetBorder(true).SetTitle("Document Preview")

	imagePreview := tview.NewImage()
	imagePreview.SetBorder(true).SetTitle("Image Preview")

	videoPreview := tview.NewTextView()
	videoPreview.SetBorder(true).SetTitle("Video Preview")

	body.AddItem(previewPane, 0, 10, false)
	body.AddItem(mediaTitleInput, 0, 1, false)

	app.SetRoot(body, true).SetFocus(mediaTitleInput)

	return app, body, mediaTitleInput, previewPane, documentPreview, imagePreview, videoPreview

}
