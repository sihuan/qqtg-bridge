# qqtg-bridge

转发同步 qq 群和 tg 群消息，支持 tg 的 group 和 supergroup。

## 特性

1. 同时同步多个 qq-tg 群对
2. 支持通过代理登陆 tg bot
3. 支持转发图片和文字
4. qq 群闪照自动转换为正常照片转发至 tg

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

+ qq 段配置用于发送消息的 qq bot 帐号以及需要转发的 qq 群号
+ tg 段配置 tg bot 的 token 以及需要转发的 tg 群 chat id
+ forwards 段具有多个，其中的 qq-tg 群 id 对用于将前面 qq 段和 tg 段提供的群一一对应
+ proxy 段用于配置登陆 tg bot 和转发 tg 上媒体经过的代理，支持带帐号密码的代理url，默认禁用

## 系统要求

golang >= 1.16

## todo

1. tg 群已发送消息编辑事件转发
2. 撤回消息（其实不能撤回也挺好的嘛）
3. tg 群 gif（.mp4）转发
4. 视频和音频转发
