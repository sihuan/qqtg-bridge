package qq

import (
	"bytes"
	"fmt"
	mirai "github.com/Mrs4s/MiraiGo/message"
	"github.com/sihuan/qqtg-bridge/message"
	"io/ioutil"
	"net/http"
)

type ChatChan struct {
	bot      *Bot
	gid      int64
	tempChan chan *mirai.GroupMessage
}

func (bot *Bot) NewGroupChan(gid int64) {
	bot.Chats[gid] = ChatChan{
		bot:      bot,
		gid:      gid,
		tempChan: make(chan *mirai.GroupMessage, 20),
	}
}

func (c ChatChan) Read() *message.Message {
	msg := <-c.tempChan
	var (
		text      string
		imageURLS []string
		replyid   int64
	)
	for _, element := range msg.Elements {
		switch e := element.(type) {
		case *mirai.TextElement:
			text += e.Content + "\n"
		case *mirai.ImageElement:
			imageURLS = append(imageURLS, e.Url)
		case *mirai.AtElement:
		case *mirai.ReplyElement:
			replyid = int64(e.ReplySeq)
		default:
			text += "\n不支持的类型的消息"
		}
	}
	return &message.Message{
		Sender:    msg.Sender.Nickname,
		ImageURLs: imageURLS,
		ID:        int64(msg.Id),
		ReplyID:   replyid,
		Text:      text,
	}
}

func (c ChatChan) Write(msg *message.Message) {
	text := fmt.Sprintf("[%s]: %s", msg.Sender, msg.Text)
	sm := mirai.NewSendingMessage()
	sm.Append(mirai.NewText(text))
	if msg.ReplyID != 0 {
		if value, ok := c.bot.cache.Get(msg.ReplyID); ok {
			sm.Append(mirai.NewReply(value.(*mirai.GroupMessage)))
		}
	}
	for _, imageURL := range msg.ImageURLs {
		if resp, err := http.Get(imageURL); err == nil && resp.StatusCode == http.StatusOK {
			if imgbyte, err := ioutil.ReadAll(resp.Body); err == nil {
				if img, err := c.bot.UploadGroupImage(c.gid, bytes.NewReader(imgbyte)); err == nil {
					sm.Append(img)
				}
			}

		}
	}

	sentMsg := c.bot.SendGroupMessage(c.gid, sm)
	c.bot.cache.Add(msg.ID, sentMsg)
}
