package main

import (
	"github.com/sihuan/qqtg-bridge/config"
	"github.com/sihuan/qqtg-bridge/message"
	"github.com/sihuan/qqtg-bridge/qq"
	"github.com/sihuan/qqtg-bridge/tg"
	"os"
	"os/signal"
)

func main() {
	//os.Setenv("HTTP_PROXY", "127.0.0.1:8889")
	//os.Setenv("HTTPS_PROXY", "127.0.0.1:8889")
	config.Init()

	qq.Init()
	qq.Login()
	qq.RefreshList()
	qq.MakeChan()
	qq.StartService()

	tg.Init()
	tg.MakeChan()
	go tg.StartService()

	forward := func(chatChanA message.MsgChan, chatChanB message.MsgChan) {
		go message.Copy(chatChanA, chatChanB)
		go message.Copy(chatChanB, chatChanA)
	}

	for _, forwardConfig := range config.GlobalConfig.Forwards {
		forward(qq.Instance.Chats[forwardConfig.QQ], tg.Instance.Chats[forwardConfig.TG])
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
}
