package utils

import (
	"fmt"
	"github.com/thoj/go-ircevent"
	"strings"
	"time"
)

func IRCAction(conn *irc.Connection, channel, action string) {
	conn.Privmsg(channel, fmt.Sprintf("\001ACTION %s\001", action))
}

func IRCInvite(conn *irc.Connection, nick, channel string) {
	conn.SendRawf("INVITE %s %s", nick, channel)
}

func IRCOper(conn *irc.Connection, user, pass string) {
	conn.SendRawf("OPER %s %s", user, pass)
}

func IRCAway(conn *irc.Connection, message ...string) {
	msg := strings.Join(message, " ")
	if msg != "" {
		msg = " :" + msg
	}
	conn.SendRawf("AWAY%s", msg)
}

func IRCTopic(conn *irc.Connection, channel string, topic ...string) {
	msg := strings.Join(topic, " ")
	if msg != "" {
		msg = " :" + msg
	}
	conn.SendRawf("TOPIC %s%s", channel, msg)
}

func IRCRedispatch(conn *irc.Connection, code, raw, nick, host, source, user string, arguments ...string) {
	// This is here to throttle redispatches
	time.Sleep(time.Second)

	// Build synthetic event
	event := &irc.Event{
		Code:      code,
		Raw:       raw,
		Nick:      nick,
		Host:      host,
		Source:    source,
		User:      user,
		Arguments: arguments,
	}

	// Dispatch it
	conn.RunCallbacks(event)
}
