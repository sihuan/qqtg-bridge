package tg

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sihuan/qqtg-bridge/message"
)

type ChatChan struct {
	bot      *Bot
	chatid   int64
	tempChan chan *tgbotapi.Message
}

func (b *Bot) NewChatChan(chatid int64) {
	b.Chats[chatid] = ChatChan{
		bot:      b,
		chatid:   chatid,
		tempChan: make(chan *tgbotapi.Message, 20),
	}
}

func (c ChatChan) Read() *message.Message {
	var (
		imageURLs []string
		videoURLs []string
		replyid   int64
	)
	msg := <-c.tempChan
	text := msg.Text
	if msg.Caption != "" {
		text += "\n" + msg.Caption
	}
	if msg.Photo != nil {
		if imageURL, err := c.bot.GetFileDirectURL(msg.Photo[len(msg.Photo)-1].FileID); err == nil {
			imageURLs = append(imageURLs, imageURL)
		}
	}

	if msg.Sticker != nil {
		if imageURL, err := c.bot.GetFileDirectURL(msg.Sticker.FileID); err == nil {
			imageURLs = append(imageURLs, imageURL)
		}
	}
	//if msg.Video != nil {
	//	if videoURL,err := c.bot.GetFileDirectURL(msg.Video.FileID); err == nil {
	//		videoURLs = append(videoURLs, videoURL)
	//	}
	//}
	//if msg.Document != nil {
	//	switch msg.Document.MimeType {
	//	case "video/mp4":
	//		if videoURL,err := c.bot.GetFileDirectURL(msg.Document.FileID); err == nil {
	//			videoURLs = append(videoURLs, videoURL)
	//		}
	//	}
	//}
	if msg.ReplyToMessage != nil {
		replyid = int64(msg.ReplyToMessage.MessageID)
	}
	return &message.Message{
		Sender:    msg.From.FirstName,
		ImageURLs: imageURLs,
		VideoURLs: videoURLs,
		ReplyID:   replyid,
		ID:        int64(msg.MessageID),
		Text:      msg.Text,
	}
}

func (c ChatChan) Write(msg *message.Message) {
	var sendingMsg tgbotapi.Chattable
	text := fmt.Sprintf("[%s]: %s", msg.Sender, msg.Text)
	var replyTgID = 0

	if msg.ReplyID != 0 {
		if value, ok := c.bot.cache.Get(msg.ReplyID); ok {
			replyTgID = value.(int)
		}
	}

	if msg.ImageURLs != nil {
		var photos []interface{}
		for i, url := range msg.ImageURLs {
			inputMediaPhoto := tgbotapi.NewInputMediaPhoto(tgbotapi.FileURL(url))
			if i == 0 {
				inputMediaPhoto.Caption = text
			}
			photos = append(photos, inputMediaPhoto)
		}
		mediaGroupMsg := tgbotapi.NewMediaGroup(c.chatid, photos)
		if replyTgID != 0 {
			mediaGroupMsg.ReplyToMessageID = replyTgID
		}
		sendingMsg = mediaGroupMsg
	} else {
		textMsg := tgbotapi.NewMessage(c.chatid, text)
		if replyTgID != 0 {
			textMsg.ReplyToMessageID = replyTgID
		}
		sendingMsg = textMsg
	}
	sentMsg, err := c.bot.Send(sendingMsg)
	if err != nil {
		logger.Errorln(err)
	}
	c.bot.cache.Add(msg.ID, sentMsg.MessageID)
}
