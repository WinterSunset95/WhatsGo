package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func drawApp() (*tview.Application, *tview.List, *tview.List, *tview.InputField, *tview.InputField, *tview.TextArea, *tview.Pages, *tview.TextView, *tview.TextView) {
	app := tview.NewApplication();

	body := tview.NewFlex().SetDirection(tview.FlexColumn);

	/*	The items on the left side of the window */
	contactsList := tview.NewList().ShowSecondaryText(false);
	searchInput := tview.NewInputField().SetLabelWidth(0);
	searchInput.SetFieldBackgroundColor(tcell.ColorBlack).SetFieldTextColor(tcell.ColorWhite)
	searchInput.SetBorder(true).SetTitle("Search");
	// Upto here

	/*	The items on the right side of the window */
	// Firstly we need the top bar
	// This top bar will have a box for notifications and a box for miscellaneous actions
	topFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	notificationsBox := tview.NewTextView();
	notificationsBox.SetBorder(true).SetTitle("Notifications");
	miscActions := tview.NewTextView();
	miscActions.SetBorder(true).SetTitle("Help");
	miscActions.SetText("Tab: Cycle through windows")
	topFlex.AddItem(notificationsBox, 0, 1, false);
	topFlex.AddItem(miscActions, 0, 1, false);

	// This is the message input field
	messageInputField := tview.NewInputField().SetLabelWidth(0);
	messageInputField.SetFieldBackgroundColor(tcell.ColorBlack).SetFieldTextColor(tcell.ColorWhite)
	messageInputField.SetBorder(true).SetTitle("Message");

	// This is the list of messages
	messageList := tview.NewList().ShowSecondaryText(true);
	messageList.SetSelectedBackgroundColor(tcell.ColorBlack).SetSelectedTextColor(tcell.ColorWhite);
	// Upto here

	// Add the items to the flex
	rightFlex := tview.NewFlex().SetDirection(tview.FlexRow);
	rightFlex.AddItem(topFlex, 0, 1, false)
	rightFlex.AddItem(messageList, 0, 10, false)
	rightFlex.AddItem(messageInputField, 0, 1, true)
	leftFlex := tview.NewFlex().SetDirection(tview.FlexRow);
	leftFlex.AddItem(contactsList, 0, 10, false)
	leftFlex.AddItem(searchInput, 0, 1, true)

	// Add the flexes to the body
	body.AddItem(leftFlex, 0, 1, false);
	body.AddItem(rightFlex, 0, 4, false);

	// The debug page
	debugPage := tview.NewTextArea().SetPlaceholder("Debug page");

	// Different pages
	pages := tview.NewPages();
	pages.AddPage("Home", body, true, true);
	pages.AddPage("Debug", debugPage, true, true);
	pages.SendToFront("Home")

	// Set the focus to the message input field
	app.SetRoot(pages, true).SetFocus(searchInput);

	// Return the app
	return app, contactsList, messageList, searchInput, messageInputField, debugPage, pages, notificationsBox, miscActions;
}
