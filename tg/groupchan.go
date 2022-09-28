package tg

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sihuan/qqtg-bridge/cache"
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
			// webp && webm
			imageURLs = append(imageURLs, imageURL)
		}
	}

	//if msg.Video != nil {
	//	if imageURL, err := c.bot.GetFileDirectURL(msg.Video.FileID); err == nil {
	//		// video will appear as video/mp4 here
	//		logger.Infof("video %s %s", msg.Video.MimeType, imageURL)
	//	}
	//}

	if msg.Document != nil {
		if docURL, err := c.bot.GetFileDirectURL(msg.Document.FileID); err == nil {
			// gif will appear as video/mp4 here
			if msg.Document.MimeType == "video/mp4" {
				imageURLs = append(imageURLs, docURL)
			}
		}
	}

	if msg.ReplyToMessage != nil {
		replyid = int64(msg.ReplyToMessage.MessageID)
	}
	if text == "" && len(imageURLs) == 0 {
		text = "不支持的类型消息"
	}
	return &message.Message{
		Sender:    msg.From.FirstName,
		ImageURLs: imageURLs,
		ReplyID:   replyid,
		ID:        int64(msg.MessageID),
		Text:      text,
	}
}

func (c ChatChan) Write(msg *message.Message) {
	var sendingMsg tgbotapi.Chattable
	text := fmt.Sprintf("[%s]: %s", msg.Sender, msg.Text)
	var replyTgID = 0

	if msg.ReplyID != 0 {
		if value, ok := cache.QQ2TGCache.Get(msg.ReplyID); ok {
			replyTgID = int(value.(int64))
		} else {
			text = "无法定位的回复\n" + text
		}
	}

	var cacheFile []string
	if msg.ImageURLs != nil {
		var photos []interface{}
		for i, url := range msg.ImageURLs {
			// url \n name
			su := strings.Split(url, "\n")
			url, name := su[0], su[1]
			suffix := filepath.Ext(name)
			switch suffix {
			case ".gif":
				if e, _ := utils.FileExist("gif"); !e {
					err := os.Mkdir("gif", 0o755)
					if err != nil {
						logger.WithError(err).Errorln("Stat gif dir failed")
						continue
					}
				}

				// download gif
				inf := filepath.Join("gif", name)
				resp, err := (&http.Client{}).Get(url)
				if err != nil {
					logger.WithError(err).Errorln("Get gif url error")
					continue
				}
				if resp.StatusCode != http.StatusOK {
					logger.WithError(err).Errorln("Get gif url not ok")
					continue
				}
				imgbyte, err := io.ReadAll(resp.Body)
				resp.Body.Close()
				if err != nil {
					logger.WithError(err).Errorln("Download gif failed")
					continue
				}
				err = os.WriteFile(inf, imgbyte, 0o755)
				if err != nil {
					logger.WithError(err).Errorln("Write gif to file failed")
					continue
				}

				// transcode for mp4
				outf := filepath.Join("gif", fmt.Sprintf("%s.mp4", name))
				if e, _ := utils.FileExist(outf); e {
					os.Remove(outf)
				}
				err = ffmpeg.Input(inf).Output(outf).WithTimeout(time.Minute * 5).Run()
				os.Remove(inf)
				if err != nil {
					logger.WithError(err).Errorln("Ffmpeg transcode to mp4 failed")
					continue
				}

				imgbyte, err = os.ReadFile(outf)
				if err != nil {
					logger.WithError(err).Errorln("Read generate mp4 file failed")
					os.Remove(outf)
					continue
				}

				// tgbotapi.NewInputMediaDocument does not function as expected
				inputDocumentMp4 := tgbotapi.NewInputMediaVideo(tgbotapi.FilePath(outf))
				if i == 0 {
					inputDocumentMp4.Caption = text
				}
				photos = append(photos, inputDocumentMp4)
				cacheFile = append(cacheFile, outf)
			default:
				inputMediaPhoto := tgbotapi.NewInputMediaPhoto(tgbotapi.FileURL(url))
				if i == 0 {
					inputMediaPhoto.Caption = text
				}
				photos = append(photos, inputMediaPhoto)
			}
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
		logger.WithError(err).Errorln("Send message failed")
	}
	cache.TG2QQCache.Add(int64(sentMsg.MessageID), msg.ID)
	cache.QQ2TGCache.Add(msg.ID, int64(sentMsg.MessageID))

	for _, f := range cacheFile {
		os.Remove(f)
	}
}
