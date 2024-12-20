package main

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	misc "github.com/RomulusH2O/forbidden/resender/misc"
	pb "github.com/RomulusH2O/forbidden/resender/protob"
)

const (
	iconFilePath     = "resource/icons8-image-file-200.png"
	iconSize         = float32(32)
	uploadPIPageSize = 3
)

type genCtrl struct {
	clientId string
	nickname string

	uiCtrl
	timeCtrl
	rpcCtrl
}

type uiCtrl struct {
	uploadProgressInfos       []UploadProgressInfo
	uploadProgressInfoHeadIdx int
	knownFiles                []string
}

type timeCtrl struct {
	pullMessageTimer     *time.Ticker
	userTypingShareTimer *time.Ticker
	userTypingTrackTimer *time.Ticker

	latestMessageTimestamp *timestamppb.Timestamp
	oldestMessageTimestamp *timestamppb.Timestamp
}

type rpcCtrl struct {
	connection *grpc.ClientConn
	client     pb.ChatFileServiceClient
}

type message struct {
	username string
	content  string
}

type attachment struct {
	username string
	title    string
}

type UploadProgressInfo struct {
	Title    string
	Progress float64
	Complete bool
}

func (c *genCtrl) initClientId() {

	c.clientId = misc.RandStringRunes(misc.CodeLength)
	c.nickname = fmt.Sprintf("Unknown (id %s)", c.clientId)
}

func (c *genCtrl) initWatchOverConnection(u *ui) {

	c.watchOverConnection(u)
}

func (c *genCtrl) initTrackAndShareUserTyping(u *ui) {

	c.userTypingTrackTimer = time.NewTicker(time.Second * 1)
	go c.trackUserTyping(u)

	c.userTypingShareTimer = time.NewTicker(time.Second * 1)
	go c.shareUserTyping(u)
}

func (c *genCtrl) watchOverConnection(u *ui) {

	go func() {

		for {

			if c.connection.WaitForStateChange(context.Background(), connectivity.Connecting) {

				connState := c.connection.GetState()

				if connState != connectivity.TransientFailure {

					continue

				} else {

					u.connState.Show()
				}
			}

			//fmt.Println(Connection.GetState())
		}
	}()

	go func() {

		for {

			if c.connection.WaitForStateChange(context.Background(), connectivity.Ready) {

				connState := c.connection.GetState()

				if connState != connectivity.TransientFailure {

					continue

				} else {

					u.connState.Show()
				}
			}
		}
	}()

	go func() {

		for {

			if c.connection.WaitForStateChange(context.Background(), connectivity.Connecting) {

				connState := c.connection.GetState()

				if connState == connectivity.Ready {

					u.connState.Hide()
				}
			}
		}
	}()
}

func (c *genCtrl) shareUserTyping(u *ui) {

	for {
		<-c.userTypingTrackTimer.C

		isTyping := false

		if len(u.messageEntry.Text) > 0 {

			isTyping = true
		}

		err, _ := c.client.UpdateSharedTypingInfo(context.Background(), &pb.UserTypingInfo{
			ClientId: c.clientId,
			Typing:   isTyping,
		})

		if err != nil {
			fmt.Printf("Failed ShareUserTyping: %v\n", err)
			continue
		}
	}
}

func (c *genCtrl) trackUserTyping(u *ui) {

	var stream pb.ChatFileService_BroadcastSharedTypingInfoClient
	var errStream error

	for {
		<-c.userTypingTrackTimer.C

		u.typingStatusLabel.Hide()

		stream, errStream = c.client.BroadcastSharedTypingInfo(context.Background(), &emptypb.Empty{})

		if errStream != nil {

			fmt.Println("Error: errStream in TrackAndShareUserTyping")
			continue
		}

		info, errRecv := stream.Recv()

		if errRecv != nil {

			fmt.Println("Error: errRecv in TrackAndShareUserTyping")
			continue
		}

		/*if info.GetFinal() {

			break
		}*/

		fmt.Println(info.GetTyping(), info.ClientId, c.clientId)
		//fmt.Println("Will I show?")
		if info.GetTyping() && info.ClientId != c.clientId {
			//fmt.Println("I will show!")
			u.typingStatusLabel.Show()
		}
	}
}

func (c *genCtrl) initStayUpdatedWithServices(u *ui) {

	c.latestMessageTimestamp = timestamppb.Now()
	c.oldestMessageTimestamp = c.latestMessageTimestamp
	c.pullMessageTimer = time.NewTicker(time.Second * 3)
	go c.stayUpdatedWithChat(u)
}

func (c *genCtrl) stayUpdatedWithChat(u *ui) {

	var stream pb.ChatFileService_GetMessageSequenceClient
	var errStream error

	for {
		<-c.pullMessageTimer.C

		stream, errStream = c.client.GetMessageSequence(context.Background())

		if errStream != nil {

			fmt.Printf("Failed to proceed get messages: %v\n", errStream)
			continue
		}

		refTimestamp := c.latestMessageTimestamp

		errSend := stream.Send(&pb.ChatMessageRequest{
			ChatId:    "0",
			Count:     0,
			NewerThan: refTimestamp,
			OlderThan: nil,
		})

		if errSend != nil {
			fmt.Println("Error: errSend in StayUpdatedWithChat")
			continue
		}

		strTemps := []string{}
		nickTemps := []string{}
		boolTemps := []bool{}

		for {
			msg, errRecv := stream.Recv()

			if errRecv != nil {

				//fmt.Println(errRecv.Error())
				break
			}

			if msg.Final {

				items := u.messagesContainer.Objects

				for i, s := range strTemps {

					if boolTemps[i] {
						items = append(items, newAttachmentCell(&attachment{username: nickTemps[i], title: s}, u, c))

					} else {
						items = append(items, newMessageCell(&message{username: nickTemps[i], content: s}))

					}
				}

				u.messagesContainer.Objects = items

				u.messagesContainer.Refresh()

				//fmt.Println(MessageScroll.Offset.X, MessageScroll.Offset.Y+MessageScroll.Size().Height, Messages.Size().Height)

				if u.chatScroll.Offset.Y+u.chatScroll.Size().Height >= u.messagesContainer.Size().Height {
					u.chatScroll.ScrollToBottom()
				}

				//fmt.Println(MessageScroll.Offset.X, MessageScroll.Offset.Y+MessageScroll.Size().Height, Messages.Size().Height)

				break
			}
			if msg.AttachmentPresent {
				strTemps = append(strTemps, msg.GetAttachmentName())
				nickTemps = append(nickTemps, msg.GetSenderNickname())
				boolTemps = append(boolTemps, true)

			} else {
				strTemps = append(strTemps, msg.GetText())
				nickTemps = append(nickTemps, msg.GetSenderNickname())
				boolTemps = append(boolTemps, false)
			}
			c.latestMessageTimestamp = msg.GetUploadTimestamp()
		}
	}
}

func (c *genCtrl) loadEarlierMessages(u *ui) {

	u.loadEarlierButton.Hide()

	stream, errStream := c.client.GetMessageSequence(context.Background())

	if errStream != nil {

		fmt.Printf("Failed to proceed get messages: %v\n", errStream)
		u.loadEarlierButton.Show()
		return
	}

	refTimestamp := c.oldestMessageTimestamp

	errSend := stream.Send(&pb.ChatMessageRequest{
		ChatId:    "0",
		Count:     3,
		NewerThan: nil,
		OlderThan: refTimestamp,
	})

	if errSend != nil {

		fmt.Println("Error: errSend in StayUpdatedWithChat")
		u.loadEarlierButton.Show()
		return
	}

	strTemps := []string{}
	nickTemps := []string{}
	boolTemps := []bool{}

	for {
		msg, errRecv := stream.Recv()

		if errRecv != nil {

			//fmt.Println(errRecv.Error())
			break
		}

		if msg.Final {

			items := u.messagesContainer.Objects

			for i, s := range strTemps {

				if boolTemps[i] {

					items = append([]fyne.CanvasObject{newAttachmentCell(&attachment{username: nickTemps[i], title: s}, u, c)}, items...)

				} else {

					items = append([]fyne.CanvasObject{newMessageCell(&message{username: nickTemps[i], content: s})}, items...)
				}
			}

			u.messagesContainer.Objects = items

			u.messagesContainer.Refresh()
			//fmt.Println(MessageScroll.Offset.X, MessageScroll.Offset.Y+MessageScroll.Size().Height, Messages.Size().Height)

			if u.chatScroll.Offset.Y <= 0 {
				u.chatScroll.ScrollToTop()
			}

			//fmt.Println(MessageScroll.Offset.X, MessageScroll.Offset.Y+MessageScroll.Size().Height, Messages.Size().Height)

			u.loadEarlierButton.Show()

			break
		}
		if msg.AttachmentPresent {

			strTemps = append(strTemps, msg.GetAttachmentName())
			nickTemps = append(nickTemps, msg.GetSenderNickname())
			boolTemps = append(boolTemps, true)

		} else {
			strTemps = append(strTemps, msg.GetText())
			nickTemps = append(nickTemps, msg.GetSenderNickname())
			boolTemps = append(boolTemps, false)
		}
		c.oldestMessageTimestamp = msg.GetUploadTimestamp()
	}
}
