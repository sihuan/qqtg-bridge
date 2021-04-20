package tg

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sihuan/qqtg-bridge/cache"
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

	if msg.ReplyToMessage != nil {
		replyid = int64(msg.ReplyToMessage.MessageID)
	}
	return &message.Message{
		Sender:    msg.From.FirstName,
		ImageURLs: imageURLs,
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
		if value, ok := cache.QQ2TGCache.Get(msg.ReplyID); ok {
			replyTgID = int(value.(int64))
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
	cache.TG2QQCache.Add(int64(sentMsg.MessageID),msg.ID)
	cache.QQ2TGCache.Add(msg.ID,int64(sentMsg.MessageID))

}
