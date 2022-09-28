package qq

import (
	"bytes"
	"errors"
	"fmt"
	mirai "github.com/Mrs4s/MiraiGo/message"
	"github.com/sihuan/qqtg-bridge/cache"
	"github.com/sihuan/qqtg-bridge/config"
	"github.com/sihuan/qqtg-bridge/message"
	"github.com/sihuan/qqtg-bridge/utils"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
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
			text += e.Content //+ "\n"
		case *mirai.FaceElement:
			text += "[" + e.Name + "]"
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
				// ImageId is the filename, we need it to identify gif images
				imageURLS = append(imageURLS, fmt.Sprintf("%s\n%s", tmpUrl, e.ImageId))
			} else {
				imageURLS = append(imageURLS, fmt.Sprintf("%s\n%s", e.Url, e.ImageId))
			}
		case *mirai.AtElement:
		case *mirai.ReplyElement:
			replyid = int64(e.ReplySeq)
		default:
			text += "\n不支持的类型消息\n"
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

	if msg.ReplyID != 0 {
		if value, ok := cache.TG2QQCache.Get(msg.ReplyID); ok {
			if groupMsg, ok := cache.QQMID2MSG.Get(value.(int64)); ok {
				sm.Append(mirai.NewReply(groupMsg.(*mirai.GroupMessage)))
			}
		} else {
			text = "无法定位的回复\n" + text
		}
	}
	sm.Append(mirai.NewText(text))
	for _, imageURL := range msg.ImageURLs {
		if img, err := c.uploadImg(imageURL); err == nil {
			sm.Append(img)
		} else {
			logger.WithError(err).Errorln("Image forward failed.")
		}
	}

	sentMsg := c.bot.SendGroupMessage(c.gid, sm)
	cache.QQ2TGCache.Add(int64(sentMsg.Id), msg.ID)
	cache.TG2QQCache.Add(msg.ID, int64(sentMsg.Id))
	cache.QQMID2MSG.Add(int64(sentMsg.Id), sentMsg)
}

func (c ChatChan) uploadImg(url string) (mirai.IMessageElement, error) {
	resp, err := proxyClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("http get not ok")
	}
	imgbyte, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// transcode
	suffix := filepath.Ext(url)
	name := filepath.Base(url)
	switch suffix {
	case ".mp4", ".webm":
		if e, _ := utils.FileExist("gif"); !e {
			err = os.Mkdir("gif", 0o755)
			if err != nil {
				return nil, err
			}
		}

		inf := filepath.Join("gif", name)
		outf := filepath.Join("gif", fmt.Sprintf("%s.gif", name))
		err = os.WriteFile(inf, imgbyte, 0o755)
		if err != nil {
			return nil, err
		}
		defer os.Remove(inf)
		err = ffmpeg.Input(inf).Output(outf).WithTimeout(time.Minute * 5).Run()
		if err != nil {
			return nil, err
		}
		imgbyte, err = os.ReadFile(outf)
		defer os.Remove(outf)
		if err != nil {
			return nil, err
		}
	}

	return c.bot.UploadImage(mirai.Source{SourceType: mirai.SourceGroup, PrimaryID: c.gid}, bytes.NewReader(imgbyte))
}
