package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	misc "github.com/RomulusH2O/forbidden/resender/misc"
	pb "github.com/RomulusH2O/forbidden/resender/protob"
)

func (s *Server) RequestRespond(ctx context.Context, req *pb.FileUploadRequest) (*pb.FileUploadResponse, error) {

	return nil, nil
}

func (s *Server) TransferChunks(stream pb.ChatFileService_TransferChunksServer) error {

	base_name := misc.RandStringRunes(misc.CodeLength)
	pending_location := fmt.Sprintf("./pending/%s.%s", base_name, "raw")

	file, errCreate := os.OpenFile(pending_location, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if errCreate != nil {
		fmt.Println(errCreate)
	}

	senderNickname := "Unknown Uploader"

	for {
		chunk, errRecv := stream.Recv()

		if errRecv != nil {
			fmt.Println(errRecv.Error())
			file.Close()
			break
		}

		if chunk.Final {
			stream.SendAndClose(&emptypb.Empty{})

			file.Close()

			final_location := fmt.Sprintf("./uploaded/%s.%s", base_name, "raw")
			errRename := os.Rename(pending_location, final_location)

			if errRename != nil {
				fmt.Println(errRename)
			}

			senderNickname = chunk.GetSenderNickname()

			break
		}

		_, errWrite := file.Write(chunk.GetData())

		if errWrite != nil {
			fmt.Println(errWrite.Error())
		}
	}

	now := timestamppb.Now()

	result := ctrl.database.Create(&ChatMessageEntity{
		Id:                     misc.RandStringRunes(misc.CodeLength),
		ChatId:                 "0",
		SenderNickname:         senderNickname,
		Text:                   "0",
		SendTimestampSeconds:   now.GetSeconds(),
		SendTimestampNanos:     now.GetNanos(),
		UploadTimestampSeconds: now.GetSeconds(),
		UploadTimestampNanos:   now.GetNanos(),
		AttachmentPresent:      true,
		AttachmentName:         fmt.Sprintf("%s.%s", base_name, "raw"),
	})

	if result.Error != nil {
		fmt.Println("Sorry, database error")
	}

	return nil
}

func (s *Server) DownloadChunks(request *pb.FileDownloadRequest, stream pb.ChatFileService_DownloadChunksServer) error {

	file, errOpen := os.Open("./uploaded/" + request.GetName())

	if errOpen != nil {
		fmt.Println("Error: errOpen in DownloadChunks")
		return errOpen
	}
	defer file.Close()

	finfo, errStat := file.Stat()

	if errStat != nil {
		fmt.Println("Error: errStat in DownloadChunks")
		return errStat
	}

	fmt.Println("Size: ", finfo.Size())

	transferMd5 := misc.RandStringRunes(misc.CodeLength)

	for {
		//time.Sleep(1 * time.Second)

		b := make([]byte, 100)
		n, errRead := file.Read(b)

		if errRead != nil {

			if errRead == io.EOF && n == 0 {

				break
			}

			fmt.Println("Error: errRead in DownloadChunks")
			return errRead
		}

		errSend := stream.Send(&pb.FileContentChunk{
			Id:    transferMd5,
			Data:  b[:n],
			Final: false,
		})

		if errSend != nil {

			fmt.Println(errSend.Error())
			return errSend
		}
	}

	errSend := stream.Send(&pb.FileContentChunk{
		Id:    transferMd5,
		Data:  []byte{},
		Final: true,
	})

	if errSend != nil {

		fmt.Println(errSend.Error())
		return errSend
	}

	return nil
}

func (s *Server) KeepInfoUpdated(request *pb.KeepUpdatedRequest, stream pb.ChatFileService_KeepInfoUpdatedServer) error {

	for i, info := range ctrl.knownFiles {

		stream.Send(&pb.KnownFileInfo{
			Name:  info,
			Final: i+1 > len(ctrl.knownFiles),
		})
	}

	return nil
}

func (s *Server) UploadChatMessage(ctx context.Context, msg *pb.ChatMessage) (*emptypb.Empty, error) {

	now := timestamppb.Now()

	result := ctrl.database.Create(&ChatMessageEntity{
		Id:                     misc.RandStringRunes(misc.CodeLength),
		ChatId:                 msg.GetChatId(),
		SenderNickname:         msg.GetSenderNickname(),
		Text:                   msg.GetText(),
		SendTimestampSeconds:   msg.GetSendTimestamp().GetSeconds(),
		SendTimestampNanos:     msg.GetSendTimestamp().GetNanos(),
		UploadTimestampSeconds: now.GetSeconds(),
		UploadTimestampNanos:   now.GetNanos(),
		AttachmentPresent:      false,
		AttachmentName:         "0",
	})

	if result.Error != nil {
		fmt.Println("Sorry, database error")
		return nil, errors.New("database error")
	}

	//fmt.Println(now.GetSeconds(), now.GetNanos())

	return &emptypb.Empty{}, nil
}

func (s *Server) GetMessageSequence(stream pb.ChatFileService_GetMessageSequenceServer) error {

	var messages []ChatMessageEntity

	for {
		request, err := stream.Recv()

		if err == io.EOF {
			return nil
		}
		if err != nil {
			return nil
		}

		count := request.GetCount()

		if count > 0 {
			ctrl.database.Where("(? > upload_timestamp_seconds OR (? = upload_timestamp_seconds AND ? > upload_timestamp_nanos)) AND chat_id = ?",
				request.OlderThan.GetSeconds(), request.OlderThan.GetSeconds(), request.OlderThan.GetNanos(), request.GetChatId()).Order("upload_timestamp_seconds DESC, upload_timestamp_nanos DESC").Limit(int(count)).Find(&messages)
		} else {
			ctrl.database.Where("? < upload_timestamp_seconds OR (? = upload_timestamp_seconds AND ? < upload_timestamp_nanos) AND chat_id = ?",
				request.GetNewerThan().GetSeconds(), request.GetNewerThan().GetSeconds(), request.GetNewerThan().GetNanos(), request.GetChatId()).Find(&messages)
		}

		//fmt.Println("Number of messages: ", len(messages))

		if len(messages) > 0 {

			for _, msg := range messages {
				fmt.Println(msg.Text)
				stream.Send(&pb.ChatMessage{
					Id:                msg.Id,
					ChatId:            "0",
					Text:              msg.Text,
					SenderNickname:    msg.SenderNickname,
					SendTimestamp:     &timestamppb.Timestamp{Seconds: msg.SendTimestampSeconds, Nanos: msg.SendTimestampNanos},
					UploadTimestamp:   &timestamppb.Timestamp{Seconds: msg.UploadTimestampSeconds, Nanos: msg.UploadTimestampNanos},
					Final:             false,
					AttachmentPresent: msg.AttachmentPresent,
					AttachmentName:    msg.AttachmentName,
				})
			}
		}
		stream.Send(&pb.ChatMessage{
			Id:                "0",
			ChatId:            "0",
			Text:              "0",
			SenderNickname:    "0",
			SendTimestamp:     nil,
			UploadTimestamp:   nil,
			Final:             true,
			AttachmentPresent: false,
			AttachmentName:    "0",
		})
	}
}

func (s *Server) BroadcastSharedTypingInfo(_ *emptypb.Empty, stream pb.ChatFileService_BroadcastSharedTypingInfoServer) error {

	ctrl.typingUsersLock.Lock()

	for k, v := range ctrl.typingUsers {

		if v {
			errSend := stream.Send(&pb.UserTypingInfo{
				ClientId: k,
				Typing:   v,
				Final:    false,
			})

			if errSend != nil {
				fmt.Println("errSend in BroadcastTypingInfo")
				continue
			}
		}
	}
	ctrl.typingUsersLock.Unlock()

	return nil
}

func (s *Server) UpdateSharedTypingInfo(ctx context.Context, info *pb.UserTypingInfo) (*emptypb.Empty, error) {

	ctrl.typingUsersLock.RLock()

	ctrl.typingUsers[info.GetClientId()] = info.Typing

	ctrl.typingUsersLock.RUnlock()

	//fmt.Println(info.GetClientId(), len(ctrl.typingUsers))

	return &emptypb.Empty{}, nil
}
