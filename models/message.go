package models

type MessageLink struct {
	OriginalChatID    int64
	OriginalMessageID int
	ForwardedMessageID int
}