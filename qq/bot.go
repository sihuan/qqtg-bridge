package qq

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/Mrs4s/MiraiGo/client"
	mirai "github.com/Mrs4s/MiraiGo/message"
	"github.com/sihuan/qqtg-bridge/config"
	"github.com/sihuan/qqtg-bridge/utils"
	"github.com/sirupsen/logrus"
	asc2art "github.com/yinghau76/go-ascii-art"
	"image"
	"io/ioutil"
	"os"
	"strings"
)

// Bot 全局 Bot
type Bot struct {
	*client.QQClient
	Chats map[int64]ChatChan
	start bool
}

// Instance Bot 实例
var Instance *Bot

var logger = logrus.WithField("qq", "internal")

// Init 快速初始化
// 使用 config.GlobalConfig 初始化账号
// 使用 ./device.json 初始化设备信息
func Init() {
	mc := make(map[int64]ChatChan)
	Instance = &Bot{
		QQClient: client.NewClient(
			config.GlobalConfig.QQ.Account,
			config.GlobalConfig.QQ.Password,
		),
		Chats: mc,
		start: false,
	}
	b, _ := utils.FileExist("./device.json")
	if !b {
		logger.Warnln("no device.json, GenRandomDevice")
		GenRandomDevice()
	}
	err := client.SystemDeviceInfo.ReadJson(utils.ReadFile("./device.json"))

	if err != nil {
		logger.WithError(err).Panic("device.json error")
	}
}

// GenRandomDevice 生成随机设备信息
func GenRandomDevice() {
	client.GenRandomDevice()
	b, _ := utils.FileExist("./device.json")
	if b {
		logger.Warn("device.json exists, will not write device to file")
	}
	err := ioutil.WriteFile("device.json", client.SystemDeviceInfo.ToJson(), os.FileMode(0755))
	if err != nil {
		logger.WithError(err).Errorf("unable to write device.json")
	}
}

// Login 登录
func Login() {
	resp, err := Instance.Login()
	console := bufio.NewReader(os.Stdin)

	for {
		if err != nil {
			logger.WithError(err).Fatal("unable to login")
		}

		var text string
		if !resp.Success {
			switch resp.Error {

			case client.NeedCaptcha:
				img, _, _ := image.Decode(bytes.NewReader(resp.CaptchaImage))
				fmt.Println(asc2art.New("image", img).Art)
				fmt.Print("please input captcha: ")
				text, _ := console.ReadString('\n')
				resp, err = Instance.SubmitCaptcha(strings.ReplaceAll(text, "\n", ""), resp.CaptchaSign)
				continue

			case client.UnsafeDeviceError:
				fmt.Printf("device lock -> %v\n", resp.VerifyUrl)
				os.Exit(4)

			case client.SMSNeededError:
				fmt.Println("device lock enabled, Need SMS Code")
				fmt.Printf("Send SMS to %s ? (yes)", resp.SMSPhone)
				t, _ := console.ReadString('\n')
				t = strings.TrimSpace(t)
				if t != "yes" {
					os.Exit(2)
				}
				if !Instance.RequestSMS() {
					logger.Warnf("unable to request SMS Code")
					os.Exit(2)
				}
				logger.Warn("please input SMS Code: ")
				text, _ = console.ReadString('\n')
				resp, err = Instance.SubmitSMS(strings.ReplaceAll(strings.ReplaceAll(text, "\n", ""), "\r", ""))
				continue

			case client.TooManySMSRequestError:
				fmt.Printf("too many SMS request, please try later.\n")
				os.Exit(6)

			case client.SMSOrVerifyNeededError:
				fmt.Println("device lock enabled, choose way to verify:")
				fmt.Println("1. Send SMS Code to ", resp.SMSPhone)
				fmt.Println("2. Scan QR Code")
				fmt.Print("input (1,2):")
				text, _ = console.ReadString('\n')
				text = strings.TrimSpace(text)
				switch text {
				case "1":
					if !Instance.RequestSMS() {
						fmt.Println("unable to request SMS Code")
						os.Exit(2)
					}
					fmt.Print("please input SMS Code: ")
					text, _ = console.ReadString('\n')
					resp, err = Instance.SubmitSMS(strings.ReplaceAll(strings.ReplaceAll(text, "\n", ""), "\r", ""))
					continue
				case "2":
					fmt.Printf("device lock -> %v\n", resp.VerifyUrl)
					os.Exit(2)
				default:
					fmt.Println("invalid input")
					os.Exit(2)
				}

			case client.SliderNeededError:
				if client.SystemDeviceInfo.Protocol == client.AndroidPhone {
					fmt.Println("Android Phone Protocol DO NOT SUPPORT Slide verify")
					fmt.Println("please use other protocol")
					os.Exit(2)
				}
				Instance.AllowSlider = false
				Instance.Disconnect()
				resp, err = Instance.Login()
				continue

			case client.OtherLoginError, client.UnknownLoginError:
				logger.Fatalf("login failed: %v", resp.ErrorMessage)
			}

		}

		break
	}

	logger.Infof("qq login: %s", Instance.Nickname)
}

// RefreshList 刷新联系人
func RefreshList() {
	logger.Info("start reload friends list")
	err := Instance.ReloadFriendList()
	if err != nil {
		logger.WithError(err).Error("unable to load friends list")
	}
	logger.Infof("load %d friends", len(Instance.FriendList))
	logger.Info("start reload groups list")
	err = Instance.ReloadGroupList()
	if err != nil {
		logger.WithError(err).Error("unable to load groups list")
	}
	logger.Infof("load %d groups", len(Instance.GroupList))
}

func MakeChan() {
	for _, gid := range config.GlobalConfig.QQ.Groups {
		Instance.NewGroupChan(gid)
	}
}

func StartService() {
	if Instance.start {
		return
	}

	Instance.start = true
	Instance.OnGroupMessage(RouteMsg)
}

func RouteMsg(c *client.QQClient, msg *mirai.GroupMessage) {
	if msgChan, ok := Instance.Chats[msg.GroupCode]; ok {
		logger.Info(msg.ToString())
		msgChan.tempChan <- msg
	}
}
