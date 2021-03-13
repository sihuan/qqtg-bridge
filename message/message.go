package message

type Message struct {
	Sender    string
	ID        int64
	ReplyID   int64
	ImageURLs []string
	VideoURLs []string
	Text      string
}

type MsgChan interface {
	Read() *Message
	Write(*Message)
}

func Copy(dst MsgChan, src MsgChan) {
	for {
		msg := src.Read()
		dst.Write(msg)
	}
}
