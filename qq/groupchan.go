package qq

import (
	"bytes"
	"errors"
	"fmt"
	mirai "github.com/Mrs4s/MiraiGo/message"
	"io/ioutil"
	"net/http"
	"qqtg-bridge/cache"
	"qqtg-bridge/config"
	"qqtg-bridge/message"
	"strings"
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
	cache.QQMID2MSG.Add(int64(msg.Id), msg)
	var (
		text      string
		imageURLS []string
		replyid   int64
	)
	for _, element := range msg.Elements {
		switch e := element.(type) {
		case *mirai.TextElement:
			text += e.Content + "\n"
		case *mirai.FaceElement:
			text += e.Name
		case *mirai.MusicShareElement:
			text += e.Title + ": " + e.MusicUrl
		case *mirai.ServiceElement:
			text += e.SubType + " " + e.Content
		case *mirai.GroupImageElement:
			if e.Flash {
				tmpUrl := "https://gchat.qpic.cn/gchatpic_new/%d/%d-1234567890-%s/0?term=3"
				tmpImg := strings.Replace(e.ImageId, "-", "", -1)
				tmpImg = strings.Replace(tmpImg, "{", "", -1)
				tmpImg = strings.Replace(tmpImg, "}", "", -1)
				tmpUrl = fmt.Sprintf(tmpUrl, config.GlobalConfig.QQ.Account, msg.GroupCode, tmpImg[:32])
				imageURLS = append(imageURLS, tmpUrl)
			} else {
				imageURLS = append(imageURLS, e.Url)
			}
		case *mirai.AtElement:
		case *mirai.ReplyElement:
			replyid = int64(e.ReplySeq)
		default:
			text += "\n不支持的类型消息"
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
		if value, ok := cache.TG2QQCache.Get(msg.ReplyID); ok {
			if groupMsg, ok := cache.QQMID2MSG.Get(value.(int64)); ok {
				sm.Append(mirai.NewReply(groupMsg.(*mirai.GroupMessage)))
			}
		}
	}
	for _, imageURL := range msg.ImageURLs {
		if img, err := c.uploadImg(imageURL); err == nil {
			sm.Append(img)
		}
	}

	sentMsg := c.bot.SendGroupMessage(c.gid, sm)
	cache.QQ2TGCache.Add(int64(sentMsg.Id), msg.ID)
	cache.TG2QQCache.Add(msg.ID, int64(sentMsg.Id))
	cache.QQMID2MSG.Add(int64(sentMsg.Id), sentMsg)
}

func (c ChatChan) uploadImg(url string) (*mirai.GroupImageElement, error) {
	resp, err := proxyClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("http get not ok")
	}
	imgbyte, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return c.bot.UploadGroupImage(c.gid, bytes.NewReader(imgbyte))
}
