package tg

import (
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sihuan/qqtg-bridge/config"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	"net/url"
)

// Bot 全局 Bot
type Bot struct {
	*tgbotapi.BotAPI
	Chats map[int64]ChatChan
	start bool
}

// Instance Bot 实例
var Instance *Bot

var logger = logrus.WithField("tg", "internal")

// 使用 config.GlobalConfig 初始化 bot
func Init() {
	var (
		bot      *tgbotapi.BotAPI
		err      error
		proxyUrl *url.URL = nil
	)
	mc := make(map[int64]ChatChan)
	if config.GlobalConfig.Proxy.Enable {
		proxyUrl, err = url.Parse(config.GlobalConfig.Proxy.URL)
		if err != nil {
			logger.WithError(err).Errorln("Configure proxy failed")
			logger.Infoln("Try to init telegram bot without proxy")
		}
	}

	if proxyUrl != nil {
		proxyTrans := &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		}
		proxyClient := &http.Client{
			Transport: proxyTrans,
		}
		bot, err = tgbotapi.NewBotAPIWithClient(config.GlobalConfig.TG.Token, tgbotapi.APIEndpoint, proxyClient)
	} else {
		bot, err = tgbotapi.NewBotAPI(config.GlobalConfig.TG.Token)
	}

	if err != nil {
		log.Panic(err)
	}
	Instance = &Bot{
		BotAPI: bot,
		Chats:  mc,
		start:  false,
	}
}

func MakeChan() {
	for _, chatid := range config.GlobalConfig.TG.Chats {
		Instance.NewChatChan(chatid)
	}
}

func StartService() {
	if Instance.start {
		return
	}

	Instance.start = true

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := Instance.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil || (!update.Message.Chat.IsGroup() && !update.Message.Chat.IsSuperGroup()) {
			continue
		}
		if chat, ok := Instance.Chats[update.Message.Chat.ID]; ok {
			logger.Infof("[%s]: %s %s", update.Message.From.FirstName, update.Message.Text, update.Message.Caption)
			chat.tempChan <- update.Message
		}
	}
}
