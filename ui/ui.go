package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var UIApp *tview.Application
var UIBody *tview.Flex
var UIContactsList *tview.List
var UIMessageList *tview.List
var UISearchInput *tview.InputField
var UIMessageInputField *tview.InputField
var UIDebugPage *tview.TextArea
var UIPages *tview.Pages
var UINotificationsBox *tview.TextView
var UIHelpBox *tview.TextView
var UIHelpPage *tview.TextView
var UIModalSelector *tview.Modal

func UIInitialize() {
	UIApp = tview.NewApplication();

	UIBody = tview.NewFlex().SetDirection(tview.FlexColumn);

	/*	The items on the left side of the window */
	UIContactsList = tview.NewList().ShowSecondaryText(false);
	UISearchInput = tview.NewInputField().SetLabelWidth(0);
	UISearchInput.SetFieldBackgroundColor(tcell.ColorBlack).SetFieldTextColor(tcell.ColorWhite)
	UISearchInput.SetBorder(true).SetTitle("Search");
	// Upto here

	/*	The items on the right side of the window */
	// Firstly we need the top bar
	// This top bar will have a box for notifications and a box for miscellaneous actions
	topFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	UINotificationsBox = tview.NewTextView();
	UINotificationsBox.SetBorder(true).SetTitle("Notifications");
	UIHelpBox = tview.NewTextView();
	UIHelpBox.SetBorder(true).SetTitle("Help");
	UIHelpBox.SetText("Tab: Cycle through windows, Esc: Open menu")
	topFlex.AddItem(UINotificationsBox, 0, 1, false);
	topFlex.AddItem(UIHelpBox, 0, 1, false);

	// This is the message input field
	UIMessageInputField = tview.NewInputField().SetLabelWidth(0);
	UIMessageInputField.SetFieldBackgroundColor(tcell.ColorBlack).SetFieldTextColor(tcell.ColorWhite)
	UIMessageInputField.SetBorder(true).SetTitle("Message");

	// This is the list of messages
	UIMessageList = tview.NewList().ShowSecondaryText(true);
	UIMessageList.SetSelectedBackgroundColor(tcell.ColorWhite).SetSelectedTextColor(tcell.ColorBlack);
	// Upto here

	// Add the items to the flex
	rightFlex := tview.NewFlex().SetDirection(tview.FlexRow);
	rightFlex.AddItem(topFlex, 0, 1, false)
	rightFlex.AddItem(UIMessageList, 0, 10, false)
	rightFlex.AddItem(UIMessageInputField, 0, 1, true)
	leftFlex := tview.NewFlex().SetDirection(tview.FlexRow);
	leftFlex.AddItem(UIContactsList, 0, 10, false)
	leftFlex.AddItem(UISearchInput, 0, 1, true)

	// Add the flexes to the body
	UIBody.AddItem(leftFlex, 0, 1, false);
	UIBody.AddItem(rightFlex, 0, 4, false);

	/////////////////////////////
	//// The pages container ////
	/////////////////////////////
	UIPages = tview.NewPages();

	///////////////////
	//// Help page ////
	///////////////////
	UIHelpPage = tview.NewTextView()
	helpText := `
	Thank you for using WhatsGo.

	# Pages:
	- Home: The main page
	- Debug: The debug page
	- Modal: The modal page
	- Explorer: Browse the filesystem to send files
	- MediaSender: Preview the media before sending, add a caption.

	# Keybinds:
	- Tab: Cycle through windows
	- Esc: Open menu, Close menu
	- Enter:
		- On 'Search': Will take you to the contacts list
		- On 'Contacts': Will select the contact, and take you to the 'Message' input
		- On 'Chat': Will open the image, video or sticker if it is on a media message
		- On 'Explorer': Will open the file in the 'MediaSender' page.
		- On 'MediaSender': Will send the media with the caption

	# Features:
	- Send media: Send images, videos, documents
	- Send messages: Send text messages
	- Download media: Download images, videos
	- Search: Search for contacts

	# Features to be added:
	- Download documents
	- Send stickers

	# Notes:
	- I don't know half of what I am doing
	- Bugs will be everywhere
	`
	UIHelpPage.SetText(helpText);

	// The debug page
	UIDebugPage = tview.NewTextArea().SetPlaceholder("Debug page");

	///////////////
	//// Modal ////
	///////////////
	UIModalSelector = tview.NewModal();

	////////////////////////////////////////
	//// Add all the items to the pages ////
	////////////////////////////////////////
	UIPages.AddPage("Home", UIBody, true, true);
	UIPages.AddPage("Debug", UIDebugPage, true, true);
	UIPages.AddPage("Modal", UIModalSelector, true, true);
	UIPages.AddPage("Help", UIHelpPage, true, true);
	UIPages.SendToFront("Home")

	// Set the focus to the message input field
	UIApp.SetRoot(UIPages, true).SetFocus(UISearchInput);
}
