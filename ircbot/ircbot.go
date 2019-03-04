package ircbot

import (
	"bytes"
	"fmt"

	"github.com/sorcix/irc"
	"github.com/voldyman/ircx"
)

type (
	Bot struct {
		channels []string
		bot      *ircx.Bot
		chEvents chan *MessageEvent
	}

	MessageEvent struct {
		Sender  string
		Text    string
		Channel string
	}
)

func New(server, nick string, channels []string) *Bot {
	return &Bot{
		channels: channels,
		bot:      ircx.Classic(server, nick),
		chEvents: make(chan *MessageEvent),
	}
}

func (i *Bot) Start() (chan *MessageEvent, error) {
	err := i.bot.Connect()
	if err != nil {
		return nil, err
	}

	i.registerHandlers()

	go i.bot.CallbackLoop()

	return i.chEvents, nil
}

func insertNBS(s string) string {
    var buffer bytes.Buffer
    for _,rune := range s {
       buffer.WriteRune(rune)
       buffer.WriteRune('\u2060')
    }
    return buffer.String()
}

func (i *Bot) SendMessage(nick, msg, channel string, flag bool) {
	msgBuf := bytes.NewBufferString("")

	if flag == true {
		fmt.Fprintf(msgBuf, "%s:%s", insertNBS(nick), msg)
	} else {
		fmt.Fprintf(msgBuf, "%s", msg)
	}

	i.bot.SendMessage(&irc.Message{
		Command:  "PRIVMSG",
		Params:   []string{channel},
		Trailing: msgBuf.String(),
	})
}

func (i *Bot) SendRawMessage(msg, to string) {
	i.bot.SendMessage(&irc.Message{
		Command:  "PRIVMSG",
		Params:   []string{to},
		Trailing: msg,
	})
}

func (i *Bot) registerHandlers() {
	// IRC Ping Pong handler
	i.bot.AddCallback(irc.PING, ircx.Callback{
		Handler: ircx.HandlerFunc(pingHandler),
	})

	// IRC register handler
	i.bot.AddCallback(irc.RPL_WELCOME, ircx.Callback{
		Handler: ircx.HandlerFunc(i.registerConnect),
	})

	i.bot.AddCallback(irc.PRIVMSG, ircx.Callback{
		Handler: ircx.HandlerFunc(i.msgHandler),
	})
}

func (i *Bot) msgHandler(s ircx.Sender, m *irc.Message) {
	ev := &MessageEvent{
		Sender:  m.Name,
		Text:    m.Trailing,
		Channel: m.Params[0],
	}
	i.chEvents <- ev
}

func (i *Bot) registerConnect(s ircx.Sender, m *irc.Message) {
	for _, ch := range i.channels {
		s.Send(&irc.Message{
			Command: irc.JOIN,
			Params:  []string{ch},
		})
	}

}

func pingHandler(s ircx.Sender, m *irc.Message) {
	s.Send(&irc.Message{
		Command:  irc.PONG,
		Params:   m.Params,
		Trailing: m.Trailing,
	})
}
