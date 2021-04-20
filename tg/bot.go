package tg

import (
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sihuan/qqtg-bridge/config"
	"github.com/sirupsen/logrus"
	"log"
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
	mc := make(map[int64]ChatChan)
	bot, err := tgbotapi.NewBotAPI(config.GlobalConfig.TG.Token)
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
		if update.Message == nil || !update.Message.Chat.IsGroup() {
			continue
		}
		if chat, ok := Instance.Chats[update.Message.Chat.ID]; ok {
			chat.tempChan <- update.Message
		}
	}
}
