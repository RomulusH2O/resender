package main

import (
	"gorm.io/gorm"
)

type ChatMessageEntity struct {
	gorm.Model
	Id                     string `gorm:"primaryKey"`
	ChatId                 string
	SenderNickname         string
	Text                   string
	SendTimestampSeconds   int64
	SendTimestampNanos     int32
	UploadTimestampSeconds int64
	UploadTimestampNanos   int32
	AttachmentPresent      bool
	AttachmentName         string
}
