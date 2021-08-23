# qqtg-bridge

转发同步qq群和tg群消息，支持tg的group和supergroup。

## 特性

1. 同时同步多给qq-tg群对
2. 支持通过proxy登陆tg bot
3. 支持转发图片和文字
4. qq群闪照自动转换为正常照片转发至tg

## 配置

下面是配置的示例文本

```
title = "configuration of qqtg-bridge"

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
```

+ qq段配置用于发送消息的qq bot帐号以及需要转发的qq群号
+ tg段配置tg bot的token以及需要转发的tg群id
+ forwards段具有多个，其中的qq-tg群id对用于将前面qq段和tg段提供的群一一对应
+ proxy段用于配置登陆tg bot和转发tg上媒体经过的代理，支持带帐号密码的代理url，默认禁用

## 系统要求

golang >= 1.16

## todo

1. tg群已发送消息编辑事件转发
2. 撤回消息（其实不能撤回也挺好的嘛）
3. tg群gif（.mp4）转发
4. 视频和音频转发

