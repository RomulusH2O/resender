package main

import (
	"context"
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"google.golang.org/protobuf/types/known/timestamppb"

	misc "github.com/RomulusH2O/forbidden/resender/misc"
	pb "github.com/RomulusH2O/forbidden/resender/protob"
)

type ui struct {
	window            fyne.Window
	connState         *canvas.Text
	messagesContainer *fyne.Container
	chatScroll        *container.Scroll
	knownFilesList    *widget.List

	uiFileUpload
	uiUpperChat
	uiLowerChat
}

type uiFileUpload struct {
	uploadProgressLHeader *widget.Label
	uploadProgressRHeader *widget.Label

	uploadProgressScrollUpButton   *widget.Button
	uploadProgressScrollDownButton *widget.Button

	uploadProgressEntryNames []*widget.Label
	uploadProgressBars       []*widget.ProgressBar
	uploadProgressControls   []*fyne.Container
}

type uiUpperChat struct {
	loadEarlierButton *widget.Button
}

type uiLowerChat struct {
	typingStatusLabel *canvas.Text
	messageEntry      *widget.Entry
}

func (u *ui) makeUI(c *genCtrl) fyne.CanvasObject {

	u.connState = canvas.NewText("CONNECTION FAILURE (RETRYING...)", color.NRGBA{220, 20, 20, 255})
	u.connState.Hide()

	u.messagesContainer = container.NewVBox()
	u.chatScroll = container.NewScroll(u.messagesContainer)

	u.knownFilesList = u.makeKnownFilesList(c)
	knownFilesListPadded := container.New(layout.NewPaddedLayout(), u.knownFilesList)

	files := container.NewBorder(u.makeUploadProgressPanel(c), u.makeFileOperations(c), nil, nil, knownFilesListPadded)

	files = container.New(layout.NewMaxLayout(), canvas.NewRectangle(color.NRGBA{60, 60, 60, 255}), files)

	chat := container.NewBorder(u.makeUpperChat(c), u.makeLowerChat(c), nil, nil, u.chatScroll)

	u.connState = canvas.NewText("CONNECTION FAILURE (RETRYING...)", color.NRGBA{220, 20, 20, 255})
	u.connState.Hide()

	return container.NewBorder(u.connState, nil, nil, files, chat)
}

func (u *ui) makeKnownFilesList(c *genCtrl) *widget.List {

	res, _ := fyne.LoadResourceFromPath(iconFilePath)

	u.knownFilesList = widget.NewList(

		func() int {

			return len(c.knownFiles)

		},

		func() fyne.CanvasObject {

			iconGrid := container.New(layout.NewGridWrapLayout(fyne.NewSize(75, 75)), widget.NewIcon(res))

			return container.New(layout.NewHBoxLayout(),
				widget.NewButtonWithIcon("Download", theme.DownloadIcon(), nil),
				iconGrid, widget.NewLabel("?FILE ENTRY?"))
		},

		func(i widget.ListItemID, o fyne.CanvasObject) {

			for _, obj := range o.(*fyne.Container).Objects {

				label, targetOk := obj.(*widget.Label)
				if targetOk {
					label.SetText(c.knownFiles[i])
				}

				button, target2Ok := obj.(*widget.Button)
				if target2Ok {
					button.OnTapped = func() {

						LetInteractWithDownloadAttachmentDialog(c.knownFiles[i], u, c)
					}
				}
			}
		})

	return u.knownFilesList
}

func (u *ui) makeFileOperations(c *genCtrl) *fyne.Container {

	removeButton := widget.NewButtonWithIcon("Remove", theme.ContentRemoveIcon(), nil)
	removeButton.Hide()

	removeButtonPadded := container.New(layout.NewPaddedLayout(), removeButton)
	removeButtonMaxed := container.New(layout.NewMaxLayout(), canvas.NewRectangle(color.NRGBA{60, 60, 60, 255}), removeButtonPadded)

	refreshButton := container.New(layout.NewPaddedLayout(),
		widget.NewButtonWithIcon("File history", theme.ViewRefreshIcon(), func() { RefreshButtonTapped(u, c) }))

	searchFileEntry := widget.NewEntry()
	searchFileEntry.SetPlaceHolder("Type here to search for files...")

	fileOperations := container.New(layout.NewFormLayout(),
		layout.NewSpacer(), container.New(layout.NewHBoxLayout(), refreshButton, removeButtonMaxed),
		widget.NewButtonWithIcon("Search", theme.SearchIcon(), nil), searchFileEntry)

	return container.New(layout.NewPaddedLayout(), fileOperations)
}

func (u *ui) makeUploadProgressPanel(c *genCtrl) *fyne.Container {

	uploadProgressEntryName1 := widget.NewLabel("   ---   ")
	uploadProgressEntryName2 := widget.NewLabel("   ---   ")
	uploadProgressEntryName3 := widget.NewLabel("   ---   ")

	u.uploadProgressEntryNames = []*widget.Label{uploadProgressEntryName1, uploadProgressEntryName2, uploadProgressEntryName3}

	uploadProgressBar1 := widget.NewProgressBar()
	uploadProgressBar2 := widget.NewProgressBar()
	uploadProgressBar3 := widget.NewProgressBar()

	u.uploadProgressBars = []*widget.ProgressBar{uploadProgressBar1, uploadProgressBar2, uploadProgressBar3}

	uploadProgressBarWithButton1 := container.New(layout.NewHBoxLayout(), uploadProgressBar1, widget.NewButton("  Abort ", nil))
	uploadProgressBarWithButton2 := container.New(layout.NewHBoxLayout(), uploadProgressBar2, widget.NewButton("  Abort ", nil))
	uploadProgressBarWithButton3 := container.New(layout.NewHBoxLayout(), uploadProgressBar3, widget.NewButton("  Abort ", nil))

	uploadProgressBarWithButton1.Hide()
	uploadProgressBarWithButton2.Hide()
	uploadProgressBarWithButton3.Hide()

	u.uploadProgressControls = []*fyne.Container{uploadProgressBarWithButton1, uploadProgressBarWithButton2, uploadProgressBarWithButton3}

	u.uploadProgressLHeader = widget.NewLabel("Attachments")
	u.uploadProgressRHeader = widget.NewLabel("")

	u.uploadProgressScrollUpButton = widget.NewButtonWithIcon("", theme.MoveUpIcon(), func() { UploadProgressScrollUpTapped(u, c) })
	u.uploadProgressScrollDownButton = widget.NewButtonWithIcon("", theme.MoveDownIcon(), func() { UploadProgressScrollDownTapped(u, c) })

	uploadProgressRows := container.New(layout.NewFormLayout(),
		u.uploadProgressLHeader, u.uploadProgressRHeader,
		uploadProgressEntryName1, uploadProgressBarWithButton1,
		uploadProgressEntryName2, uploadProgressBarWithButton2,
		uploadProgressEntryName3, uploadProgressBarWithButton3)

	return container.New(layout.NewFormLayout(),
		container.New(layout.NewCenterLayout(),
			container.New(layout.NewVBoxLayout(),
				u.uploadProgressScrollUpButton, u.uploadProgressScrollDownButton)),
		uploadProgressRows)
}

func (u *ui) makeUpperChat(c *genCtrl) *fyne.Container {

	u.loadEarlierButton = widget.NewButtonWithIcon("Earlier messages", theme.HistoryIcon(), func() { c.loadEarlierMessages(u) })

	loadEarlierButtonPadded := container.New(layout.NewPaddedLayout(), u.loadEarlierButton)

	buttonRow := container.New(layout.NewHBoxLayout(), loadEarlierButtonPadded)

	buttonRowPadded := container.New(layout.NewPaddedLayout(), buttonRow)
	buttonRowWithBackground := container.New(layout.NewMaxLayout(), canvas.NewRectangle(color.NRGBA{60, 60, 60, 0}), buttonRowPadded)

	buttonRowMaxed := container.NewMax(buttonRowWithBackground)

	return container.NewCenter(buttonRowMaxed)
}

func (u *ui) makeLowerChat(c *genCtrl) *fyne.Container {

	u.messageEntry = widget.NewMultiLineEntry()
	u.messageEntry.SetPlaceHolder("Enter your message...")

	nicknameButton := container.New(layout.NewPaddedLayout(), widget.NewButtonWithIcon("Set nickname", theme.LoginIcon(), func() {

		dialog.ShowEntryDialog("Set your nickname (1-30 letters!)", "", func(nickname string) {

			if len(nickname) > 0 && len(nickname) < 31 {
				c.nickname = fmt.Sprintf("%s (id %s)", nickname, c.clientId)
				u.messageEntry.SetPlaceHolder(fmt.Sprintf("Enter your message as %s...", c.nickname))
			}
		}, u.window)
	}))

	nicknameButtonPadded := container.New(layout.NewPaddedLayout(), canvas.NewRectangle(color.NRGBA{60, 60, 60, 255}), nicknameButton)

	uploadButton := container.New(layout.NewPaddedLayout(), widget.NewButtonWithIcon("Attachment", theme.ContentAddIcon(), func() { UploadButtonTapped(u, c) }))
	uploadButtonPadded := container.New(layout.NewPaddedLayout(), canvas.NewRectangle(color.NRGBA{60, 60, 60, 255}), uploadButton)

	u.typingStatusLabel = canvas.NewText("SOMEONE IS TYPING...", color.White)
	u.typingStatusLabel.Hide()

	messageForm := container.New(layout.NewFormLayout(),
		layout.NewSpacer(), u.typingStatusLabel,
		layout.NewSpacer(), container.New(layout.NewHBoxLayout(), nicknameButtonPadded, uploadButtonPadded),
		GetSendButton(u, c), u.messageEntry)

	messageFormPadded := container.New(layout.NewPaddedLayout(), messageForm)

	return container.New(layout.NewMaxLayout(), canvas.NewRectangle(color.NRGBA{60, 60, 60, 255}), messageFormPadded)
}

func GetSendButton(u *ui, c *genCtrl) *widget.Button {

	button := widget.NewButtonWithIcon("Send", theme.MailSendIcon(), func() {

		//		go func() {

		if len(u.messageEntry.Text) < 1 {
			return
		}

		fmt.Println()
		fmt.Println(c.connection.GetState())

		text := u.messageEntry.Text

		fmt.Println(c.connection.GetState())

		_, err := c.client.UploadChatMessage(context.Background(), &pb.ChatMessage{
			Id:              misc.RandStringRunes(misc.CodeLength),
			ChatId:          "0",
			Text:            text,
			SenderNickname:  c.nickname,
			SendTimestamp:   timestamppb.Now(),
			UploadTimestamp: nil,
		})
		if err != nil {
			fmt.Printf("Failed to upload a message: %v\n", err)
			return
		}
		u.messageEntry.SetText("")
		fmt.Println("Message uploaded")
		fmt.Println(c.connection.GetState())
		fmt.Println()
		//		}()
	})

	return button
}
