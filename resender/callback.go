package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"

	misc "github.com/RomulusH2O/forbidden/resender/misc"
	pb "github.com/RomulusH2O/forbidden/resender/protob"
)

func UploadButtonTapped(u *ui, c *genCtrl) {

	fileToUploadDialog := dialog.NewFileOpen(func(urc fyne.URIReadCloser, errDialogClose error) {
		FileToUploadDialogCallback(urc, errDialogClose, u, c)
	}, u.window)
	fileToUploadDialog.Show()
}

func RefreshKnownFiles(c *genCtrl) {

	stream, errStream := c.client.KeepInfoUpdated(context.Background(), &pb.KeepUpdatedRequest{})

	if errStream != nil {
		fmt.Println("Error: errStream in RefreshButtonTapped")
		return
	}

	knownFilesTmp := []string{}

	for {

		info, errRecv := stream.Recv()

		if errRecv == io.EOF {

			break
		}

		if errRecv != nil {

			fmt.Println(errRecv.Error())
			return
		}

		knownFilesTmp = append(knownFilesTmp, info.GetName())
	}
	// mutex?
	c.knownFiles = knownFilesTmp
}

func RefreshButtonTapped(u *ui, c *genCtrl) {

	RefreshKnownFiles(c)

	u.knownFilesList.Refresh()
}

func FileToUploadDialogCallback(urc fyne.URIReadCloser, errDialogClose error, u *ui, c *genCtrl) {

	if c.uploadProgressInfos == nil {
		c.uploadProgressInfos = []UploadProgressInfo{}
		c.uploadProgressInfoHeadIdx = 0
	}

	if errDialogClose != nil {
		fmt.Println("Error: errDialogClose in FileToUploadDialogCallback")
	}

	if urc == nil {
		return
	}

	fmt.Println("Path: ", urc.URI().Path())

	go func() {

		file, errOpen := os.Open(urc.URI().Path())

		if errOpen != nil {
			fmt.Println("Error: errOpen in FileToUploadDialogCallback")
			return
		}
		defer file.Close()

		finfo, errStat := file.Stat()

		if errStat != nil {
			fmt.Println("Error: errStat in FileToUploadDialogCallback")
			return
		}

		fmt.Println("Size: ", finfo.Size())

		stream, errStream := c.client.TransferChunks(context.Background())

		if errStream != nil {
			fmt.Println("Error: errStream in FileToUploadDialogCallback")
			return
		}

		c.uploadProgressInfos = append(c.uploadProgressInfos, UploadProgressInfo{
			Title:    urc.URI().Name(),
			Progress: 0.0,
			Complete: false,
		})

		idx := len(c.uploadProgressInfos) - 1

		UpdateUploadProgressPanel(u, c)

		nRead := 0

		transferMd5 := misc.RandStringRunes(misc.CodeLength)

		for {

			b := make([]byte, 100)
			n, errRead := file.Read(b)

			if errRead != nil {

				if errRead == io.EOF && n == 0 {

					break
				}

				fmt.Println("Error: errRead in FileToUploadDialogCallback")
				return
			}

			errSend := stream.Send(&pb.FileContentChunk{
				Id:    transferMd5,
				Data:  b[:n],
				Final: false,
			})

			if errSend != nil {

				fmt.Println(errSend.Error())
				return
			}

			nRead += n

			c.uploadProgressInfos[idx].Progress = float64(nRead) / float64(finfo.Size())

			UpdateUploadProgressPanel(u, c)
		}

		errSend := stream.Send(&pb.FileContentChunk{
			Id:             transferMd5,
			Data:           []byte{},
			SenderNickname: c.nickname,
			Final:          true,
		})

		if errSend != nil {

			fmt.Println(errSend.Error())
			return
		}

		stream.CloseAndRecv()

		c.uploadProgressInfos[idx].Complete = true

		UpdateUploadProgressPanel(u, c)
	}()
}

func UpdateUploadProgressPanel(u *ui, c *genCtrl) {

	if len(c.uploadProgressInfos) > 0 {

		u.uploadProgressLHeader.Show()
		u.uploadProgressRHeader.Show()

		u.uploadProgressScrollUpButton.Show()
		u.uploadProgressScrollDownButton.Show()

		idx := 0

		for i := range u.uploadProgressControls {

			idx++

			if c.uploadProgressInfoHeadIdx+i > len(c.uploadProgressInfos)-1 {
				break
			}
			u.uploadProgressEntryNames[i].SetText(fmt.Sprintf("%d", c.uploadProgressInfoHeadIdx+i))
			u.uploadProgressBars[i].SetValue(float64(c.uploadProgressInfos[c.uploadProgressInfoHeadIdx+i].Progress))
			u.uploadProgressControls[i].Show()
		}

		for _, entry := range u.uploadProgressControls[idx:] {

			entry.Hide()
		}

	} else {

		u.uploadProgressLHeader.Hide()
		u.uploadProgressRHeader.Hide()

		u.uploadProgressScrollUpButton.Hide()
		u.uploadProgressScrollDownButton.Hide()
	}
}

func UploadProgressScrollUpTapped(u *ui, c *genCtrl) {

	if c.uploadProgressInfoHeadIdx-1 > -1 {

		c.uploadProgressInfoHeadIdx--
		UpdateUploadProgressPanel(u, c)
	}
}

func UploadProgressScrollDownTapped(u *ui, c *genCtrl) {

	if c.uploadProgressInfoHeadIdx+uploadPIPageSize < len(c.uploadProgressInfos) {

		c.uploadProgressInfoHeadIdx++
		UpdateUploadProgressPanel(u, c)
	}
}

func LetInteractWithDownloadAttachmentDialog(title string, u *ui, c *genCtrl) {

	dialog.NewFolderOpen(func(uri fyne.ListableURI, errDialogClose error) {

		if errDialogClose != nil {
			fmt.Println("Error: errDialogClose in CreateRenderer")
		}

		if uri == nil {
			return
		}

		fmt.Println("Path: ", uri.Path())

		stream, errStream := c.client.DownloadChunks(context.Background(), &pb.FileDownloadRequest{Name: title})

		if errStream != nil {
			fmt.Println("Error: errStream in CreateRenderer")
			return
		}

		file, errCreate := os.OpenFile(uri.Path()+"/"+title, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)

		if errCreate != nil {
			fmt.Println(errCreate)
		}

		for {
			chunk, errRecv := stream.Recv()

			if errRecv != nil {
				fmt.Println(errRecv.Error())
				break
			}

			if chunk.Final {
				stream.CloseSend()

				//final_location := fmt.Sprintf("./uploaded/%s.%s", base_name, "raw")
				//os.Rename(pending_location, final_location)

				break
			}

			_, errWrite := file.Write(chunk.GetData())

			if errWrite != nil {
				fmt.Println(errWrite.Error())
			}
		}

	}, u.window).Show()
}
