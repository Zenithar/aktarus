package state

import (
	"github.com/thoj/go-ircevent"
	"strings"
)

func (st *StateTracker) joined(event *irc.Event) {
	// Wait until ready
	st.mutex.Lock()

	channel := st.channels[event.Arguments[0]]
	nick := st.nicks[event.Nick]

	// Haven't seen this channel before
	if channel == nil {
		// Create a channel object
		channel = &Channel{
			Name:  event.Arguments[0],
			Nicks: make(map[string]*ChannelPrivileges),
		}

		// Put it in the channels map
		st.channels[event.Arguments[0]] = channel

		// Get some initial info about it
		st.conn.Mode(channel.Name)
		st.conn.Who(channel.Name)
	}

	// Not seen this nick before
	if nick == nil {
		// Create a nick object
		nick = &Nick{
			Nick:     event.Nick,
			User:     event.User,
			Host:     event.Host,
			Channels: make(map[string]*ChannelPrivileges),
		}

		// Put it in the nicks map
		st.nicks[event.Nick] = nick

		// Get some inital info about it
		st.conn.Who(nick.Nick)
	}

	// Associate the nick with the channel
	st.associate(event.Nick, event.Arguments[0])

	// Ready for next event
	st.mutex.Unlock()
}

func (st *StateTracker) kicked(event *irc.Event) {
	st.mutex.Lock()
	st.disassociate(event.Nick, event.Arguments[0])
	st.mutex.Unlock()
}

func (st *StateTracker) nickChanged(event *irc.Event) {
	st.mutex.Lock()
	st.changeNick(event.Nick, event.Arguments[0])
	st.mutex.Unlock()
}

func (st *StateTracker) parted(event *irc.Event) {
	st.mutex.Lock()
	st.disassociate(event.Nick, event.Arguments[0])
	st.mutex.Unlock()
}

func (st *StateTracker) quitted(event *irc.Event) {
	st.mutex.Lock()
	st.deleteNick(event.Nick)
	st.mutex.Unlock()
}

func (st *StateTracker) topicSet(event *irc.Event) {
	st.mutex.Lock()
	st.setTopic(event.Arguments[0], event.Arguments[1])
	st.mutex.Unlock()
}

func (st *StateTracker) whoisReply(event *irc.Event) {
	st.mutex.Lock()
	nick := st.nicks[event.Arguments[1]]
	if nick != nil && nick != st.GetNick(st.conn.GetNick()) {
		nick.User = event.Arguments[2]
		nick.Host = event.Arguments[3]
		nick.Name = event.Arguments[5]
	}
	st.mutex.Unlock()
}

func (st *StateTracker) modeReply(event *irc.Event) {
	st.mutex.Lock()
	if channel, ok := st.channels[event.Arguments[0]]; ok {
		channel.ParseModes(event.Arguments[1], event.Arguments[2:]...)
	}
	st.mutex.Unlock()
}

func (st *StateTracker) topicReply(event *irc.Event) {
	st.mutex.Lock()
	if channel := st.channels[event.Arguments[0]]; channel != nil {
		st.setTopic(channel.Name, event.Arguments[1])
	}
	st.mutex.Unlock()
}

func (st *StateTracker) whoReply(event *irc.Event) {
	st.mutex.Lock()
	if nick, ok := st.nicks[event.Arguments[5]]; ok {
		nick.User = event.Arguments[2]
		nick.Host = event.Arguments[3]
		if idx := strings.Index(event.Arguments[6], "*"); idx != -1 {
			nick.Modes.Oper = true
		}
		if idx := strings.Index(event.Arguments[6], "H"); idx != -1 {
			nick.Modes.Invisible = true
		}
	}
	st.mutex.Unlock()
}

func (st *StateTracker) namesReply(event *irc.Event) {
	st.mutex.Lock()
	if channel, ok := st.channels[event.Arguments[2]]; ok {
		names := strings.Split(strings.TrimSpace(event.Arguments[len(event.Arguments)-1]), " ")
		for _, name := range names {
			switch priv := name[0]; priv {
			case '~', '&', '@', '%', '+':
				name = name[1:]
				fallthrough
			default:
				nick := st.nicks[name]

				if nick == nil {
					st.nicks[name] = &Nick{
						Nick:     name,
						Channels: make(map[string]*ChannelPrivileges),
					}
				}
				privs, ok := channel.Nicks[name]
				if !ok {
					privs = st.associate(name, channel.Name)
				}

				switch priv {
				case '~':
					privs.Owner = true
				case '&':
					privs.Admin = true
				case '@':
					privs.Op = true
				case '%':
					privs.HalfOp = true
				case '+':
					privs.Voice = true
				}
			}
		}
	}
	st.mutex.Unlock()
}

func (st *StateTracker) whoisReplySSL(event *irc.Event) {
	st.mutex.Lock()
	if nick, ok := st.nicks[event.Arguments[0]]; ok && nick != st.GetNick(st.conn.GetNick()) {
		nick.User = event.Arguments[1]
		nick.Host = event.Arguments[2]
		nick.Name = event.Arguments[4]
		nick.Modes.SSL = true
	}
	st.mutex.Unlock()
}

func (st *StateTracker) InitStateCallbacks() {
	st.conn.AddCallback("JOIN", st.joined)
	st.conn.AddCallback("KICK", st.kicked)
	st.conn.AddCallback("NICK", st.nickChanged)
	st.conn.AddCallback("PART", st.parted)
	st.conn.AddCallback("QUIT", st.quitted)
	st.conn.AddCallback("TOPIC", st.topicSet)
	st.conn.AddCallback("311", st.whoisReply)
	st.conn.AddCallback("MODE", st.modeReply)
	st.conn.AddCallback("332", st.topicReply)
	st.conn.AddCallback("352", st.whoReply)
	st.conn.AddCallback("353", st.namesReply)
	st.conn.AddCallback("671", st.whoisReplySSL)
}
