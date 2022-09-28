package config

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"log"
	"os"
)

type QQConfig struct {
	Account  int64
	Password string
	Groups   []int64
}

type TGConfig struct {
	Token string
	Chats []int64
}

type ForwardConfig struct {
	QQ int64
	TG int64
}

type ProxyConfig struct {
	Enable bool
	URL    string
}

type Config struct {
	QQ       QQConfig
	TG       TGConfig
	Forwards []ForwardConfig
	Proxy    ProxyConfig
}

// GlobalConfig 默认全局配置
var GlobalConfig *Config

// Init 使用 ./config.toml 初始化全局配置
func Init() {
	c := viper.New()
	c.SetConfigName("config")
	c.SetConfigType("toml")
	c.AddConfigPath(".")
	c.AddConfigPath("./config")

	err := c.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config File not found
			createSampleConfig()
			logrus.Fatalln("Config does not exist, a sample of config was created")
		} else {
			logrus.WithField("config", "GlobalConfig").WithError(err).Panicf("unable to read global config")
		}
	}
	err = c.Unmarshal(&GlobalConfig)
	if err != nil {
		logrus.WithError(err).Panicln("unmarshal global config err")
	}
}

func createSampleConfig() {
	confSample := []byte(
		`title = "configuration of qqtg-bridge"

[qq]
  account=10086
  password="qq password"
  groups=[1111111,2222222]

[tg]
  token="1658565726:AAGugcmaKbYbBqKV7Kx4mUVYSTGzTq4UDUo"
  chats=[-12345,-98765]

[[forwards]]
  qq=1111111
  tg=-98765

[[forwards]]
  qq=2222222
  tg=-12345

[proxy]
  enable=false
  url="socks5://127.0.0.1:7891"
`)

	if os.WriteFile("config.toml", confSample, 0o644) != nil {
		log.Fatalln("Can't create a sample of config")
	}

}
